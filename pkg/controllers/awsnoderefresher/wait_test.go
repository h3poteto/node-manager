package awsnoderefresher

import (
	"context"
	"log"
	"testing"
	"time"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	cloudaws "github.com/h3poteto/node-manager/pkg/cloud/aws"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRefreshAWSWait(t *testing.T) {
	cases := []struct {
		title     string
		refresher *operatorv1alpha1.AWSNodeRefresher
	}{
		{
			title: "Phase is not replacing",
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
					AWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherUpdateIncreasing,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(-15 * time.Minute),
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
						CreationTimestamp: metav1.Time{
							Time: time.Time{},
						},
					},
				},
			},
		},
		{
			title: "Node still living",
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
								Time: time.Now().Add(-24 * time.Hour),
							},
						},
						{
							Name:                 "worker-2",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-24 * time.Hour),
							},
						},
						{
							Name:                 "worker-3",
							InstanceID:           "instanceId-3",
							AvailabilityZone:     "us-east-1d",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-24 * time.Hour),
							},
						},
					},
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherUpdateReplacing,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(-15 * time.Minute),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					ReplaceTargetNode: &operatorv1alpha1.AWSNode{
						Name:                 "worker-3",
						InstanceID:           "instanceId-3",
						AvailabilityZone:     "us-east-1d",
						InstanceType:         "t3.small",
						AutoScalingGroupName: "asg-1",
						CreationTimestamp: metav1.Time{
							Time: time.Now().Add(-24 * time.Hour),
						},
					},
				},
			},
		},
		{
			title: "Should wait",
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
								Time: time.Now().Add(-24 * time.Hour),
							},
						},
						{
							Name:                 "worker-2",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-24 * time.Hour),
							},
						},
					},
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherUpdateReplacing,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(-15 * time.Minute),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-10 * time.Minute),
					},
					ReplaceTargetNode: &operatorv1alpha1.AWSNode{
						Name:                 "worker-3",
						InstanceID:           "instanceId-3",
						AvailabilityZone:     "us-east-1d",
						InstanceType:         "t3.small",
						AutoScalingGroupName: "asg-1",
						CreationTimestamp: metav1.Time{
							Time: time.Now().Add(-24 * time.Hour),
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		log.Printf("Running CASE: %s", c.title)
		ctx := context.Background()

		r := &AWSNodeRefresherReconciler{
			cloud:    &cloudaws.AWS{},
			Client:   &mockedClient{},
			Recorder: &mockedRecorder{},
		}

		err := r.refreshAWSWait(ctx, c.refresher)
		if err != nil {
			t.Errorf("CASE: %s : Failed to refresh wait: %v", c.title, err)
		}
	}
}

func TestCheckInstances(t *testing.T) {
	cases := []struct {
		title     string
		refresher *operatorv1alpha1.AWSNodeRefresher
		waiting   bool
		enough    bool
	}{
		{
			title: "Waiting cooltime",
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
					AWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-1 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherUpdateAWSWaiting,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(-20 * time.Minute),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-20 * time.Minute),
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
			enough:  false,
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
						Time: time.Now().Add(-15 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherUpdateAWSWaiting,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(-20 * time.Minute),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-20 * time.Minute),
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
			enough:  false,
		},
		{
			title: "Enough instances",
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
							AvailabilityZone:     "us-east-1d",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-10 * time.Hour),
							},
						},
					},
					LastASGModifiedTime: &metav1.Time{
						Time: time.Now().Add(-15 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherUpdateAWSWaiting,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(-20 * time.Minute),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-20 * time.Minute),
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
			enough:  true,
		},
	}

	for _, c := range cases {
		log.Printf("Running CASE: %s", c.title)
		ctx := context.Background()

		r := &AWSNodeRefresherReconciler{
			cloud:    &cloudaws.AWS{},
			Client:   &mockedClient{},
			Recorder: &mockedRecorder{},
		}

		waiting, enough := r.checkInstances(ctx, c.refresher)
		if waiting != c.waiting {
			t.Errorf("CASE: %s : Waiting is not matched, expected %t, returned %t", c.title, c.waiting, waiting)
		}
		if enough != c.enough {
			t.Errorf("CASE: %s : Enough is not matched, expected %t, returned %t", c.title, c.enough, enough)
		}
	}
}

func TestAllReplaced(t *testing.T) {
	cases := []struct {
		title     string
		refresher *operatorv1alpha1.AWSNodeRefresher
		replaced  bool
	}{
		{
			title: "All nodes are not replaced",
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
								Time: time.Now().Add(-10 * time.Minute),
							},
						},
						{
							Name:                 "worker-2",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-15 * time.Minute),
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
						Time: time.Now().Add(-10 * time.Minute),
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherUpdateAWSWaiting,
					NextUpdateTime: &metav1.Time{
						Time: time.Now().Add(-30 * time.Minute),
					},
					UpdateStartTime: &metav1.Time{
						Time: time.Now().Add(-1 * time.Minute),
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
			replaced: false,
		},
		{
			title: "All nodes are replaced",
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
								Time: time.Now().Add(-10 * time.Minute),
							},
						},
						{
							Name:                 "worker-2",
							InstanceID:           "instanceId-2",
							AvailabilityZone:     "us-east-1c",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-15 * time.Minute),
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
						Time: time.Now().Add(-10 * time.Minute),
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
			replaced: true,
		},
	}

	for _, c := range cases {
		log.Printf("Running CASE: %s", c.title)
		ctx := context.Background()

		r := &AWSNodeRefresherReconciler{
			cloud:    &cloudaws.AWS{},
			Client:   &mockedClient{},
			Recorder: &mockedRecorder{},
		}

		replaced := r.allReplaced(ctx, c.refresher)
		if replaced != c.replaced {
			t.Errorf("CASE: %s : Replaced is not matched, expected %t, returned %t", c.title, c.replaced, replaced)
		}
	}
}
