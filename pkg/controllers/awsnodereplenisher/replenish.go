package awsnodereplenisher

import (
	"context"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	"github.com/h3poteto/node-manager/pkg/util/klog"
	corev1 "k8s.io/api/core/v1"
)

func (r *AWSNodeReplenisherReconciler) syncAWSNodes(ctx context.Context, replenisher *operatorv1alpha1.AWSNodeReplenisher) (bool, error) {
	if len(replenisher.Status.AWSNodes) > int(replenisher.Spec.Desired) {
		klog.Infof(ctx, "nodes count is %d, but desired count is %d, so deleting nodes", len(replenisher.Status.AWSNodes), replenisher.Spec.Desired)
		if err := r.deleteNode(ctx, replenisher, len(replenisher.Status.AWSNodes)); err != nil {
			return true, err
		}
		r.Recorder.Event(replenisher, corev1.EventTypeNormal, "Delete nodes", "Delete member nodes")
		return true, nil
	} else if len(replenisher.Status.AWSNodes) < int(replenisher.Spec.Desired) {
		klog.Infof(ctx, "nodes count is %d, but desired count is %d, so adding nodes", len(replenisher.Status.AWSNodes), replenisher.Spec.Desired)
		if err := r.addNode(ctx, replenisher, len(replenisher.Status.AWSNodes)); err != nil {
			return true, err
		}
		r.Recorder.Event(replenisher, corev1.EventTypeNormal, "Add nodes", "Add member nodes")
		return true, nil
	}
	return false, nil
}
