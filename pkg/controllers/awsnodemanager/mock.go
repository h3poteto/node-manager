package awsnodemanager

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
	DescribeAutoScalingGroupsOutput *autoscaling.DescribeAutoScalingGroupsOutput
}

func (m *mockedASGAPI) DescribeAutoScalingGroups(in *autoscaling.DescribeAutoScalingGroupsInput) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	return m.DescribeAutoScalingGroupsOutput, nil
}

type mockedEC2API struct {
	ec2iface.EC2API
	describeFunc func(in *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
}

func (m *mockedEC2API) DescribeInstances(in *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	return m.describeFunc(in)
}

type mockedClient struct {
	client.Client
	getFunc    func(obj client.Object) error
	updatedObj client.Object
}

func (m *mockedClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	m.updatedObj = obj
	return nil
}

func (m *mockedClient) Get(ctx context.Context, key types.NamespacedName, obj client.Object) error {
	return m.getFunc(obj)
}

type mockedRecorder struct {
	record.EventRecorder
}

func (m *mockedRecorder) Event(object runtime.Object, eventtype, reason, messageFmt string) {
}

func (m *mockedRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}) {
}
