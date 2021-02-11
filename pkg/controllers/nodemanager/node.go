package nodemanager

import (
	"context"
	"errors"
	"reflect"
	"sort"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *NodeManagerReconciler) syncNode(ctx context.Context, resourceName types.NamespacedName) error {
	klog.Info("finding nodeManager resources")
	list := operatorv1alpha1.NodeManagerList{}
	if err := r.Client.List(ctx, &list); err != nil {
		klog.Errorf("failed to list nodeManagers: %v", err)
		return err
	}
	if len(list.Items) == 0 {
		klog.Warning("cloud not find any nodeManager resources")
		return nil
	}
	if len(list.Items) > 1 {
		return errors.New("found multiple nodeManager resources")
	}
	nodeManager := list.Items[0]

	klog.Infof("finding %s in node resources", resourceName.Name)
	node := &corev1.Node{}
	if err := r.Client.Get(ctx, resourceName, node); err != nil && apierrors.IsNotFound(err) {
		masterNode := findNameInList(nodeManager.Status.MasterNodes, resourceName.Name)
		workerNode := findNameInList(nodeManager.Status.WorkerNodes, resourceName.Name)
		if masterNode == "" && workerNode == "" {
			klog.Warningf("resource %s is not node", resourceName.Name)
			return err
		}
		klog.Infof("Node %s has been deleted", resourceName.Name)
	} else if err != nil {
		klog.Errorf("failed to get node: %v", err)
		return err
	}

	// Node is added, updated or deleted
	klog.Info("fetching all node resources")
	nodeList := corev1.NodeList{}
	if err := r.Client.List(ctx, &nodeList); err != nil {
		klog.Errorf("failed to list nodes: %v", err)
		return err
	}

	var masterNames, workerNames []string
	for i := range nodeList.Items {
		node := &nodeList.Items[i]
		if _, ok := node.Labels[NodeMasterLabel]; ok {
			masterNames = append(masterNames, node.Name)
			continue
		}
		if _, ok := node.Labels[NodeWorkerLabel]; ok {
			workerNames = append(workerNames, node.Name)
			continue
		}
		klog.Warningf("node %s does not have any node-role.kubernetes.io labels", node.Name)
	}
	sort.SliceStable(masterNames, func(i, j int) bool { return masterNames[i] < masterNames[j] })
	sort.SliceStable(workerNames, func(i, j int) bool { return workerNames[i] < workerNames[j] })

	klog.Infof("checking nodeManager status: %q/%q", nodeManager.Namespace, nodeManager.Name)
	status := nodeManager.Status.DeepCopy()
	sort.SliceStable(status.MasterNodes, func(i, j int) bool { return status.MasterNodes[i] < status.MasterNodes[j] })
	sort.SliceStable(status.WorkerNodes, func(i, j int) bool { return status.WorkerNodes[i] < status.WorkerNodes[j] })
	if reflect.DeepEqual(status.MasterNodes, masterNames) && reflect.DeepEqual(status.WorkerNodes, workerNames) {
		klog.Info("all nodes are already synced in nodeManager status")
		// AWSNodeManager have to handle updating node event, because sometimes it have to check current state of instance in order to add/delete instances.
		return r.updateAWSNodeManagerRevision(ctx, &nodeManager)
	}
	status.MasterNodes = masterNames
	status.WorkerNodes = workerNames

	klog.Infof("need status update, current status is %#v, node status is %#v", nodeManager.Status, *status)
	nodeManager.Status = *status
	if err := r.Client.Update(ctx, &nodeManager); err != nil {
		klog.Errorf("failed to update nodeManager status %s/%s: %v", nodeManager.Namespace, nodeManager.Name, err)
		return err
	}
	klog.Infof("success to update nodeManager status %s/%s for all nodes", nodeManager.Namespace, nodeManager.Name)
	return nil
}

func (r *NodeManagerReconciler) updateAWSNodeManagerRevision(ctx context.Context, nodeManager *operatorv1alpha1.NodeManager) error {
	if nodeManager.Status.MasterAWSNodeManager != nil {
		awsNodeManager := operatorv1alpha1.AWSNodeManager{}
		if err := r.Client.Get(
			ctx,
			client.ObjectKey{
				Namespace: nodeManager.Status.MasterAWSNodeManager.Namespace,
				Name:      nodeManager.Status.MasterAWSNodeManager.Name,
			},
			&awsNodeManager,
		); err != nil {
			klog.Errorf("failed to get aws node manager for master: %v", err)
			return err
		}
		awsNodeManager.Status.Revision += 1
		klog.Infof("updating revision for aws node manager %s/%s", awsNodeManager.Namespace, awsNodeManager.Name)
		if err := r.Client.Update(ctx, &awsNodeManager); err != nil {
			klog.Errorf("failed to update aws node manager %s/%s: %v", awsNodeManager.Namespace, awsNodeManager.Name, err)
			return err
		}
	}
	if nodeManager.Status.WorkerAWSNodeManager != nil {
		awsNodeManager := operatorv1alpha1.AWSNodeManager{}
		if err := r.Client.Get(
			ctx,
			client.ObjectKey{
				Namespace: nodeManager.Status.WorkerAWSNodeManager.Namespace,
				Name:      nodeManager.Status.WorkerAWSNodeManager.Name,
			},
			&awsNodeManager,
		); err != nil {
			klog.Errorf("failed to get aws node manager for worker: %v", err)
			return err
		}
		awsNodeManager.Status.Revision += 1
		klog.Infof("updating revision for aws node manager %s/%s", awsNodeManager.Namespace, awsNodeManager.Name)
		if err := r.Client.Update(ctx, &awsNodeManager); err != nil {
			klog.Errorf("failed to update aws node manager %s/%s: %v", awsNodeManager.Namespace, awsNodeManager.Name, err)
			return err
		}
	}
	return nil
}

func findNameInList(list []string, targetName string) string {
	for i := range list {
		if list[i] == targetName {
			return list[i]
		}
	}
	return ""
}
