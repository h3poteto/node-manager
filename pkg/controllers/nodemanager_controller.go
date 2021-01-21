/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"errors"
	"reflect"
	"sort"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
)

const (
	NodeMasterLabel = "node-role.kubernetes.io/master"
	NodeWorkerLabel = "node-role.kubernetes.io/node"
)

// NodeManagerReconciler reconciles a NodeManager object
type NodeManagerReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=operator.h3poteto.dev,resources=nodemanagers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.h3poteto.dev,resources=nodemanagers/status,verbs=get;update;patch

func (r *NodeManagerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("nodemanager", req.NamespacedName)

	nodeManager := operatorv1alpha1.NodeManager{}
	// We watch NodeManager resources and Node resources.
	// So, the request can contains both of them.
	klog.Infof("fetching NodeManager resources: %s", req.NamespacedName.Name)
	if err := r.Client.Get(ctx, req.NamespacedName, &nodeManager); err == nil {
		err := r.syncNodeManager(ctx, &nodeManager)
		return ctrl.Result{}, err
	}
	if err := r.syncNode(ctx, req.NamespacedName); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	return ctrl.Result{}, nil
}

func (r *NodeManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.NodeManager{}).
		Owns(&operatorv1alpha1.AWSNodeReplenisher{}).
		Watches(&source.Kind{Type: &corev1.Node{}}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}

func (r *NodeManagerReconciler) syncNodeManager(ctx context.Context, nodeManager *operatorv1alpha1.NodeManager) error {
	klog.Info("fetching node resources")
	nodeList := corev1.NodeList{}
	if err := r.Client.List(ctx, &nodeList); err != nil {
		klog.Errorf("failed to list nodes: %v", err)
		return err
	}
	var masterNames []string
	var workerNames []string
	for i := range nodeList.Items {
		if _, ok := nodeList.Items[i].Labels[NodeMasterLabel]; ok {
			masterNames = append(masterNames, nodeList.Items[i].Name)
		}
		if _, ok := nodeList.Items[i].Labels[NodeWorkerLabel]; ok {
			workerNames = append(workerNames, nodeList.Items[i].Name)
		}
	}

	switch nodeManager.Spec.CloudProvider {
	case "aws":
		if nodeManager.Spec.Aws == nil {
			err := errors.New("please specify spec.aws when cloudProvider is aws")
			klog.Error(err)
			return err
		}
		masterName, workerName, err := r.syncAWSNodeReplenisher(ctx, nodeManager, masterNames, workerNames)
		if err != nil {
			return err
		}
		nodeManager.Status.MasterNodeReplenisherName = masterName
		nodeManager.Status.WorkerNodeReplenisherName = workerName
		nodeManager.Status.MasterNodes = masterNames
		nodeManager.Status.WorkerNodes = workerNames

		currentNodeManager := operatorv1alpha1.NodeManager{}
		if err := r.Client.Get(ctx, client.ObjectKey{Namespace: nodeManager.Namespace, Name: nodeManager.Name}, &currentNodeManager); err != nil {
			klog.Errorf("failed to get NodeManager %s/%s: %v", nodeManager.Namespace, nodeManager.Name, err)
			return err
		}
		if reflect.DeepEqual(nodeManager.Status, currentNodeManager.Status) {
			klog.Infof("NodeManager %s/%s is already synced", nodeManager.Namespace, nodeManager.Name)
			return nil
		}
		currentNodeManager.Status = nodeManager.Status
		if err := r.Client.Update(ctx, &currentNodeManager); err != nil {
			klog.Errorf("failed to update nodeManager %q/%q: %v", currentNodeManager.Namespace, currentNodeManager.Name, err)
			return err
		}
		klog.Infof("updated NodeManager status %q/%q", currentNodeManager.Namespace, currentNodeManager.Name)
		return nil
	default:
		klog.Info("could not find cloud provider in NodeManager resource")
		return nil
	}
}

func (r *NodeManagerReconciler) syncAWSNodeReplenisher(ctx context.Context, nodeManager *operatorv1alpha1.NodeManager, masterNames, workerNames []string) (string, string, error) {
	klog.Info("Syncing AWSNodeReplenisher from NodeManager")
	var masterName, workerName string
	if nodeManager.Spec.Aws.Masters != nil {
		masterReplenisher, err := r.syncMasterAWSNodeReplenisher(ctx, nodeManager, masterNames)
		if err != nil {
			return "", "", err
		}
		masterName = masterReplenisher.Name
	}
	if nodeManager.Spec.Aws.Workers != nil {
		workerReplenisher, err := r.syncWorkerAWSNodeReplenisher(ctx, nodeManager, workerNames)
		if err != nil {
			return "", "", err
		}
		workerName = workerReplenisher.Name
	}
	return masterName, workerName, nil
}

