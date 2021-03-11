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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	cloudaws "github.com/h3poteto/node-manager/pkg/cloud/aws"
	pkgctx "github.com/h3poteto/node-manager/pkg/util/context"
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
	Session  *session.Session
}

// +kubebuilder:rbac:groups=operator.h3poteto.dev,resources=awsnodereplenishers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.h3poteto.dev,resources=awsnodereplenishers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

func (r *AWSNodeReplenisherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("awsnodereplenisher", req.NamespacedName)
	ctx = pkgctx.SetController(ctx, "nodemanager")
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
	r.Session = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	if err := r.syncReplenisher(ctx, &replenisher); err != nil {
		klog.Errorf(ctx, "failed to sync AWSNodeReplenisher: %v", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *AWSNodeReplenisherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.AWSNodeReplenisher{}).
		Complete(r)
}

// syncReplenisher checks nodes and replenish AWS instances when node resources are not enough.
func (r *AWSNodeReplenisherReconciler) syncReplenisher(ctx context.Context, replenisher *operatorv1alpha1.AWSNodeReplenisher) error {
	klog.Info(ctx, "reading node info from status")

	if len(replenisher.Status.AWSNodes) == int(replenisher.Spec.Desired) {
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

	if len(replenisher.Status.AWSNodes) > int(replenisher.Spec.Desired) {
		klog.Infof(ctx, "nodes count is %d, but desired count is %d, so deleting nodes", len(replenisher.Status.AWSNodes), replenisher.Spec.Desired)
		if err := r.deleteNode(ctx, replenisher, len(replenisher.Status.AWSNodes)); err != nil {
			return err
		}
	} else {
		klog.Infof(ctx, "nodes count is %d, but desired count is %d, so adding nodes", len(replenisher.Status.AWSNodes), replenisher.Spec.Desired)
		if err := r.addNode(ctx, replenisher, len(replenisher.Status.AWSNodes)); err != nil {
			return err
		}
	}
	return nil
}

func (r *AWSNodeReplenisherReconciler) addNode(ctx context.Context, replenisher *operatorv1alpha1.AWSNodeReplenisher, currentNodesCount int) error {
	if err := r.updateStatusAWSUpdating(ctx, replenisher); err != nil {
		return err
	}
	cloud := cloudaws.New(r.Session, replenisher.Spec.Region)
	return cloud.AddInstancesToAutoScalingGroups(replenisher.Spec.AutoScalingGroups, int(replenisher.Spec.Desired), currentNodesCount)
}

func (r *AWSNodeReplenisherReconciler) deleteNode(ctx context.Context, replenisher *operatorv1alpha1.AWSNodeReplenisher, currentNodesCount int) error {
	if err := r.updateStatusAWSUpdating(ctx, replenisher); err != nil {
		return err
	}

	cloud := cloudaws.New(r.Session, replenisher.Spec.Region)
	return cloud.DeleteInstancesToAutoScalingGroups(replenisher.Spec.AutoScalingGroups, int(replenisher.Spec.Desired), currentNodesCount)
}

func (r *AWSNodeReplenisherReconciler) updateLatestTimestamp(ctx context.Context, replenisher *operatorv1alpha1.AWSNodeReplenisher, now metav1.Time) error {
	// We have to retry when this update function is failed.
	// If we don't update LastASGModifiedTime after modify some ASGs,
	// the next process in reconcile will add/delete instances without wait.
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		currentReplenisher := operatorv1alpha1.AWSNodeReplenisher{}
		if err := r.Client.Get(ctx, client.ObjectKey{Namespace: replenisher.Namespace, Name: replenisher.Name}, &currentReplenisher); err != nil {
			klog.Errorf(ctx, "failed to get AWSNodeReplenisher %s/%s: %v", replenisher.Namespace, replenisher.Name, err)
			return err
		}
		currentReplenisher.Status.LastASGModifiedTime = &now
		currentReplenisher.Status.Revision += 1
		if err := r.Client.Update(ctx, &currentReplenisher); err != nil {
			klog.Errorf(ctx, "failed to update AWSNodeReplenisher status %s/%s: %v", replenisher.Namespace, replenisher.Name, err)
			return err
		}
		r.Recorder.Eventf(&currentReplenisher, corev1.EventTypeNormal, "Updated", "Updated AWSNodeReplenisher status %s/%s", currentReplenisher.Namespace, currentReplenisher.Name)
		return nil
	})
}

func (r *AWSNodeReplenisherReconciler) updateStatusSynced(ctx context.Context, replenisher *operatorv1alpha1.AWSNodeReplenisher) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		currentReplenisher := operatorv1alpha1.AWSNodeReplenisher{}
		if err := r.Client.Get(ctx, client.ObjectKey{Namespace: replenisher.Namespace, Name: replenisher.Name}, &currentReplenisher); err != nil {
			klog.Errorf(ctx, "failed to get AWSNodeReplenisher %s/%s: %v", replenisher.Namespace, replenisher.Name, err)
			return err
		}
		if currentReplenisher.Status.Phase == operatorv1alpha1.AWSNodeReplenisherSynced {
			klog.Infof(ctx, "AWSNodeReplenisher %s/%s is already synced", currentReplenisher.Namespace, currentReplenisher.Name)
			return nil
		}
		currentReplenisher.Status.Phase = operatorv1alpha1.AWSNodeReplenisherSynced
		currentReplenisher.Status.Revision += 1
		if err := r.Client.Update(ctx, &currentReplenisher); err != nil {
			klog.Errorf(ctx, "failed to update AWSNodeReplenisher status %s/%s: %v", replenisher.Namespace, replenisher.Name, err)
			return err
		}
		r.Recorder.Eventf(&currentReplenisher, corev1.EventTypeNormal, "Updated", "Updated AWSNodeReplenisher status %s/%s", currentReplenisher.Namespace, currentReplenisher.Name)
		return nil
	})
}

func (r *AWSNodeReplenisherReconciler) updateStatusAWSUpdating(ctx context.Context, replenisher *operatorv1alpha1.AWSNodeReplenisher) error {
	now := metav1.Now()
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		currentReplenisher := operatorv1alpha1.AWSNodeReplenisher{}
		if err := r.Client.Get(ctx, client.ObjectKey{Namespace: replenisher.Namespace, Name: replenisher.Name}, &currentReplenisher); err != nil {
			klog.Errorf(ctx, "failed to get AWSNodeReplenisher %s/%s: %v", replenisher.Namespace, replenisher.Name, err)
			return err
		}
		currentReplenisher.Status.Phase = operatorv1alpha1.AWSNodeReplenisherAWSUpdating
		currentReplenisher.Status.LastASGModifiedTime = &now
		currentReplenisher.Status.Revision += 1
		if err := r.Client.Update(ctx, &currentReplenisher); err != nil {
			klog.Errorf(ctx, "failed to update AWSNodeReplenisher status %s/%s: %v", replenisher.Namespace, replenisher.Name, err)
			return err
		}
		r.Recorder.Eventf(&currentReplenisher, corev1.EventTypeNormal, "Updated", "Updated AWSNodeReplenisher status %s/%s", currentReplenisher.Namespace, currentReplenisher.Name)
		return nil
	})
}
