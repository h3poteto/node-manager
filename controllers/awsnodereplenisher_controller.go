/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
)

const (
	AnnotationKey = "managed.aws-node-replenisher.operator.h3poteto.dev"
)

// AWSNodeReplenisherReconciler reconciles a AWSNodeReplenisher object
type AWSNodeReplenisherReconciler struct {
	client.Client
	Log     logr.Logger
	Scheme  *runtime.Scheme
	Session *session.Session
}

// +kubebuilder:rbac:groups=operator.h3poteto.dev,resources=awsnodereplenishers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.h3poteto.dev,resources=awsnodereplenishers/status,verbs=get;update;patch

func (r *AWSNodeReplenisherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("awsnodereplenisher", req.NamespacedName)

	klog.Info("fetching AWSNodeReplenisher resources")
	replenisher := operatorv1alpha1.AWSNodeReplenisher{}
	if err := r.Client.Get(ctx, req.NamespacedName, &replenisher); err != nil {
		klog.Infof("failed to get AWSNodeReplenisher resources: %v", err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Generate aws client
	r.Session = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	if err := r.syncReplenisher(ctx, &replenisher); err != nil {
		klog.Errorf("failed to sync AWSNodeReplenisher: %v", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *AWSNodeReplenisherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.AWSNodeReplenisher{}).
		Complete(r)
}

// syncReplenisher checks nodes and replenish AWS instances when node resources are not enough.
func (r *AWSNodeReplenisherReconciler) syncReplenisher(ctx context.Context, replenisher *operatorv1alpha1.AWSNodeReplenisher) error {
	klog.Info("syncing nodes and aws instances")
	if err := r.syncAWSNodes(ctx, replenisher); err != nil {
		return err
	}

	klog.Info("reading node info from status")

	if len(replenisher.Status.AWSNodes) == int(replenisher.Spec.Desired) {
		klog.Info("nodes count is same as desired count")
		return r.updateStatusSynced(ctx, replenisher)
	}

	now := time.Now()
	if replenisher.Status.Phase == operatorv1alpha1.AWSNodeReplenisherAWSUpdating && replenisher.Status.LastASGModifiedTime != nil && now.Before(replenisher.Status.LastASGModifiedTime.Add(10*time.Minute)) {
		klog.Info("Waiting cool time")
		return nil
	}

	if len(replenisher.Status.AWSNodes) > int(replenisher.Spec.Desired) {
		klog.Infof("nodes count is %d, but desired count is %d, so deleting nodes", len(replenisher.Status.AWSNodes), replenisher.Spec.Desired)
		// TODO: deleting
	} else {
		klog.Infof("nodes count is %d, but desired count is %d, so adding nodes", len(replenisher.Status.AWSNodes), replenisher.Spec.Desired)
		if err := r.addNode(ctx, replenisher, int(replenisher.Spec.Desired)-len(replenisher.Status.AWSNodes)); err != nil {
			return err
		}
	}
	return nil
}

func (r *AWSNodeReplenisherReconciler) syncAWSNodes(ctx context.Context, replenisher *operatorv1alpha1.AWSNodeReplenisher) error {
	for i, node := range replenisher.Status.AWSNodes {
		if node.InstanceID != "" {
			continue
		}
		svc := ec2.New(r.Session, aws.NewConfig().WithRegion(replenisher.Spec.Region))
		input := &ec2.DescribeInstancesInput{
			DryRun: nil,
			Filters: []*ec2.Filter{
				{
					Name: aws.String("private-dns-name"),
					Values: []*string{
						aws.String(node.Name),
					},
				},
			},
		}
		output, err := svc.DescribeInstances(input)
		if err != nil {
			klog.Errorf("failed to describe aws instances: %v", err)
			return err
		}
		if len(output.Reservations) < 1 || len(output.Reservations[0].Instances) < 1 {
			klog.Warningf("could not find aws instance %s", node.Name)
			continue
		}
		instance := output.Reservations[0].Instances[0]
		replenisher.Status.AWSNodes[i].InstanceID = *instance.InstanceId
		replenisher.Status.AWSNodes[i].InstanceType = *instance.InstanceType
		replenisher.Status.AWSNodes[i].AvailabilityZone = *instance.Placement.AvailabilityZone
		// Normally auto scaling group name is filled in name tag of instances.
		tag := findTag(instance.Tags, "Name")
		if tag == nil {
			klog.Warningf("could not find Name tag in aws instance %s", *instance.InstanceId)
			continue
		}
		replenisher.Status.AWSNodes[i].AutoScalingGroupName = *tag.Value
	}
	currentReplenisher := operatorv1alpha1.AWSNodeReplenisher{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: replenisher.Namespace, Name: replenisher.Name}, &currentReplenisher); err != nil {
		klog.Errorf("failed to get AWSNodeReplenisher: %v", err)
		return err
	}
	if reflect.DeepEqual(currentReplenisher.Status, replenisher.Status) {
		klog.Infof("AWSNodeReplenisher %s/%s is already synced", replenisher.Namespace, replenisher.Name)
		return nil
	}
	currentReplenisher.Status = replenisher.Status
	currentReplenisher.Status.Revision += 1
	// update replenisher status
	klog.Infof("updating replenisher status: %s/%s", currentReplenisher.Namespace, currentReplenisher.Name)
	if err := r.Client.Update(ctx, &currentReplenisher); err != nil {
		klog.Errorf("failed to update replenisher %s/%s: %v", currentReplenisher.Namespace, currentReplenisher.Name, err)
		return err
	}
	klog.Infof("success to update repelnisher %s/%s", currentReplenisher.Namespace, currentReplenisher.Name)
	return nil
}

func (r *AWSNodeReplenisherReconciler) addNode(ctx context.Context, replenisher *operatorv1alpha1.AWSNodeReplenisher, count int) error {
	if err := r.updateStatusAWSUpdating(ctx, replenisher); err != nil {
		return err
	}

	// Check desired capacity of each AutScalingGroups
	var asgNameList []*string
	for i := range replenisher.Spec.AutoScalingGroups {
		asgNameList = append(asgNameList, aws.String(replenisher.Spec.AutoScalingGroups[i].Name))
	}
	svc := autoscaling.New(r.Session, aws.NewConfig().WithRegion(replenisher.Spec.Region))
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: asgNameList,
	}

	output, err := svc.DescribeAutoScalingGroups(input)
	if err != nil {
		klog.Errorf("failed to describe AutoScalingGroups: %v", err)
		return err
	}
	sumCurrentDesired := 0
	var asgs []*autoscaling.Group
	// safetyASGs have same value desired capacity and current instances count.
	var safetyASGs []*autoscaling.Group
	for _, asg := range output.AutoScalingGroups {
		sumCurrentDesired += int(*asg.DesiredCapacity)
		asgs = append(asgs, asg)
		if int(*asg.DesiredCapacity) == len(asg.Instances) {
			safetyASGs = append(safetyASGs, asg)
		}
	}
	sort.SliceStable(asgs, func(i, j int) bool {
		return *asgs[i].DesiredCapacity < *asgs[j].DesiredCapacity
	})
	sort.SliceStable(safetyASGs, func(i, j int) bool {
		return (*safetyASGs[i].MaxSize - *safetyASGs[i].DesiredCapacity) > (*safetyASGs[j].MaxSize - *safetyASGs[j].DesiredCapacity)
	})

	// Increment smallest desired ASG when spec desired and current desired are different.
	if int(replenisher.Spec.Desired) != sumCurrentDesired {
		// Smallest desired ASG
		targetASG := asgs[0]
		newDesired := int(*targetASG.DesiredCapacity) + count
		klog.Infof("spec desired is %d, and current ASG desired is %d, so increment desired capacity of smallest ASG: %s", replenisher.Spec.Desired, sumCurrentDesired, *targetASG.AutoScalingGroupName)
		return updateASGCapacity(svc, targetASG, newDesired)
	} else {
		// Add instance in safety ASG when spec desired and current desired are same.
		if len(safetyASGs) < 1 {
			err := errors.New("there are no safety AutoScalingGroups, so could not add instances")
			klog.Error(err)
			return err
		}
		targetASG := safetyASGs[0]
		newDesired := int(*targetASG.DesiredCapacity) + count
		klog.Infof("spec desired is %d, and current ASG desired is %d, so increment desired capacity of safety ASG: %s", replenisher.Spec.Desired, sumCurrentDesired, *targetASG.AutoScalingGroupName)
		return updateASGCapacity(svc, targetASG, newDesired)
	}
}

func (r *AWSNodeReplenisherReconciler) deleteNode(ctx context.Context, replenisher *operatorv1alpha1.AWSNodeReplenisher, count int) error {
	return nil
}

func (r *AWSNodeReplenisherReconciler) updateLatestTimestamp(ctx context.Context, replenisher *operatorv1alpha1.AWSNodeReplenisher, now metav1.Time) error {
	// We have to retry when this update function is failed.
	// If we don't update LastASGModifiedTime after modify some ASGs,
	// the next process in reconcile will add/delete instances without wait.
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		currentReplenisher := operatorv1alpha1.AWSNodeReplenisher{}
		if err := r.Client.Get(ctx, client.ObjectKey{Namespace: replenisher.Namespace, Name: replenisher.Name}, &currentReplenisher); err != nil {
			klog.Errorf("failed to get AWSNodeReplenisher %s/%s: %v", replenisher.Namespace, replenisher.Name, err)
			return err
		}
		currentReplenisher.Status.LastASGModifiedTime = &now
		currentReplenisher.Status.Revision += 1
		if err := r.Client.Update(ctx, &currentReplenisher); err != nil {
			klog.Errorf("failed to update AWSNodeReplenisher status %s/%s: %v", replenisher.Namespace, replenisher.Name, err)
			return err
		}
		return nil
	})
}

