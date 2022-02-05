package awsnoderefresher

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	cloudaws "github.com/h3poteto/node-manager/pkg/cloud/aws"
)

func TestRefreshReplace(t *testing.T) {
	cases := []struct {
		title             string
		refresher         *operatorv1alpha1.AWSNodeRefresher
		terminateResp     *ec2.TerminateInstancesOutput
		terminateTargetID *string
	}{
		{
			title: "Instance is not enough",
			refresher: &operatorv1alpha1.AWSNodeRefresher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-refresher",
				},
				Spec: operatorv1alpha1.AWSNodeRefresherSpec{
					Region:                   "us-east-1",
					AutoScalingGroups:        nil,
					Desired:                  2,
					ASGModifyCoolTimeSeconds: 300,
					Role:                     operatorv1alpha1.Worker,
					Schedule:                 "* * * * *",
					SurplusNodes:             1,
					DrainGracePeriodSeconds:  300,
				},
				Status: operatorv1alpha1.AWSNodeRefresherStatus{
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "worker-1",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "autoscaling-group",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-10 * time.Hour),
							},
						},
						{
							Name:                 "worker-2",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "autoscaling-group",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-10 * time.Hour),
							},
						},
					},
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherUpdateIncreasing,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(24 * time.Hour),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					ReplaceTargetNode: nil,
				},
			},
		},
		{
			title: "Instance is enough",
			refresher: &operatorv1alpha1.AWSNodeRefresher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-refresher",
				},
				Spec: operatorv1alpha1.AWSNodeRefresherSpec{
					Region:                   "us-east-1",
					AutoScalingGroups:        nil,
					Desired:                  2,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
					Schedule:                 "* * * * *",
					SurplusNodes:             1,
					DrainGracePeriodSeconds:  300,
				},
				Status: operatorv1alpha1.AWSNodeRefresherStatus{
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "worker-1",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "autoscaling-group",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-10 * time.Hour),
							},
						},
						{
							Name:                 "worker-2",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "autoscaling-group",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-9 * time.Hour),
							},
						},
						{
							Name:                 "worker-3",
							InstanceID:           "instanceId-3",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "autoscaling-group",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Minute),
							},
						},
					},
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherUpdateIncreasing,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(24 * time.Hour),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					ReplaceTargetNode: nil,
				},
			},
			terminateResp: &ec2.TerminateInstancesOutput{
				TerminatingInstances: []*ec2.InstanceStateChange{
					&ec2.InstanceStateChange{
						CurrentState: &ec2.InstanceState{
							Code: aws.Int64(48),
							Name: aws.String(ec2.InstanceStateNameTerminated),
						},
						InstanceId: nil,
						PreviousState: &ec2.InstanceState{
							Code: aws.Int64(16),
							Name: aws.String(ec2.InstanceStateNameRunning),
						},
					},
				},
			},
			terminateTargetID: aws.String("instanceId-1"),
		},
	}

	for _, c := range cases {
		log.Printf("Running CASE: %s", c.title)
		ctx := context.Background()

		mockedEC2 := &mockedEC2API{
			TerminateInstancesResp: c.terminateResp,
			terminateInstanceID:    c.terminateTargetID,
		}
		mockedAWS := &cloudaws.AWS{
			EC2: mockedEC2,
		}
		r := &AWSNodeRefresherReconciler{
			cloud:    mockedAWS,
			Client:   &mockedClient{},
			Recorder: &mockedRecorder{},
		}

		err := r.refreshReplace(ctx, c.refresher)
		if err != nil {
			t.Errorf("CASE: %s : %v", c.title, err)
			continue
		}
	}
}

