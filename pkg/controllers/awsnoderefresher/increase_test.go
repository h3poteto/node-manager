package awsnoderefresher

import (
	"context"
	"log"
	"testing"
	"time"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	cloudaws "github.com/h3poteto/node-manager/pkg/cloud/aws"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type mockedASGAPI struct {
	autoscalingiface.AutoScalingAPI
	DescribeAutoScalingGroupsOutput *autoscaling.DescribeAutoScalingGroupsOutput
	UpdateAutoScalingGroupOutput    *autoscaling.UpdateAutoScalingGroupOutput
}

func (m *mockedASGAPI) DescribeAutoScalingGroups(in *autoscaling.DescribeAutoScalingGroupsInput) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	return m.DescribeAutoScalingGroupsOutput, nil
}

func (m *mockedASGAPI) UpdateAutoScalingGroup(in *autoscaling.UpdateAutoScalingGroupInput) (*autoscaling.UpdateAutoScalingGroupOutput, error) {
	return m.UpdateAutoScalingGroupOutput, nil
}

func TestRetryIncrease(t *testing.T) {
	cases := []struct {
		title        string
		refresher    *operatorv1alpha1.AWSNodeRefresher
		describeResp *autoscaling.DescribeAutoScalingGroupsOutput
		updateResp   *autoscaling.UpdateAutoScalingGroupOutput
		retried      bool
		waiting      bool
	}{
		{
			title: "No surplus",
			refresher: &operatorv1alpha1.AWSNodeRefresher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-refresher",
				},
				Spec: operatorv1alpha1.AWSNodeRefresherSpec{
					Region: "us-east-1",
					AutoScalingGroups: []operatorv1alpha1.AutoScalingGroup{
						{
							Name: "asg-1",
						},
					},
					Desired:                  2,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
					Schedule:                 "* * * * *",
					SurplusNodes:             0,
				},
				Status: operatorv1alpha1.AWSNodeRefresherStatus{
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "worker-1",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-10 * time.Hour),
							},
						},
						{
							Name:                 "worker-2",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-10 * time.Hour),
							},
						},
					},
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-11 * time.Minute),
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
			describeResp: &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []*autoscaling.Group{
					&autoscaling.Group{
						AutoScalingGroupName: aws.String("asg-1"),
						AvailabilityZones: []*string{
							aws.String("us-east-1a"),
							aws.String("us-east-1c"),
						},
						DesiredCapacity: aws.Int64(2),
						Instances: []*autoscaling.Instance{
							&autoscaling.Instance{
								AvailabilityZone:        aws.String("us-east-1a"),
								InstanceId:              aws.String("instanceId-1"),
								InstanceType:            aws.String("t3.small"),
								LaunchConfigurationName: nil,
							},
							&autoscaling.Instance{
								AvailabilityZone:        aws.String("us-east-1c"),
								InstanceId:              aws.String("instanceId-2"),
								InstanceType:            aws.String("t3.small"),
								LaunchConfigurationName: nil,
							},
						},
						LaunchConfigurationName: nil,
						MaxSize:                 aws.Int64(3),
						MinSize:                 aws.Int64(0),
					},
				},
				NextToken: nil,
			},
			updateResp: &autoscaling.UpdateAutoScalingGroupOutput{},
			waiting:    false,
			retried:    false,
		},
		{
			title: "Node is enough",
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
				},
				Status: operatorv1alpha1.AWSNodeRefresherStatus{
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "worker-1",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-10 * time.Hour),
							},
						},
						{
							Name:                 "worker-2",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-10 * time.Hour),
							},
						},
						{
							Name:                 "worker-3",
							InstanceID:           "instanceId-3",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
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
			describeResp: &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []*autoscaling.Group{
					&autoscaling.Group{
						AutoScalingGroupName: aws.String("asg-1"),
						AvailabilityZones: []*string{
							aws.String("us-east-1a"),
							aws.String("us-east-1c"),
						},
						DesiredCapacity: aws.Int64(2),
						Instances: []*autoscaling.Instance{
							&autoscaling.Instance{
								AvailabilityZone:        aws.String("us-east-1a"),
								InstanceId:              aws.String("instanceId-1"),
								InstanceType:            aws.String("t3.small"),
								LaunchConfigurationName: nil,
							},
							&autoscaling.Instance{
								AvailabilityZone:        aws.String("us-east-1c"),
								InstanceId:              aws.String("instanceId-2"),
								InstanceType:            aws.String("t3.small"),
								LaunchConfigurationName: nil,
							},
							&autoscaling.Instance{
								AvailabilityZone:        aws.String("us-east-1c"),
								InstanceId:              aws.String("instanceId-3"),
								InstanceType:            aws.String("t3.small"),
								LaunchConfigurationName: nil,
							},
						},
						LaunchConfigurationName: nil,
						MaxSize:                 aws.Int64(3),
						MinSize:                 aws.Int64(0),
					},
				},
				NextToken: nil,
			},
			updateResp: &autoscaling.UpdateAutoScalingGroupOutput{},
			waiting:    false,
			retried:    false,
		},
		{
			title: "Node is not enough and during cooltime",
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
						Time: time.Now().Add(-5 * time.Minute),
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
			waiting: true,
			retried: false,
		},
		{
			title: "Node is not enough and finished cooltime",
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
				},
				Status: operatorv1alpha1.AWSNodeRefresherStatus{
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "worker-1",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-10 * time.Hour),
							},
						},
						{
							Name:                 "worker-2",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-10 * time.Hour),
							},
						},
					},
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-11 * time.Minute),
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
			describeResp: &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []*autoscaling.Group{
					&autoscaling.Group{
						AutoScalingGroupName: aws.String("asg-1"),
						AvailabilityZones: []*string{
							aws.String("us-east-1a"),
							aws.String("us-east-1c"),
						},
						DesiredCapacity: aws.Int64(2),
						Instances: []*autoscaling.Instance{
							&autoscaling.Instance{
								AvailabilityZone:        aws.String("us-east-1a"),
								InstanceId:              aws.String("instanceId-1"),
								InstanceType:            aws.String("t3.small"),
								LaunchConfigurationName: nil,
							},
							&autoscaling.Instance{
								AvailabilityZone:        aws.String("us-east-1c"),
								InstanceId:              aws.String("instanceId-2"),
								InstanceType:            aws.String("t3.small"),
								LaunchConfigurationName: nil,
							},
						},
						LaunchConfigurationName: nil,
						MaxSize:                 aws.Int64(3),
						MinSize:                 aws.Int64(0),
					},
				},
				NextToken: nil,
			},
			updateResp: &autoscaling.UpdateAutoScalingGroupOutput{},
			waiting:    false,
			retried:    true,
		},
	}

	for _, c := range cases {
		log.Printf("Running CASE: %s", c.title)
		ctx := context.Background()

		mockedASG := &mockedASGAPI{
			DescribeAutoScalingGroupsOutput: c.describeResp,
			UpdateAutoScalingGroupOutput:    c.updateResp,
		}
		mockedAWS := &cloudaws.AWS{
			Autoscaling: mockedASG,
		}
		r := &AWSNodeRefresherReconciler{
			cloud:    mockedAWS,
			Client:   &mockedClient{},
			Recorder: &mockedRecorder{},
		}

		waiting, increased, err := r.retryIncrease(ctx, c.refresher)
		if err != nil {
			t.Errorf("CASE: %s : %v", c.title, err)
			continue
		}
		if waiting != c.waiting {
			t.Errorf("CASE: %s : waiting is not matched, expected %t, but returned %t", c.title, c.waiting, waiting)
		}
		if increased != c.retried {
			t.Errorf("CASE: %s : increased is not matched, expected %t, but returned %t", c.title, c.retried, increased)
		}
	}
}

