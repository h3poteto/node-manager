package awsnoderefresher

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	"github.com/h3poteto/node-manager/pkg/util/klog"
)

func (r *AWSNodeRefresherReconciler) refreshIncrease(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) error {
	owner, err := r.ownerAWSNodeManager(ctx, refresher)
	if err != nil {
		return err
	}
	now := metav1.Now()
	if !shouldIncrease(ctx, refresher, &now, owner) {
		return nil
	}

	refresher.Status.Phase = operatorv1alpha1.AWSNodeRefresherUpdateIncreasing
	refresher.Status.UpdateStartTime = &now
	refresher.Status.LastASGModifiedTime = &now
	refresher.Status.Revision += 1
	if err := r.Client.Update(ctx, refresher); err != nil {
		klog.Errorf(ctx, "failed to update refresher: %v", err)
		return err
	}
	r.Recorder.Event(refresher, corev1.EventTypeNormal, "Increase instance", "Increase instance to ASG for refresh")

	return r.cloud.AddInstancesToAutoScalingGroups(refresher.Spec.AutoScalingGroups, int(refresher.Spec.Desired)+1, len(refresher.Status.AWSNodes))
}

func shouldIncrease(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher, now *metav1.Time, owner *operatorv1alpha1.AWSNodeManager) bool {
	if refresher.Status.Phase != operatorv1alpha1.AWSNodeRefresherScheduled {
		klog.Warningf(ctx, "AWSNodeRefresher phase is not matched: %s, so should not increase", refresher.Status.Phase)
		return false
	}
	if owner.Status.Phase == operatorv1alpha1.AWSNodeManagerReplenishing {
		klog.Info(ctx, "Now replenishing, so skip refresh")
		return false
	}
	if refresher.Status.NextUpdateTime.Before(now) {
		return true
	}
	return false
}

func (r *AWSNodeRefresherReconciler) retryIncrease(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) (bool, bool, error) {
	now := metav1.Now()
	if waitingIncrease(ctx, refresher, &now) {
		return true, false, nil
	}
	if !shouldRetryIncrease(ctx, refresher, &now) {
		return false, false, nil
	}

	refresher.Status.LastASGModifiedTime = &now
	refresher.Status.Revision += 1
	if err := r.Client.Update(ctx, refresher); err != nil {
		klog.Errorf(ctx, "failed to update refresher: %v", err)
		return false, false, err
	}
	r.Recorder.Event(refresher, corev1.EventTypeNormal, "Retry increase", "Retry to increase instance to ASG for refresh")

	err := r.cloud.AddInstancesToAutoScalingGroups(
		refresher.Spec.AutoScalingGroups,
		int(refresher.Spec.Desired)+int(refresher.Spec.SurplusNodes),
		len(refresher.Status.AWSNodes),
	)
	return false, true, err
}

func waitingIncrease(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher, now *metav1.Time) bool {
	if len(refresher.Status.AWSNodes) >= int(refresher.Spec.Desired)+int(refresher.Spec.SurplusNodes) {
		return false
	}
	if now.Time.After(refresher.Status.LastASGModifiedTime.Add(time.Duration(refresher.Spec.ASGModifyCoolTimeSeconds) * time.Second)) {
		return false
	}
	klog.Info(ctx, "Waiting cooltime")
	return true
}

func shouldRetryIncrease(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher, now *metav1.Time) bool {
	if refresher.Status.Phase != operatorv1alpha1.AWSNodeRefresherUpdateIncreasing {
		klog.Warningf(ctx, "AWSNodeRefresher phase is not matched: %s, so should not retry to increase", refresher.Status.Phase)
		return false
	}
	if len(refresher.Status.AWSNodes) < int(refresher.Spec.Desired)+int(refresher.Spec.SurplusNodes) {
		return true
	}
	return false
}
