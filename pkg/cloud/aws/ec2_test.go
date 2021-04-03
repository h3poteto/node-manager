package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
)

type mockedEC2API struct {
	ec2iface.EC2API
	Resp ec2.DescribeInstancesOutput
}

func (m *mockedEC2API) DescribeInstances(in *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	return &m.Resp, nil
}

func TestGetInstanceInformation(t *testing.T) {

	replenisher := &operatorv1alpha1.AWSNodeManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-manager",
			Namespace: "default",
		},
		Spec: operatorv1alpha1.AWSNodeManagerSpec{
			Region:                   "us-east-1",
			AutoScalingGroups:        nil,
			Desired:                  0,
			ASGModifyCoolTimeSeconds: 600,
			Role:                     "master",
		},
		Status: operatorv1alpha1.AWSNodeManagerStatus{
			AWSNodes: []operatorv1alpha1.AWSNode{
				{
					Name:                 "ip-172-32-16-0",
					InstanceID:           "",
					AvailabilityZone:     "",
					InstanceType:         "",
					AutoScalingGroupName: "",
				},
			},
			Phase: operatorv1alpha1.AWSNodeManagerInit,
		},
	}

	resp := ec2.DescribeInstancesOutput{
		NextToken: nil,
		Reservations: []*ec2.Reservation{
			&ec2.Reservation{
				Groups: nil,
				Instances: []*ec2.Instance{
					&ec2.Instance{
						InstanceId:   aws.String("instanceId-1"),
						InstanceType: aws.String(ec2.InstanceTypeT3Small),
						Placement: &ec2.Placement{
							AvailabilityZone: aws.String("us-east-1"),
						},
						PrivateDnsName:   aws.String("ip-172-32-16-0"),
						PrivateIpAddress: aws.String("172.32.16.0"),
						Tags: []*ec2.Tag{
							&ec2.Tag{
								Key:   aws.String("Name"),
								Value: aws.String("autoscaling-group-name"),
							},
						},
					},
				},
			},
		},
	}

	mocked := &mockedEC2API{
		Resp: resp,
	}

	a := &AWS{
		EC2: mocked,
	}

	if err := a.ReflectInstancesInformation(replenisher); err != nil {
		t.Error(err)
		return
	}

	if replenisher.Status.AWSNodes[0].InstanceID != "instanceId-1" {
		t.Errorf("InstanceID is not matched, expected: %s, return: %s", "instanceId-1", replenisher.Status.AWSNodes[0].InstanceID)
		return
	}

	if replenisher.Status.AWSNodes[0].InstanceType != ec2.InstanceTypeT3Small {
		t.Errorf("InstanceType is not matched, expected: %s, return: %s", ec2.InstanceTypeT3Small, replenisher.Status.AWSNodes[0].InstanceType)
		return
	}

	if replenisher.Status.AWSNodes[0].AvailabilityZone != "us-east-1" {
		t.Errorf("AvailabilityZone is not matched, expected: %s, return: %s", "us-east-1", replenisher.Status.AWSNodes[0].AvailabilityZone)
		return
	}

	if replenisher.Status.AWSNodes[0].AutoScalingGroupName != "autoscaling-group-name" {
		t.Errorf("AutoScalingGroupName is not matched, expected: %s, return: %s", "autoscaling-group-name", replenisher.Status.AWSNodes[0].AutoScalingGroupName)
		return
	}
}
