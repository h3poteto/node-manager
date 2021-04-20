package awsnoderefresher

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	cloudaws "github.com/h3poteto/node-manager/pkg/cloud/aws"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRefreshDecrease(t *testing.T) {
	cases := []struct {
		title        string
		refresher    *operatorv1alpha1.AWSNodeRefresher
		describeResp *autoscaling.DescribeAutoScalingGroupsOutput
		updateResp   *autoscaling.UpdateAutoScalingGroupOutput
	}{
		{
			title: "Phase is not matched",
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
								Time: time.Now().Add(-30 * time.Minute),
							},
						},
						{
							Name:                 "worker-2",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-30 * time.Minute),
							},
						},
					},
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-30 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherUpdateReplacing,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(-40 * time.Minute),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-40 * time.Minute),
					},
					ReplaceTargetNode: &operatorv1alpha1.AWSNode{
						Name:                 "",
						InstanceID:           "",
						AvailabilityZone:     "",
						InstanceType:         "",
						AutoScalingGroupName: "",
						CreationTimestamp: metav1.Time{
							Time: time.Time{},
						},
					},
				},
			},
		},
		{
			title: "Instances are not enough",
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
								Time: time.Now().Add(-30 * time.Minute),
							},
						},
						{
							Name:                 "worker-2",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-30 * time.Minute),
							},
						},
					},
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-30 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherUpdateAWSWaiting,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(-40 * time.Minute),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-40 * time.Minute),
					},
					ReplaceTargetNode: &operatorv1alpha1.AWSNode{
						Name:                 "",
						InstanceID:           "",
						AvailabilityZone:     "",
						InstanceType:         "",
						AutoScalingGroupName: "",
						CreationTimestamp: metav1.Time{
							Time: time.Time{},
						},
					},
				},
			},
		},
		{
			title: "Instances are enough",
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
								Time: time.Now().Add(-30 * time.Minute),
							},
						},
						{
							Name:                 "worker-2",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-30 * time.Minute),
							},
						},
						{
							Name:                 "worker-3",
							InstanceID:           "instanceId-3",
							AvailabilityZone:     "us-east-1d",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-30 * time.Minute),
							},
						},
					},
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-30 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherUpdateAWSWaiting,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(-40 * time.Minute),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-40 * time.Minute),
					},
					ReplaceTargetNode: &operatorv1alpha1.AWSNode{
						Name:                 "",
						InstanceID:           "",
						AvailabilityZone:     "",
						InstanceType:         "",
						AutoScalingGroupName: "",
						CreationTimestamp: metav1.Time{
							Time: time.Time{},
						},
					},
				},
			},
			describeResp: &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []*autoscaling.Group{
					&autoscaling.Group{
						AutoScalingGroupName: aws.String("asg-1"),
						AvailabilityZones: []*string{
							aws.String("us-east-1a"),
							aws.String("us-east-1c"),
							aws.String("us-east-1d"),
						},
						DesiredCapacity: aws.Int64(3),
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
								AvailabilityZone:        aws.String("us-east-1d"),
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
		},
		{
			title: "Instance is enough and no surplus",
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
					Desired:                  3,
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
								Time: time.Now().Add(-30 * time.Minute),
							},
						},
						{
							Name:                 "worker-2",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-30 * time.Minute),
							},
						},
						{
							Name:                 "worker-3",
							InstanceID:           "instanceId-3",
							AvailabilityZone:     "us-east-1d",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-30 * time.Minute),
							},
						},
					},
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-30 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherUpdateAWSWaiting,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(-40 * time.Minute),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-40 * time.Minute),
					},
					ReplaceTargetNode: &operatorv1alpha1.AWSNode{
						Name:                 "",
						InstanceID:           "",
						AvailabilityZone:     "",
						InstanceType:         "",
						AutoScalingGroupName: "",
						CreationTimestamp: metav1.Time{
							Time: time.Time{},
						},
					},
				},
			},
			describeResp: &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []*autoscaling.Group{
					&autoscaling.Group{
						AutoScalingGroupName: aws.String("asg-1"),
						AvailabilityZones: []*string{
							aws.String("us-east-1a"),
							aws.String("us-east-1c"),
							aws.String("us-east-1d"),
						},
						DesiredCapacity: aws.Int64(3),
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
								AvailabilityZone:        aws.String("us-east-1d"),
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
		},
	}

	for _, c := range cases {
		log.Printf("Running CASE: %s", c.title)
		ctx := context.Background()

		r := &AWSNodeRefresherReconciler{
			cloud: &cloudaws.AWS{
				Autoscaling: &mockedASGAPI{
					DescribeAutoScalingGroupsOutput: c.describeResp,
					UpdateAutoScalingGroupOutput:    c.updateResp,
				},
			},
			Client:   &mockedClient{},
			Recorder: &mockedRecorder{},
		}

		err := r.refreshDecrease(ctx, c.refresher)
		if err != nil {
			t.Errorf("CASE: %s : Failed to decrease: %v", c.title, err)
		}
	}
}

