package awsnodemanager

import (
	"context"
	"reflect"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	"github.com/h3poteto/node-manager/pkg/util/klog"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *AWSNodeManagerReconciler) syncAWSNodeRefresher(ctx context.Context, awsNodeManager *operatorv1alpha1.AWSNodeManager) (*operatorv1alpha1.AWSNodeRefresher, error) {
	if awsNodeManager.Spec.RefreshSchedule == "" {
		return nil, nil
	}
	klog.Info(ctx, "checking if an existing AWSNodeRefresher")
	existingRefresher, err := r.fetchExistingRefresher(ctx, awsNodeManager)
	if apierrors.IsNotFound(err) || existingRefresher == nil {
		klog.Info(ctx, "AWSNodeRefresher does not exist, so create it")
		return r.createAWSNodeRefresher(ctx, awsNodeManager)
	}
	if err != nil {
		klog.Errorf(ctx, "failed to get AWSNodeRefresher: %v", err)
		return nil, err
	}

	return r.updateAWSNodeRefresher(ctx, existingRefresher, awsNodeManager)
}

func (r *AWSNodeManagerReconciler) fetchExistingRefresher(ctx context.Context, awsNodeManager *operatorv1alpha1.AWSNodeManager) (*operatorv1alpha1.AWSNodeRefresher, error) {
	existingRefresher := operatorv1alpha1.AWSNodeRefresher{}
	if awsNodeManager.Status.NodeRefresher != nil {
		err := r.Client.Get(
			ctx,
			client.ObjectKey{
				Namespace: awsNodeManager.Status.NodeRefresher.Namespace,
				Name:      awsNodeManager.Status.NodeRefresher.Name,
			},
			&existingRefresher,
		)
		if err != nil {
			return nil, err
		}
		return &existingRefresher, nil
	}
	newRefresher := generateAWSNodeRefresher(awsNodeManager)
	err := r.Client.Get(
		ctx,
		client.ObjectKey{
			Namespace: newRefresher.Namespace,
			Name:      newRefresher.Name,
		},
		&existingRefresher,
	)
	if err != nil {
		return nil, err
	}
	return &existingRefresher, nil
}

func (r *AWSNodeManagerReconciler) createAWSNodeRefresher(ctx context.Context, awsNodeManager *operatorv1alpha1.AWSNodeManager) (*operatorv1alpha1.AWSNodeRefresher, error) {
	klog.Infof(ctx, "creating AWSNodeRefresher")
	newRefresher := generateAWSNodeRefresher(awsNodeManager)
	if err := r.Client.Create(ctx, newRefresher); err != nil {
		klog.Errorf(ctx, "failed to create AWSNodeRefresher: %v", err)
		return nil, err
	}
	r.Recorder.Eventf(newRefresher, corev1.EventTypeNormal, "Created", "Created AWSNodeRefresher %s/%s", newRefresher.Namespace, newRefresher.Name)
	return newRefresher, nil
}

func (r *AWSNodeManagerReconciler) updateAWSNodeRefresher(ctx context.Context, existingRefresher *operatorv1alpha1.AWSNodeRefresher, awsNodeManager *operatorv1alpha1.AWSNodeManager) (*operatorv1alpha1.AWSNodeRefresher, error) {
	newRefresher := generateAWSNodeRefresher(awsNodeManager)
	if reflect.DeepEqual(existingRefresher.Spec, newRefresher.Spec) && reflect.DeepEqual(existingRefresher.Status.AWSNodes, newRefresher.Status.AWSNodes) {
		return existingRefresher, nil
	}
	existingRefresher.Spec = newRefresher.Spec
	existingRefresher.Status.AWSNodes = newRefresher.Status.AWSNodes
	existingRefresher.Status.Revision += 1
	if err := r.Client.Update(ctx, existingRefresher); err != nil {
		klog.Errorf(ctx, "failed to update existing AWSNodeRefresher %s/%s: %v", existingRefresher.Namespace, existingRefresher.Name, err)
		return nil, err
	}
	klog.Infof(ctx, "updated AWSNodeRefresher %s/%s", existingRefresher.Namespace, existingRefresher.Name)
	r.Recorder.Eventf(existingRefresher, corev1.EventTypeNormal, "Updated", "Updated AWSNodeRefresher %s/%s", existingRefresher.Namespace, existingRefresher.Name)
	return existingRefresher, nil
}

func generateAWSNodeRefresher(awsNodeManager *operatorv1alpha1.AWSNodeManager) *operatorv1alpha1.AWSNodeRefresher {
	return &operatorv1alpha1.AWSNodeRefresher{
		ObjectMeta: metav1.ObjectMeta{
			Name:        awsNodeManager.Name,
			Namespace:   awsNodeManager.Namespace,
			Labels:      awsNodeManager.GetLabels(),
			Annotations: awsNodeManager.GetAnnotations(),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(awsNodeManager, awsNodeManager.GroupVersionKind()),
			},
		},
		Spec: operatorv1alpha1.AWSNodeRefresherSpec{
			Region:                   awsNodeManager.Spec.Region,
			AutoScalingGroups:        awsNodeManager.Spec.AutoScalingGroups,
			Desired:                  awsNodeManager.Spec.Desired,
			ASGModifyCoolTimeSeconds: awsNodeManager.Spec.ASGModifyCoolTimeSeconds,
			Role:                     awsNodeManager.Spec.Role,
			Schedule:                 awsNodeManager.Spec.RefreshSchedule,
			SurplusNodes:             awsNodeManager.Spec.SurplusNodes,
		},
		Status: operatorv1alpha1.AWSNodeRefresherStatus{
			AWSNodes: awsNodeManager.Status.AWSNodes,
			Revision: 0,
			Phase:    operatorv1alpha1.AWSNodeRefresherInit,
		},
	}
}
