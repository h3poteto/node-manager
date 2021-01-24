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

type TestTargetASG struct {
	asgName         *string
	regions         []*string
	instances       []*autoscaling.Instance
	maxSize         *int64
	minSize         *int64
	desiredCapacity *int64
	expectedDesired int
}

func TestInstancesToAutoScalingGroups(t *testing.T) {
	cases := []struct {
		asgs             []TestTargetASG
		specDesiredTotal int
		currentNodeCount int
	}{
		// Single ASG, and increment 1 node
		{
			asgs: []TestTargetASG{
				{
					asgName: aws.String("nodes-ap-northeast-1a"),
					regions: []*string{
						aws.String("ap-northeast-1a"),
					},
					instances: []*autoscaling.Instance{
						{
							AvailabilityZone: aws.String("ap-northeast-1a"),
							InstanceId:       aws.String("test1"),
							InstanceType:     aws.String("t3.medium"),
						},
					},
					maxSize:         aws.Int64(2),
					minSize:         aws.Int64(0),
					desiredCapacity: aws.Int64(1),
					expectedDesired: 2,
				},
			},
			specDesiredTotal: 2,
			currentNodeCount: 1,
		},
	}

	for _, c := range cases {
		var asgs []*autoscaling.Group
		var groups []operatorv1alpha1.AutoScalingGroup
		for _, ca := range c.asgs {
			asg := &autoscaling.Group{
				AutoScalingGroupName: ca.asgName,
				AvailabilityZones:    ca.regions,
				DesiredCapacity:      ca.desiredCapacity,
				Instances:            ca.instances,
				MaxSize:              ca.maxSize,
				MinSize:              ca.minSize,
			}
			asgs = append(asgs, asg)
			groups = append(groups, operatorv1alpha1.AutoScalingGroup{
				Name: *ca.asgName,
			})
		}
		resp := autoscaling.DescribeAutoScalingGroupsOutput{
			AutoScalingGroups: asgs,
			NextToken:         nil,
		}
		mocked := &mockedAutoScalingAPI{
			Resp: resp,
		}
		a := &AWS{
			autoscaling: mocked,
		}

		err := a.AddInstancesToAutoScalingGroups(groups, c.specDesiredTotal, c.specDesiredTotal-c.currentNodeCount)
		if err != nil {
			t.Error(err)
			return
		}

		for _, ca := range c.asgs {
			if val, ok := mocked.RequestASGDesired[*ca.asgName]; !ok {
				t.Errorf("%s does not exist in request", *ca.asgName)
				return
			} else if val != int64(ca.expectedDesired) {
				t.Errorf("%s desired capacity is not matched, expected %d, but returned %d", *ca.asgName, ca.expectedDesired, val)
			}
		}

	}
}
