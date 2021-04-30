package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
)

func (a *AWS) ReflectInstancesInformation(awsNodeManager *operatorv1alpha1.AWSNodeManager) error {
	for i, node := range awsNodeManager.Status.AWSNodes {
		if node.InstanceID != "" {
			continue
		}
		input := &ec2.DescribeInstancesInput{
			DryRun: nil,
			Filters: []*ec2.Filter{
				{
					Name: aws.String("private-dns-name"),
					Values: []*string{
						aws.String(node.Name),
					},
				},
			},
		}
		output, err := a.EC2.DescribeInstances(input)
		if err != nil {
			klog.Errorf("failed to describe aws instances: %v", err)
			return err
		}
		if len(output.Reservations) < 1 || len(output.Reservations[0].Instances) < 1 {
			klog.Warningf("could not find aws instance %s", node.Name)
			continue
		}
		instance := output.Reservations[0].Instances[0]
		awsNodeManager.Status.AWSNodes[i].InstanceID = *instance.InstanceId
		awsNodeManager.Status.AWSNodes[i].InstanceType = *instance.InstanceType
		awsNodeManager.Status.AWSNodes[i].AvailabilityZone = *instance.Placement.AvailabilityZone
		// Normally auto scaling group name is filled in name tag of instances.
		tag := findTag(instance.Tags, "Name")
		if tag == nil {
			klog.Warningf("could not find Name tag in aws instance %s", *instance.InstanceId)
			continue
		}
		awsNodeManager.Status.AWSNodes[i].AutoScalingGroupName = *tag.Value
	}
	return nil
}

func findTag(tags []*ec2.Tag, key string) *ec2.Tag {
	for i := range tags {
		if *tags[i].Key == key {
			return tags[i]
		}
	}
	return nil
}

func (a *AWS) DeleteInstance(node *operatorv1alpha1.AWSNode) error {
	input := &ec2.TerminateInstancesInput{
		DryRun: nil,
		InstanceIds: []*string{
			aws.String(node.InstanceID),
		},
	}
	_, err := a.EC2.TerminateInstances(input)
	if err != nil {
		klog.Errorf("failed to terminate instance %s: %v", node.InstanceID, err)
		return err
	}
	return nil
}

func (a *AWS) DescribeInstance(node *operatorv1alpha1.AWSNode) (*ec2.Instance, error) {
	input := &ec2.DescribeInstancesInput{
		DryRun: nil,
		Filters: []*ec2.Filter{
			{
				Name: aws.String("private-dns-name"),
				Values: []*string{
					aws.String(node.Name),
				},
			},
		},
	}
	output, err := a.EC2.DescribeInstances(input)
	if err != nil {
		klog.Errorf("failed to describe aws instances: %v", err)
		return nil, err
	}
	if len(output.Reservations) < 1 || len(output.Reservations[0].Instances) < 1 {
		err := fmt.Errorf("could not find aws instance %s", node.Name)
		klog.Error(err)
		return nil, err
	}
	instance := output.Reservations[0].Instances[0]
	return instance, nil
}

func (a *AWS) GetAWSNodes(instanceIDs []*string) ([]operatorv1alpha1.AWSNode, error) {
	input := &ec2.DescribeInstancesInput{
		DryRun:      nil,
		InstanceIds: instanceIDs,
	}
	output, err := a.EC2.DescribeInstances(input)
	if err != nil {
		klog.Errorf("failed to describe ec2 instances: %v", err)
		return nil, err
	}
	var nodes []operatorv1alpha1.AWSNode
	for _, r := range output.Reservations {
		for _, instance := range r.Instances {
			tag := findTag(instance.Tags, "Name")
			n := operatorv1alpha1.AWSNode{
				Name:                 *instance.PrivateDnsName,
				InstanceID:           *instance.InstanceId,
				AvailabilityZone:     *instance.Placement.AvailabilityZone,
				InstanceType:         *instance.InstanceType,
				AutoScalingGroupName: *tag.Value,
				CreationTimestamp:    metav1.NewTime(instance.LaunchTime.In(time.Local)),
			}
			nodes = append(nodes, n)
		}
	}
	return nodes, nil
}