func TestRetryDecrease(t *testing.T) {
	cases := []struct {
		title        string
		refresher    *operatorv1alpha1.AWSNodeRefresher
		describeResp *autoscaling.DescribeAutoScalingGroupsOutput
		updateResp   *autoscaling.UpdateAutoScalingGroupOutput
		waiting      bool
		retried      bool
	}{
		{
			title: "During cooltime",
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
					SurplusNodes:             1,
				},
				Status: operatorv1alpha1.AWSNodeRefresherStatus{
					AWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-1 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherUpdateDecreasing,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(-30 * time.Minute),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-30 * time.Minute),
					},
					ReplaceTargetNode: &operatorv1alpha1.AWSNode{
						Name:                 "",
						InstanceID:           "",
						AvailabilityZone:     "",
						InstanceType:         "",
						AutoScalingGroupName: "",
						CreationTimestamp: metav1.Time{
							Time: time.Time{},
						},
					},
				},
			},
			waiting: true,
			retried: false,
		},
		{
			title: "Phase is not matched",
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
					SurplusNodes:             1,
				},
				Status: operatorv1alpha1.AWSNodeRefresherStatus{
					AWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-20 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherUpdateAWSWaiting,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(-30 * time.Minute),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-30 * time.Minute),
					},
					ReplaceTargetNode: &operatorv1alpha1.AWSNode{
						Name:                 "",
						InstanceID:           "",
						AvailabilityZone:     "",
						InstanceType:         "",
						AutoScalingGroupName: "",
						CreationTimestamp: metav1.Time{
							Time: time.Time{},
						},
					},
				},
			},
			waiting: false,
			retried: false,
		},
		{
			title: "Instances are not enough",
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
								Time: time.Now().Add(-20 * time.Minute),
							},
						},
						{
							Name:                 "worker-2",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-20 * time.Minute),
							},
						},
					},
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-15 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherUpdateDecreasing,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(-30 * time.Minute),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-30 * time.Minute),
					},
					ReplaceTargetNode: &operatorv1alpha1.AWSNode{
						Name:                 "",
						InstanceID:           "",
						AvailabilityZone:     "",
						InstanceType:         "",
						AutoScalingGroupName: "",
						CreationTimestamp: metav1.Time{
							Time: time.Time{},
						},
					},
				},
			},
			waiting: false,
			retried: false,
		},
		{
			title: "Instances are not enough and no surplus",
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
								Time: time.Now().Add(-20 * time.Minute),
							},
						},
						{
							Name:                 "worker-2",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-20 * time.Minute),
							},
						},
					},
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-15 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherUpdateDecreasing,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(-30 * time.Minute),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-30 * time.Minute),
					},
					ReplaceTargetNode: &operatorv1alpha1.AWSNode{
						Name:                 "",
						InstanceID:           "",
						AvailabilityZone:     "",
						InstanceType:         "",
						AutoScalingGroupName: "",
						CreationTimestamp: metav1.Time{
							Time: time.Time{},
						},
					},
				},
			},
			describeResp: &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []*autoscaling.Group{
					&autoscaling.Group{
						AutoScalingGroupName: aws.String("asg-1"),
						AvailabilityZones: []*string{
							aws.String("us-east-1a"),
							aws.String("us-east-1c"),
							aws.String("us-east-1d"),
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
			title: "Finished cooltime and enough instances",
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
								Time: time.Now().Add(-20 * time.Minute),
							},
						},
						{
							Name:                 "worker-2",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-20 * time.Minute),
							},
						},
						{
							Name:                 "worker-3",
							InstanceID:           "instanceId-3",
							AvailabilityZone:     "us-east-1d",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-20 * time.Minute),
							},
						},
					},
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-15 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherUpdateDecreasing,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(-30 * time.Minute),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-30 * time.Minute),
					},
					ReplaceTargetNode: &operatorv1alpha1.AWSNode{
						Name:                 "",
						InstanceID:           "",
						AvailabilityZone:     "",
						InstanceType:         "",
						AutoScalingGroupName: "",
						CreationTimestamp: metav1.Time{
							Time: time.Time{},
						},
					},
				},
			},
			describeResp: &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []*autoscaling.Group{
					&autoscaling.Group{
						AutoScalingGroupName: aws.String("asg-1"),
						AvailabilityZones: []*string{
							aws.String("us-east-1a"),
							aws.String("us-east-1c"),
							aws.String("us-east-1d"),
						},
						DesiredCapacity: aws.Int64(3),
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
								AvailabilityZone:        aws.String("us-east-1d"),
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
			retried:    true,
		},
	}

	for _, c := range cases {
		log.Printf("Running CASE: %s", c.title)
		ctx := context.Background()

		r := &AWSNodeRefresherReconciler{
			cloud: &cloudaws.AWS{
				Autoscaling: &mockedASGAPI{
					DescribeAutoScalingGroupsOutput: c.describeResp,
					UpdateAutoScalingGroupOutput:    c.updateResp,
				},
			},
			Client:   &mockedClient{},
			Recorder: &mockedRecorder{},
		}

		waiting, retried, err := r.retryDecrease(ctx, c.refresher)
		if err != nil {
			t.Errorf("CASE: %s : Failed to decrease: %v", c.title, err)
		}
		if waiting != c.waiting {
			t.Errorf("CASE: %s : Waiting is not matched, expected %t, returned %t", c.title, c.waiting, waiting)
		}
		if retried != c.retried {
			t.Errorf("CASE: %s : Retried is not matched, expected %t, returned %t", c.title, c.retried, retried)
		}
	}
}
