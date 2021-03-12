package awsnoderefresher

import (
	"context"
	"time"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	"github.com/h3poteto/node-manager/pkg/util/klog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *AWSNodeRefresherReconciler) refreshAWSWait(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) error {
	if !shouldAWSWait(ctx, refresher) {
		return nil
	}
	refresher.Status.Phase = operatorv1alpha1.AWSNodeRefresherUpdateAWSWaiting
	if err := r.Client.Update(ctx, refresher); err != nil {
		klog.Errorf(ctx, "failed to update refresher: %v", err)
		return err
	}
	return nil
}

func shouldAWSWait(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) bool {
	if refresher.Status.Phase != operatorv1alpha1.AWSNodeRefresherUpdateReplacing {
		klog.Warningf(ctx, "AWSNodeRefresher phase is not matched: %s, so should not aws wait", refresher.Status.Phase)
		return false
	}
	if nodeStillLiving(refresher.Status.AWSNodes, refresher.Status.ReplaceTargetNode) {
		return false
	}
	return true
}

func (r *AWSNodeRefresherReconciler) stillWaiting(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) bool {
	now := metav1.Now()
	if now.Time.Before(refresher.Status.LastASGModifiedTime.Add(time.Duration(refresher.Spec.ASGModifyCoolTimeSeconds) * time.Second)) {
		klog.Info(ctx, "Waiting cooltime")
		return true
	}
	return false
}

func (r *AWSNodeRefresherReconciler) enoughInstances(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) bool {
	if len(refresher.Status.AWSNodes) < int(refresher.Spec.Desired)+IncreaseInstanceCount {
		klog.Infof(ctx, "Instance is not enough, current: %d, expected: %d + %d", len(refresher.Status.AWSNodes), refresher.Spec.Desired, IncreaseInstanceCount)
		return false
	}
	return true
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
