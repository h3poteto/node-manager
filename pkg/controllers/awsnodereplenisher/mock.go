package awsnodereplenisher

import (
	"context"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type mockedASGAPI struct {
	autoscalingiface.AutoScalingAPI
	detachInstancesOutput *autoscaling.DetachInstancesOutput
	describeResp          *autoscaling.DescribeAutoScalingGroupsOutput
	updatedASG            map[string]*autoscaling.UpdateAutoScalingGroupInput
}

func (m *mockedASGAPI) DetachInstances(in *autoscaling.DetachInstancesInput) (*autoscaling.DetachInstancesOutput, error) {
	return m.detachInstancesOutput, nil
}

func (m *mockedASGAPI) DescribeAutoScalingGroups(in *autoscaling.DescribeAutoScalingGroupsInput) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	return m.describeResp, nil
}

func (m *mockedASGAPI) UpdateAutoScalingGroup(in *autoscaling.UpdateAutoScalingGroupInput) (*autoscaling.UpdateAutoScalingGroupOutput, error) {
	if len(m.updatedASG) > 0 {
		m.updatedASG[*in.AutoScalingGroupName] = in
	} else {
		m.updatedASG = map[string]*autoscaling.UpdateAutoScalingGroupInput{
			*in.AutoScalingGroupName: in,
		}
	}
	return &autoscaling.UpdateAutoScalingGroupOutput{}, nil
}

type mockedEC2API struct {
	ec2iface.EC2API
	terminatedInstances []*string
}

func (m *mockedEC2API) TerminateInstances(in *ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error) {
	m.terminatedInstances = in.InstanceIds
	return &ec2.TerminateInstancesOutput{}, nil
}

type mockedClient struct {
	client.Client
	getFunc    func(obj client.Object) error
	updatedObj client.Object
}

func (m *mockedClient) Get(ctx context.Context, key types.NamespacedName, obj client.Object) error {
	return m.getFunc(obj)
}

func (m *mockedClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	m.updatedObj = obj
	return nil
}

type mockedRecorder struct {
	record.EventRecorder
}

func (m *mockedRecorder) Event(object runtime.Object, eventtype, reason, messageFmt string) {
}

func (m *mockedRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}) {
}
