package awsnodemanager

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	cloudaws "github.com/h3poteto/node-manager/pkg/cloud/aws"
	"github.com/h3poteto/node-manager/pkg/util/klog"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *AWSNodeManagerReconciler) syncAWSNodes(ctx context.Context, awsNodeManager *operatorv1alpha1.AWSNodeManager) (bool, error) {
	cloud := cloudaws.New(r.Session, awsNodeManager.Spec.Region)
	if err := reflectInstances(ctx, cloud, awsNodeManager); err != nil {
		return false, err
	}
	klog.Info(ctx, "Checking not joined instances")
	if err := reflectNotJoinedInstances(cloud, awsNodeManager); err != nil {
		return false, err
	}

	currentManager := operatorv1alpha1.AWSNodeManager{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: awsNodeManager.Namespace, Name: awsNodeManager.Name}, &currentManager); err != nil {
		klog.Errorf(ctx, "failed to get AWSNodeManager: %v", err)
		return false, err
	}
	if reflect.DeepEqual(currentManager.Status, awsNodeManager.Status) {
		klog.Infof(ctx, "AWSNodeManager %s/%s is already synced", awsNodeManager.Namespace, awsNodeManager.Name)
		return false, nil
	}
	currentManager.Status = awsNodeManager.Status
	currentManager.Status.Revision += 1
	if currentManager.Status.Phase == operatorv1alpha1.AWSNodeManagerInit {
		currentManager.Status.Phase = operatorv1alpha1.AWSNodeManagerSynced
	}
	// update awsNodeManager status
	klog.Infof(ctx, "updating AWSNodeManager status: %s/%s", currentManager.Namespace, currentManager.Name)
	if err := r.Client.Update(ctx, &currentManager); err != nil {
		klog.Errorf(ctx, "failed to update AWSNodeManager %s/%s: %v", currentManager.Namespace, currentManager.Name, err)
		return false, err
	}
	klog.Infof(ctx, "success to update AWSNodeManager %s/%s", currentManager.Namespace, currentManager.Name)
	r.Recorder.Eventf(&currentManager, corev1.EventTypeNormal, "Updated", "Updated AWSNodeManager %s/%s", currentManager.Namespace, currentManager.Name)
	return true, nil
}

func reflectInstances(ctx context.Context, cloud *cloudaws.AWS, awsNodeManager *operatorv1alpha1.AWSNodeManager) error {
	for i := range awsNodeManager.Status.AWSNodes {
		node := &awsNodeManager.Status.AWSNodes[i]
		if node.InstanceID != "" {
			continue
		}
		instance, err := cloud.DescribeInstance(node)
		if err != nil {
			return err
		}
		n, err := cloudaws.ConvertInstanceToAWSNode(instance)
		if err != nil {
			klog.Warning(ctx, err)
			continue
		}
		awsNodeManager.Status.AWSNodes[i].InstanceID = n.InstanceID
		awsNodeManager.Status.AWSNodes[i].InstanceType = n.InstanceType
		awsNodeManager.Status.AWSNodes[i].AvailabilityZone = n.AvailabilityZone
		awsNodeManager.Status.AWSNodes[i].AutoScalingGroupName = n.AutoScalingGroupName
	}
	return nil
}

func reflectNotJoinedInstances(cloud *cloudaws.AWS, awsNodeManager *operatorv1alpha1.AWSNodeManager) error {
	groups, err := cloud.DescribeAutoScalingGroups(awsNodeManager.Spec.AutoScalingGroups)
	if err != nil {
		return err
	}
	var instanceIDs []*string
	for _, group := range groups {
		for _, instance := range group.Instances {
			if includedCluster(instance, awsNodeManager.Status.AWSNodes) {
				continue
			}
			instanceIDs = append(instanceIDs, instance.InstanceId)
		}
	}
	if len(instanceIDs) == 0 {
		return nil
	}
	nodes, err := cloud.GetAWSNodes(instanceIDs)
	if err != nil {
		return err
	}
	awsNodeManager.Status.NotJoinedAWSNodes = nodes
	return nil
}

func includedCluster(instance *autoscaling.Instance, awsNodes []operatorv1alpha1.AWSNode) bool {
	for _, node := range awsNodes {
		if *instance.InstanceId == node.InstanceID {
			return true
		}
	}
	return false
}
