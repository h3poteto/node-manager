package aws

import (
	"errors"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"k8s.io/klog/v2"

	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
)

func (a *AWS) AddInstancesToAutoScalingGroups(groups []operatorv1alpha1.AutoScalingGroup, totalDesired int, currentNodesCount int) error {
	if totalDesired <= currentNodesCount {
		return NewDesiredInvalidErrorf("desired does not exceed current, totalDesired: %d, currentNodesCount: %d", totalDesired, currentNodesCount)
	}
	var asgNameList []*string
	for i := range groups {
		asgNameList = append(asgNameList, aws.String(groups[i].Name))
	}
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: asgNameList,
	}
	output, err := a.Autoscaling.DescribeAutoScalingGroups(input)
	if err != nil {
		klog.Errorf("failed to describe AutoScalingGroups: %v", err)
		return err
	}

	sumASGDesired := 0
	sumASGInstances := 0
	// safetyASGs have same value desired capacity and current instances count.
	var safetyASGs []*autoscaling.Group
	for _, asg := range output.AutoScalingGroups {
		sumASGDesired += int(*asg.DesiredCapacity)
		sumASGInstances += len(asg.Instances)
		if int(*asg.DesiredCapacity) == len(asg.Instances) {
			safetyASGs = append(safetyASGs, asg)
		}
	}

	if currentNodesCount != sumASGInstances {
		err := NewInstanceNotYetJoinErrorf("not all instances join the cluster as nodes, all instances: %d, current nodes: %d", sumASGInstances, currentNodesCount)
		return err
	}

	if len(safetyASGs) < 1 {
		err := errors.New("there are no safety AutoScalingGroups, so could not add instances")
		klog.Error(err)
		return err
	}

	sort.SliceStable(safetyASGs, func(i, j int) bool {
		return (*safetyASGs[i].MaxSize - *safetyASGs[i].DesiredCapacity) > (*safetyASGs[j].MaxSize - *safetyASGs[j].DesiredCapacity)
	})

	// Increase desired capacity equally across all ASGs.
	surplus := totalDesired - currentNodesCount
	if surplus == 0 {
		klog.Info("Don't need to increase desired capacity, so skip it")
		return nil
	}
	for surplus > 0 {
		// Exit this loop when all ASGs capacity is fullfilled.
		if fullfilled := allASGIsFullfilled(safetyASGs); fullfilled {
			break
		}
		for j := range safetyASGs {
			if surplus < 1 {
				break
			}
			if int(*safetyASGs[j].MaxSize-*safetyASGs[j].DesiredCapacity) > 0 {
				*safetyASGs[j].DesiredCapacity += 1
				surplus--
			}
		}
	}
	klog.Infof("spec desired is %d, and current ASGs desired is %d, so increase desired capacity for all ASGs", totalDesired, sumASGDesired)
	if surplus > 0 {
		klog.Warningf("all ASGs is already fullfilled so could not add %d instances", surplus)
	}
	return a.updateASGsDesired(safetyASGs)
}

