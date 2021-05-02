package awsnodemanager

import (
	"context"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	cloudaws "github.com/h3poteto/node-manager/pkg/cloud/aws"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestSyncAWSNodes(t *testing.T) {
	creationTimestamp := time.Now().Add(-1 * time.Hour)
	cases := []struct {
		title                 string
		describeInstanceResp  *ec2.DescribeInstancesOutput
		getAWSNodesResp       *ec2.DescribeInstancesOutput
		asgDescribeResp       *autoscaling.DescribeAutoScalingGroupsOutput
		awsNodeManager        *operatorv1alpha1.AWSNodeManager
		currentAWSNodeManager *operatorv1alpha1.AWSNodeManager
		expectedUpdated       bool
		expectedStatus        *operatorv1alpha1.AWSNodeManagerStatus
	}{
		{
			title: "Contains not joined instances",
			describeInstanceResp: &ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{
					&ec2.Reservation{
						Instances: []*ec2.Instance{
							&ec2.Instance{
								InstanceId:   aws.String("instanceId-1"),
								InstanceType: aws.String(ec2.InstanceTypeT3Nano),
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
						},
					},
				},
			},
			getAWSNodesResp: &ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{
					&ec2.Reservation{
						Instances: []*ec2.Instance{
							&ec2.Instance{
								InstanceId:   aws.String("instanceId-2"),
								InstanceType: aws.String(ec2.InstanceTypeT3Nano),
								LaunchTime:   aws.Time(creationTimestamp),
								Placement: &ec2.Placement{
									AvailabilityZone: aws.String("us-east-1a"),
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
						},
					},
				},
			},
			asgDescribeResp: &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []*autoscaling.Group{
					&autoscaling.Group{
						AutoScalingGroupName: aws.String("asg-1"),
						AvailabilityZones: []*string{
							aws.String("us-east-1a"),
						},
						DesiredCapacity: aws.Int64(2),
						Instances: []*autoscaling.Instance{
							&autoscaling.Instance{
								AvailabilityZone: aws.String("us-east-1a"),
								InstanceId:       aws.String("instanceId-1"),
								InstanceType:     aws.String(ec2.InstanceTypeT3Nano),
							},
							&autoscaling.Instance{
								AvailabilityZone: aws.String("us-east-1a"),
								InstanceId:       aws.String("instanceId-2"),
								InstanceType:     aws.String(ec2.InstanceTypeT3Nano),
							},
						},
						MaxSize: aws.Int64(3),
						MinSize: aws.Int64(0),
					},
				},
			},
			awsNodeManager: &operatorv1alpha1.AWSNodeManager{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-manager",
				},
				Spec: operatorv1alpha1.AWSNodeManagerSpec{
					Region: "us-east-1",
					AutoScalingGroups: []operatorv1alpha1.AutoScalingGroup{
						{
							Name: "asg-1",
						},
					},
					Desired:                  2,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     "worker",
					EnableReplenish:          true,
					RefreshSchedule:          "",
					SurplusNodes:             1,
				},
				Status: operatorv1alpha1.AWSNodeManagerStatus{
					NodeReplenisher: &operatorv1alpha1.AWSNodeReplenisherRef{
						Namespace: "",
						Name:      "",
					},
					NodeRefresher: &operatorv1alpha1.AWSNodeRefresherRef{
						Namespace: "",
						Name:      "",
					},
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "ip-172-32-16-0",
							InstanceID:           "",
							AvailabilityZone:     "",
							InstanceType:         "",
							AutoScalingGroupName: "",
							CreationTimestamp: metav1.Time{
								Time: creationTimestamp,
							},
						},
					},
					NotJoinedAWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Time{},
					},
					Revision: 1,
					Phase:    operatorv1alpha1.AWSNodeManagerInit,
				},
			},
			currentAWSNodeManager: &operatorv1alpha1.AWSNodeManager{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-manager",
				},
				Spec: operatorv1alpha1.AWSNodeManagerSpec{
					Region: "us-east-1",
					AutoScalingGroups: []operatorv1alpha1.AutoScalingGroup{
						{
							Name: "asg-1",
						},
					},
					Desired:                  2,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     "worker",
					EnableReplenish:          true,
					RefreshSchedule:          "",
					SurplusNodes:             1,
				},
				Status: operatorv1alpha1.AWSNodeManagerStatus{
					NodeReplenisher: &operatorv1alpha1.AWSNodeReplenisherRef{
						Namespace: "",
						Name:      "",
					},
					NodeRefresher: &operatorv1alpha1.AWSNodeRefresherRef{
						Namespace: "",
						Name:      "",
					},
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "ip-172-32-16-0",
							InstanceID:           "",
							AvailabilityZone:     "",
							InstanceType:         "",
							AutoScalingGroupName: "",
							CreationTimestamp: metav1.Time{
								Time: creationTimestamp,
							},
						},
					},
					NotJoinedAWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Time{},
					},
					Revision: 1,
					Phase:    operatorv1alpha1.AWSNodeManagerInit,
				},
			},
			expectedUpdated: true,
			expectedStatus: &operatorv1alpha1.AWSNodeManagerStatus{
				NodeReplenisher: &operatorv1alpha1.AWSNodeReplenisherRef{
					Namespace: "",
					Name:      "",
				},
				NodeRefresher: &operatorv1alpha1.AWSNodeRefresherRef{
					Namespace: "",
					Name:      "",
				},
				AWSNodes: []operatorv1alpha1.AWSNode{
					{
						Name:                 "ip-172-32-16-0",
						InstanceID:           "instanceId-1",
						AvailabilityZone:     "us-east-1a",
						InstanceType:         ec2.InstanceTypeT3Nano,
						AutoScalingGroupName: "asg-1",
						CreationTimestamp: metav1.Time{
							Time: creationTimestamp,
						},
					},
				},
				NotJoinedAWSNodes: []operatorv1alpha1.AWSNode{
					{
						Name:                 "ip-172-32-16-1",
						InstanceID:           "instanceId-2",
						AvailabilityZone:     "us-east-1a",
						InstanceType:         ec2.InstanceTypeT3Nano,
						AutoScalingGroupName: "asg-1",
						CreationTimestamp: metav1.Time{
							Time: creationTimestamp.In(time.Local),
						},
					},
				},
				LastASGModifiedTime: &metav1.Time{
					Time: time.Time{},
				},
				Revision: 2,
				Phase:    operatorv1alpha1.AWSNodeManagerSynced,
			},
		},
		{
			title: "Does not contain not joined instances",
			describeInstanceResp: &ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{
					&ec2.Reservation{
						Instances: []*ec2.Instance{
							&ec2.Instance{
								InstanceId:   aws.String("instanceId-1"),
								InstanceType: aws.String(ec2.InstanceTypeT3Nano),
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
						},
					},
				},
			},
			getAWSNodesResp: &ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{
					&ec2.Reservation{
						Instances: []*ec2.Instance{
							&ec2.Instance{
								InstanceId:   aws.String("instanceId-2"),
								InstanceType: aws.String(ec2.InstanceTypeT3Nano),
								LaunchTime:   aws.Time(creationTimestamp),
								Placement: &ec2.Placement{
									AvailabilityZone: aws.String("us-east-1a"),
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
						},
					},
				},
			},
			asgDescribeResp: &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []*autoscaling.Group{
					&autoscaling.Group{
						AutoScalingGroupName: aws.String("asg-1"),
						AvailabilityZones: []*string{
							aws.String("us-east-1a"),
						},
						DesiredCapacity: aws.Int64(1),
						Instances: []*autoscaling.Instance{
							&autoscaling.Instance{
								AvailabilityZone: aws.String("us-east-1a"),
								InstanceId:       aws.String("instanceId-1"),
								InstanceType:     aws.String(ec2.InstanceTypeT3Nano),
							},
						},
						MaxSize: aws.Int64(3),
						MinSize: aws.Int64(0),
					},
				},
			},
			awsNodeManager: &operatorv1alpha1.AWSNodeManager{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-manager",
				},
				Spec: operatorv1alpha1.AWSNodeManagerSpec{
					Region: "us-east-1",
					AutoScalingGroups: []operatorv1alpha1.AutoScalingGroup{
						{
							Name: "asg-1",
						},
					},
					Desired:                  2,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     "worker",
					EnableReplenish:          true,
					RefreshSchedule:          "",
					SurplusNodes:             1,
				},
				Status: operatorv1alpha1.AWSNodeManagerStatus{
					NodeReplenisher: &operatorv1alpha1.AWSNodeReplenisherRef{
						Namespace: "",
						Name:      "",
					},
					NodeRefresher: &operatorv1alpha1.AWSNodeRefresherRef{
						Namespace: "",
						Name:      "",
					},
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "ip-172-32-16-0",
							InstanceID:           "",
							AvailabilityZone:     "",
							InstanceType:         "",
							AutoScalingGroupName: "",
							CreationTimestamp: metav1.Time{
								Time: creationTimestamp,
							},
						},
					},
					NotJoinedAWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Time{},
					},
					Revision: 1,
					Phase:    operatorv1alpha1.AWSNodeManagerInit,
				},
			},
			currentAWSNodeManager: &operatorv1alpha1.AWSNodeManager{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-manager",
				},
				Spec: operatorv1alpha1.AWSNodeManagerSpec{
					Region: "us-east-1",
					AutoScalingGroups: []operatorv1alpha1.AutoScalingGroup{
						{
							Name: "asg-1",
						},
					},
					Desired:                  2,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     "worker",
					EnableReplenish:          true,
					RefreshSchedule:          "",
					SurplusNodes:             1,
				},
				Status: operatorv1alpha1.AWSNodeManagerStatus{
					NodeReplenisher: &operatorv1alpha1.AWSNodeReplenisherRef{
						Namespace: "",
						Name:      "",
					},
					NodeRefresher: &operatorv1alpha1.AWSNodeRefresherRef{
						Namespace: "",
						Name:      "",
					},
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "ip-172-32-16-0",
							InstanceID:           "",
							AvailabilityZone:     "",
							InstanceType:         "",
							AutoScalingGroupName: "",
							CreationTimestamp: metav1.Time{
								Time: creationTimestamp,
							},
						},
					},
					NotJoinedAWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Time{},
					},
					Revision: 1,
					Phase:    operatorv1alpha1.AWSNodeManagerInit,
				},
			},
			expectedUpdated: true,
			expectedStatus: &operatorv1alpha1.AWSNodeManagerStatus{
				NodeReplenisher: &operatorv1alpha1.AWSNodeReplenisherRef{
					Namespace: "",
					Name:      "",
				},
				NodeRefresher: &operatorv1alpha1.AWSNodeRefresherRef{
					Namespace: "",
					Name:      "",
				},
				AWSNodes: []operatorv1alpha1.AWSNode{
					{
						Name:                 "ip-172-32-16-0",
						InstanceID:           "instanceId-1",
						AvailabilityZone:     "us-east-1a",
						InstanceType:         ec2.InstanceTypeT3Nano,
						AutoScalingGroupName: "asg-1",
						CreationTimestamp: metav1.Time{
							Time: creationTimestamp,
						},
					},
				},
				NotJoinedAWSNodes: nil,
				LastASGModifiedTime: &metav1.Time{
					Time: time.Time{},
				},
				Revision: 2,
				Phase:    operatorv1alpha1.AWSNodeManagerSynced,
			},
		},
		{
			title:                "Don't need update",
			describeInstanceResp: nil,
			getAWSNodesResp:      nil,
			asgDescribeResp: &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []*autoscaling.Group{
					&autoscaling.Group{
						AutoScalingGroupName: aws.String("asg-1"),
						AvailabilityZones: []*string{
							aws.String("us-east-1a"),
						},
						DesiredCapacity: aws.Int64(1),
						Instances: []*autoscaling.Instance{
							&autoscaling.Instance{
								AvailabilityZone: aws.String("us-east-1a"),
								InstanceId:       aws.String("instanceId-1"),
								InstanceType:     aws.String(ec2.InstanceTypeT3Nano),
							},
						},
						MaxSize: aws.Int64(3),
						MinSize: aws.Int64(0),
					},
				},
			},
			awsNodeManager: &operatorv1alpha1.AWSNodeManager{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-manager",
				},
				Spec: operatorv1alpha1.AWSNodeManagerSpec{
					Region: "us-east-1",
					AutoScalingGroups: []operatorv1alpha1.AutoScalingGroup{
						{
							Name: "asg-1",
						},
					},
					Desired:                  2,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     "worker",
					EnableReplenish:          true,
					RefreshSchedule:          "",
					SurplusNodes:             1,
				},
				Status: operatorv1alpha1.AWSNodeManagerStatus{
					NodeReplenisher: &operatorv1alpha1.AWSNodeReplenisherRef{
						Namespace: "",
						Name:      "",
					},
					NodeRefresher: &operatorv1alpha1.AWSNodeRefresherRef{
						Namespace: "",
						Name:      "",
					},
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "ip-172-32-16-0",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         ec2.InstanceTypeT3Nano,
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: creationTimestamp,
							},
						},
					},
					NotJoinedAWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Time{},
					},
					Revision: 1,
					Phase:    operatorv1alpha1.AWSNodeManagerSynced,
				},
			},
			currentAWSNodeManager: &operatorv1alpha1.AWSNodeManager{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-manager",
				},
				Spec: operatorv1alpha1.AWSNodeManagerSpec{
					Region: "us-east-1",
					AutoScalingGroups: []operatorv1alpha1.AutoScalingGroup{
						{
							Name: "asg-1",
						},
					},
					Desired:                  2,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     "worker",
					EnableReplenish:          true,
					RefreshSchedule:          "",
					SurplusNodes:             1,
				},
				Status: operatorv1alpha1.AWSNodeManagerStatus{
					NodeReplenisher: &operatorv1alpha1.AWSNodeReplenisherRef{
						Namespace: "",
						Name:      "",
					},
					NodeRefresher: &operatorv1alpha1.AWSNodeRefresherRef{
						Namespace: "",
						Name:      "",
					},
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "ip-172-32-16-0",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         ec2.InstanceTypeT3Nano,
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: creationTimestamp,
							},
						},
					},
					NotJoinedAWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Time{},
					},
					Revision: 1,
					Phase:    operatorv1alpha1.AWSNodeManagerSynced,
				},
			},
			expectedUpdated: false,
			expectedStatus: &operatorv1alpha1.AWSNodeManagerStatus{
				NodeReplenisher: &operatorv1alpha1.AWSNodeReplenisherRef{
					Namespace: "",
					Name:      "",
				},
				NodeRefresher: &operatorv1alpha1.AWSNodeRefresherRef{
					Namespace: "",
					Name:      "",
				},
				AWSNodes: []operatorv1alpha1.AWSNode{
					{
						Name:                 "ip-172-32-16-0",
						InstanceID:           "instanceId-1",
						AvailabilityZone:     "us-east-1a",
						InstanceType:         ec2.InstanceTypeT3Nano,
						AutoScalingGroupName: "asg-1",
						CreationTimestamp: metav1.Time{
							Time: creationTimestamp,
						},
					},
				},
				NotJoinedAWSNodes: nil,
				LastASGModifiedTime: &metav1.Time{
					Time: time.Time{},
				},
				Revision: 1,
				Phase:    operatorv1alpha1.AWSNodeManagerSynced,
			},
		},
	}

	for _, c := range cases {
		log.Printf("Running CASE: %s", c.title)
		ctx := context.Background()

		mockedASG := &mockedASGAPI{
			DescribeAutoScalingGroupsOutput: c.asgDescribeResp,
		}
		mockedEC2 := &mockedEC2API{
			describeFunc: func(in *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
				if in.InstanceIds != nil {
					return c.getAWSNodesResp, nil
				}
				return c.describeInstanceResp, nil
			},
		}
		mockedAWS := &cloudaws.AWS{
			EC2:         mockedEC2,
			Autoscaling: mockedASG,
		}
		cli := &mockedClient{
			getFunc: func(obj client.Object) error {
				obj.(*operatorv1alpha1.AWSNodeManager).ObjectMeta = c.currentAWSNodeManager.ObjectMeta
				obj.(*operatorv1alpha1.AWSNodeManager).Spec = c.currentAWSNodeManager.Spec
				obj.(*operatorv1alpha1.AWSNodeManager).Status = c.currentAWSNodeManager.Status
				return nil
			},
		}
		r := &AWSNodeManagerReconciler{
			cloud:    mockedAWS,
			Client:   cli,
			Recorder: &mockedRecorder{},
		}

		updated, err := r.syncAWSNodes(ctx, c.awsNodeManager)
		if err != nil {
			t.Errorf("CASE: %s : error has occur: %v", c.title, err)
			continue
		}
		if updated != c.expectedUpdated {
			t.Errorf("CASE: %s : updated is not matched, expected %t, returned %t", c.title, c.expectedUpdated, updated)
			continue
		}
		if !updated {
			continue
		}
		updatedObj := cli.updatedObj.(*operatorv1alpha1.AWSNodeManager)
		if !reflect.DeepEqual(updatedObj.Status, *c.expectedStatus) {
			t.Errorf("CASE: %s : status is not matched,\n expected %+v,\n returned %+v", c.title, *c.expectedStatus, updatedObj.Status)
		}

	}
}
