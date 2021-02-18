package awsnoderefresher

import (
	"context"
	"errors"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	cloudaws "github.com/h3poteto/node-manager/pkg/cloud/aws"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func (r *AWSNodeRefresherReconciler) refreshReplace(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) error {
	if !shouldReplace(refresher) {
		return nil
	}

	target, err := findDeleteTarget(refresher.Status.AWSNodes)
	if err != nil {
		return err
	}

	now := metav1.Now()
	refresher.Status.Phase = operatorv1alpha1.AWSNodeRefresherUpdateReplacing
	refresher.Status.LastASGModifiedTime = &now
	refresher.Status.ReplaceTargetNode = target
	if err := r.Client.Update(ctx, refresher); err != nil {
		klog.Errorf("failed to update refresher: %v", err)
		return err
	}

	cloud := cloudaws.New(r.Session, refresher.Spec.Region)
	return cloud.DeleteInstance(target)
}

func shouldReplace(refresher *operatorv1alpha1.AWSNodeRefresher) bool {
	if refresher.Status.Phase != operatorv1alpha1.AWSNodeRefresherUpdateIncreasing {
		return false
	}
	if len(refresher.Status.AWSNodes) != int(refresher.Spec.Desired)+IncreaseInstanceCount {
		klog.Infof("Node is not enough, current: %d, desired: %d + %d", len(refresher.Status.AWSNodes), refresher.Spec.Desired, IncreaseInstanceCount)
		return false
	}
	return true
}

func findDeleteTarget(nodes []operatorv1alpha1.AWSNode) (*operatorv1alpha1.AWSNode, error) {
	var target *operatorv1alpha1.AWSNode
	for i := range nodes {
		node := &nodes[i]
		if target == nil {
			target = node
			continue
		}
		if node.CreationTimestamp.Before(&target.CreationTimestamp) {
			target = node
		}
	}
	if target == nil {
		return nil, errors.New("No nodes are running")
	}
	return target, nil
}

func (r *AWSNodeRefresherReconciler) refreshNextReplace(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) error {
	refresher.Status.Phase = operatorv1alpha1.AWSNodeRefresherUpdateIncreasing
	if err := r.Client.Update(ctx, refresher); err != nil {
		klog.Errorf("failed to update refresher: %v", err)
		return err
	}
	return nil
}
