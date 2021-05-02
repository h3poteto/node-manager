package awsnodereplenisher

import (
	"context"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
)

func (r *AWSNodeReplenisherReconciler) addNode(ctx context.Context, replenisher *operatorv1alpha1.AWSNodeReplenisher, currentNodesCount int) error {
	if err := r.updateStatusAWSUpdating(ctx, replenisher); err != nil {
		return err
	}

	return r.cloud.AddInstancesToAutoScalingGroups(replenisher.Spec.AutoScalingGroups, int(replenisher.Spec.Desired), currentNodesCount)
}

func (r *AWSNodeReplenisherReconciler) deleteNode(ctx context.Context, replenisher *operatorv1alpha1.AWSNodeReplenisher, currentNodesCount int) error {
	if err := r.updateStatusAWSUpdating(ctx, replenisher); err != nil {
		return err
	}

	return r.cloud.DeleteInstancesToAutoScalingGroups(replenisher.Spec.AutoScalingGroups, int(replenisher.Spec.Desired), currentNodesCount)
}
