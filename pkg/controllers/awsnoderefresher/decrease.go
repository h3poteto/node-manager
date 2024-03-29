package awsnoderefresher

import (
	"context"
	"time"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	"github.com/h3poteto/node-manager/pkg/util/klog"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *AWSNodeRefresherReconciler) refreshDecrease(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) error {
	now := metav1.Now()
	should, skip := r.shouldDecrease(ctx, refresher)
	if skip {
		refresher.Status.Phase = operatorv1alpha1.AWSNodeRefresherUpdateDecreasing
		refresher.Status.Revision += 1
		if err := r.Client.Update(ctx, refresher); err != nil {
			klog.Errorf(ctx, "failed to update refresher: %v", err)
			return err
		}
		r.Recorder.Event(refresher, corev1.EventTypeNormal, "Skip decrease instance", "Skip decrease instance in ASG for refresh")
	}
	if !should {
		return nil
	}

	refresher.Status.Phase = operatorv1alpha1.AWSNodeRefresherUpdateDecreasing
	refresher.Status.LastASGModifiedTime = &now
	refresher.Status.Revision += 1
	if err := r.Client.Update(ctx, refresher); err != nil {
		klog.Errorf(ctx, "failed to update refresher: %v", err)
		return err
	}
	r.Recorder.Event(refresher, corev1.EventTypeNormal, "Decrease instance", "Decrease instance in ASG for refresh")

	return r.cloud.DeleteInstancesToAutoScalingGroups(refresher.Spec.AutoScalingGroups, int(refresher.Spec.Desired), len(refresher.Status.AWSNodes))
}

func (r *AWSNodeRefresherReconciler) shouldDecrease(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) (bool, bool) {
	if refresher.Status.Phase != operatorv1alpha1.AWSNodeRefresherUpdateAWSWaiting {
		klog.Warningf(ctx, "AWSNodeRefresher phase is not matched: %s, so should not decrease", refresher.Status.Phase)
		return false, false
	}
	if refresher.Spec.SurplusNodes == 0 {
		return false, true
	}
	if len(refresher.Status.AWSNodes) >= (int(refresher.Spec.Desired) + int(refresher.Spec.SurplusNodes)) {
		return true, false
	}
	return false, false
}

func (r *AWSNodeRefresherReconciler) retryDecrease(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) (bool, bool, error) {
	now := metav1.Now()
	if waitingDecrease(ctx, refresher, &now) {
		return true, false, nil
	}
	if !shouldRetryDecrease(ctx, refresher, &now) {
		return false, false, nil
	}

	refresher.Status.Phase = operatorv1alpha1.AWSNodeRefresherUpdateDecreasing
	refresher.Status.LastASGModifiedTime = &now
	refresher.Status.Revision += 1
	if err := r.Client.Update(ctx, refresher); err != nil {
		klog.Errorf(ctx, "failed to update refresher: %v", err)
		return false, false, err
	}
	r.Recorder.Eventf(refresher, corev1.EventTypeNormal, "Retry decrease", "Retry to decrease instances for AWSNodeRefresher %s/%s", refresher.Namespace, refresher.Name)

	err := r.cloud.DeleteInstancesToAutoScalingGroups(
		refresher.Spec.AutoScalingGroups,
		int(refresher.Spec.Desired),
		len(refresher.Status.AWSNodes),
	)
	return false, true, err
}

func shouldRetryDecrease(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher, now *metav1.Time) bool {
	if refresher.Status.Phase != operatorv1alpha1.AWSNodeRefresherUpdateDecreasing {
		klog.Warningf(ctx, "AWSNodeRefresher phase is not matched: %s, so should not retry to decrease", refresher.Status.Phase)
		return false
	}
	if len(refresher.Status.AWSNodes) > int(refresher.Spec.Desired) {
		return true
	}
	return false
}

func waitingDecrease(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher, now *metav1.Time) bool {
	if now.Time.Before(refresher.Status.LastASGModifiedTime.Add(time.Duration(refresher.Spec.ASGModifyCoolTimeSeconds) * time.Second)) {
		klog.Info(ctx, "Waiting cooltime")
		return true
	}
	return false
}
