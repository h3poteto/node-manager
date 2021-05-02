package awsnodereplenisher

import (
	"context"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	"github.com/h3poteto/node-manager/pkg/util/klog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
