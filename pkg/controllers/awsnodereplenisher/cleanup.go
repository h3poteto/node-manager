package awsnodereplenisher

import (
	"context"
	"time"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	"github.com/h3poteto/node-manager/pkg/util/klog"
	corev1 "k8s.io/api/core/v1"
)

func (r *AWSNodeReplenisherReconciler) syncNotJoinedAWSNodes(ctx context.Context, replenisher *operatorv1alpha1.AWSNodeReplenisher) error {
	if !shouldClean(replenisher) {
		return nil
	}
	now := time.Now()

	for _, node := range replenisher.Status.NotJoinedAWSNodes {
		if shouldWait(&node, now) {
			continue
		}
		if err := r.updateStatusAWSUpdating(ctx, replenisher); err != nil {
			return err
		}
		err := r.cloud.DetachInstanceFromASG(node.InstanceID, node.AutoScalingGroupName)
		if err != nil {
			klog.Errorf(ctx, "failed to detach instance %s from ASG %s: %v", node.InstanceID, node.AutoScalingGroupName, err)
			return err
		}

		err = r.cloud.DeleteInstance(&node)
		if err != nil {
			klog.Errorf(ctx, "failed to delete instance %s: %v", node.InstanceID, err)
			return err
		}
		klog.Infof(ctx, "detach and terminate instance %s from %s", node.InstanceID, node.AutoScalingGroupName)
		r.Recorder.Eventf(replenisher, corev1.EventTypeNormal, "Delete instance", "Detach and terminate instance %s from %s", node.InstanceID, node.AutoScalingGroupName)
	}

	return nil
}

func shouldClean(replenisher *operatorv1alpha1.AWSNodeReplenisher) bool {
	return len(replenisher.Status.NotJoinedAWSNodes) > 0
}

func shouldWait(node *operatorv1alpha1.AWSNode, now time.Time) bool {
	if now.Before(node.CreationTimestamp.Time.Add(1 * time.Hour)) {
		return true
	}
	return false
}