func (r *NodeManagerReconciler) syncMasterAWSNodeReplenisher(ctx context.Context, nodeManager *operatorv1alpha1.NodeManager, masterNames []string) (*operatorv1alpha1.AWSNodeReplenisher, error) {
	klog.Info("checking if an existing AWSNodeReplenisher for master")
	existingNodeReplenisher := operatorv1alpha1.AWSNodeReplenisher{}
	err := r.Client.Get(ctx, client.ObjectKey{Namespace: nodeManager.Namespace, Name: nodeManager.Status.MasterNodeReplenisherName}, &existingNodeReplenisher)

	if nodeManager.Status.MasterNodeReplenisherName == "" || apierrors.IsNotFound(err) {
		klog.Info("AWSNodeReplenisher for master does not exist, so create it")
		nodeReplenisher := generateMasterAWSNodeReplenisher(nodeManager)
		for i := range masterNames {
			a := operatorv1alpha1.AWSNode{
				Name: masterNames[i],
			}
			nodeReplenisher.Status.AWSNodes = append(nodeReplenisher.Status.AWSNodes, a)
		}
		if err := r.Client.Create(ctx, nodeReplenisher); err != nil {
			return nil, err
		}

		r.Recorder.Eventf(nodeReplenisher, corev1.EventTypeNormal, "Created", "Created AWSNodeReplenisher for master %s/%s", nodeReplenisher.Namespace, nodeReplenisher.Name)
		klog.Infof("created AWSNodeReplenisher for master %q/%q", nodeReplenisher.Namespace, nodeReplenisher.Name)

		return nodeReplenisher, nil
	}
	if err != nil {
		klog.Errorf("failed to get AWSNodeReplenisher for master: %v", err)
		return nil, err
	}

	nodeReplenisher := generateMasterAWSNodeReplenisher(nodeManager)
	var nodeNames []string
	for _, node := range existingNodeReplenisher.Status.AWSNodes {
		nodeNames = append(nodeNames, node.Name)
	}
	if reflect.DeepEqual(existingNodeReplenisher.Spec, nodeReplenisher.Spec) && reflect.DeepEqual(nodeNames, nodeManager.Status.MasterNodes) {
		klog.Infof("AWSNodeReplenisher %q/%q is already synced", existingNodeReplenisher.Namespace, existingNodeReplenisher.Name)
		return &existingNodeReplenisher, nil
	}
	existingNodeReplenisher.Spec = nodeReplenisher.Spec
	existingNodeReplenisher.Status.AWSNodes = []operatorv1alpha1.AWSNode{}
	for i := range nodeManager.Status.MasterNodes {
		a := operatorv1alpha1.AWSNode{
			Name: nodeManager.Status.MasterNodes[i],
		}
		existingNodeReplenisher.Status.AWSNodes = append(existingNodeReplenisher.Status.AWSNodes, a)
	}
	existingNodeReplenisher.Status.Revision += 1
	if err := r.Client.Update(ctx, &existingNodeReplenisher); err != nil {
		klog.Errorf("failed to update existing AWSNodeRepelnisher %q/%q : %v", existingNodeReplenisher.Namespace, existingNodeReplenisher.Name, err)
		return nil, err
	}
	klog.Infof("updated AWSNodeReplenisher spec for master %q/%q", existingNodeReplenisher.Namespace, existingNodeReplenisher.Name)

	return &existingNodeReplenisher, nil
}

