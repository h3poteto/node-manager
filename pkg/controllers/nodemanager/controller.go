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

package nodemanager

import (
	"context"
	"errors"
	"reflect"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	pkgctx "github.com/h3poteto/node-manager/pkg/util/context"
	"github.com/h3poteto/node-manager/pkg/util/klog"
	"github.com/h3poteto/node-manager/pkg/util/requestid"
)

const (
	NodeMasterLabel = "node-role.kubernetes.io/control-plane"
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
// +kubebuilder:rbac:groups=operator.h3poteto.dev,resources=awsnodemanagers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.h3poteto.dev,resources=awsnodemanagers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

func (r *NodeManagerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("nodemanager", req.NamespacedName)
	ctx = pkgctx.SetController(ctx, "nodemanager")
	id, err := requestid.RequestID()
	if err != nil {
		return ctrl.Result{}, err
	}
	ctx = pkgctx.SetRequestID(ctx, id)

	nodeManager := operatorv1alpha1.NodeManager{}
	// We watch NodeManager resources and Node resources.
	// So, the request can contains both of them.
	klog.Infof(ctx, "fetching NodeManager resources: %s", req.NamespacedName.Name)
	if err := r.Client.Get(ctx, req.NamespacedName, &nodeManager); err == nil {
		err := r.syncNodeManager(ctx, &nodeManager)
		if err != nil {
			r.Recorder.Eventf(&nodeManager, corev1.EventTypeWarning, "Error", "Failed to sync: %v", err)
		}
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
		Owns(&operatorv1alpha1.AWSNodeManager{}).
		Watches(&corev1.Node{}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}

func (r *NodeManagerReconciler) syncNodeManager(ctx context.Context, nodeManager *operatorv1alpha1.NodeManager) error {
	klog.Info(ctx, "fetching node resources")
	nodeList := corev1.NodeList{}
	if err := r.Client.List(ctx, &nodeList); err != nil {
		klog.Errorf(ctx, "failed to list nodes: %v", err)
		return err
	}
	var masterNodes, workerNodes []*corev1.Node
	for i := range nodeList.Items {
		node := &nodeList.Items[i]
		if _, ok := node.Labels[NodeMasterLabel]; ok {
			masterNodes = append(masterNodes, node)
		}
		if _, ok := node.Labels[NodeWorkerLabel]; ok {
			workerNodes = append(workerNodes, node)
		}
	}

	updated, err := r.reflectNodes(ctx, nodeManager, masterNodes, workerNodes)
	if err != nil {
		return err
	}
	if updated {
		return nil
	}

	switch nodeManager.Spec.CloudProvider {
	case "aws":
		if nodeManager.Spec.Aws == nil {
			err := errors.New("please specify spec.aws when cloudProvider is aws")
			klog.Error(ctx, err)
			return err
		}
		masterManager, workerManager, err := r.syncAWSNodeManager(ctx, nodeManager, masterNodes, workerNodes)
		if err != nil {
			return err
		}
		newStatus := nodeManager.Status.DeepCopy()
		if masterManager != nil {
			newStatus.MasterAWSNodeManager = &operatorv1alpha1.AWSNodeManagerRef{
				Namespace: masterManager.Namespace,
				Name:      masterManager.Name,
			}
		}
		if workerManager != nil {
			newStatus.WorkerAWSNodeManager = &operatorv1alpha1.AWSNodeManagerRef{
				Namespace: workerManager.Namespace,
				Name:      workerManager.Name,
			}
		}

		if reflect.DeepEqual(nodeManager.Status, newStatus) {
			klog.Infof(ctx, "NodeManager %s/%s is already synced", nodeManager.Namespace, nodeManager.Name)
			return nil
		}
		nodeManager.Status = *newStatus
		if err := r.Client.Update(ctx, nodeManager); err != nil {
			klog.Errorf(ctx, "failed to update nodeManager %q/%q: %v", nodeManager.Namespace, nodeManager.Name, err)
			return err
		}
		r.Recorder.Eventf(nodeManager, corev1.EventTypeNormal, "Updated", "Updated NodeManager %s/%s", nodeManager.Namespace, nodeManager.Name)
		klog.Infof(ctx, "updated NodeManager status %q/%q", nodeManager.Namespace, nodeManager.Name)
		return nil
	default:
		klog.Info(ctx, "could not find cloud provider in NodeManager resource")
		return nil
	}
}