func TestShouldReplace(t *testing.T) {
	cases := []struct {
		title     string
		refresher *operatorv1alpha1.AWSNodeRefresher
		expected  bool
	}{
		{
			title: "Instance is not enough",
			refresher: &operatorv1alpha1.AWSNodeRefresher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-refresher",
				},
				Spec: operatorv1alpha1.AWSNodeRefresherSpec{
					Region:                   "us-east-1",
					AutoScalingGroups:        nil,
					Desired:                  2,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
					Schedule:                 "* * * * *",
					SurplusNodes:             1,
					DrainGracePeriodSeconds:  300,
				},
				Status: operatorv1alpha1.AWSNodeRefresherStatus{
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "worker-1",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "autoscaling-group",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-10 * time.Hour),
							},
						},
						{
							Name:                 "worker-2",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "autoscaling-group",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-10 * time.Hour),
							},
						},
					},
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherDraining,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(24 * time.Hour),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					ReplaceTargetNode: nil,
				},
			},
			expected: false,
		},
		{
			title: "Instance is enough",
			refresher: &operatorv1alpha1.AWSNodeRefresher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-refresher",
				},
				Spec: operatorv1alpha1.AWSNodeRefresherSpec{
					Region:                   "us-east-1",
					AutoScalingGroups:        nil,
					Desired:                  2,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
					Schedule:                 "* * * * *",
					SurplusNodes:             1,
					DrainGracePeriodSeconds:  300,
				},
				Status: operatorv1alpha1.AWSNodeRefresherStatus{
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "worker-1",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "autoscaling-group",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-10 * time.Hour),
							},
						},
						{
							Name:                 "worker-2",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "autoscaling-group",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-10 * time.Hour),
							},
						},
						{
							Name:                 "worker-3",
							InstanceID:           "instanceId-3",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "autoscaling-group",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Minute),
							},
						},
					},
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherDraining,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(24 * time.Hour),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					ReplaceTargetNode: nil,
				},
			},
			expected: true,
		},
		{
			title: "Instance is enough with surplus 0",
			refresher: &operatorv1alpha1.AWSNodeRefresher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-refresher",
				},
				Spec: operatorv1alpha1.AWSNodeRefresherSpec{
					Region:                   "us-east-1",
					AutoScalingGroups:        nil,
					Desired:                  2,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
					Schedule:                 "* * * * *",
					SurplusNodes:             0,
					DrainGracePeriodSeconds:  300,
				},
				Status: operatorv1alpha1.AWSNodeRefresherStatus{
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "worker-1",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "autoscaling-group",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-10 * time.Hour),
							},
						},
						{
							Name:                 "worker-2",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "autoscaling-group",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-10 * time.Hour),
							},
						},
					},
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherDraining,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(24 * time.Hour),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					ReplaceTargetNode: nil,
				},
			},
			expected: true,
		},
		{
			title: "Instance is not enough with surplus 0",
			refresher: &operatorv1alpha1.AWSNodeRefresher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-refresher",
				},
				Spec: operatorv1alpha1.AWSNodeRefresherSpec{
					Region:                   "us-east-1",
					AutoScalingGroups:        nil,
					Desired:                  3,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
					Schedule:                 "* * * * *",
					SurplusNodes:             0,
					DrainGracePeriodSeconds:  300,
				},
				Status: operatorv1alpha1.AWSNodeRefresherStatus{
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "worker-1",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "autoscaling-group",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-10 * time.Hour),
							},
						},
						{
							Name:                 "worker-2",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "autoscaling-group",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-10 * time.Hour),
							},
						},
					},
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherDraining,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(24 * time.Hour),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					ReplaceTargetNode: nil,
				},
			},
			expected: false,
		},
	}

	for _, c := range cases {
		log.Printf("Running CASE: %s", c.title)
		ctx := context.Background()

		result := shouldReplace(ctx, c.refresher)
		if result != c.expected {
			t.Errorf("CASE: %s : result is not matched, expected %t, but returned %t", c.title, c.expected, result)
		}
	}
}

