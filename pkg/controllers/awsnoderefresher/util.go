package awsnoderefresher

import (
	"context"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/h3poteto/node-manager/pkg/util/klog"
)

func (r *AWSNodeRefresherReconciler) ownerAWSNodeManager(ctx context.Context, refresher *operatorv1alpha1.AWSNodeRefresher) (*operatorv1alpha1.AWSNodeManager, error) {
	ref := metav1.GetControllerOf(refresher)
	if ref == nil || ref.APIVersion != operatorv1alpha1.SchemeBuilder.GroupVersion.String() || ref.Kind != "AWSNodeManager" {
		klog.Warningf(ctx, "could not find owner of %s/%s", refresher.Namespace, refresher.Name)
		return nil, nil
	}

	manager := operatorv1alpha1.AWSNodeManager{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: refresher.Namespace, Name: ref.Name}, &manager); err != nil {
		klog.Errorf(ctx, "failed to get AWSNodeMAnager %s/%s: %v", refresher.Namespace, ref.Name, err)
		return nil, err
	}
	return &manager, nil
}
