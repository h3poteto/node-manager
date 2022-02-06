package awsnoderefresher

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestShouldDrain(t *testing.T) {
	cases := []struct {
		title     string
		refresher *operatorv1alpha1.AWSNodeRefresher
		expected  bool
	}{
		{
			title: "Node count is not match",
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
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:             "node-1",
							InstanceID:       "instance-1",
							AvailabilityZone: "ap-northeast-1a",
							InstanceType:     "t2.small",
						},
						{
							Name:             "node-2",
							InstanceID:       "instance-2",
							AvailabilityZone: "ap-northeast-1c",
							InstanceType:     "t2.small",
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
				},
			},
			expected: false,
		},
		{
			title: "Node count is match",
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
					AWSNodes: []operatorv1alpha1.AWSNode{
						{
							Name:             "node-1",
							InstanceID:       "instance-1",
							AvailabilityZone: "ap-northeast-1a",
							InstanceType:     "t2.small",
						},
						{
							Name:             "node-2",
							InstanceID:       "instance-2",
							AvailabilityZone: "ap-northeast-1c",
							InstanceType:     "t2.small",
						},
						{
							Name:             "node-3",
							InstanceID:       "instance-3",
							AvailabilityZone: "ap-northeast-1d",
							InstanceType:     "t2.small",
						},
						{
							Name:             "node-4",
							InstanceID:       "instance-4",
							AvailabilityZone: "ap-northeast-1c",
							InstanceType:     "t2.small",
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
				},
			},
			expected: true,
		},
	}
	for _, c := range cases {
		log.Printf("Running CASE: %s", c.title)

		result := shouldDrain(context.Background(), c.refresher)
		if result != c.expected {
			t.Errorf("CASE: %s : result is not matched, expected %t, vut returned %t", c.title, c.expected, result)
		}
	}
}

func TestShouldRetryDrain(t *testing.T) {
	cases := []struct {
		title     string
		refresher *operatorv1alpha1.AWSNodeRefresher
		pods      []corev1.Pod
		expected  bool
	}{
		{
			title: "There is not pod in the node",
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
					Phase:    operatorv1alpha1.AWSNodeRefresherDraining,
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
			pods:     []corev1.Pod{},
			expected: false,
		},
		{
			title: "Pods are running on another node",
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
					Phase:    operatorv1alpha1.AWSNodeRefresherDraining,
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
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod1",
						Namespace: "kube-system",
					},
					Spec: corev1.PodSpec{
						NodeName: "node-2",
					},
					Status: corev1.PodStatus{},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod2",
						Namespace: "default",
					},
					Spec: corev1.PodSpec{
						NodeName: "node-3",
					},
					Status: corev1.PodStatus{},
				},
			},
			expected: false,
		},
		{
			title: "Pods are running on the node, but DaemonSet or Static Pod",
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
					Phase:    operatorv1alpha1.AWSNodeRefresherDraining,
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
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod1",
						Namespace: "kube-system",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "apps/v1",
								Kind:       "DaemonSet",
							},
						},
					},
					Spec: corev1.PodSpec{
						NodeName: "node-1",
					},
					Status: corev1.PodStatus{},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod2",
						Namespace: "default",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion: "v1",
								Kind:       "Node",
							},
						},
					},
					Spec: corev1.PodSpec{
						NodeName: "node-1",
					},
					Status: corev1.PodStatus{},
				},
			},
			expected: false,
		},

		{
			title: "Pods are running on the node",
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
					Phase:    operatorv1alpha1.AWSNodeRefresherDraining,
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
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod1",
						Namespace: "kube-system",
					},
					Spec: corev1.PodSpec{
						NodeName: "node-1",
					},
					Status: corev1.PodStatus{},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod2",
						Namespace: "default",
					},
					Spec: corev1.PodSpec{
						NodeName: "node-1",
					},
					Status: corev1.PodStatus{},
				},
			},
			expected: true,
		},
	}

	for _, c := range cases {
		log.Printf("Running CASE: %s", c.title)

		cli := &mockedClient{
			listFunc: func(listObj client.ObjectList) error {
				listObj.(*corev1.PodList).Items = c.pods
				return nil
			},
		}
		r := &AWSNodeRefresherReconciler{
			Client:   cli,
			Recorder: &mockedRecorder{},
		}

		result := r.shouldRetryDrain(context.Background(), c.refresher, "node-1")
		if result != c.expected {
			t.Errorf("CASE: %s : result is not matched, expected %t, vut returned %t", c.title, c.expected, result)
		}
	}
}
