package awsnoderefresher

import (
	"context"
	"time"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	cloudaws "github.com/h3poteto/node-manager/pkg/cloud/aws"
	"github.com/h3poteto/node-manager/pkg/util/klog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *AWSNodeRefresherReconciler) refreshDecrease(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) error {
	now := metav1.Now()
	if !r.shouldDecrease(ctx, refresher) {
		return nil
	}

	refresher.Status.Phase = operatorv1alpha1.AWSNodeRefresherUpdateDecreasing
	refresher.Status.LastASGModifiedTime = &now
	if err := r.Client.Update(ctx, refresher); err != nil {
		klog.Errorf(ctx, "failed to update refresher: %v", err)
		return err
	}

	cloud := cloudaws.New(r.Session, refresher.Spec.Region)
	return cloud.DeleteInstancesToAutoScalingGroups(refresher.Spec.AutoScalingGroups, int(refresher.Spec.Desired), len(refresher.Status.AWSNodes))
}

func (r *AWSNodeRefresherReconciler) shouldDecrease(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) bool {
	if refresher.Status.Phase != operatorv1alpha1.AWSNodeRefresherUpdateAWSWaiting {
		klog.Warningf(ctx, "AWSNodeRefresher phase is not matched: %s, so should not decrease", refresher.Status.Phase)
		return false
	}
	if len(refresher.Status.AWSNodes) > int(refresher.Spec.Desired) {
		return true
	}
	return false
}

func (r *AWSNodeRefresherReconciler) retryDecrease(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) (bool, error) {
	now := metav1.Now()
	if !shouldRetryDecrease(ctx, refresher, &now) {
		return false, nil
	}

	refresher.Status.Phase = operatorv1alpha1.AWSNodeRefresherUpdateDecreasing
	refresher.Status.LastASGModifiedTime = &now
	if err := r.Client.Update(ctx, refresher); err != nil {
		klog.Errorf(ctx, "failed to update refresher: %v", err)
		return false, err
	}

	cloud := cloudaws.New(r.Session, refresher.Spec.Region)
	err := cloud.DeleteInstancesToAutoScalingGroups(
		refresher.Spec.AutoScalingGroups,
		int(refresher.Spec.Desired),
		len(refresher.Status.AWSNodes),
	)
	return true, err
}

func shouldRetryDecrease(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher, now *metav1.Time) bool {
	if refresher.Status.Phase != operatorv1alpha1.AWSNodeRefresherUpdateDecreasing {
		klog.Warningf(ctx, "AWSNodeRefresher phase is not matched: %s, so should not retry to decrease", refresher.Status.Phase)
		return false
	}
	if len(refresher.Status.AWSNodes) == int(refresher.Spec.Desired) {
		return false
	}
	if now.Time.After(refresher.Status.LastASGModifiedTime.Add(time.Duration(refresher.Spec.ASGModifyCoolTimeSeconds) * time.Second)) {
		return true
	}
	klog.Info(ctx, "Waiting cooltime")
	return false
}
