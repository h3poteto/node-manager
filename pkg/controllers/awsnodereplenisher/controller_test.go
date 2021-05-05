package awsnodereplenisher

import (
	"context"
	"errors"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	cloudaws "github.com/h3poteto/node-manager/pkg/cloud/aws"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilpointer "k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestSyncReplenisher(t *testing.T) {
	cases := []struct {
		title             string
		replenisher       *operatorv1alpha1.AWSNodeReplenisher
		ownerManager      *operatorv1alpha1.AWSNodeManager
		describeASGResp   *autoscaling.DescribeAutoScalingGroupsOutput
		expectedPhase     operatorv1alpha1.AWSNodeReplenisherPhase
		expectedUpdated   map[string]*autoscaling.UpdateAutoScalingGroupInput
		expectedTerminate []*string
	}{
		{
			title: "Nodes count is same as desired count",
			replenisher: &operatorv1alpha1.AWSNodeReplenisher{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "test-replenisher",
					OwnerReferences: nil,
				},
				Spec: operatorv1alpha1.AWSNodeReplenisherSpec{
					Region: "us-east-1",
					AutoScalingGroups: []operatorv1alpha1.AutoScalingGroup{
						{
							Name: "asg-1",
						},
						{
							Name: "asg-2",
						},
						{
							Name: "asg-3",
						},
					},
					Desired:                  3,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
				},
				Status: operatorv1alpha1.AWSNodeReplenisherStatus{
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "172-32-16-0",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
						{
							Name:                 "172-32-16-1",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-2",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
						{
							Name:                 "172-32-16-2",
							InstanceID:           "instanceId-3",
							AvailabilityZone:     "us-east-1d",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-3",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
					},
					NotJoinedAWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Time{},
					},
					Revision: 1,
					Phase:    operatorv1alpha1.AWSNodeReplenisherInit,
				},
			},
			expectedPhase: operatorv1alpha1.AWSNodeReplenisherSynced,
		},
		{
			title: "Phase is refreshing",
			replenisher: &operatorv1alpha1.AWSNodeReplenisher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-replenisher",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         operatorv1alpha1.SchemeBuilder.GroupVersion.String(),
							Kind:               "AWSNodeManager",
							Name:               "test-aws-nodemanager",
							UID:                "uid-1",
							Controller:         utilpointer.BoolPtr(true),
							BlockOwnerDeletion: utilpointer.BoolPtr(true),
						},
					},
				},
				Spec: operatorv1alpha1.AWSNodeReplenisherSpec{
					Region: "us-east-1",
					AutoScalingGroups: []operatorv1alpha1.AutoScalingGroup{
						{
							Name: "asg-1",
						},
						{
							Name: "asg-2",
						},
						{
							Name: "asg-3",
						},
					},
					Desired:                  4,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
				},
				Status: operatorv1alpha1.AWSNodeReplenisherStatus{
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "172-32-16-0",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
						{
							Name:                 "172-32-16-1",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-2",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
						{
							Name:                 "172-32-16-2",
							InstanceID:           "instanceId-3",
							AvailabilityZone:     "us-east-1d",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-3",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
					},
					NotJoinedAWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Time{},
					},
					Revision: 1,
					Phase:    operatorv1alpha1.AWSNodeReplenisherSynced,
				},
			},
			ownerManager: &operatorv1alpha1.AWSNodeManager{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-aws-nodemanager",
				},
				Spec: operatorv1alpha1.AWSNodeManagerSpec{
					Region: "us-east-1",
					AutoScalingGroups: []operatorv1alpha1.AutoScalingGroup{
						{
							Name: "asg-1",
						},
						{
							Name: "asg-2",
						},
						{
							Name: "asg-3",
						},
					},
					Desired:                  4,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
					EnableReplenish:          true,
					RefreshSchedule:          "",
					SurplusNodes:             1,
				},
				Status: operatorv1alpha1.AWSNodeManagerStatus{
					NodeReplenisher: &operatorv1alpha1.AWSNodeReplenisherRef{
						Namespace: "",
						Name:      "test-replenisher",
					},
					NodeRefresher: &operatorv1alpha1.AWSNodeRefresherRef{
						Namespace: "",
						Name:      "",
					},
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "172-32-16-0",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
						{
							Name:                 "172-32-16-1",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-2",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
						{
							Name:                 "172-32-16-2",
							InstanceID:           "instanceId-3",
							AvailabilityZone:     "us-east-1d",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-3",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
					},
					NotJoinedAWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Time{},
					},
					Revision: 1,
					Phase:    operatorv1alpha1.AWSNodeManagerRefreshing,
				},
			},
			expectedPhase: operatorv1alpha1.AWSNodeReplenisherSynced,
		},
		{
			title: "Phase is waiting",
			replenisher: &operatorv1alpha1.AWSNodeReplenisher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-replenisher",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         operatorv1alpha1.SchemeBuilder.GroupVersion.String(),
							Kind:               "AWSNodeManager",
							Name:               "test-aws-nodemanager",
							UID:                "uid-1",
							Controller:         utilpointer.BoolPtr(true),
							BlockOwnerDeletion: utilpointer.BoolPtr(true),
						},
					},
				},
				Spec: operatorv1alpha1.AWSNodeReplenisherSpec{
					Region: "us-east-1",
					AutoScalingGroups: []operatorv1alpha1.AutoScalingGroup{
						{
							Name: "asg-1",
						},
						{
							Name: "asg-2",
						},
						{
							Name: "asg-3",
						},
					},
					Desired:                  4,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
				},
				Status: operatorv1alpha1.AWSNodeReplenisherStatus{
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "172-32-16-0",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
						{
							Name:                 "172-32-16-1",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-2",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
						{
							Name:                 "172-32-16-2",
							InstanceID:           "instanceId-3",
							AvailabilityZone:     "us-east-1d",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-3",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
					},
					NotJoinedAWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-5 * time.Minute),
					},
					Revision: 1,
					Phase:    operatorv1alpha1.AWSNodeReplenisherAWSUpdating,
				},
			},
			ownerManager: &operatorv1alpha1.AWSNodeManager{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-aws-nodemanager",
				},
				Spec: operatorv1alpha1.AWSNodeManagerSpec{
					Region: "us-east-1",
					AutoScalingGroups: []operatorv1alpha1.AutoScalingGroup{
						{
							Name: "asg-1",
						},
						{
							Name: "asg-2",
						},
						{
							Name: "asg-3",
						},
					},
					Desired:                  4,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
					EnableReplenish:          true,
					RefreshSchedule:          "",
					SurplusNodes:             1,
				},
				Status: operatorv1alpha1.AWSNodeManagerStatus{
					NodeReplenisher: &operatorv1alpha1.AWSNodeReplenisherRef{
						Namespace: "",
						Name:      "test-replenisher",
					},
					NodeRefresher: &operatorv1alpha1.AWSNodeRefresherRef{
						Namespace: "",
						Name:      "",
					},
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "172-32-16-0",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
						{
							Name:                 "172-32-16-1",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-2",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
						{
							Name:                 "172-32-16-2",
							InstanceID:           "instanceId-3",
							AvailabilityZone:     "us-east-1d",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-3",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
					},
					NotJoinedAWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-5 * time.Minute),
					},
					Revision: 1,
					Phase:    operatorv1alpha1.AWSNodeManagerReplenishing,
				},
			},
			expectedPhase: operatorv1alpha1.AWSNodeReplenisherAWSUpdating,
		},
		{
			title: "AWSNodes are added",
			replenisher: &operatorv1alpha1.AWSNodeReplenisher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-replenisher",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         operatorv1alpha1.SchemeBuilder.GroupVersion.String(),
							Kind:               "AWSNodeManager",
							Name:               "test-aws-nodemanager",
							UID:                "uid-1",
							Controller:         utilpointer.BoolPtr(true),
							BlockOwnerDeletion: utilpointer.BoolPtr(true),
						},
					},
				},
				Spec: operatorv1alpha1.AWSNodeReplenisherSpec{
					Region: "us-east-1",
					AutoScalingGroups: []operatorv1alpha1.AutoScalingGroup{
						{
							Name: "asg-1",
						},
						{
							Name: "asg-2",
						},
						{
							Name: "asg-3",
						},
					},
					Desired:                  4,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
				},
				Status: operatorv1alpha1.AWSNodeReplenisherStatus{
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "172-32-16-0",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
						{
							Name:                 "172-32-16-1",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-2",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
						{
							Name:                 "172-32-16-2",
							InstanceID:           "instanceId-3",
							AvailabilityZone:     "us-east-1d",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-3",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
					},
					NotJoinedAWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-15 * time.Minute),
					},
					Revision: 1,
					Phase:    operatorv1alpha1.AWSNodeReplenisherAWSUpdating,
				},
			},
			ownerManager: &operatorv1alpha1.AWSNodeManager{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-aws-nodemanager",
				},
				Spec: operatorv1alpha1.AWSNodeManagerSpec{
					Region: "us-east-1",
					AutoScalingGroups: []operatorv1alpha1.AutoScalingGroup{
						{
							Name: "asg-1",
						},
						{
							Name: "asg-2",
						},
						{
							Name: "asg-3",
						},
					},
					Desired:                  4,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
					EnableReplenish:          true,
					RefreshSchedule:          "",
					SurplusNodes:             1,
				},
				Status: operatorv1alpha1.AWSNodeManagerStatus{
					NodeReplenisher: &operatorv1alpha1.AWSNodeReplenisherRef{
						Namespace: "",
						Name:      "test-replenisher",
					},
					NodeRefresher: &operatorv1alpha1.AWSNodeRefresherRef{
						Namespace: "",
						Name:      "",
					},
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "172-32-16-0",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
						{
							Name:                 "172-32-16-1",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-2",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
						{
							Name:                 "172-32-16-2",
							InstanceID:           "instanceId-3",
							AvailabilityZone:     "us-east-1d",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-3",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
					},
					NotJoinedAWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-15 * time.Minute),
					},
					Revision: 1,
					Phase:    operatorv1alpha1.AWSNodeManagerReplenishing,
				},
			},
			expectedPhase: operatorv1alpha1.AWSNodeReplenisherAWSUpdating,
			describeASGResp: &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []*autoscaling.Group{
					&autoscaling.Group{
						AutoScalingGroupName: aws.String("asg-1"),
						AvailabilityZones:    []*string{aws.String("us-east-1a")},
						DesiredCapacity:      aws.Int64(1),
						Instances: []*autoscaling.Instance{
							&autoscaling.Instance{
								AvailabilityZone: aws.String("us-east-1a"),
								InstanceId:       aws.String("instanceId-1"),
								InstanceType:     aws.String("t3.small"),
							},
						},
						MaxSize: aws.Int64(3),
						MinSize: aws.Int64(1),
					},
					&autoscaling.Group{
						AutoScalingGroupName: aws.String("asg-2"),
						AvailabilityZones:    []*string{aws.String("us-east-1c")},
						DesiredCapacity:      aws.Int64(1),
						Instances: []*autoscaling.Instance{
							&autoscaling.Instance{
								AvailabilityZone: aws.String("us-east-1c"),
								InstanceId:       aws.String("instanceId-2"),
								InstanceType:     aws.String("t3.small"),
							},
						},
						MaxSize: aws.Int64(2),
						MinSize: aws.Int64(1),
					},
					&autoscaling.Group{
						AutoScalingGroupName: aws.String("asg-3"),
						AvailabilityZones:    []*string{aws.String("us-east-1d")},
						DesiredCapacity:      aws.Int64(1),
						Instances: []*autoscaling.Instance{
							&autoscaling.Instance{
								AvailabilityZone: aws.String("us-east-1d"),
								InstanceId:       aws.String("instanceId-3"),
								InstanceType:     aws.String("t3.small"),
							},
						},
						MaxSize: aws.Int64(2),
						MinSize: aws.Int64(1),
					},
				},
				NextToken: nil,
			},
			expectedUpdated: map[string]*autoscaling.UpdateAutoScalingGroupInput{
				"asg-1": &autoscaling.UpdateAutoScalingGroupInput{
					AutoScalingGroupName: aws.String("asg-1"),
					DesiredCapacity:      aws.Int64(2),
				},
				"asg-2": &autoscaling.UpdateAutoScalingGroupInput{
					AutoScalingGroupName: aws.String("asg-2"),
					DesiredCapacity:      aws.Int64(1),
				},
				"asg-3": &autoscaling.UpdateAutoScalingGroupInput{
					AutoScalingGroupName: aws.String("asg-3"),
					DesiredCapacity:      aws.Int64(1),
				},
			},
		},
		{
			title: "Not joined instances will be deleted",
			replenisher: &operatorv1alpha1.AWSNodeReplenisher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-replenisher",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         operatorv1alpha1.SchemeBuilder.GroupVersion.String(),
							Kind:               "AWSNodeManager",
							Name:               "test-aws-nodemanager",
							UID:                "uid-1",
							Controller:         utilpointer.BoolPtr(true),
							BlockOwnerDeletion: utilpointer.BoolPtr(true),
						},
					},
				},
				Spec: operatorv1alpha1.AWSNodeReplenisherSpec{
					Region: "us-east-1",
					AutoScalingGroups: []operatorv1alpha1.AutoScalingGroup{
						{
							Name: "asg-1",
						},
						{
							Name: "asg-2",
						},
						{
							Name: "asg-3",
						},
					},
					Desired:                  3,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
				},
				Status: operatorv1alpha1.AWSNodeReplenisherStatus{
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "172-32-16-0",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
						{
							Name:                 "172-32-16-1",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-2",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
						{
							Name:                 "172-32-16-2",
							InstanceID:           "instanceId-3",
							AvailabilityZone:     "us-east-1d",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-3",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
					},
					NotJoinedAWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "172-32-16-4",
							InstanceID:           "instanceId-4",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-2 * time.Hour),
							},
						},
					},
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-1 * time.Hour),
					},
					Revision: 1,
					Phase:    operatorv1alpha1.AWSNodeReplenisherSynced,
				},
			},
			ownerManager: &operatorv1alpha1.AWSNodeManager{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-aws-nodemanager",
				},
				Spec: operatorv1alpha1.AWSNodeManagerSpec{
					Region: "us-east-1",
					AutoScalingGroups: []operatorv1alpha1.AutoScalingGroup{
						{
							Name: "asg-1",
						},
						{
							Name: "asg-2",
						},
						{
							Name: "asg-3",
						},
					},
					Desired:                  3,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
					EnableReplenish:          true,
					RefreshSchedule:          "",
					SurplusNodes:             1,
				},
				Status: operatorv1alpha1.AWSNodeManagerStatus{
					NodeReplenisher: &operatorv1alpha1.AWSNodeReplenisherRef{
						Namespace: "",
						Name:      "test-replenisher",
					},
					NodeRefresher: &operatorv1alpha1.AWSNodeRefresherRef{
						Namespace: "",
						Name:      "",
					},
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "172-32-16-0",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
						{
							Name:                 "172-32-16-1",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-2",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
						{
							Name:                 "172-32-16-2",
							InstanceID:           "instanceId-3",
							AvailabilityZone:     "us-east-1d",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-3",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
					},
					NotJoinedAWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "172-32-16-4",
							InstanceID:           "instanceId-4",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-2 * time.Hour),
							},
						},
					},
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-1 * time.Hour),
					},
					Revision: 1,
					Phase:    operatorv1alpha1.AWSNodeManagerReplenishing,
				},
			},
			expectedPhase: operatorv1alpha1.AWSNodeReplenisherAWSUpdating,
			describeASGResp: &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []*autoscaling.Group{
					&autoscaling.Group{
						AutoScalingGroupName: aws.String("asg-1"),
						AvailabilityZones:    []*string{aws.String("us-east-1a")},
						DesiredCapacity:      aws.Int64(1),
						Instances: []*autoscaling.Instance{
							&autoscaling.Instance{
								AvailabilityZone: aws.String("us-east-1a"),
								InstanceId:       aws.String("instanceId-1"),
								InstanceType:     aws.String("t3.small"),
							},
							&autoscaling.Instance{
								AvailabilityZone: aws.String("us-east-1a"),
								InstanceId:       aws.String("instanceId-4"),
								InstanceType:     aws.String("t3.small"),
							},
						},
						MaxSize: aws.Int64(3),
						MinSize: aws.Int64(1),
					},
					&autoscaling.Group{
						AutoScalingGroupName: aws.String("asg-2"),
						AvailabilityZones:    []*string{aws.String("us-east-1c")},
						DesiredCapacity:      aws.Int64(1),
						Instances: []*autoscaling.Instance{
							&autoscaling.Instance{
								AvailabilityZone: aws.String("us-east-1c"),
								InstanceId:       aws.String("instanceId-2"),
								InstanceType:     aws.String("t3.small"),
							},
						},
						MaxSize: aws.Int64(2),
						MinSize: aws.Int64(1),
					},
					&autoscaling.Group{
						AutoScalingGroupName: aws.String("asg-3"),
						AvailabilityZones:    []*string{aws.String("us-east-1d")},
						DesiredCapacity:      aws.Int64(1),
						Instances: []*autoscaling.Instance{
							&autoscaling.Instance{
								AvailabilityZone: aws.String("us-east-1d"),
								InstanceId:       aws.String("instanceId-3"),
								InstanceType:     aws.String("t3.small"),
							},
						},
						MaxSize: aws.Int64(2),
						MinSize: aws.Int64(1),
					},
				},
				NextToken: nil,
			},
			expectedTerminate: []*string{
				aws.String("instanceId-4"),
			},
		},
	}

	for _, c := range cases {
		log.Printf("Running CASE %s", c.title)
		ctx := context.Background()

		a := &mockedASGAPI{
			describeResp: c.describeASGResp,
		}
		e := &mockedEC2API{}
		var currentReplenisher *operatorv1alpha1.AWSNodeReplenisher
		cli := &mockedClient{
			getFunc: func(obj client.Object) error {
				_, ok := obj.(*operatorv1alpha1.AWSNodeReplenisher)
				if ok {
					obj.(*operatorv1alpha1.AWSNodeReplenisher).ObjectMeta = c.replenisher.ObjectMeta
					obj.(*operatorv1alpha1.AWSNodeReplenisher).Spec = c.replenisher.Spec
					obj.(*operatorv1alpha1.AWSNodeReplenisher).Status = c.replenisher.Status
					currentReplenisher = obj.(*operatorv1alpha1.AWSNodeReplenisher)
					return nil
				}
				_, ok = obj.(*operatorv1alpha1.AWSNodeManager)
				if ok {
					obj.(*operatorv1alpha1.AWSNodeManager).ObjectMeta = c.ownerManager.ObjectMeta
					obj.(*operatorv1alpha1.AWSNodeManager).Spec = c.ownerManager.Spec
					obj.(*operatorv1alpha1.AWSNodeManager).Status = c.ownerManager.Status
					return nil
				}
				return errors.New("unknown object")
			},
		}
		r := AWSNodeReplenisherReconciler{
			Client:   cli,
			Recorder: &mockedRecorder{},
			cloud: &cloudaws.AWS{
				EC2:         e,
				Autoscaling: a,
			},
		}

		err := r.syncReplenisher(ctx, c.replenisher)
		if err != nil {
			t.Errorf("CASE: %s : error has occur: %v", c.title, err)
		}
		if currentReplenisher != nil {
			if currentReplenisher.Status.Phase != c.expectedPhase {
				t.Errorf("CASE: %s : phase is not matched, expected %s, returned %s", c.title, c.expectedPhase, currentReplenisher.Status.Phase)
			}
		} else if cli.updatedObj != nil {
			if cli.updatedObj.(*operatorv1alpha1.AWSNodeReplenisher).Status.Phase != c.expectedPhase {
				t.Errorf("CASE: %s : phase is not matched, expected %s, returned %s", c.title, c.expectedPhase, cli.updatedObj.(*operatorv1alpha1.AWSNodeReplenisher).Status.Phase)
			}
		} else {
			if c.replenisher.Status.Phase != c.expectedPhase {
				t.Errorf("CASE: %s : phase is not matched, expected %s, returned %s", c.title, c.expectedPhase, c.replenisher.Status.Phase)
			}
		}
		for key, val := range c.expectedUpdated {
			if !reflect.DeepEqual(*a.updatedASG[key], *val) {
				t.Errorf("CASE: %s : updated object is not matched, expected %+v, returned %+v", c.title, *val, *a.updatedASG[key])
			}
		}

		if c.expectedTerminate != nil {
			if !reflect.DeepEqual(e.terminatedInstances, c.expectedTerminate) {
				t.Errorf("CASE: %s : terminated instances are not matched, expected %+v, returned %+v", c.title, c.expectedTerminate, e.terminatedInstances)
			}
		}
	}
}

