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
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	cloudaws "github.com/h3poteto/node-manager/pkg/cloud/aws"
	pkgctx "github.com/h3poteto/node-manager/pkg/util/context"
	"github.com/h3poteto/node-manager/pkg/util/externalevent"
	"github.com/h3poteto/node-manager/pkg/util/klog"
	"github.com/h3poteto/node-manager/pkg/util/requestid"
)

// AWSNodeRefresherReconciler reconciles a AWSNodeRefresher object
type AWSNodeRefresherReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
	cloud    *cloudaws.AWS
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
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	r.cloud = cloudaws.New(sess, refresher.Spec.Region)

	if err := r.syncRefresher(ctx, &refresher); err != nil {
		klog.Errorf(ctx, "failed to sync AWSNodeRefresher: %v", err)
		r.Recorder.Eventf(&refresher, corev1.EventTypeWarning, "Error", "Failed to sync: %v", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *AWSNodeRefresherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	external := externalevent.NewExternalEventWatcher(1*time.Minute, func(ctx context.Context, c client.Client) ([]client.Object, error) {
		var refreshers operatorv1alpha1.AWSNodeRefresherList
		err := c.List(ctx, &refreshers)
		if err != nil {
			return nil, err
		}
		var list []client.Object
		for i := range refreshers.Items {
			item := &refreshers.Items[i]
			list = append(list, item)
		}
		return list, nil
	})
	err := mgr.Add(external)
	if err != nil {
		return err
	}
	src := source.Channel{
		Source: external.Channel,
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
		waiting, retried, err := r.retryIncrease(ctx, refresher)
		if err != nil {
			return err
		}
		if waiting {
			return nil
		}
		if retried {
			return nil
		}
		return r.refreshReplace(ctx, refresher)
	case operatorv1alpha1.AWSNodeRefresherUpdateReplacing:
		waiting, retried, err := r.retryReplace(ctx, refresher)
		if err != nil {
			return err
		}
		if waiting {
			return nil
		}
		if retried {
			return nil
		}
		return r.refreshAWSWait(ctx, refresher)
	case operatorv1alpha1.AWSNodeRefresherUpdateAWSWaiting:
		waiting, enough := r.checkInstances(ctx, refresher)
		if waiting {
			return nil
		}
		if !enough {
			return nil
		}
		klog.Info(ctx, "finish waiting")
		if r.allReplaced(ctx, refresher) {
			return r.refreshDecrease(ctx, refresher)
		} else {
			return r.refreshNextReplace(ctx, refresher)
		}
	case operatorv1alpha1.AWSNodeRefresherUpdateDecreasing:
		waiting, retried, err := r.retryDecrease(ctx, refresher)
		if err != nil {
			return err
		}
		if waiting {
			return nil
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
