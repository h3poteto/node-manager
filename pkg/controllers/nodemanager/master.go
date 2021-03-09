package nodemanager

import (
	"context"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	"github.com/h3poteto/node-manager/pkg/util/klog"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *NodeManagerReconciler) syncMasterAWSNodeManager(ctx context.Context, nodeManager *operatorv1alpha1.NodeManager, masterNodes []*corev1.Node) (*operatorv1alpha1.AWSNodeManager, error) {
	klog.Info(ctx, "checking if an existing AWSNodeManager for master")
	if nodeManager.Status.MasterAWSNodeManager == nil {
		return r.createAWSNodeManager(ctx, nodeManager, masterNodes, operatorv1alpha1.Master)
	}
	existingAWSNodeManager := operatorv1alpha1.AWSNodeManager{}
	err := r.Client.Get(
		ctx,
		client.ObjectKey{
			Namespace: nodeManager.Status.MasterAWSNodeManager.Namespace,
			Name:      nodeManager.Status.MasterAWSNodeManager.Name,
		},
		&existingAWSNodeManager,
	)
	if apierrors.IsNotFound(err) {
		klog.Info(ctx, "AWSNodeManager for master does not exist, so create it")
		return r.createAWSNodeManager(ctx, nodeManager, masterNodes, operatorv1alpha1.Master)
	}
	if err != nil {
		klog.Errorf(ctx, "failed to get AWSNodeManager for master :%v", err)
		return nil, err
	}

	return r.updateAWSNodeManager(ctx, &existingAWSNodeManager, nodeManager, masterNodes, operatorv1alpha1.Master)
}