func (r *NodeManagerReconciler) syncWorkerAWSNodeReplenisher(ctx context.Context, nodeManager *operatorv1alpha1.NodeManager, workerNames []string) (*operatorv1alpha1.AWSNodeReplenisher, error) {
	klog.Info("checking if an existing AWSNodeReplenisher for worker")
	existingNodeReplenisher := operatorv1alpha1.AWSNodeReplenisher{}
	err := r.Client.Get(ctx, client.ObjectKey{Namespace: nodeManager.Namespace, Name: nodeManager.Status.WorkerNodeReplenisherName}, &existingNodeReplenisher)

	if nodeManager.Status.WorkerNodeReplenisherName == "" || apierrors.IsNotFound(err) {
		klog.Info("AWSNodeReplenisher for worker does not exist, so create it")
		nodeReplenisher := generateWorkerAWSNodeReplenisher(nodeManager)
		for i := range workerNames {
			a := operatorv1alpha1.AWSNode{
				Name: workerNames[i],
			}
			nodeReplenisher.Status.AWSNodes = append(nodeReplenisher.Status.AWSNodes, a)
		}
		if err := r.Client.Create(ctx, nodeReplenisher); err != nil {
			klog.Errorf("failed to create AWSNodeReplenisher %q/%q: %v", nodeReplenisher.Namespace, nodeReplenisher.Name, err)
			return nil, err
		}

		r.Recorder.Eventf(nodeReplenisher, corev1.EventTypeNormal, "Created", "Created AWSNodeReplenisher for worker %s/%s", nodeReplenisher.Namespace, nodeReplenisher.Name)
		klog.Infof("created AWSNodeReplenisher for worker %q/%q", nodeReplenisher.Namespace, nodeReplenisher.Name)

		return nodeReplenisher, nil
	}
	if err != nil {
		klog.Errorf("failed to get AWSNodeReplenisher for worker: %v", err)
		return nil, err
	}

	nodeReplenisher := generateWorkerAWSNodeReplenisher(nodeManager)
	var nodeNames []string
	for _, node := range existingNodeReplenisher.Status.AWSNodes {
		nodeNames = append(nodeNames, node.Name)
	}
	if reflect.DeepEqual(existingNodeReplenisher.Spec, nodeReplenisher.Spec) && reflect.DeepEqual(nodeNames, nodeManager.Status.WorkerNodes) {
		klog.Infof("AWSNodeReplenisher %q/%q is already synced", existingNodeReplenisher.Namespace, existingNodeReplenisher.Name)
		return &existingNodeReplenisher, nil
	}
	existingNodeReplenisher.Spec = nodeReplenisher.Spec
	existingNodeReplenisher.Status.AWSNodes = []operatorv1alpha1.AWSNode{}
	for i := range nodeManager.Status.WorkerNodes {
		a := operatorv1alpha1.AWSNode{
			Name: nodeManager.Status.WorkerNodes[i],
		}
		existingNodeReplenisher.Status.AWSNodes = append(existingNodeReplenisher.Status.AWSNodes, a)
	}
	existingNodeReplenisher.Status.Revision += 1
	if err := r.Client.Update(ctx, &existingNodeReplenisher); err != nil {
		klog.Errorf("failed to update AWSNodeReplenisher %s/%s: %v", existingNodeReplenisher.Namespace, existingNodeReplenisher.Name, err)
		return nil, err
	}
	klog.Infof("updated AWSNodeReplenisher spec for worker %q/%q", existingNodeReplenisher.Namespace, existingNodeReplenisher.Name)
	return &existingNodeReplenisher, nil
}

func generateMasterAWSNodeReplenisher(nodeManager *operatorv1alpha1.NodeManager) *operatorv1alpha1.AWSNodeReplenisher {
	replenisher := &operatorv1alpha1.AWSNodeReplenisher{
		ObjectMeta: metav1.ObjectMeta{
			Name:            nodeManager.Name + "-master",
			Namespace:       nodeManager.Namespace,
			Labels:          nodeManager.GetLabels(),
			Annotations:     nodeManager.GetAnnotations(),
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(nodeManager, operatorv1alpha1.GroupVersion.WithKind("NodeManager"))},
		},
		Spec: operatorv1alpha1.AWSNodeReplenisherSpec{
			Region:                   nodeManager.Spec.Aws.Region,
			Role:                     operatorv1alpha1.Worker,
			AutoScalingGroups:        nodeManager.Spec.Aws.Masters.AutoScalingGroups,
			Desired:                  nodeManager.Spec.Aws.Masters.Desired,
			ASGModifyCoolTimeSeconds: nodeManager.Spec.Aws.Masters.ASGModifyCoolTimeSeconds,
		},
	}
	return replenisher
}

func generateWorkerAWSNodeReplenisher(nodeManager *operatorv1alpha1.NodeManager) *operatorv1alpha1.AWSNodeReplenisher {
	replenisher := &operatorv1alpha1.AWSNodeReplenisher{
		ObjectMeta: metav1.ObjectMeta{
			Name:            nodeManager.Name + "-worker",
			Namespace:       nodeManager.Namespace,
			Labels:          nodeManager.GetLabels(),
			Annotations:     nodeManager.GetAnnotations(),
			OwnerReferences: []metav1.OwnerReference{*metav1.NewControllerRef(nodeManager, operatorv1alpha1.GroupVersion.WithKind("NodeManager"))},
		},
		Spec: operatorv1alpha1.AWSNodeReplenisherSpec{
			Region:                   nodeManager.Spec.Aws.Region,
			Role:                     operatorv1alpha1.Worker,
			AutoScalingGroups:        nodeManager.Spec.Aws.Workers.AutoScalingGroups,
			Desired:                  nodeManager.Spec.Aws.Workers.Desired,
			ASGModifyCoolTimeSeconds: nodeManager.Spec.Aws.Workers.ASGModifyCoolTimeSeconds,
		},
	}
	return replenisher
}

