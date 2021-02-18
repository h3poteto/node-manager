package awsnoderefresher

import (
	"context"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	"k8s.io/klog/v2"
)

func (r *AWSNodeRefresherReconciler) refreshComplete(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) error {
	if !shouldComplete(refresher) {
		return nil
	}

	refresher.Status.Phase = operatorv1alpha1.AWSNodeRefresherCompleted
	refresher.Status.UpdateStartTime = nil
	refresher.Status.ReplaceTargetNode = nil
	if err := r.Client.Update(ctx, refresher); err != nil {
		klog.Errorf("failed to update refresher: %v", err)
		return err
	}
	return nil
}

func shouldComplete(refresher *operatorv1alpha1.AWSNodeRefresher) bool {
	if refresher.Status.Phase != operatorv1alpha1.AWSNodeRefresherUpdateDecreasing {
		klog.Warningf("AWSNodeRefresher phase is not matched: %s, so should not complete", refresher.Status.Phase)
		return false
	}
	return true
}