func TestShouldSync(t *testing.T) {
	cases := []struct {
		title        string
		replenisher  *operatorv1alpha1.AWSNodeReplenisher
		expectedSync bool
	}{
		{
			title: "Nodes are insufficient",
			replenisher: &operatorv1alpha1.AWSNodeReplenisher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-replenisher",
				},
				Spec: operatorv1alpha1.AWSNodeReplenisherSpec{
					Region:                   "us-east-1",
					AutoScalingGroups:        nil,
					Desired:                  2,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
				},
				Status: operatorv1alpha1.AWSNodeReplenisherStatus{
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "172-32-16-0",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
					},
					NotJoinedAWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Time{},
					},
					Revision: 1,
					Phase:    operatorv1alpha1.AWSNodeReplenisherSynced,
				},
			},
			expectedSync: true,
		},
		{
			title: "Nodes are enough",
			replenisher: &operatorv1alpha1.AWSNodeReplenisher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-replenisher",
				},
				Spec: operatorv1alpha1.AWSNodeReplenisherSpec{
					Region:                   "us-east-1",
					AutoScalingGroups:        nil,
					Desired:                  2,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
				},
				Status: operatorv1alpha1.AWSNodeReplenisherStatus{
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "172-32-16-0",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
						{
							Name:                 "172-32-16-1",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-2",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
					},
					NotJoinedAWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Time{},
					},
					Revision: 1,
					Phase:    operatorv1alpha1.AWSNodeReplenisherSynced,
				},
			},
			expectedSync: false,
		},
		{
			title: "Nodes are excessive",
			replenisher: &operatorv1alpha1.AWSNodeReplenisher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-replenisher",
				},
				Spec: operatorv1alpha1.AWSNodeReplenisherSpec{
					Region:                   "us-east-1",
					AutoScalingGroups:        nil,
					Desired:                  2,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
				},
				Status: operatorv1alpha1.AWSNodeReplenisherStatus{
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "172-32-16-0",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
						{
							Name:                 "172-32-16-1",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-2",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
						{
							Name:                 "172-32-16-2",
							InstanceID:           "instanceId-3",
							AvailabilityZone:     "us-east-1d",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-3",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
					},
					NotJoinedAWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Time{},
					},
					Revision: 1,
					Phase:    operatorv1alpha1.AWSNodeReplenisherSynced,
				},
			},
			expectedSync: true,
		},
		{
			title: "Not joine instances exist",
			replenisher: &operatorv1alpha1.AWSNodeReplenisher{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-replenisher",
				},
				Spec: operatorv1alpha1.AWSNodeReplenisherSpec{
					Region:                   "us-east-1",
					AutoScalingGroups:        nil,
					Desired:                  1,
					ASGModifyCoolTimeSeconds: 600,
					Role:                     operatorv1alpha1.Worker,
				},
				Status: operatorv1alpha1.AWSNodeReplenisherStatus{
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "172-32-16-0",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
					},
					NotJoinedAWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "172-32-16-1",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1d",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-3",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-1 * time.Hour),
							},
						},
					},
					LastASGModifiedTime: &metav1.Time{
						Time: time.Time{},
					},
					Revision: 1,
					Phase:    operatorv1alpha1.AWSNodeReplenisherSynced,
				},
			},
			expectedSync: true,
		},
	}
	for _, c := range cases {
		log.Printf("Running CASE %s", c.title)

		result := shouldSync(c.replenisher)
		if result != c.expectedSync {
			t.Errorf("CASE: %s : sync is not matched, expected %t, returned %t", c.title, c.expectedSync, result)
		}
	}
}