func (r *NodeManagerReconciler) syncNode(ctx context.Context, resourceName types.NamespacedName) error {
	klog.Info("finding nodeManager resources")
	list := operatorv1alpha1.NodeManagerList{}
	if err := r.Client.List(ctx, &list); err != nil {
		klog.Errorf("failed to list nodeManagers: %v", err)
		return err
	}
	if len(list.Items) == 0 {
		klog.Warning("could not find any nodeManager resources")
		return nil
	}
	if len(list.Items) > 1 {
		return errors.New("found multiple nodeManager resources")
	}
	nodeManager := list.Items[0]

	klog.Infof("finding %q in node resources", resourceName.Name)
	node := &corev1.Node{}
	if err := r.Client.Get(ctx, resourceName, node); err != nil && apierrors.IsNotFound(err) {
		mNode := findNameInList(nodeManager.Status.MasterNodes, resourceName.Name)
		wNode := findNameInList(nodeManager.Status.WorkerNodes, resourceName.Name)
		if mNode == "" && wNode == "" {
			// resource is not node
			return err
		}
		// Node is deleted
	} else if err != nil {
		klog.Errorf("failed to get node: %v", err)
		return err
	}

	// Node is added, updated or deleted
	klog.Info("fetching node resources")
	nodeList := corev1.NodeList{}
	if err := r.Client.List(ctx, &nodeList); err != nil {
		klog.Errorf("failed to list nodes: %v", err)
		return err
	}

	var masterNames, workerNames []string
	for i := range nodeList.Items {
		if _, ok := nodeList.Items[i].Labels[NodeMasterLabel]; ok {
			masterNames = append(masterNames, nodeList.Items[i].Name)
			continue
		}
		if _, ok := nodeList.Items[i].Labels[NodeWorkerLabel]; ok {
			workerNames = append(workerNames, nodeList.Items[i].Name)
			continue
		}
		klog.Warningf("node %q does not have any node-role.kubernetes.io labels", nodeList.Items[i].Name)
	}
	sort.SliceStable(masterNames, func(i, j int) bool { return masterNames[i] < masterNames[j] })
	sort.SliceStable(workerNames, func(i, j int) bool { return workerNames[i] < workerNames[j] })

	klog.Infof("checking nodeManager status: %q/%q", nodeManager.Namespace, nodeManager.Name)
	status := nodeManager.Status.DeepCopy()
	sort.SliceStable(status.MasterNodes, func(i, j int) bool { return status.MasterNodes[i] < status.MasterNodes[j] })
	sort.SliceStable(status.WorkerNodes, func(i, j int) bool { return status.WorkerNodes[i] < status.WorkerNodes[j] })
	if reflect.DeepEqual(status.MasterNodes, masterNames) && reflect.DeepEqual(status.WorkerNodes, workerNames) {
		klog.Info("all nodes are already synced in nodeManager status")
		// Node replenisher have to handle updating node event, because sometimes it have to check current state of instance in order to add/delete instances.
		return r.updateReplenisherRevision(ctx, &nodeManager)
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

func (r *NodeManagerReconciler) updateReplenisherRevision(ctx context.Context, nodeManager *operatorv1alpha1.NodeManager) error {
	if nodeManager.Status.MasterNodeReplenisherName != "" {
		replenisher := operatorv1alpha1.AWSNodeReplenisher{}
		err := r.Client.Get(ctx, client.ObjectKey{Namespace: nodeManager.Namespace, Name: nodeManager.Status.MasterNodeReplenisherName}, &replenisher)
		if err != nil {
			klog.Errorf("failed to get node replenisher for master: %v", err)
			return err
		}
		replenisher.Status.Revision += 1
		klog.Infof("updating revision of node replenisher %s", replenisher.Name)
		if err := r.Client.Update(ctx, &replenisher); err != nil {
			klog.Errorf("failed to update node replenisher %s: %v", replenisher.Name, err)
			return err
		}
	}
	if nodeManager.Status.WorkerNodeReplenisherName != "" {
		replenisher := operatorv1alpha1.AWSNodeReplenisher{}
		err := r.Client.Get(ctx, client.ObjectKey{Namespace: nodeManager.Namespace, Name: nodeManager.Status.WorkerNodeReplenisherName}, &replenisher)
		if err != nil {
			klog.Errorf("failed to get node replenisher for worker: %v", err)
			return err
		}
		replenisher.Status.Revision += 1
		klog.Infof("updating revision of node replenisher %s", replenisher.Name)
		if err := r.Client.Update(ctx, &replenisher); err != nil {
			klog.Errorf("failed to update node replenisher %s: %v", replenisher.Name, err)
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
