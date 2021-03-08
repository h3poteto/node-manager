package awsnoderefresher

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	cloudaws "github.com/h3poteto/node-manager/pkg/cloud/aws"
)

const IncreaseInstanceCount int = 1

func (r *AWSNodeRefresherReconciler) refreshIncrease(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) error {
	owner, err := r.ownerAWSNodeManager(ctx, refresher)
	if err != nil {
		return err
	}
	now := metav1.Now()
	if !shouldIncrease(refresher, &now, owner) {
		return nil
	}

	refresher.Status.Phase = operatorv1alpha1.AWSNodeRefresherUpdateIncreasing
	refresher.Status.UpdateStartTime = &now
	refresher.Status.LastASGModifiedTime = &now
	if err := r.Client.Update(ctx, refresher); err != nil {
		klog.Errorf("failed to update refresher: %v", err)
		return err
	}

	cloud := cloudaws.New(r.Session, refresher.Spec.Region)
	return cloud.AddInstancesToAutoScalingGroups(refresher.Spec.AutoScalingGroups, int(refresher.Spec.Desired)+1, len(refresher.Status.AWSNodes))
}

func shouldIncrease(refresher *operatorv1alpha1.AWSNodeRefresher, now *metav1.Time, owner *operatorv1alpha1.AWSNodeManager) bool {
	if refresher.Status.Phase != operatorv1alpha1.AWSNodeRefresherScheduled {
		klog.Warningf("AWSNodeRefresher phase is not matched: %s, so should not increase", refresher.Status.Phase)
		return false
	}
	if owner.Status.Phase == operatorv1alpha1.AWSNodeManagerReplenishing {
		klog.Info("Now replenishing, so skip refresh")
		return false
	}
	if refresher.Status.NextUpdateTime.Before(now) {
		return true
	}
	return false
}

func (r *AWSNodeRefresherReconciler) retryIncrease(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) (bool, error) {
	now := metav1.Now()
	if !shouldRetryIncrease(refresher, &now) {
		return false, nil
	}

	refresher.Status.LastASGModifiedTime = &now
	if err := r.Client.Update(ctx, refresher); err != nil {
		klog.Errorf("failed to update refresher: %v", err)
		return false, err
	}

	cloud := cloudaws.New(r.Session, refresher.Spec.Region)
	err := cloud.AddInstancesToAutoScalingGroups(
		refresher.Spec.AutoScalingGroups,
		int(refresher.Spec.Desired)+IncreaseInstanceCount,
		len(refresher.Status.AWSNodes),
	)
	return true, err
}

func shouldRetryIncrease(refresher *operatorv1alpha1.AWSNodeRefresher, now *metav1.Time) bool {
	if refresher.Status.Phase != operatorv1alpha1.AWSNodeRefresherUpdateIncreasing {
		klog.Warningf("AWSNodeRefresher phase is not matched: %s, so should not retry to increase", refresher.Status.Phase)
		return false
	}
	if len(refresher.Status.AWSNodes) < int(refresher.Spec.Desired)+IncreaseInstanceCount {
		if now.Time.After(refresher.Status.LastASGModifiedTime.Add(time.Duration(refresher.Spec.ASGModifyCoolTimeSeconds) * time.Second)) {
			return true
		}
		klog.Info("Waiting cooltime")
	}
	return false
}
