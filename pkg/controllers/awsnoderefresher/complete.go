package awsnoderefresher

import (
	"context"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	"github.com/h3poteto/node-manager/pkg/util/klog"
	corev1 "k8s.io/api/core/v1"
)

func (r *AWSNodeRefresherReconciler) refreshComplete(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) error {
	if !shouldComplete(ctx, refresher) {
		return nil
	}

	refresher.Status.Phase = operatorv1alpha1.AWSNodeRefresherCompleted
	refresher.Status.UpdateStartTime = nil
	refresher.Status.ReplaceTargetNode = nil
	if err := r.Client.Update(ctx, refresher); err != nil {
		klog.Errorf(ctx, "failed to update refresher: %v", err)
		return err
	}
	r.Recorder.Event(refresher, corev1.EventTypeNormal, "Completed refresh", "Completed to refresh")
	return nil
}

func shouldComplete(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) bool {
	if refresher.Status.Phase != operatorv1alpha1.AWSNodeRefresherUpdateDecreasing {
		klog.Warningf(ctx, "AWSNodeRefresher phase is not matched: %s, so should not complete", refresher.Status.Phase)
		return false
	}
	return true
}