func (a *AWS) DeleteInstancesToAutoScalingGroups(groups []operatorv1alpha1.AutoScalingGroup, totalDesired int, currentNodesCount int) error {
	if totalDesired >= currentNodesCount {
		return NewDesiredInvalidErrorf("desired exceeds current, totalDesired: %d, currentNodesCount: %d", totalDesired, currentNodesCount)
	}

	// Check desired capacity of each AutScalingGroups
	var asgNameList []*string
	for i := range groups {
		asgNameList = append(asgNameList, aws.String(groups[i].Name))
	}
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: asgNameList,
	}
	output, err := a.Autoscaling.DescribeAutoScalingGroups(input)
	if err != nil {
		klog.Errorf("failed to describe AutoScalingGroups: %v", err)
		return err
	}

	sumASGDesired := 0
	sumASGInstances := 0
	var safetyASGs []*autoscaling.Group
	// unsafetyASGs have different value desired capacity and current instances count.
	var unsafetyASGs []*autoscaling.Group
	for _, asg := range output.AutoScalingGroups {
		sumASGDesired += int(*asg.DesiredCapacity)
		sumASGInstances += len(asg.Instances)
		if int(*asg.DesiredCapacity) != len(asg.Instances) {
			unsafetyASGs = append(unsafetyASGs, asg)
		} else {
			safetyASGs = append(safetyASGs, asg)
		}
	}
	if currentNodesCount != sumASGInstances {
		err := NewInstanceNotYetJoinErrorf("not all instances join the cluster as nodes, all instances: %d, current nodes: %d", sumASGInstances, currentNodesCount)
		return err
	}

	sort.SliceStable(unsafetyASGs, func(i, j int) bool {
		return (*unsafetyASGs[i].DesiredCapacity - *unsafetyASGs[i].MinSize) > (*unsafetyASGs[j].DesiredCapacity - *unsafetyASGs[j].MinSize)
	})

	if len(unsafetyASGs) > 0 {
		targetASG := unsafetyASGs[0]
		newDesired := len(targetASG.Instances)
		klog.Infof("there is invalid AutoScalingGroup %s, so decrement desired count: %d", *targetASG.AutoScalingGroupName, newDesired)
		_ = a.updateASGCapacity(targetASG, newDesired)
	}

	if len(safetyASGs) < 1 {
		err := errors.New("there are no safety AutoScalingGroups, so could not delete instances")
		klog.Error(err)
		return err
	}

	sort.SliceStable(safetyASGs, func(i, j int) bool {
		return (*safetyASGs[i].DesiredCapacity - *safetyASGs[i].MinSize) > (*safetyASGs[j].DesiredCapacity - *safetyASGs[j].MinSize)
	})

	// Decrease desired capacity equally across all ASGs.
	surplus := currentNodesCount - totalDesired
	for surplus > 0 {
		// Exit this loop when all ASGs capacity is minimized.
		if minimized := allASGIsMinimized(safetyASGs); minimized {
			break
		}
		for j := range safetyASGs {
			if surplus < 1 {
				break
			}
			if int(*safetyASGs[j].DesiredCapacity-*safetyASGs[j].MinSize) > 0 {
				*safetyASGs[j].DesiredCapacity -= 1
				surplus--
			}
		}
	}
	klog.Infof("spec desired is %d, and current ASG desired is %d, so decrement desired capacity for all ASGs", totalDesired, sumASGDesired)
	if surplus > 0 {
		klog.Warningf("all ASGs is already minimized so could not delete %d instances", surplus)
	}
	return a.updateASGsDesired(safetyASGs)
}

func (a *AWS) updateASGsDesired(asgs []*autoscaling.Group) error {
	var err []error
	for i := range asgs {
		if e := a.updateASGCapacity(asgs[i], int(*asgs[i].DesiredCapacity)); e != nil {
			err = append(err, e)
		}
	}
	if len(err) > 0 {
		return err[0]
	}
	return nil
}

func (a *AWS) updateASGCapacity(asg *autoscaling.Group, newDesired int) error {
	if newDesired > int(*asg.MaxSize) {
		klog.Warningf("AutoScalingGroup %s has reached capacity limit, new desired: %d, but max: %d, so reduce new desired", *asg.AutoScalingGroupName, newDesired, *asg.MaxSize)
		newDesired = int(*asg.MaxSize)
	}
	if newDesired < int(*asg.MinSize) {
		klog.Warningf("AutoScalingGroup %s has reached capacity minimize, new desired: %d, but min: %d so increase new desired", *asg.AutoScalingGroupName, newDesired, *asg.MinSize)
		newDesired = int(*asg.MinSize)
	}
	updateInput := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: asg.AutoScalingGroupName,
		DesiredCapacity:      aws.Int64(int64(newDesired)),
	}
	_, err := a.Autoscaling.UpdateAutoScalingGroup(updateInput)
	if err != nil {
		klog.Errorf("failed to update AutoScalingGroups: %v", err)
		return err
	}
	klog.Infof("updated desired capacity of AutoScalingGroup %s", *asg.AutoScalingGroupName)
	return nil
}

func allASGIsFullfilled(asgs []*autoscaling.Group) bool {
	for i := range asgs {
		if int(*asgs[i].MaxSize-*asgs[i].DesiredCapacity) > 0 {
			return false
		}
	}
	return true
}

func allASGIsMinimized(asgs []*autoscaling.Group) bool {
	for i := range asgs {
		if int(*asgs[i].DesiredCapacity-*asgs[i].MinSize) > 0 {
			return false
		}
	}
	return true
}
