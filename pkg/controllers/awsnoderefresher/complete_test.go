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

func TestRefreshComplete(t *testing.T) {
	cases := []struct {
		title     string
		refresher *operatorv1alpha1.AWSNodeRefresher
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
		},
		{
			title: "Should complete",
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

		err := r.refreshComplete(ctx, c.refresher)
		if err != nil {
			t.Errorf("CASE: %s : Failed to complete: %v", c.title, err)
		}
	}

}
