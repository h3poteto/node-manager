package awsnoderefresher

import (
	"context"
	"log"
	"testing"
	"time"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	cloudaws "github.com/h3poteto/node-manager/pkg/cloud/aws"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRefreshIncrease(t *testing.T) {
	cases := []struct {
		title          string
		refresher      *operatorv1alpha1.AWSNodeRefresher
		awsNodeManager *operatorv1alpha1.AWSNodeManager
		describeResp   *autoscaling.DescribeAutoScalingGroupsOutput
		updateResp     *autoscaling.UpdateAutoScalingGroupOutput
		expectedPhase  operatorv1alpha1.AWSNodeRefresherPhase
	}{
		{
			title: "Cluster is replenishing",
			refresher: &operatorv1alpha1.AWSNodeRefresher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-refresher",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         operatorv1alpha1.SchemeBuilder.GroupVersion.String(),
							Kind:               "AWSNodeManager",
							Name:               "test-awsnodemanager",
							UID:                "",
							Controller:         aws.Bool(true),
							BlockOwnerDeletion: aws.Bool(true),
						},
					},
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
						Time: time.Time{},
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherScheduled,
					NextUpdateTime: &metav1.Time{
						Time: time.Time{},
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Time{},
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
			awsNodeManager: &operatorv1alpha1.AWSNodeManager{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-awsnodemanager",
				},
				Spec: operatorv1alpha1.AWSNodeManagerSpec{
					Region:                   "",
					AutoScalingGroups:        nil,
					Desired:                  0,
					ASGModifyCoolTimeSeconds: 0,
					Role:                     "",
					EnableReplenish:          false,
					RefreshSchedule:          "",
					SurplusNodes:             0,
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
					AWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Time{},
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeManagerReplenishing,
				},
			},
			expectedPhase: operatorv1alpha1.AWSNodeRefresherScheduled,
		},
		{
			title: "Before next update time",
			refresher: &operatorv1alpha1.AWSNodeRefresher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-refresher",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         operatorv1alpha1.SchemeBuilder.GroupVersion.String(),
							Kind:               "AWSNodeManager",
							Name:               "test-awsnodemanager",
							UID:                "",
							Controller:         aws.Bool(true),
							BlockOwnerDeletion: aws.Bool(true),
						},
					},
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
						Time: time.Now().Add(-1 * time.Hour),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherScheduled,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(10 * time.Minute),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-1 * time.Hour),
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
			awsNodeManager: &operatorv1alpha1.AWSNodeManager{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-awsnodemanager",
				},
				Spec: operatorv1alpha1.AWSNodeManagerSpec{
					Region:                   "",
					AutoScalingGroups:        nil,
					Desired:                  0,
					ASGModifyCoolTimeSeconds: 0,
					Role:                     "",
					EnableReplenish:          false,
					RefreshSchedule:          "",
					SurplusNodes:             0,
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
					AWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Time{},
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeManagerSynced,
				},
			},
			expectedPhase: operatorv1alpha1.AWSNodeRefresherScheduled,
		},
		{
			title: "Before next update time with surplus is 0",
			refresher: &operatorv1alpha1.AWSNodeRefresher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-refresher",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         operatorv1alpha1.SchemeBuilder.GroupVersion.String(),
							Kind:               "AWSNodeManager",
							Name:               "test-awsnodemanager",
							UID:                "",
							Controller:         aws.Bool(true),
							BlockOwnerDeletion: aws.Bool(true),
						},
					},
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
						Time: time.Now().Add(-1 * time.Hour),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherScheduled,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(10 * time.Minute),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-1 * time.Hour),
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
			awsNodeManager: &operatorv1alpha1.AWSNodeManager{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-awsnodemanager",
				},
				Spec: operatorv1alpha1.AWSNodeManagerSpec{
					Region:                   "",
					AutoScalingGroups:        nil,
					Desired:                  0,
					ASGModifyCoolTimeSeconds: 0,
					Role:                     "",
					EnableReplenish:          false,
					RefreshSchedule:          "",
					SurplusNodes:             0,
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
					AWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Time{},
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeManagerSynced,
				},
			},
			expectedPhase: operatorv1alpha1.AWSNodeRefresherScheduled,
		},
		{
			title: "After next update time",
			refresher: &operatorv1alpha1.AWSNodeRefresher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-refresher",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         operatorv1alpha1.SchemeBuilder.GroupVersion.String(),
							Kind:               "AWSNodeManager",
							Name:               "test-awsnodemanager",
							UID:                "",
							Controller:         aws.Bool(true),
							BlockOwnerDeletion: aws.Bool(true),
						},
					},
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
						Time: time.Now().Add(-1 * time.Hour),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherScheduled,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-1 * time.Hour),
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
			awsNodeManager: &operatorv1alpha1.AWSNodeManager{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-awsnodemanager",
				},
				Spec: operatorv1alpha1.AWSNodeManagerSpec{
					Region:                   "",
					AutoScalingGroups:        nil,
					Desired:                  0,
					ASGModifyCoolTimeSeconds: 0,
					Role:                     "",
					EnableReplenish:          false,
					RefreshSchedule:          "",
					SurplusNodes:             0,
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
					AWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Time{},
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeManagerSynced,
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
			updateResp:    &autoscaling.UpdateAutoScalingGroupOutput{},
			expectedPhase: operatorv1alpha1.AWSNodeRefresherUpdateIncreasing,
		},
		{
			title: "After next update time and surplus is 0",
			refresher: &operatorv1alpha1.AWSNodeRefresher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-refresher",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         operatorv1alpha1.SchemeBuilder.GroupVersion.String(),
							Kind:               "AWSNodeManager",
							Name:               "test-awsnodemanager",
							UID:                "",
							Controller:         aws.Bool(true),
							BlockOwnerDeletion: aws.Bool(true),
						},
					},
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
						Time: time.Now().Add(-1 * time.Hour),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherScheduled,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-1 * time.Hour),
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
			awsNodeManager: &operatorv1alpha1.AWSNodeManager{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-awsnodemanager",
				},
				Spec: operatorv1alpha1.AWSNodeManagerSpec{
					Region:                   "",
					AutoScalingGroups:        nil,
					Desired:                  0,
					ASGModifyCoolTimeSeconds: 0,
					Role:                     "",
					EnableReplenish:          false,
					RefreshSchedule:          "",
					SurplusNodes:             0,
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
					AWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Time{},
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeManagerSynced,
				},
			},
			expectedPhase: operatorv1alpha1.AWSNodeRefresherUpdateIncreasing,
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
		cli := &mockedClient{
			getFunc: func(obj client.Object) error {
				obj.(*operatorv1alpha1.AWSNodeManager).ObjectMeta = c.awsNodeManager.ObjectMeta
				obj.(*operatorv1alpha1.AWSNodeManager).Spec = c.awsNodeManager.Spec
				obj.(*operatorv1alpha1.AWSNodeManager).Status = c.awsNodeManager.Status
				return nil
			},
		}
		r := &AWSNodeRefresherReconciler{
			cloud:    mockedAWS,
			Client:   cli,
			Recorder: &mockedRecorder{},
		}

		err := r.refreshIncrease(ctx, c.refresher)
		if err != nil {
			t.Errorf("CASE : %s : Failed to increase: %v", c.title, err)
			continue
		}

		if c.refresher.Status.Phase != c.expectedPhase {
			t.Errorf("CASE : %s : Phase is not matched, expected: %s, returned: %s", c.title, c.expectedPhase, c.refresher.Status.Phase)
		}
	}
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