func (r *AWSNodeReplenisherReconciler) updateStatusSynced(ctx context.Context, replenisher *operatorv1alpha1.AWSNodeReplenisher) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		currentReplenisher := operatorv1alpha1.AWSNodeReplenisher{}
		if err := r.Client.Get(ctx, client.ObjectKey{Namespace: replenisher.Namespace, Name: replenisher.Name}, &currentReplenisher); err != nil {
			klog.Errorf("failed to get AWSNodeReplenisher %s/%s: %v", replenisher.Namespace, replenisher.Name, err)
			return err
		}
		if currentReplenisher.Status.Phase == operatorv1alpha1.AWSNodeReplenisherSynced {
			klog.Infof("AWSNodeReplenisher %s/%s is already synced", currentReplenisher.Namespace, currentReplenisher.Name)
			return nil
		}
		currentReplenisher.Status.Phase = operatorv1alpha1.AWSNodeReplenisherSynced
		currentReplenisher.Status.Revision += 1
		if err := r.Client.Update(ctx, &currentReplenisher); err != nil {
			klog.Errorf("failed to update AWSNodeReplenisher status %s/%s: %v", replenisher.Namespace, replenisher.Name, err)
			return err
		}
		return nil
	})
}

func (r *AWSNodeReplenisherReconciler) updateStatusAWSUpdating(ctx context.Context, replenisher *operatorv1alpha1.AWSNodeReplenisher) error {
	now := metav1.Now()
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		currentReplenisher := operatorv1alpha1.AWSNodeReplenisher{}
		if err := r.Client.Get(ctx, client.ObjectKey{Namespace: replenisher.Namespace, Name: replenisher.Name}, &currentReplenisher); err != nil {
			klog.Errorf("failed to get AWSNodeReplenisher %s/%s: %v", replenisher.Namespace, replenisher.Name, err)
			return err
		}
		currentReplenisher.Status.Phase = operatorv1alpha1.AWSNodeReplenisherAWSUpdating
		currentReplenisher.Status.LastASGModifiedTime = &now
		currentReplenisher.Status.Revision += 1
		if err := r.Client.Update(ctx, &currentReplenisher); err != nil {
			klog.Errorf("failed to update AWSNodeReplenisher status %s/%s: %v", replenisher.Namespace, replenisher.Name, err)
			return err
		}
		return nil
	})
}

