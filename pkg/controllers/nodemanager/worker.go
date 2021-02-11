package nodemanager

import (
	"context"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *NodeManagerReconciler) syncWorkerAWSNodeManager(ctx context.Context, nodeManager *operatorv1alpha1.NodeManager, workerNodes []*corev1.Node) (*operatorv1alpha1.AWSNodeManager, error) {
	klog.Info("checking if an existing AWSNodeManager for worker")
	if nodeManager.Status.WorkerAWSNodeManager == nil {
		return r.createAWSNodeManager(ctx, nodeManager, workerNodes, operatorv1alpha1.Worker)
	}
	existingAWSNodeManager := operatorv1alpha1.AWSNodeManager{}
	err := r.Client.Get(
		ctx,
		client.ObjectKey{
			Namespace: nodeManager.Status.WorkerAWSNodeManager.Namespace,
			Name:      nodeManager.Status.WorkerAWSNodeManager.Name,
		},
		&existingAWSNodeManager,
	)
	if apierrors.IsNotFound(err) {
		klog.Info("AWSNodeManager for worker does not exist, so create it")
		return r.createAWSNodeManager(ctx, nodeManager, workerNodes, operatorv1alpha1.Worker)
	}
	if err != nil {
		klog.Errorf("failed to get AWSNodeManager for worker :%v", err)
		return nil, err
	}

	return r.updateAWSNodeManager(ctx, &existingAWSNodeManager, nodeManager, workerNodes, operatorv1alpha1.Worker)
}
