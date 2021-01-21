package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	operatorv1alpha1 "github.com/h3poteto/node-manager/api/v1alpha1"
)

type mockedAutoScalingAPI struct {
	autoscalingiface.AutoScalingAPI
	Resp              autoscaling.DescribeAutoScalingGroupsOutput
	RequestASGDesired map[string]int64
}

func (m *mockedAutoScalingAPI) DescribeAutoScalingGroups(in *autoscaling.DescribeAutoScalingGroupsInput) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
	return &m.Resp, nil
}

func (m *mockedAutoScalingAPI) UpdateAutoScalingGroup(in *autoscaling.UpdateAutoScalingGroupInput) (*autoscaling.UpdateAutoScalingGroupOutput, error) {
	m.RequestASGDesired = map[string]int64{
		*in.AutoScalingGroupName: *in.DesiredCapacity,
	}
	return &autoscaling.UpdateAutoScalingGroupOutput{}, nil
}

func TestInstancesToAutoScalingGroups(t *testing.T) {
	asgName := "nodes-ap-northeast-1a"
	region := "ap-northeast-1a"
	instances := []*autoscaling.Instance{
		{
			AvailabilityZone: aws.String(region),
			InstanceId:       aws.String("test1"),
			InstanceType:     aws.String("t3.medium"),
		},
	}
	resp := autoscaling.DescribeAutoScalingGroupsOutput{
		AutoScalingGroups: []*autoscaling.Group{
			&autoscaling.Group{
				AutoScalingGroupName: aws.String(asgName),
				AvailabilityZones: []*string{
					aws.String(region),
				},
				DesiredCapacity: aws.Int64(1),
				Instances:       instances,
				MaxSize:         aws.Int64(2),
				MinSize:         aws.Int64(0),
			},
		},
		NextToken: nil,
	}

	mocked := &mockedAutoScalingAPI{
		Resp: resp,
	}
	a := &AWS{
		autoscaling: mocked,
	}

	groups := []operatorv1alpha1.AutoScalingGroup{
		{
			Name: asgName,
		},
	}

	err := a.AddInstancesToAutoScalingGroups(groups, 2, 1)
	if err != nil {
		t.Error(err)
		return
	}
	if val, ok := mocked.RequestASGDesired[asgName]; !ok {
		t.Errorf("%s does not exist in request", asgName)
		return
	} else if val != int64(2) {
		t.Errorf("%s desired capacity is not matched, expected %d, but returned %d", asgName, 2, val)
	}
}
