package aws

import (
	"errors"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"k8s.io/klog/v2"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
)

func (a *AWS) AddInstancesToAutoScalingGroups(groups []operatorv1alpha1.AutoScalingGroup, totalDesired int, count int) error {
	var asgNameList []*string
	for i := range groups {
		asgNameList = append(asgNameList, aws.String(groups[i].Name))
	}
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: asgNameList,
	}
	output, err := a.autoscaling.DescribeAutoScalingGroups(input)
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
		return (*asgs[i].MaxSize - *asgs[i].DesiredCapacity) > (*asgs[j].MaxSize - *asgs[j].DesiredCapacity)
	})
	sort.SliceStable(safetyASGs, func(i, j int) bool {
		return (*safetyASGs[i].MaxSize - *safetyASGs[i].DesiredCapacity) > (*safetyASGs[j].MaxSize - *safetyASGs[j].DesiredCapacity)
	})

	// Increment smallest desired ASG when spec desired and current desired are different.
	if totalDesired != sumCurrentDesired {
		// Smallest desired ASG
		targetASG := asgs[0]
		newDesired := int(*targetASG.DesiredCapacity) + count
		klog.Infof("spec desired is %d, and current ASG desired is %d, so increment desired capacity of smallest ASG: %s", totalDesired, sumCurrentDesired, *targetASG.AutoScalingGroupName)
		return updateASGCapacity(a.autoscaling, targetASG, newDesired)
	} else {
		// Add instance in safety ASG when spec desired and current desired are same.
		if len(safetyASGs) < 1 {
			err := errors.New("there are no safety AutoScalingGroups, so could not add instances")
			klog.Error(err)
			return err
		}
		targetASG := safetyASGs[0]
		newDesired := int(*targetASG.DesiredCapacity) + count
		klog.Infof("spec desired is %d, and current ASG desired is %d, so increment desired capacity of safety ASG: %s", totalDesired, sumCurrentDesired, *targetASG.AutoScalingGroupName)
		return updateASGCapacity(a.autoscaling, targetASG, newDesired)
	}
}

func (a *AWS) DeleteInstancesToAutoScalingGroups(groups []operatorv1alpha1.AutoScalingGroup, totalDesired int, count int) error {
	// Check desired capacity of each AutScalingGroups
	var asgNameList []*string
	for i := range groups {
		asgNameList = append(asgNameList, aws.String(groups[i].Name))
	}
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: asgNameList,
	}
	output, err := a.autoscaling.DescribeAutoScalingGroups(input)
	if err != nil {
		klog.Errorf("failed to describe AutoScalingGroups: %v", err)
		return err
	}
	sumCurrentDesired := 0
	var safetyASGs []*autoscaling.Group
	// unsafetyASGs have different value desired capacity and current instances count.
	var unsafetyASGs []*autoscaling.Group
	for _, asg := range output.AutoScalingGroups {
		sumCurrentDesired += int(*asg.DesiredCapacity)
		if int(*asg.DesiredCapacity) != len(asg.Instances) {
			unsafetyASGs = append(unsafetyASGs, asg)
		} else {
			safetyASGs = append(safetyASGs, asg)
		}
	}
	sort.SliceStable(safetyASGs, func(i, j int) bool {
		return (*safetyASGs[i].DesiredCapacity - *safetyASGs[i].MinSize) > (*safetyASGs[j].DesiredCapacity - *safetyASGs[j].MinSize)
	})
	sort.SliceStable(unsafetyASGs, func(i, j int) bool {
		return (*unsafetyASGs[i].DesiredCapacity - *unsafetyASGs[i].MinSize) > (*unsafetyASGs[j].DesiredCapacity - *unsafetyASGs[j].MinSize)
	})

	if len(unsafetyASGs) > 0 {
		targetASG := unsafetyASGs[0]
		newDesired := len(targetASG.Instances)
		klog.Infof("there is invalid AutoScalingGroup %s, so decrement desired count: %d", *targetASG.AutoScalingGroupName, newDesired)
		_ = updateASGCapacity(a.autoscaling, targetASG, newDesired)
	}

	// Decrement largest desired ASG
	if len(safetyASGs) < 1 {
		err := errors.New("there are no AutoScalingGroups, so could not delete instances")
		klog.Error(err)
		return err
	}
	targetASG := safetyASGs[0]
	newDesired := int(*targetASG.DesiredCapacity) - count
	klog.Infof("spec desired is %d, and current ASG desired is %d, so decrement desired capacity of largest ASG: %s", totalDesired, sumCurrentDesired, *targetASG.AutoScalingGroupName)
	return updateASGCapacity(a.autoscaling, targetASG, newDesired)
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
	if newDesired < int(*asg.MinSize) {
		klog.Warningf("AutoScalingGroup %s has reached capacity minimize, new desired: %d, but min: %d so increase new desired", *asg.AutoScalingGroupName, newDesired, *asg.MinSize)
		newDesired = int(*asg.MinSize)
	}
	if newDesired == int(*asg.DesiredCapacity) {
		err := fmt.Errorf("AutoScalingGroup %s has already minimized, so could not update desired capacity", *asg.AutoScalingGroupName)
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
