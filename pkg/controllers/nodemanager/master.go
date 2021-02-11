package nodemanager

import (
	"context"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *NodeManagerReconciler) syncMasterAWSNodeManager(ctx context.Context, nodeManager *operatorv1alpha1.NodeManager, masterNames []string) (*operatorv1alpha1.AWSNodeManager, error) {
	klog.Info("checking if an existing AWSNodeManager for master")
	if nodeManager.Status.MasterAWSNodeManager == nil {
		return r.createAWSNodeManager(ctx, nodeManager, masterNames, operatorv1alpha1.Master)
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
		klog.Info("AWSNodeManager for master does not exist, so create it")
		return r.createAWSNodeManager(ctx, nodeManager, masterNames, operatorv1alpha1.Master)
	}
	if err != nil {
		klog.Errorf("failed to get AWSNodeManager for master :%v", err)
		return nil, err
	}

	return r.updateAWSNodeManager(ctx, &existingAWSNodeManager, nodeManager, masterNames, operatorv1alpha1.Master)
}
