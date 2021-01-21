package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/klog/v2"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
)

func (a *AWS) GetInstancesInformation(replenisher *operatorv1alpha1.AWSNodeReplenisher) error {
	for i, node := range replenisher.Status.AWSNodes {
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
		output, err := a.ec2.DescribeInstances(input)
		if err != nil {
			klog.Errorf("failed to describe aws instances: %v", err)
			return err
		}
		if len(output.Reservations) < 1 || len(output.Reservations[0].Instances) < 1 {
			klog.Warningf("could not find aws instance %s", node.Name)
			continue
		}
		instance := output.Reservations[0].Instances[0]
		replenisher.Status.AWSNodes[i].InstanceID = *instance.InstanceId
		replenisher.Status.AWSNodes[i].InstanceType = *instance.InstanceType
		replenisher.Status.AWSNodes[i].AvailabilityZone = *instance.Placement.AvailabilityZone
		// Normally auto scaling group name is filled in name tag of instances.
		tag := findTag(instance.Tags, "Name")
		if tag == nil {
			klog.Warningf("could not find Name tag in aws instance %s", *instance.InstanceId)
			continue
		}
		replenisher.Status.AWSNodes[i].AutoScalingGroupName = *tag.Value
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
