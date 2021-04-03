package awsnoderefresher

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	"github.com/h3poteto/node-manager/pkg/util/klog"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *AWSNodeRefresherReconciler) refreshReplace(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) error {
	if !shouldReplace(ctx, refresher) {
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
		klog.Errorf(ctx, "failed to update refresher: %v", err)
		return err
	}
	r.Recorder.Event(refresher, corev1.EventTypeNormal, "Replace instance", "Replace instance in ASG for refresh")

	return r.cloud.DeleteInstance(target)
}

func shouldReplace(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) bool {
	if refresher.Status.Phase != operatorv1alpha1.AWSNodeRefresherUpdateIncreasing {
		klog.Warningf(ctx, "AWSNodeRefresher phase is not matched: %s, so should not replace", refresher.Status.Phase)
		return false
	}
	if len(refresher.Status.AWSNodes) != int(refresher.Spec.Desired)+IncreaseInstanceCount {
		klog.Infof(ctx, "Node is not enough, current: %d, desired: %d + %d", len(refresher.Status.AWSNodes), refresher.Spec.Desired, IncreaseInstanceCount)
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
		klog.Errorf(ctx, "failed to update refresher: %v", err)
		return err
	}
	r.Recorder.Event(refresher, corev1.EventTypeNormal, "Start next replace", "Start to next replace in ASG")
	return nil
}

func (r *AWSNodeRefresherReconciler) retryReplace(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) (bool, error) {
	now := metav1.Now()
	if !r.shouldRetryReplace(ctx, refresher) {
		return false, nil
	}
	target := refresher.Status.ReplaceTargetNode
	refresher.Status.Phase = operatorv1alpha1.AWSNodeRefresherUpdateReplacing
	refresher.Status.LastASGModifiedTime = &now
	refresher.Status.ReplaceTargetNode = target
	if err := r.Client.Update(ctx, refresher); err != nil {
		klog.Errorf(ctx, "failed to update refresher: %v", err)
		return false, err
	}
	r.Recorder.Event(refresher, corev1.EventTypeNormal, "Retry replace", "Retry to replace instance in ASG for refresh")

	err := r.cloud.DeleteInstance(target)
	return true, err
}

func (r *AWSNodeRefresherReconciler) shouldRetryReplace(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) bool {
	if refresher.Status.Phase != operatorv1alpha1.AWSNodeRefresherUpdateReplacing {
		klog.Warningf(ctx, "AWSNodeRefresher phase is not matched: %s, so should not retry to replace", refresher.Status.Phase)
		return false
	}

	instance, err := r.cloud.DescribeInstance(refresher.Status.ReplaceTargetNode)
	if err != nil {
		klog.Warning(ctx, err)
		return false
	}

	if *instance.State.Name != ec2.InstanceStateNameTerminated {
		klog.Infof(ctx, "Instance %s state is %s, so retry to replace it", *instance.InstanceId, *instance.State.Name)
		return true
	}

	return false
}

func (r *AWSNodeRefresherReconciler) replaceWait(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) bool {
	now := metav1.Now()
	if now.Time.Before(refresher.Status.LastASGModifiedTime.Add(1 * time.Minute)) {
		return true
	}

	return false
}
