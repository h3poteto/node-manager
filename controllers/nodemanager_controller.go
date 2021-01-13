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

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	node := corev1.Node{}

	// We watch NodeManager resources and Node resources.
	// So, the request can contains both of them.
	var managerErr error
	var nodeErr error

	klog.Info("fetching NodeManager resources")
	managerErr = r.Client.Get(ctx, req.NamespacedName, &nodeManager)
	if managerErr == nil {
		err := r.syncNodeManager(ctx, &nodeManager)
		return ctrl.Result{}, err
	}
	nodeErr = r.Client.Get(ctx, req.NamespacedName, &node)
	if nodeErr == nil {
		err := r.syncNode(ctx, &node)
		return ctrl.Result{}, err
	}

	if managerErr != nil {
		klog.Errorf("failed to get NodeManager resources: %v", managerErr)
		return ctrl.Result{}, client.IgnoreNotFound(managerErr)
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
		newStatus := nodeManager.Status.DeepCopy()
		newStatus.MasterNodeReplenisherName = masterName
		newStatus.WorkerNodeReplenisherName = workerName
		newStatus.MasterNodes = masterNames
		newStatus.WorkerNodes = workerNames
		if reflect.DeepEqual(nodeManager.Status, newStatus) {
			return nil
		}
		nodeManager.Status = *newStatus
		if err := r.Client.Update(ctx, nodeManager); err != nil {
			klog.Errorf("failed to update nodeManager %q/%q: %v", nodeManager.Namespace, nodeManager.Name, err)
			return err
		}
		klog.Infof("updated NodeManager status %q/%q", nodeManager.Namespace, nodeManager.Name)
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
		nodeReplenisher.Status.Nodes = masterNames
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
	if reflect.DeepEqual(existingNodeReplenisher.Spec, nodeReplenisher.Spec) && reflect.DeepEqual(existingNodeReplenisher.Status.Nodes, nodeManager.Status.MasterNodes) {
		klog.Infof("AWSNodeReplenisher %q/%q is already synced", existingNodeReplenisher.Namespace, existingNodeReplenisher.Name)
		return &existingNodeReplenisher, nil
	}
	existingNodeReplenisher.Spec = nodeReplenisher.Spec
	existingNodeReplenisher.Status.Nodes = nodeManager.Status.MasterNodes
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
		nodeReplenisher.Status.Nodes = workerNames
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
	if reflect.DeepEqual(existingNodeReplenisher.Spec, nodeReplenisher.Spec) && reflect.DeepEqual(existingNodeReplenisher.Status.Nodes, nodeManager.Status.WorkerNodes) {
		klog.Infof("AWSNodeReplenisher %q/%q is already synced", existingNodeReplenisher.Namespace, existingNodeReplenisher.Name)
		return &existingNodeReplenisher, nil
	}
	existingNodeReplenisher.Spec = nodeReplenisher.Spec
	existingNodeReplenisher.Status.Nodes = nodeManager.Status.WorkerNodes
	if err := r.Client.Update(ctx, &existingNodeReplenisher); err != nil {
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
			AutoScalingGroups:  nodeManager.Spec.Aws.Masters.AutoScalingGroups,
			Desired:            nodeManager.Spec.Aws.Masters.Desired,
			ScaleInWaitSeconds: nodeManager.Spec.Aws.Masters.ScaleInWaitSeconds,
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
			AutoScalingGroups:  nodeManager.Spec.Aws.Workers.AutoScalingGroups,
			Desired:            nodeManager.Spec.Aws.Workers.Desired,
			ScaleInWaitSeconds: nodeManager.Spec.Aws.Workers.ScaleInWaitSeconds,
		},
	}
	return replenisher
}

func (r *NodeManagerReconciler) syncNode(ctx context.Context, node *corev1.Node) error {
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

	klog.Infof("checking nodeManager status: %q/%q", nodeManager.Namespace, nodeManager.Name)
	if _, ok := node.Labels[NodeMasterLabel]; ok {
		if name := findNameInList(nodeManager.Status.MasterNodes, node.Name); name != "" {
			klog.Infof("node %s is already included in nodeManager status", node.Name)
			return nil
		}
		nodeManager.Status.MasterNodes = append(nodeManager.Status.MasterNodes, node.Name)
		if err := r.Client.Update(ctx, &nodeManager); err != nil {
			klog.Errorf("failed to update nodeManager %q/%q: %v", nodeManager.Namespace, nodeManager.Name, err)
			return err
		}
		return nil
	}
	if _, ok := node.Labels[NodeWorkerLabel]; ok {
		if name := findNameInList(nodeManager.Status.WorkerNodes, node.Name); name != "" {
			klog.Infof("node %s is already included in nodeManager status", node.Name)
			return nil
		}
		nodeManager.Status.WorkerNodes = append(nodeManager.Status.WorkerNodes, node.Name)
		if err := r.Client.Update(ctx, &nodeManager); err != nil {
			klog.Errorf("failed to update nodeManager %q/%q: %v", nodeManager.Namespace, nodeManager.Name, err)
			return err
		}
		return nil
	}
	klog.Warningf("node %s does not have any node-role.kubernetes.io labels", node.Name)
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