func findTag(tags []*ec2.Tag, key string) *ec2.Tag {
	for i := range tags {
		if *tags[i].Key == key {
			return tags[i]
		}
	}
	return nil
}

func updateASGCapacity(client *autoscaling.AutoScaling, asg *autoscaling.Group, newDesired int) error {
	if newDesired > int(*asg.MaxSize) {
		klog.Warningf("AutoScalingGroup %s has reached capacity limit, new desired: %d, but max: %d, so reduce new desired", *asg.AutoScalingGroupName, newDesired, *asg.MaxSize)
		newDesired = int(*asg.MaxSize)
	}
	if newDesired == int(*asg.DesiredCapacity) {
		err := fmt.Errorf("AutoScalingGroup %s has already fullfilled, so could not update desired capacity", *asg.AutoScalingGroupName)
		klog.Error(err)
		return err
	}
	updateInput := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: asg.AutoScalingGroupName,
		DesiredCapacity:      aws.Int64(int64(newDesired)),
	}
	_, err := client.UpdateAutoScalingGroup(updateInput)
	if err != nil {
		klog.Errorf("failed to update AutoScalingGroups: %v", err)
		return err
	}
	klog.Infof("updated desired capacity of AutoScalingGroup %s", *asg.AutoScalingGroupName)
	return nil
}
