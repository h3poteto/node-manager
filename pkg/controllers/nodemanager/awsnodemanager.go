package nodemanager

import (
	"context"
	"fmt"
	"reflect"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	"github.com/h3poteto/node-manager/pkg/util/klog"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *NodeManagerReconciler) syncAWSNodeManager(ctx context.Context, nodeManager *operatorv1alpha1.NodeManager, masterNodes, workerNodes []*corev1.Node) (*operatorv1alpha1.AWSNodeManager, *operatorv1alpha1.AWSNodeManager, error) {
	var masterNodeManager, workerNodeManager *operatorv1alpha1.AWSNodeManager
	var err error
	if nodeManager.Spec.Aws.Masters != nil {
		masterNodeManager, err = r.syncMasterAWSNodeManager(ctx, nodeManager, masterNodes)
		if err != nil {
			return nil, nil, err
		}
	}
	if nodeManager.Spec.Aws.Workers != nil {
		workerNodeManager, err = r.syncWorkerAWSNodeManager(ctx, nodeManager, workerNodes)
		if err != nil {
			return nil, nil, err
		}
	}
	return masterNodeManager, workerNodeManager, nil

}

func (r *NodeManagerReconciler) createAWSNodeManager(ctx context.Context, nodeManager *operatorv1alpha1.NodeManager, nodes []*corev1.Node, role operatorv1alpha1.NodeRole) (*operatorv1alpha1.AWSNodeManager, error) {
	klog.Infof(ctx, "creating AWSNodeManager for %s", role)
	switch role {
	case operatorv1alpha1.Master:
		newMasterManager := generateMasterAWSNodeManager(nodeManager)
		for i := range nodes {
			node := nodes[i]
			a := operatorv1alpha1.AWSNode{
				Name:              node.Name,
				CreationTimestamp: node.CreationTimestamp,
			}
			newMasterManager.Status.AWSNodes = append(newMasterManager.Status.AWSNodes, a)
		}
		if err := r.Client.Create(ctx, newMasterManager); err != nil {
			klog.Errorf(ctx, "failed to create AWSNodemanager for %s", role)
			return nil, err
		}

		r.Recorder.Eventf(newMasterManager, corev1.EventTypeNormal, "Created", "Created AWSNodeManager for master %s/%s", newMasterManager.Namespace, newMasterManager.Name)
		return newMasterManager, nil
	case operatorv1alpha1.Worker:
		newWorkerManager := generateWorkerAWSNodeManager(nodeManager)
		for i := range nodes {
			node := nodes[i]
			a := operatorv1alpha1.AWSNode{
				Name:              node.Name,
				CreationTimestamp: node.CreationTimestamp,
			}
			newWorkerManager.Status.AWSNodes = append(newWorkerManager.Status.AWSNodes, a)
		}
		if err := r.Client.Create(ctx, newWorkerManager); err != nil {
			klog.Errorf(ctx, "failed to create AWSNodeManager for %s", role)
			return nil, err
		}

		r.Recorder.Eventf(newWorkerManager, corev1.EventTypeNormal, "Created", "Created AWSNodeManager for worker %s/%s", newWorkerManager.Namespace, newWorkerManager.Name)
		return newWorkerManager, nil
	default:
		return nil, fmt.Errorf("Role %s is not acceptable", role)
	}
}

