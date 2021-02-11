package nodemanager

import (
	"context"
	"fmt"
	"reflect"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func (r *NodeManagerReconciler) syncAWSNodeManager(ctx context.Context, nodeManager *operatorv1alpha1.NodeManager, masterNames, workerNames []string) (*operatorv1alpha1.AWSNodeManager, *operatorv1alpha1.AWSNodeManager, error) {
	var masterNodeManager, workerNodeManager *operatorv1alpha1.AWSNodeManager
	var err error
	if nodeManager.Spec.Aws.Masters != nil {
		masterNodeManager, err = r.syncMasterAWSNodeManager(ctx, nodeManager, masterNames)
		if err != nil {
			return nil, nil, err
		}
	}
	if nodeManager.Spec.Aws.Workers != nil {
		workerNodeManager, err = r.syncWorkerAWSNodeManager(ctx, nodeManager, workerNames)
		if err != nil {
			return nil, nil, err
		}
	}
	return masterNodeManager, workerNodeManager, nil

}

func (r *NodeManagerReconciler) createAWSNodeManager(ctx context.Context, nodeManager *operatorv1alpha1.NodeManager, nodeNames []string, role operatorv1alpha1.NodeRole) (*operatorv1alpha1.AWSNodeManager, error) {
	klog.Infof("creating AWSNodeManager for %s", role)
	switch role {
	case operatorv1alpha1.Master:
		newMasterManager := generateMasterAWSNodeManager(nodeManager)
		for i := range nodeNames {
			a := operatorv1alpha1.AWSNode{
				Name: nodeNames[i],
			}
			newMasterManager.Status.AWSNodes = append(newMasterManager.Status.AWSNodes, a)
		}
		if err := r.Client.Create(ctx, newMasterManager); err != nil {
			klog.Errorf("failed to create AWSNodemanager for %s", role)
			return nil, err
		}

		r.Recorder.Eventf(newMasterManager, corev1.EventTypeNormal, "Created", "Created AWSNodeManager for master %s/%s", newMasterManager.Namespace, newMasterManager.Name)
		return newMasterManager, nil
	case operatorv1alpha1.Worker:
		newWorkerManager := generateWorkerAWSNodeManager(nodeManager)
		for i := range nodeNames {
			a := operatorv1alpha1.AWSNode{
				Name: nodeNames[i],
			}
			newWorkerManager.Status.AWSNodes = append(newWorkerManager.Status.AWSNodes, a)
		}
		if err := r.Client.Create(ctx, newWorkerManager); err != nil {
			klog.Errorf("failed to create AWSNodeManager for %s", role)
			return nil, err
		}

		r.Recorder.Eventf(newWorkerManager, corev1.EventTypeNormal, "Created", "Created AWSNodeManager for worker %s/%s", newWorkerManager.Namespace, newWorkerManager.Name)
		return newWorkerManager, nil
	default:
		return nil, fmt.Errorf("Role %s is not acceptable", role)
	}
}

func (r *NodeManagerReconciler) updateAWSNodeManager(ctx context.Context, existing *operatorv1alpha1.AWSNodeManager, nodeManager *operatorv1alpha1.NodeManager, nodeNames []string, role operatorv1alpha1.NodeRole) (*operatorv1alpha1.AWSNodeManager, error) {
	switch role {
	case operatorv1alpha1.Master:
		newMasterManager := generateMasterAWSNodeManager(nodeManager)
		var names []string
		for _, node := range existing.Status.AWSNodes {
			names = append(names, node.Name)
		}
		if reflect.DeepEqual(existing.Spec, newMasterManager.Spec) && reflect.DeepEqual(names, nodeManager.Status.MasterNodes) {
			klog.Infof("AWSNodeManager %s/%s is already synced", existing.Namespace, existing.Name)
			return existing, nil
		}
		existing.Spec = newMasterManager.Spec
		existing.Status.AWSNodes = []operatorv1alpha1.AWSNode{}
		for i := range nodeManager.Status.MasterNodes {
			a := operatorv1alpha1.AWSNode{
				Name: nodeManager.Status.MasterNodes[i],
			}
			existing.Status.AWSNodes = append(existing.Status.AWSNodes, a)
		}
		existing.Status.Revision += 1
		if err := r.Client.Update(ctx, existing); err != nil {
			klog.Errorf("failed to update existing AWSNodeManager %s/%s: %v", existing.Namespace, existing.Name, err)
			return nil, err
		}
		klog.Infof("updated AWSNodeManager spec for master %s/%s", existing.Namespace, existing.Name)
		r.Recorder.Eventf(existing, corev1.EventTypeNormal, "Updated", "Updated AWSNodeManager %s/%s", existing.Namespace, existing.Name)
		return existing, nil
	case operatorv1alpha1.Worker:
		newWorkerManager := generateWorkerAWSNodeManager(nodeManager)
		var names []string
		for _, node := range existing.Status.AWSNodes {
			names = append(names, node.Name)
		}
		if reflect.DeepEqual(existing.Spec, newWorkerManager.Spec) && reflect.DeepEqual(names, nodeManager.Status.WorkerNodes) {
			klog.Infof("AWSNodeManager %s/%s is already synced", existing.Namespace, existing.Name)
			return existing, nil
		}
		existing.Spec = newWorkerManager.Spec
		existing.Status.AWSNodes = []operatorv1alpha1.AWSNode{}
		for i := range nodeManager.Status.WorkerNodes {
			a := operatorv1alpha1.AWSNode{
				Name: nodeManager.Status.WorkerNodes[i],
			}
			existing.Status.AWSNodes = append(existing.Status.AWSNodes, a)
		}
		existing.Status.Revision += 1
		if err := r.Client.Update(ctx, existing); err != nil {
			klog.Errorf("failed to update existing AWSNodeManager %s/%s: %v", existing.Namespace, existing.Name, err)
			return nil, err
		}
		klog.Infof("updated AWSNodeManager spec for worker %s/%s", existing.Namespace, existing.Name)
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
			Desired:                  nodeManager.Spec.Aws.Masters.Desired,
			Role:                     operatorv1alpha1.Master,
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
			Desired:                  nodeManager.Spec.Aws.Workers.Desired,
			Role:                     operatorv1alpha1.Worker,
		},
		Status: operatorv1alpha1.AWSNodeManagerStatus{
			Phase: operatorv1alpha1.AWSNodeManagerInit,
		},
	}
}
