/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package awsnoderefresher

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	pkgctx "github.com/h3poteto/node-manager/pkg/util/context"
	"github.com/h3poteto/node-manager/pkg/util/klog"
	"github.com/h3poteto/node-manager/pkg/util/requestid"
)

// AWSNodeRefresherReconciler reconciles a AWSNodeRefresher object
type AWSNodeRefresherReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
	Session  *session.Session
}

// +kubebuilder:rbac:groups=operator.h3poteto.dev,resources=awsnoderefreshers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.h3poteto.dev,resources=awsnoderefreshers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

func (r *AWSNodeRefresherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("awsnoderefresher", req.NamespacedName)
	ctx = pkgctx.SetController(ctx, "awsnoderefresher")
	id, err := requestid.RequestID()
	if err != nil {
		return ctrl.Result{}, err
	}
	ctx = pkgctx.SetRequestID(ctx, id)

	klog.Infof(ctx, "fetching AWSNodeRefresher %s", req.NamespacedName)
	refresher := operatorv1alpha1.AWSNodeRefresher{}
	if err := r.Client.Get(ctx, req.NamespacedName, &refresher); err != nil {
		klog.Infof(ctx, "failed to get AWSNodeRefresher resources: %v", err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Generate aws client
	r.Session = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	if err := r.syncRefresher(ctx, &refresher); err != nil {
		klog.Errorf(ctx, "failed to sync AWSNodeRefresher: %v", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *AWSNodeRefresherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	external := newExternalEventWatcher()
	err := mgr.Add(external)
	if err != nil {
		return err
	}
	src := source.Channel{
		Source: external.channel,
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.AWSNodeRefresher{}).
		Watches(&src, &handler.EnqueueRequestForObject{}).
		Complete(r)
}

func (r *AWSNodeRefresherReconciler) syncRefresher(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) error {
	switch refresher.Status.Phase {
	case operatorv1alpha1.AWSNodeRefresherInit:
		return r.scheduleNext(ctx, refresher)
	case operatorv1alpha1.AWSNodeRefresherScheduled:
		return r.refreshIncrease(ctx, refresher)
	case operatorv1alpha1.AWSNodeRefresherUpdateIncreasing:
		retried, err := r.retryIncrease(ctx, refresher)
		if err != nil {
			return err
		}
		if retried {
			return nil
		}
		return r.refreshReplace(ctx, refresher)
	case operatorv1alpha1.AWSNodeRefresherUpdateReplacing:
		return r.refreshAWSWait(ctx, refresher)
	case operatorv1alpha1.AWSNodeRefresherUpdateAWSWaiting:
		if r.stillWaiting(ctx, refresher) {
			return nil
		}
		if !r.enoughInstances(ctx, refresher) {
			return nil
		}
		klog.Info(ctx, "finish waiting")
		if r.allReplaced(ctx, refresher) {
			return r.refreshDecrease(ctx, refresher)
		} else {
			return r.refreshNextReplace(ctx, refresher)
		}
	case operatorv1alpha1.AWSNodeRefresherUpdateDecreasing:
		if r.stillWaiting(ctx, refresher) {
			return nil
		}
		klog.Info(ctx, "finish waiting")
		retried, err := r.retryDecrease(ctx, refresher)
		if err != nil {
			return err
		}
		if retried {
			return nil
		}
		return r.refreshComplete(ctx, refresher)
	case operatorv1alpha1.AWSNodeRefresherCompleted:
		return r.scheduleNext(ctx, refresher)
	default:
		klog.Warningf(ctx, "Unknown phase %s for AWSNodeRefrehser", refresher.Status.Phase)
		return nil
	}
}
