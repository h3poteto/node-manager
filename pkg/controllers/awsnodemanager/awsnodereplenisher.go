package awsnodemanager

import (
	"context"
	"reflect"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *AWSNodeManagerReconciler) syncAWSNodeReplenisher(ctx context.Context, awsNodeManager *operatorv1alpha1.AWSNodeManager) (*operatorv1alpha1.AWSNodeReplenisher, error) {
	if !awsNodeManager.Spec.EnableReplenish {
		return nil, nil
	}
	existingReplenisher, err := r.fetchExistingReplenisher(ctx, awsNodeManager)
	if apierrors.IsNotFound(err) || existingReplenisher == nil {
		klog.Info("AWSNodeReplenisher does not exist, so create it")
		return r.createAWSNodeReplenisher(ctx, awsNodeManager)
	}
	if err != nil {
		klog.Errorf("failed to get AWSNodeReplenisher: %v", err)
		return nil, err
	}

	return r.updateAWSNodeReplenisher(ctx, existingReplenisher, awsNodeManager)
}

func (r *AWSNodeManagerReconciler) fetchExistingReplenisher(ctx context.Context, awsNodeManager *operatorv1alpha1.AWSNodeManager) (*operatorv1alpha1.AWSNodeReplenisher, error) {
	existingReplenisher := operatorv1alpha1.AWSNodeReplenisher{}
	if awsNodeManager.Status.NodeReplenisher != nil {
		err := r.Client.Get(
			ctx,
			client.ObjectKey{
				Namespace: awsNodeManager.Status.NodeReplenisher.Namespace,
				Name:      awsNodeManager.Status.NodeReplenisher.Name,
			},
			&existingReplenisher,
		)
		if err != nil {
			return nil, err
		}
		return &existingReplenisher, nil
	}
	newReplenisher := generateAWSNodeReplenisher(awsNodeManager)
	err := r.Client.Get(
		ctx,
		client.ObjectKey{
			Namespace: newReplenisher.Namespace,
			Name:      newReplenisher.Name,
		},
		&existingReplenisher,
	)
	if err != nil {
		return nil, err
	}
	return &existingReplenisher, nil
}

func (r *AWSNodeManagerReconciler) createAWSNodeReplenisher(ctx context.Context, awsNodeManager *operatorv1alpha1.AWSNodeManager) (*operatorv1alpha1.AWSNodeReplenisher, error) {
	klog.Info("creating AWSNodeReplenisher")
	newReplenisher := generateAWSNodeReplenisher(awsNodeManager)
	if err := r.Client.Create(ctx, newReplenisher); err != nil {
		klog.Errorf("failed to create AWSNodeReplenisher: %v", err)
		return nil, err
	}
	r.Recorder.Eventf(newReplenisher, corev1.EventTypeNormal, "Created", "Created AWSNodeReplenisher %s/%s", newReplenisher.Namespace, newReplenisher.Name)
	return newReplenisher, nil
}

func (r *AWSNodeManagerReconciler) updateAWSNodeReplenisher(ctx context.Context, existingReplenisher *operatorv1alpha1.AWSNodeReplenisher, awsNodeManager *operatorv1alpha1.AWSNodeManager) (*operatorv1alpha1.AWSNodeReplenisher, error) {
	newReplenisher := generateAWSNodeReplenisher(awsNodeManager)
	if reflect.DeepEqual(existingReplenisher.Spec, newReplenisher.Spec) && reflect.DeepEqual(existingReplenisher.Status.AWSNodes, newReplenisher.Status.AWSNodes) {
		klog.Infof("AWSNodeReplenisher %s/%s is already synced", existingReplenisher.Namespace, existingReplenisher.Name)
		return existingReplenisher, nil
	}
	existingReplenisher.Spec = newReplenisher.Spec
	existingReplenisher.Status.AWSNodes = newReplenisher.Status.AWSNodes
	existingReplenisher.Status.Revision += 1
	if err := r.Client.Update(ctx, existingReplenisher); err != nil {
		klog.Errorf("failed to update existing AWSNodeReplenisher %s/%s: %v", existingReplenisher.Namespace, existingReplenisher.Name, err)
		return nil, err
	}
	klog.Infof("updated AWSNodeReplenisher %s/%s", existingReplenisher.Namespace, existingReplenisher.Name)
	r.Recorder.Eventf(existingReplenisher, corev1.EventTypeNormal, "Updated", "Updated AWSNodeReplenisher %s/%s", existingReplenisher.Namespace, existingReplenisher.Name)
	return existingReplenisher, nil
}

func generateAWSNodeReplenisher(awsNodeManager *operatorv1alpha1.AWSNodeManager) *operatorv1alpha1.AWSNodeReplenisher {
	return &operatorv1alpha1.AWSNodeReplenisher{
		ObjectMeta: metav1.ObjectMeta{
			Name:        awsNodeManager.Name,
			Namespace:   awsNodeManager.Namespace,
			Labels:      awsNodeManager.GetLabels(),
			Annotations: awsNodeManager.GetAnnotations(),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(awsNodeManager, awsNodeManager.GroupVersionKind()),
			},
		},
		Spec: operatorv1alpha1.AWSNodeReplenisherSpec{
			Region:                   awsNodeManager.Spec.Region,
			AutoScalingGroups:        awsNodeManager.Spec.AutoScalingGroups,
			Desired:                  awsNodeManager.Spec.Desired,
			ASGModifyCoolTimeSeconds: awsNodeManager.Spec.ASGModifyCoolTimeSeconds,
			Role:                     awsNodeManager.Spec.Role,
		},
		Status: operatorv1alpha1.AWSNodeReplenisherStatus{
			AWSNodes: awsNodeManager.Status.AWSNodes,
			Revision: 0,
			Phase:    operatorv1alpha1.AWSNodeReplenisherInit,
		},
	}
}
