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

package awsnodereplenisher

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

const (
	AnnotationKey = "managed.aws-node-replenisher.operator.h3poteto.dev"
)

// AWSNodeReplenisherReconciler reconciles a AWSNodeReplenisher object
type AWSNodeReplenisherReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	cloud    *cloudaws.AWS
}

// +kubebuilder:rbac:groups=operator.h3poteto.dev,resources=awsnodereplenishers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.h3poteto.dev,resources=awsnodereplenishers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

func (r *AWSNodeReplenisherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("awsnodereplenisher", req.NamespacedName)
	ctx = pkgctx.SetController(ctx, "awsnodereplenisher")
	id, err := requestid.RequestID()
	if err != nil {
		return ctrl.Result{}, err
	}
	ctx = pkgctx.SetRequestID(ctx, id)

	klog.Info(ctx, "fetching AWSNodeReplenisher resources")
	replenisher := operatorv1alpha1.AWSNodeReplenisher{}
	if err := r.Client.Get(ctx, req.NamespacedName, &replenisher); err != nil {
		klog.Infof(ctx, "failed to get AWSNodeReplenisher resources: %v", err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Generate aws client
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	r.cloud = cloudaws.New(sess, replenisher.Spec.Region)

	if err := r.syncReplenisher(ctx, &replenisher); err != nil {
		klog.Errorf(ctx, "failed to sync AWSNodeReplenisher: %v", err)
		r.Recorder.Eventf(&replenisher, corev1.EventTypeWarning, "Error", "Failed to sync: %v", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *AWSNodeReplenisherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	external := externalevent.NewExternalEventWatcher[*operatorv1alpha1.AWSNodeReplenisher](5*time.Minute, func(ctx context.Context, c client.Client) ([]*operatorv1alpha1.AWSNodeReplenisher, error) {
		var replenishers operatorv1alpha1.AWSNodeReplenisherList
		err := c.List(ctx, &replenishers)
		if err != nil {
			return nil, err
		}
		var list []*operatorv1alpha1.AWSNodeReplenisher
		for i := range replenishers.Items {
			item := &replenishers.Items[i]
			list = append(list, item)
		}
		return list, nil
	})
	err := mgr.Add(external)
	if err != nil {
		return err
	}
	src := source.Channel(external.Channel, &handler.TypedEnqueueRequestForObject[*operatorv1alpha1.AWSNodeReplenisher]{})

	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.AWSNodeReplenisher{}).
		WatchesRawSource(src).
		Complete(r)
}

// syncReplenisher checks nodes and replenish AWS instances when node resources are not enough.
func (r *AWSNodeReplenisherReconciler) syncReplenisher(ctx context.Context, replenisher *operatorv1alpha1.AWSNodeReplenisher) error {

	if !shouldSync(replenisher) {
		klog.Info(ctx, "nodes count is same as desired count")
		return r.updateStatusSynced(ctx, replenisher)
	}

	owner, err := r.ownerAWSNodeManager(ctx, replenisher)
	if err != nil {
		return err
	}
	if owner.Status.Phase == operatorv1alpha1.AWSNodeManagerRefreshing {
		klog.Info(ctx, "Now refreshing, so skip replenish")
		return nil
	}

	now := time.Now()
	if replenisher.Status.Phase == operatorv1alpha1.AWSNodeReplenisherAWSUpdating &&
		replenisher.Status.LastASGModifiedTime != nil &&
		now.Before(replenisher.Status.LastASGModifiedTime.Add(time.Duration(replenisher.Spec.ASGModifyCoolTimeSeconds)*time.Second)) {
		klog.Info(ctx, "Waiting cool time")
		return nil
	}

	updated, err := r.syncAWSNodes(ctx, replenisher)
	if err != nil {
		return err
	}
	if updated {
		return nil
	}

	return r.syncNotJoinedAWSNodes(ctx, replenisher)
}

func shouldSync(replenisher *operatorv1alpha1.AWSNodeReplenisher) bool {
	return len(replenisher.Status.AWSNodes) != int(replenisher.Spec.Desired) || len(replenisher.Status.NotJoinedAWSNodes) > 0
}
