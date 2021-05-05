package awsnodereplenisher

import (
	"context"
	"log"
	"testing"
	"time"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	"github.com/h3poteto/node-manager/pkg/cloud/aws"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestSyncNotJoinedAWSNodes(t *testing.T) {
	cases := []struct {
		title          string
		replenisher    *operatorv1alpha1.AWSNodeReplenisher
		expectedStatus operatorv1alpha1.AWSNodeReplenisherPhase
	}{
		{
			title: "All instances are joined",
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
					AWSNodes:          nil,
					NotJoinedAWSNodes: nil,
					LastASGModifiedTime: &metav1.Time{
						Time: time.Time{},
					},
					Revision: 1,
					Phase:    operatorv1alpha1.AWSNodeReplenisherSynced,
				},
			},
			expectedStatus: operatorv1alpha1.AWSNodeReplenisherSynced,
		},
		{
			title: "Not joined instances exist, and should wait",
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
					AWSNodes: nil,
					NotJoinedAWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "172-32-16-0",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-10 * time.Minute),
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
			expectedStatus: operatorv1alpha1.AWSNodeReplenisherAWSUpdating,
		},
		{
			title: "Not joined instances exist, and should wait",
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
					AWSNodes: nil,
					NotJoinedAWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:                 "172-32-16-0",
							InstanceID:           "instanceId-1",
							AvailabilityZone:     "us-east-1a",
							InstanceType:         "t3.small",
							AutoScalingGroupName: "asg-1",
							CreationTimestamp: metav1.Time{
								Time: time.Now().Add(-2 * time.Hour),
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
			expectedStatus: operatorv1alpha1.AWSNodeReplenisherAWSUpdating,
		},
	}

	for _, c := range cases {
		log.Printf("Running CASE: %s", c.title)
		ctx := context.Background()

		r := &AWSNodeReplenisherReconciler{
			Client: &mockedClient{
				getFunc: func(obj client.Object) error {
					obj.(*operatorv1alpha1.AWSNodeReplenisher).ObjectMeta = c.replenisher.ObjectMeta
					obj.(*operatorv1alpha1.AWSNodeReplenisher).Spec = c.replenisher.Spec
					obj.(*operatorv1alpha1.AWSNodeReplenisher).Status = c.replenisher.Status
					return nil
				},
			},
			Recorder: &mockedRecorder{},
			cloud: &aws.AWS{
				EC2:         &mockedEC2API{},
				Autoscaling: &mockedASGAPI{},
			},
		}
		err := r.syncNotJoinedAWSNodes(ctx, c.replenisher)
		if err != nil {
			t.Errorf("CASE: %s : error has occur: %v", c.title, err)
		}
	}
}