func TestRetryReplace(t *testing.T) {
	cases := []struct {
		title         string
		refresher     *operatorv1alpha1.AWSNodeRefresher
		describeResp  *ec2.DescribeInstancesOutput
		terminateResp *ec2.TerminateInstancesOutput
		waiting       bool
		replace       bool
	}{
		{
			title: "Status phase is not replacing",
			refresher: &operatorv1alpha1.AWSNodeRefresher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-refresher",
				},
				Spec: operatorv1alpha1.AWSNodeRefresherSpec{
					Region:                   "us-east-1",
					AutoScalingGroups:        nil,
					Desired:                  3,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
					Schedule:                 "* * * * *",
					SurplusNodes:             1,
					DrainGracePeriodSeconds:  300,
				},
				Status: operatorv1alpha1.AWSNodeRefresherStatus{
					AWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-5 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherUpdateAWSWaiting,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(24 * time.Hour),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					ReplaceTargetNode: &operatorv1alpha1.AWSNode{
						Name:                 "node-1",
						InstanceID:           "instanceId-1",
						AvailabilityZone:     "us-east-1",
						InstanceType:         ec2.InstanceTypeT3Small,
						AutoScalingGroupName: "autoscaling-group-name",
					},
				},
			},
			replace: false,
			waiting: false,
		},
		{
			title: "Instance is stopping",
			refresher: &operatorv1alpha1.AWSNodeRefresher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-refresher",
				},
				Spec: operatorv1alpha1.AWSNodeRefresherSpec{
					Region:                   "us-east-1",
					AutoScalingGroups:        nil,
					Desired:                  3,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
					Schedule:                 "* * * * *",
					SurplusNodes:             1,
					DrainGracePeriodSeconds:  300,
				},
				Status: operatorv1alpha1.AWSNodeRefresherStatus{
					AWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-5 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherUpdateReplacing,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(24 * time.Hour),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					ReplaceTargetNode: &operatorv1alpha1.AWSNode{
						Name:                 "node-1",
						InstanceID:           "instanceId-1",
						AvailabilityZone:     "us-east-1",
						InstanceType:         ec2.InstanceTypeT3Small,
						AutoScalingGroupName: "autoscaling-group-name",
					},
				},
			},
			describeResp: &ec2.DescribeInstancesOutput{
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
								State: &ec2.InstanceState{
									Code: aws.Int64(64),
									Name: aws.String(ec2.InstanceStateNameStopping),
								},
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
			},
			terminateResp: &ec2.TerminateInstancesOutput{
				TerminatingInstances: []*ec2.InstanceStateChange{
					&ec2.InstanceStateChange{
						CurrentState: &ec2.InstanceState{
							Code: aws.Int64(48),
							Name: aws.String(ec2.InstanceStateNameTerminated),
						},
						InstanceId: aws.String("instanceId-1"),
						PreviousState: &ec2.InstanceState{
							Code: aws.Int64(64),
							Name: aws.String(ec2.InstanceStateNameStopping),
						},
					},
				},
			},
			replace: true,
			waiting: false,
		},
		{
			title: "Instance is terminated",
			refresher: &operatorv1alpha1.AWSNodeRefresher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-refresher",
				},
				Spec: operatorv1alpha1.AWSNodeRefresherSpec{
					Region:                   "us-east-1",
					AutoScalingGroups:        nil,
					Desired:                  3,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
					Schedule:                 "* * * * *",
					SurplusNodes:             1,
					DrainGracePeriodSeconds:  300,
				},
				Status: operatorv1alpha1.AWSNodeRefresherStatus{
					AWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-5 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherUpdateReplacing,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(24 * time.Hour),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					ReplaceTargetNode: &operatorv1alpha1.AWSNode{
						Name:                 "node-1",
						InstanceID:           "instanceId-1",
						AvailabilityZone:     "us-east-1",
						InstanceType:         ec2.InstanceTypeT3Small,
						AutoScalingGroupName: "autoscaling-group-name",
					},
				},
			},
			describeResp: &ec2.DescribeInstancesOutput{
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
								State: &ec2.InstanceState{
									Code: aws.Int64(48),
									Name: aws.String(ec2.InstanceStateNameTerminated),
								},
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
			},
			replace: false,
			waiting: false,
		},
		{
			title: "Before time wait",
			refresher: &operatorv1alpha1.AWSNodeRefresher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-refresher",
				},
				Spec: operatorv1alpha1.AWSNodeRefresherSpec{
					Region:                   "us-east-1",
					AutoScalingGroups:        nil,
					Desired:                  3,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
					Schedule:                 "* * * * *",
					SurplusNodes:             1,
					DrainGracePeriodSeconds:  300,
				},
				Status: operatorv1alpha1.AWSNodeRefresherStatus{
					AWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now(),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherUpdateReplacing,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(24 * time.Hour),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					ReplaceTargetNode: &operatorv1alpha1.AWSNode{
						Name:                 "",
						InstanceID:           "",
						AvailabilityZone:     "",
						InstanceType:         "",
						AutoScalingGroupName: "",
					},
				},
			},
			waiting: true,
			replace: false,
		},
	}

	for _, c := range cases {
		log.Printf("Running CASE: %s", c.title)
		ctx := context.Background()

		mockedEC2 := &mockedEC2API{
			DescribeInstancesResp:  c.describeResp,
			TerminateInstancesResp: c.terminateResp,
		}
		mockedAWS := &cloudaws.AWS{
			EC2: mockedEC2,
		}
		r := &AWSNodeRefresherReconciler{
			cloud:    mockedAWS,
			Client:   &mockedClient{},
			Recorder: &mockedRecorder{},
		}

		waiting, replace, err := r.retryReplace(ctx, c.refresher)
		if err != nil {
			t.Errorf("CASES: %s : %v", c.title, err)
			continue
		}
		if waiting != c.waiting {
			t.Errorf("CASE: %s : waiting is not matched, expected %t, but returned %t", c.title, c.waiting, waiting)
		}
		if replace != c.replace {
			t.Errorf("CASE: %s : replace is not matched, expected %t, but returned %t", c.title, c.replace, replace)
		}
	}
}

