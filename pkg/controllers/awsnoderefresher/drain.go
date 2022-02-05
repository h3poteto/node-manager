package awsnoderefresher

import (
	"context"
	"time"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	"github.com/h3poteto/node-manager/pkg/util/klog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *AWSNodeRefresherReconciler) refreshDrain(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) error {
	if !shouldDrain(ctx, refresher) {
		return nil
	}

	target, err := findDeleteTarget(refresher.Status.AWSNodes)
	if err != nil {
		return err
	}

	now := metav1.Now()
	refresher.Status.Phase = operatorv1alpha1.AWSNodeRefresherDraining
	refresher.Status.LastASGModifiedTime = &now
	refresher.Status.ReplaceTargetNode = target
	refresher.Status.Revision += 1
	if err := r.Client.Update(ctx, refresher); err != nil {
		klog.Errorf(ctx, "failed to update refresher: %v", err)
		return err
	}
	r.Recorder.Eventf(refresher, corev1.EventTypeNormal, "Drain node", "Drain node %s", target.Name)

	return r.drain(ctx, target.Name)
}

func shouldDrain(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) bool {
	if refresher.Status.Phase != operatorv1alpha1.AWSNodeRefresherUpdateIncreasing {
		klog.Warningf(ctx, "AWSNodeRefresher phase is not matched: %s, so should not drain", refresher.Status.Phase)
		return false
	}
	if len(refresher.Status.AWSNodes) != int(refresher.Spec.Desired)+int(refresher.Spec.SurplusNodes) {
		klog.Infof(ctx, "Node is not enough, current: %d, desired: %d + %d", len(refresher.Status.AWSNodes), refresher.Spec.Desired, refresher.Spec.SurplusNodes)
		return false
	}
	return true
}

func (r *AWSNodeRefresherReconciler) retryDrain(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) (bool, bool, error) {
	if checkDrainTimeout(refresher) {
		return true, false, nil
	}

	target := refresher.Status.ReplaceTargetNode

	if !r.shouldRetryDrain(ctx, refresher, target.Name) {
		return false, false, nil
	}

	r.Recorder.Event(refresher, corev1.EventTypeNormal, "Retry drain", "Drain to replace instance in ASG for refresh")

	err := r.drain(ctx, target.Name)
	return false, true, err
}

func (r *AWSNodeRefresherReconciler) drain(ctx context.Context, nodeName string) error {
	// Node
	var node corev1.Node
	if err := r.Client.Get(ctx, client.ObjectKey{
		Name: nodeName,
	}, &node); err != nil {
		klog.Errorf(ctx, "Failed to get node: %v", err)
		return err
	}
	node.Spec.Unschedulable = true
	if err := r.Client.Update(ctx, &node); err != nil {
		klog.Errorf(ctx, "Failed to update node: %v", err)
		return err
	}

	// Pods
	var podList corev1.PodList
	if err := r.Client.List(ctx, &podList); err != nil {
		klog.Errorf(ctx, "Failed to list pods: %v", err)
		return err
	}

	for i := range podList.Items {
		pod := podList.Items[i]
		if pod.Spec.NodeName != nodeName {
			continue
		}
		// Ignore DaemonSet and Static pods
		if podIsDaemonSet(pod) || podIsStaticPod(pod) {
			continue
		}
		if err := r.Client.Delete(ctx, &pod); err != nil {
			klog.Errorf(ctx, "Failed to delete pod: %v", err)
			return err
		}
	}

	return nil
}

func (r *AWSNodeRefresherReconciler) shouldRetryDrain(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher, nodeName string) bool {
	var podList corev1.PodList
	if err := r.Client.List(ctx, &podList); err != nil {
		klog.Errorf(ctx, "Failed to list pods: %v", err)
		return true
	}
	var pods []*corev1.Pod
	for i := range podList.Items {
		pod := podList.Items[i]
		if pod.Spec.NodeName != nodeName {
			continue
		}
		if podIsDaemonSet(pod) || podIsStaticPod(pod) {
			continue
		}
		pods = append(pods, &pod)
	}
	if len(pods) > 0 {
		return true
	}
	return false
}

func checkDrainTimeout(refresher *operatorv1alpha1.AWSNodeRefresher) bool {
	now := metav1.Now()
	if now.Time.After(refresher.Status.LastASGModifiedTime.Add(time.Duration(refresher.Spec.DrainGracePeriodSeconds) * time.Second)) {
		return true
	}

	return false
}

func podIsDaemonSet(pod corev1.Pod) bool {
	for _, o := range pod.OwnerReferences {
		if o.Kind == "DaemonSet" {
			return true
		}
	}
	return false
}

func podIsStaticPod(pod corev1.Pod) bool {
	for _, o := range pod.OwnerReferences {
		if o.Kind == "Node" {
			return true
		}
	}
	return false
}