func (r *NodeManagerReconciler) updateAWSNodeManager(ctx context.Context, existing *operatorv1alpha1.AWSNodeManager, nodeManager *operatorv1alpha1.NodeManager, nodes []*corev1.Node, role operatorv1alpha1.NodeRole) (*operatorv1alpha1.AWSNodeManager, error) {
	switch role {
	case operatorv1alpha1.Master:
		newMasterManager := generateMasterAWSNodeManager(nodeManager)
		var currentNames, nodeNames []string
		for _, node := range existing.Status.AWSNodes {
			currentNames = append(currentNames, node.Name)
		}
		for _, node := range nodes {
			nodeNames = append(nodeNames, node.Name)
		}
		if reflect.DeepEqual(existing.Spec, newMasterManager.Spec) && reflect.DeepEqual(currentNames, nodeNames) {
			klog.Infof(ctx, "AWSNodeManager %s/%s is already synced", existing.Namespace, existing.Name)
			return existing, nil
		}
		existing.Spec = newMasterManager.Spec
		existing.Status.AWSNodes = []operatorv1alpha1.AWSNode{}
		for _, node := range nodes {
			a := operatorv1alpha1.AWSNode{
				Name:              node.Name,
				CreationTimestamp: node.CreationTimestamp,
			}
			existing.Status.AWSNodes = append(existing.Status.AWSNodes, a)
		}
		existing.Status.Revision += 1
		if err := r.Client.Update(ctx, existing); err != nil {
			klog.Errorf(ctx, "failed to update existing AWSNodeManager %s/%s: %v", existing.Namespace, existing.Name, err)
			return nil, err
		}
		klog.Infof(ctx, "updated AWSNodeManager spec for master %s/%s", existing.Namespace, existing.Name)
		r.Recorder.Eventf(existing, corev1.EventTypeNormal, "Updated", "Updated AWSNodeManager %s/%s", existing.Namespace, existing.Name)
		return existing, nil
	case operatorv1alpha1.Worker:
		newWorkerManager := generateWorkerAWSNodeManager(nodeManager)
		var currentNames, nodeNames []string
		for _, node := range existing.Status.AWSNodes {
			currentNames = append(currentNames, node.Name)
		}
		for _, node := range nodes {
			nodeNames = append(nodeNames, node.Name)
		}
		if reflect.DeepEqual(existing.Spec, newWorkerManager.Spec) && reflect.DeepEqual(currentNames, nodeNames) {
			klog.Infof(ctx, "AWSNodeManager %s/%s is already synced", existing.Namespace, existing.Name)
			return existing, nil
		}
		existing.Spec = newWorkerManager.Spec
		existing.Status.AWSNodes = []operatorv1alpha1.AWSNode{}
		for _, node := range nodes {
			a := operatorv1alpha1.AWSNode{
				Name:              node.Name,
				CreationTimestamp: node.CreationTimestamp,
			}
			existing.Status.AWSNodes = append(existing.Status.AWSNodes, a)
		}
		existing.Status.Revision += 1
		if err := r.Client.Update(ctx, existing); err != nil {
			klog.Errorf(ctx, "failed to update existing AWSNodeManager %s/%s: %v", existing.Namespace, existing.Name, err)
			return nil, err
		}
		klog.Infof(ctx, "updated AWSNodeManager spec for worker %s/%s", existing.Namespace, existing.Name)
		r.Recorder.Eventf(existing, corev1.EventTypeNormal, "Updated", "Updated AWSNodeManager %s/%s", existing.Namespace, existing.Name)
		return existing, nil
	default:
		return nil, fmt.Errorf("Role %s is not acceptable", role)
	}
}

func generateMasterAWSNodeManager(nodeManager *operatorv1alpha1.NodeManager) *operatorv1alpha1.AWSNodeManager {
	return &operatorv1alpha1.AWSNodeManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:        nodeManager.Name + "-master",
			Namespace:   nodeManager.Namespace,
			Labels:      nodeManager.GetLabels(),
			Annotations: nodeManager.GetAnnotations(),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(nodeManager, nodeManager.GroupVersionKind()),
			},
		},
		Spec: operatorv1alpha1.AWSNodeManagerSpec{
			Region:                   nodeManager.Spec.Aws.Region,
			AutoScalingGroups:        nodeManager.Spec.Aws.Masters.AutoScalingGroups,
			ASGModifyCoolTimeSeconds: nodeManager.Spec.Aws.Masters.ASGModifyCoolTimeSeconds,
			DrainGracePeriodSeconds:  nodeManager.Spec.Aws.Masters.DrainGracePeriodSeconds,
			Desired:                  nodeManager.Spec.Aws.Masters.Desired,
			Role:                     operatorv1alpha1.Master,
			EnableReplenish:          nodeManager.Spec.Aws.Masters.EnableReplenish,
			RefreshSchedule:          nodeManager.Spec.Aws.Masters.RefreshSchedule,
			SurplusNodes:             nodeManager.Spec.Aws.Masters.SurplusNodes,
		},
		Status: operatorv1alpha1.AWSNodeManagerStatus{
			Phase: operatorv1alpha1.AWSNodeManagerInit,
		},
	}
}

func generateWorkerAWSNodeManager(nodeManager *operatorv1alpha1.NodeManager) *operatorv1alpha1.AWSNodeManager {
	return &operatorv1alpha1.AWSNodeManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:        nodeManager.Name + "-worker",
			Namespace:   nodeManager.Namespace,
			Labels:      nodeManager.GetLabels(),
			Annotations: nodeManager.GetAnnotations(),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(nodeManager, nodeManager.GroupVersionKind()),
			},
		},
		Spec: operatorv1alpha1.AWSNodeManagerSpec{
			Region:                   nodeManager.Spec.Aws.Region,
			AutoScalingGroups:        nodeManager.Spec.Aws.Workers.AutoScalingGroups,
			ASGModifyCoolTimeSeconds: nodeManager.Spec.Aws.Workers.ASGModifyCoolTimeSeconds,
			DrainGracePeriodSeconds:  nodeManager.Spec.Aws.Workers.DrainGracePeriodSeconds,
			Desired:                  nodeManager.Spec.Aws.Workers.Desired,
			Role:                     operatorv1alpha1.Worker,
			EnableReplenish:          nodeManager.Spec.Aws.Workers.EnableReplenish,
			RefreshSchedule:          nodeManager.Spec.Aws.Workers.RefreshSchedule,
			SurplusNodes:             nodeManager.Spec.Aws.Workers.SurplusNodes,
		},
		Status: operatorv1alpha1.AWSNodeManagerStatus{
			Phase: operatorv1alpha1.AWSNodeManagerInit,
		},
	}
}
