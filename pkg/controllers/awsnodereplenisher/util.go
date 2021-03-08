package awsnodereplenisher

import (
	"context"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *AWSNodeReplenisherReconciler) ownerAWSNodeManager(ctx context.Context, replenisher *operatorv1alpha1.AWSNodeReplenisher) (*operatorv1alpha1.AWSNodeManager, error) {
	ref := metav1.GetControllerOf(replenisher)
	if ref == nil || ref.APIVersion != operatorv1alpha1.SchemeBuilder.GroupVersion.String() || ref.Kind != "AWSNodeManager" {
		klog.Warningf("could not find owner of %s/%s", replenisher.Namespace, replenisher.Name)
		return nil, nil
	}

	manager := operatorv1alpha1.AWSNodeManager{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: replenisher.Namespace, Name: ref.Name}, &manager); err != nil {
		klog.Errorf("failed to get AWSNodeManager %s/%s: %v", replenisher.Namespace, ref.Name, err)
		return nil, err
	}
	return &manager, nil
}