func TestWaitingIncrease(t *testing.T) {
	cases := []struct {
		title     string
		refresher *operatorv1alpha1.AWSNodeRefresher
		expected  bool
	}{
		{
			title: "Node is not enough and during cooltime",
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
						Time: time.Now().Add(-5 * time.Minute),
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
			expected: true,
		},
		{
			title: "Node is not enough and finished cooltime",
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
						Time: time.Now().Add(-11 * time.Minute),
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
			expected: false,
		},
	}

	for _, c := range cases {
		log.Printf("Running CASE: %s", c.title)
		ctx := context.Background()
		now := metav1.Now()
		result := waitingIncrease(ctx, c.refresher, &now)
		if result != c.expected {
			t.Errorf("CASE: %s : result is not matched, expected %t, but returned %t", c.title, c.expected, result)
		}
	}
}

func TestShouldRetryIncrease(t *testing.T) {
	cases := []struct {
		title     string
		refresher *operatorv1alpha1.AWSNodeRefresher
		expected  bool
	}{
		{
			title: "Node is enough",
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
			expected: false,
		},
		{
			title: "Node is not enough and finished cooltime",
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
						Time: time.Now().Add(-11 * time.Minute),
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
			expected: true,
		},
	}

	for _, c := range cases {
		log.Printf("Running CASE: %s", c.title)
		ctx := context.Background()
		now := metav1.Now()
		result := shouldRetryIncrease(ctx, c.refresher, &now)
		if result != c.expected {
			t.Errorf("CASE: %s : result is not matched, expected %t, but returned %t", c.title, c.expected, result)
		}
	}
}
