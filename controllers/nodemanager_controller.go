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

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
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
	log := r.Log.WithValues("nodemanager", req.NamespacedName)

	log.Info("fetching NodeManager resources")
	kind := operatorv1alpha1.NodeManager{}
	if err := r.Client.Get(ctx, req.NamespacedName, &kind); err != nil {
		log.Error(err, "failed to get NodeManager resources")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	switch kind.Spec.CloudProvider {
	case "aws":
		if kind.Spec.Aws == nil {
			err := errors.New("please specify spec.aws when cloudProvider is aws")
			return ctrl.Result{}, err
		}
		err := r.syncAWSNodeReplenisher(ctx, log, &kind)
		if err != nil {
			log.Error(err, "failed to create AWSNodeReplenisher resource")
			return ctrl.Result{}, err
		}
	default:
		log.Info("could not find cloud provider in NodeManager resource")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *NodeManagerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.NodeManager{}).
		Owns(&operatorv1alpha1.AWSNodeReplenisher{}).
		Complete(r)
}

func (r *NodeManagerReconciler) syncAWSNodeReplenisher(ctx context.Context, log logr.Logger, nodeManager *operatorv1alpha1.NodeManager) error {
	if nodeManager.Spec.Aws.Masters != nil {
		err := r.syncMasterAWSNodeReplenisher(ctx, log, nodeManager)
		if err != nil {
			return err
		}
	}
	if nodeManager.Spec.Aws.Workers != nil {
		err := r.syncWorkerAWSNodeReplenisher(ctx, log, nodeManager)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *NodeManagerReconciler) syncMasterAWSNodeReplenisher(ctx context.Context, log logr.Logger, nodeManager *operatorv1alpha1.NodeManager) error {
	log.Info("checking if an existing AWSNodeReplenisher for master")
	existingNodeReplenisher := operatorv1alpha1.AWSNodeReplenisher{}
	err := r.Client.Get(ctx, client.ObjectKey{Namespace: nodeManager.Namespace, Name: nodeManager.Status.MasterNodeReplenisherName}, &existingNodeReplenisher)

	if nodeManager.Status.MasterNodeReplenisherName == "" || apierrors.IsNotFound(err) {
		nodeReplenisher := generateMasterAWSNodeReplenisher(nodeManager)
		if err := r.Client.Create(ctx, nodeReplenisher); err != nil {
			return err
		}

		r.Recorder.Eventf(nodeReplenisher, corev1.EventTypeNormal, "Created", "Created AWSNodeReplenisher for master %q/%q", nodeReplenisher.Namespace, nodeReplenisher.Name)
		log.Info("created AWSNodeReplenisher for master %q/%q", nodeReplenisher.Namespace, nodeReplenisher.Name)

		nodeManager.Status.MasterNodeReplenisherName = nodeReplenisher.Name
		if err := r.Client.Update(ctx, nodeManager); err != nil {
			log.Error(err, "failed to update NodeManager status")
			return err
		}
		log.Info("updated NodeManager status")
		return nil
	}
	if err != nil {
		return err
	}

	nodeReplenisher := generateMasterAWSNodeReplenisher(nodeManager)
	existingNodeReplenisher.Spec = nodeReplenisher.Spec
	if err := r.Client.Update(ctx, &existingNodeReplenisher); err != nil {
		return err
	}
	log.Info("updated AWSNodeReplenisher spec for master %q/%q", existingNodeReplenisher.Namespace, existingNodeReplenisher.Name)

	return nil
}

func (r *NodeManagerReconciler) syncWorkerAWSNodeReplenisher(ctx context.Context, log logr.Logger, nodeManager *operatorv1alpha1.NodeManager) error {
	log.Info("checking if an existing AWSNodeReplenisher for worker")
	existingNodeReplenisher := operatorv1alpha1.AWSNodeReplenisher{}
	err := r.Client.Get(ctx, client.ObjectKey{Namespace: nodeManager.Namespace, Name: nodeManager.Status.WorkerNodeReplenisherName}, &existingNodeReplenisher)

	if nodeManager.Status.WorkerNodeReplenisherName == "" || apierrors.IsNotFound(err) {
		nodeReplenisher := generateWorkerAWSNodeReplenisher(nodeManager)
		if err := r.Client.Create(ctx, nodeReplenisher); err != nil {
			return err
		}

		r.Recorder.Eventf(nodeReplenisher, corev1.EventTypeNormal, "Created", "Created AWSNodeReplenisher for worker %q/%q", nodeReplenisher.Namespace, nodeReplenisher.Name)
		log.Info("created AWSNodeReplenisher for worker %q/%q", nodeReplenisher.Namespace, nodeReplenisher.Name)

		nodeManager.Status.MasterNodeReplenisherName = nodeReplenisher.Name
		if err := r.Client.Update(ctx, nodeManager); err != nil {
			log.Error(err, "failed to update NodeManager status")
			return err
		}
		log.Info("updated NodeManager status")
		return nil
	}
	if err != nil {
		return err
	}

	nodeReplenisher := generateWorkerAWSNodeReplenisher(nodeManager)
	existingNodeReplenisher.Spec = nodeReplenisher.Spec
	if err := r.Client.Update(ctx, &existingNodeReplenisher); err != nil {
		return err
	}
	log.Info("updated AWSNodeReplenisher spec for worker %q/%q", existingNodeReplenisher.Namespace, existingNodeReplenisher.Name)
	return nil
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
			Name:            nodeManager.Name + "-master",
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
