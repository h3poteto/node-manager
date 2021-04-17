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

func TestScheduleNext(t *testing.T) {
	now := metav1.Now()
	cases := []struct {
		title     string
		refresher *operatorv1alpha1.AWSNodeRefresher
	}{
		{
			title: "Next schedule",
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
					Schedule:                 "3 10 * * *",
					SurplusNodes:             1,
				},
				Status: operatorv1alpha1.AWSNodeRefresherStatus{
					AWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Time{},
					},
					Revision: 0,
					Phase:    operatorv1alpha1.AWSNodeRefresherInit,
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

		err := r.scheduleNext(ctx, c.refresher)
		if err != nil {
			t.Errorf("CASE: %s : Failed to next schedule: %v", c.title, err)
		}
		n := metav1.Time{
			Time: time.Time{},
		}
		if *c.refresher.Status.NextUpdateTime == n {
			t.Errorf("CASE: %s : Failed to update next update time: %v", c.title, *c.refresher.Status.NextUpdateTime)
		}
		if c.refresher.Status.NextUpdateTime.Before(&now) {
			t.Errorf("CASE: %s : NextUpdateTime is not after from now: %v", c.title, *c.refresher.Status.NextUpdateTime)
		}
	}
}
