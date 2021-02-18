package awsnoderefresher

import (
	"context"
	"time"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func (r *AWSNodeRefresherReconciler) refreshAWSWait(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) error {
	if !shouldAWSWait(refresher) {
		return nil
	}
	refresher.Status.Phase = operatorv1alpha1.AWSNodeRefresherUpdateAWSWaiting
	if err := r.Client.Update(ctx, refresher); err != nil {
		klog.Errorf("failed to update refresher: %v", err)
		return err
	}
	return nil
}

func shouldAWSWait(refresher *operatorv1alpha1.AWSNodeRefresher) bool {
	if refresher.Status.Phase != operatorv1alpha1.AWSNodeRefresherUpdateReplacing {
		return false
	}
	if nodeStillLiving(refresher.Status.AWSNodes, refresher.Status.ReplaceTargetNode) {
		return false
	}
	return true
}

func (r *AWSNodeRefresherReconciler) stillWaiting(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) bool {
	now := metav1.Now()
	if now.Time.After(refresher.Status.LastASGModifiedTime.Add(time.Duration(refresher.Spec.ASGModifyCoolTimeSeconds) * time.Second)) {
		klog.Info("Waiting cooltime")
		return true
	}
	if len(refresher.Status.AWSNodes) < int(refresher.Spec.Desired)+IncreaseInstanceCount {
		klog.Infof("Instance is not enough, current: %d, expected: %d + %d", len(refresher.Status.AWSNodes), refresher.Spec.Desired, IncreaseInstanceCount)
		return true
	}
	return false
}

func (r *AWSNodeRefresherReconciler) allReplaced(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) bool {
	return allNodesNewer(refresher.Status.AWSNodes, refresher.Status.UpdateStartTime)
}

// allNodesNewer returns whether all nodes are newer than start timestamp.
func allNodesNewer(nodes []operatorv1alpha1.AWSNode, start *metav1.Time) bool {
	for i := range nodes {
		node := &nodes[i]
		if node.CreationTimestamp.Before(start) {
			return false
		}
	}
	return true
}

func nodeStillLiving(nodes []operatorv1alpha1.AWSNode, target *operatorv1alpha1.AWSNode) bool {
	for i := range nodes {
		node := &nodes[i]
		if node.Name == target.Name {
			return true
		}
	}
	return false
}
