package aws

import (
	"errors"
	"log"
	"reflect"
	"testing"
	"time"

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

func TestGetAWSNodes(t *testing.T) {
	creationTimestamp := time.Now().Add(-1 * time.Hour)
	cases := []struct {
		title            string
		instanceIDs      []*string
		describeResponse ec2.DescribeInstancesOutput
		expectedNodes    []operatorv1alpha1.AWSNode
	}{
		{
			title: "Multiple instances",
			instanceIDs: []*string{
				aws.String("instanceId-1"),
				aws.String("instanceId-2"),
				aws.String("instanceId-3"),
			},
			describeResponse: ec2.DescribeInstancesOutput{
				NextToken: nil,
				Reservations: []*ec2.Reservation{
					&ec2.Reservation{
						Instances: []*ec2.Instance{
							&ec2.Instance{
								InstanceId:   aws.String("instanceId-1"),
								InstanceType: aws.String(ec2.InstanceTypeT3Small),
								LaunchTime:   aws.Time(creationTimestamp),
								Placement: &ec2.Placement{
									AvailabilityZone: aws.String("us-east-1a"),
								},
								PrivateDnsName:   aws.String("ip-172-32-16-0"),
								PrivateIpAddress: aws.String("172.32.16.0"),
								Tags: []*ec2.Tag{
									&ec2.Tag{
										Key:   aws.String("Name"),
										Value: aws.String("asg-1"),
									},
								},
							},
							&ec2.Instance{
								InstanceId:   aws.String("instanceId-2"),
								InstanceType: aws.String(ec2.InstanceTypeT3Small),
								LaunchTime:   aws.Time(creationTimestamp),
								Placement: &ec2.Placement{
									AvailabilityZone: aws.String("us-east-1c"),
								},
								PrivateDnsName:   aws.String("ip-172-32-16-1"),
								PrivateIpAddress: aws.String("172.32.16.1"),
								Tags: []*ec2.Tag{
									&ec2.Tag{
										Key:   aws.String("Name"),
										Value: aws.String("asg-1"),
									},
								},
							},
							&ec2.Instance{
								InstanceId:   aws.String("instanceId-3"),
								InstanceType: aws.String(ec2.InstanceTypeT3Small),
								LaunchTime:   aws.Time(creationTimestamp),
								Placement: &ec2.Placement{
									AvailabilityZone: aws.String("us-east-1d"),
								},
								PrivateDnsName:   aws.String("ip-172-32-16-2"),
								PrivateIpAddress: aws.String("172.32.16.2"),
								Tags: []*ec2.Tag{
									&ec2.Tag{
										Key:   aws.String("Name"),
										Value: aws.String("asg-1"),
									},
								},
							},
						},
					},
				},
			},
			expectedNodes: []operatorv1alpha1.AWSNode{
				{
					Name:                 "ip-172-32-16-0",
					InstanceID:           "instanceId-1",
					AvailabilityZone:     "us-east-1a",
					InstanceType:         ec2.InstanceTypeT3Small,
					AutoScalingGroupName: "asg-1",
					CreationTimestamp: metav1.Time{
						Time: creationTimestamp.In(time.Local),
					},
				},
				{
					Name:                 "ip-172-32-16-1",
					InstanceID:           "instanceId-2",
					AvailabilityZone:     "us-east-1c",
					InstanceType:         ec2.InstanceTypeT3Small,
					AutoScalingGroupName: "asg-1",
					CreationTimestamp: metav1.Time{
						Time: creationTimestamp.In(time.Local),
					},
				},
				{
					Name:                 "ip-172-32-16-2",
					InstanceID:           "instanceId-3",
					AvailabilityZone:     "us-east-1d",
					InstanceType:         ec2.InstanceTypeT3Small,
					AutoScalingGroupName: "asg-1",
					CreationTimestamp: metav1.Time{
						Time: creationTimestamp.In(time.Local),
					},
				},
			},
		},
	}

	for _, c := range cases {
		log.Printf("Running CASE %s", c.title)

		a := &AWS{
			EC2: &mockedEC2API{
				Resp: c.describeResponse,
			},
		}
		node, err := a.GetAWSNodes(c.instanceIDs)
		if err != nil {
			t.Errorf("CASE: %s : error has occur: %v", c.title, err)
			continue
		}
		if !reflect.DeepEqual(node, c.expectedNodes) {
			t.Errorf("CASE: %s : node is not matched, expected %+v, returned %+v", c.title, c.expectedNodes, node)
		}
	}
}

func TestConvertInstanceToAWSNode(t *testing.T) {
	creationTimestamp := time.Now().Add(-1 * time.Hour)
	cases := []struct {
		title         string
		instance      *ec2.Instance
		awsNode       *operatorv1alpha1.AWSNode
		expectedError error
	}{
		{
			title: "Does not include Name tag",
			instance: &ec2.Instance{
				InstanceId:   aws.String("instanceId-1"),
				InstanceType: aws.String(ec2.InstanceTypeT3Small),
				Placement: &ec2.Placement{
					AvailabilityZone: aws.String("us-east-1a"),
				},
				PrivateDnsName:   aws.String("ip-172-32-16-0"),
				PrivateIpAddress: aws.String("172.32.16.0"),
			},
			expectedError: NewCouldNotFoundNameTagError("could not find Name tag in aws instances %s", "instanceId-1"),
		},
		{
			title: "Include Name tag",
			instance: &ec2.Instance{
				InstanceId:   aws.String("instanceId-1"),
				InstanceType: aws.String(ec2.InstanceTypeT3Small),
				LaunchTime:   aws.Time(creationTimestamp),
				Placement: &ec2.Placement{
					AvailabilityZone: aws.String("us-east-1a"),
				},
				PrivateDnsName:   aws.String("ip-172-32-16-0"),
				PrivateIpAddress: aws.String("172.32.16.0"),
				Tags: []*ec2.Tag{
					&ec2.Tag{
						Key:   aws.String("Name"),
						Value: aws.String("asg-1"),
					},
				},
			},
			awsNode: &operatorv1alpha1.AWSNode{
				Name:                 "ip-172-32-16-0",
				InstanceID:           "instanceId-1",
				AvailabilityZone:     "us-east-1a",
				InstanceType:         ec2.InstanceTypeT3Small,
				AutoScalingGroupName: "asg-1",
				CreationTimestamp: metav1.Time{
					Time: creationTimestamp.In(time.Local),
				},
			},
			expectedError: nil,
		},
	}

	for _, c := range cases {
		log.Printf("Running CASE: %s", c.title)

		node, err := ConvertInstanceToAWSNode(c.instance)
		if err != nil {
			if !errors.Is(err, c.expectedError) {
				t.Errorf("CASE: %s : error is not matched, expected %v, returned %v", c.title, c.expectedError, err)
			}
			continue
		}
		if !reflect.DeepEqual(*node, *c.awsNode) {
			t.Errorf("CASE: %s : node is not matched, expected %+v, returned %+v", c.title, *c.awsNode, *node)
		}
	}
}