func TestShouldRetryReplace(t *testing.T) {
	cases := []struct {
		title     string
		refresher *operatorv1alpha1.AWSNodeRefresher
		resp      *ec2.DescribeInstancesOutput
		expected  bool
	}{
		{
			title: "Terminated instance",
			refresher: &operatorv1alpha1.AWSNodeRefresher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-refresher",
				},
				Spec: operatorv1alpha1.AWSNodeRefresherSpec{
					Region:                   "us-east-1",
					AutoScalingGroups:        nil,
					Desired:                  3,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
					Schedule:                 "* * * * *",
					SurplusNodes:             1,
					DrainGracePeriodSeconds:  300,
				},
				Status: operatorv1alpha1.AWSNodeRefresherStatus{
					AWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-5 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherUpdateReplacing,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(24 * time.Hour),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					ReplaceTargetNode: &operatorv1alpha1.AWSNode{
						Name:                 "node-1",
						InstanceID:           "instanceId-1",
						AvailabilityZone:     "us-east-1",
						InstanceType:         ec2.InstanceTypeT3Small,
						AutoScalingGroupName: "autoscaling-group-name",
					},
				},
			},
			resp: &ec2.DescribeInstancesOutput{
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
								State: &ec2.InstanceState{
									Code: aws.Int64(48),
									Name: aws.String(ec2.InstanceStateNameTerminated),
								},
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
			},
			expected: false,
		},
		{
			title: "Running instance",
			refresher: &operatorv1alpha1.AWSNodeRefresher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-refresher",
				},
				Spec: operatorv1alpha1.AWSNodeRefresherSpec{
					Region:                   "us-east-1",
					AutoScalingGroups:        nil,
					Desired:                  3,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
					Schedule:                 "* * * * *",
					SurplusNodes:             1,
					DrainGracePeriodSeconds:  300,
				},
				Status: operatorv1alpha1.AWSNodeRefresherStatus{
					AWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-5 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherUpdateReplacing,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(24 * time.Hour),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					ReplaceTargetNode: &operatorv1alpha1.AWSNode{
						Name:                 "node-1",
						InstanceID:           "instanceId-1",
						AvailabilityZone:     "us-east-1",
						InstanceType:         ec2.InstanceTypeT3Small,
						AutoScalingGroupName: "autoscaling-group-name",
					},
				},
			},
			resp: &ec2.DescribeInstancesOutput{
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
								State: &ec2.InstanceState{
									Code: aws.Int64(16),
									Name: aws.String(ec2.InstanceStateNameRunning),
								},
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
			},
			expected: true,
		},
	}

	for _, c := range cases {
		log.Printf("Running CASE: %s", c.title)
		ctx := context.Background()

		mockedEC2 := &mockedEC2API{
			DescribeInstancesResp: c.resp,
		}

		mockedAWS := &cloudaws.AWS{
			EC2: mockedEC2,
		}

		r := &AWSNodeRefresherReconciler{
			cloud:  mockedAWS,
			Client: &mockedClient{},
		}

		result := r.shouldRetryReplace(ctx, c.refresher)
		if result != c.expected {
			t.Errorf("CASE: %s : result is not matched, expected %t, but returned %t", c.title, c.expected, result)
		}
	}
}
