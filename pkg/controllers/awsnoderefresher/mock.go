package awsnoderefresher

import (
	"context"
	"fmt"

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
	DescribeAutoScalingGroupsOutput *autoscaling.DescribeAutoScalingGroupsOutput
	UpdateAutoScalingGroupOutput    *autoscaling.UpdateAutoScalingGroupOutput
}

func (m *mockedASGAPI) DescribeAutoScalingGroups(in *autoscaling.DescribeAutoScalingGroupsInput) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	return m.DescribeAutoScalingGroupsOutput, nil
}

func (m *mockedASGAPI) UpdateAutoScalingGroup(in *autoscaling.UpdateAutoScalingGroupInput) (*autoscaling.UpdateAutoScalingGroupOutput, error) {
	return m.UpdateAutoScalingGroupOutput, nil
}

type mockedEC2API struct {
	ec2iface.EC2API
	DescribeInstancesResp  *ec2.DescribeInstancesOutput
	TerminateInstancesResp *ec2.TerminateInstancesOutput
	terminateInstanceID    *string
}

func (m *mockedEC2API) DescribeInstances(in *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	return m.DescribeInstancesResp, nil
}

func (m *mockedEC2API) TerminateInstances(in *ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error) {
	if m.terminateInstanceID != nil {
		if *m.terminateInstanceID != *in.InstanceIds[0] {
			return nil, fmt.Errorf("Terminate target instance id is not matched: %v", in.InstanceIds)
		}
	}
	return m.TerminateInstancesResp, nil
}

type mockedClient struct {
	client.Client
	getFunc  func(obj client.Object) error
	listFunc func(listObj client.ObjectList) error
}

func (m *mockedClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return nil
}

func (m *mockedClient) Get(ctx context.Context, key types.NamespacedName, obj client.Object) error {
	return m.getFunc(obj)
}

func (m *mockedClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return m.listFunc(list)
}

type mockedRecorder struct {
	record.EventRecorder
}

func (m *mockedRecorder) Event(object runtime.Object, eventtype, reason, messageFmt string) {
}

func (m *mockedRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}) {
}
