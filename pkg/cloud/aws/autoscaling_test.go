package aws

import (
	"errors"
	"log"
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
	if len(m.RequestASGDesired) > 0 {
		m.RequestASGDesired[*in.AutoScalingGroupName] = *in.DesiredCapacity
	} else {
		m.RequestASGDesired = map[string]int64{
			*in.AutoScalingGroupName: *in.DesiredCapacity,
		}
	}
	log.Printf("%#v", m.RequestASGDesired)
	return &autoscaling.UpdateAutoScalingGroupOutput{}, nil
}

type TestTargetASG struct {
	asgName         *string
	regions         []*string
	instances       []*autoscaling.Instance
	maxSize         *int64
	minSize         *int64
	desiredCapacity *int64
	expectedDesired *int
}

func TestAddInstancesToAutoScalingGroups(t *testing.T) {
	cases := []struct {
		asgs             []TestTargetASG
		specDesiredTotal int
		currentNodeCount int
		title            string
		expectedError    error
	}{
		// Single ASG, and increment 1 node
		{
			title: "Single ASG, and increment 1 node",
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
					expectedDesired: aws.Int(2),
				},
			},
			specDesiredTotal: 2,
			currentNodeCount: 1,
			expectedError:    nil,
		},
		// Single ASG, and increase multiple nodes
		{
			title: "Single ASG, and increase multiple nodes",
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
					maxSize:         aws.Int64(4),
					minSize:         aws.Int64(0),
					desiredCapacity: aws.Int64(1),
					expectedDesired: aws.Int(4),
				},
			},
			specDesiredTotal: 4,
			currentNodeCount: 1,
			expectedError:    nil,
		},
		// Multiple ASGs, and increase multiple nodes
		{
			title: "Multiple ASGs, and increase multiple nodes",
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
						{
							AvailabilityZone: aws.String("ap-northeast-1a"),
							InstanceId:       aws.String("test2"),
							InstanceType:     aws.String("t3.medium"),
						},
					},
					maxSize:         aws.Int64(5),
					minSize:         aws.Int64(0),
					desiredCapacity: aws.Int64(2),
					expectedDesired: aws.Int(4),
				},
				{
					asgName: aws.String("nodes-ap-northeast-1c"),
					regions: []*string{
						aws.String("ap-northeast-1c"),
					},
					instances:       []*autoscaling.Instance{},
					maxSize:         aws.Int64(5),
					minSize:         aws.Int64(0),
					desiredCapacity: aws.Int64(0),
					expectedDesired: aws.Int(3),
				},
			},
			specDesiredTotal: 7,
			currentNodeCount: 2,
			expectedError:    nil,
		},
		// Multiple ASGs which include fullfilled ASGs, and increase multiple nodes
		{
			title: "Multiple ASGs which include fullfilled ASGs, and increase multiple nodes",
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
						{
							AvailabilityZone: aws.String("ap-northeast-1a"),
							InstanceId:       aws.String("test2"),
							InstanceType:     aws.String("t3.medium"),
						},
					},
					maxSize:         aws.Int64(2),
					minSize:         aws.Int64(0),
					desiredCapacity: aws.Int64(2),
					expectedDesired: aws.Int(2),
				},
				{
					asgName: aws.String("nodes-ap-northeast-1c"),
					regions: []*string{
						aws.String("ap-northeast-1c"),
					},
					instances: []*autoscaling.Instance{
						{
							AvailabilityZone: aws.String("ap-northeast-1c"),
							InstanceId:       aws.String("test3"),
							InstanceType:     aws.String("t3.medium"),
						},
					},
					maxSize:         aws.Int64(2),
					minSize:         aws.Int64(0),
					desiredCapacity: aws.Int64(1),
					expectedDesired: aws.Int(2),
				},
				{
					asgName: aws.String("nodes-ap-northeast-1d"),
					regions: []*string{
						aws.String("ap-northeast-1d"),
					},
					instances:       []*autoscaling.Instance{},
					maxSize:         aws.Int64(2),
					minSize:         aws.Int64(0),
					desiredCapacity: aws.Int64(0),
					expectedDesired: aws.Int(1),
				},
			},
			specDesiredTotal: 5,
			currentNodeCount: 3,
			expectedError:    nil,
		},
		// Multiple ASGs which are already fullfilled, and increase multiple nodes
		{
			title: "Multiple ASGs which are already fullfilled, and increase multiple nodes",
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
					maxSize:         aws.Int64(1),
					minSize:         aws.Int64(0),
					desiredCapacity: aws.Int64(1),
					expectedDesired: aws.Int(1),
				},
				{
					asgName: aws.String("nodes-ap-northeast-1c"),
					regions: []*string{
						aws.String("ap-northeast-1c"),
					},
					instances: []*autoscaling.Instance{
						{
							AvailabilityZone: aws.String("ap-northeast-1c"),
							InstanceId:       aws.String("test2"),
							InstanceType:     aws.String("t3.medium"),
						},
					},
					maxSize:         aws.Int64(1),
					minSize:         aws.Int64(0),
					desiredCapacity: aws.Int64(1),
					expectedDesired: aws.Int(1),
				},
				{
					asgName: aws.String("nodes-ap-northeast-1d"),
					regions: []*string{
						aws.String("ap-northeast-1d"),
					},
					instances: []*autoscaling.Instance{
						{
							AvailabilityZone: aws.String("ap-northeast-1d"),
							InstanceId:       aws.String("test3"),
							InstanceType:     aws.String("t3.medium"),
						},
					},
					maxSize:         aws.Int64(1),
					minSize:         aws.Int64(0),
					desiredCapacity: aws.Int64(1),
					expectedDesired: aws.Int(1),
				},
			},
			specDesiredTotal: 5,
			currentNodeCount: 3,
			expectedError:    nil,
		},
		// Multiple ASGs which include updating ASG, and increase multiple nodes
		{
			title: "Multiple ASGs which include updating ASG, and increase multiple nodes",
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
					expectedDesired: aws.Int(1),
				},
				{
					asgName: aws.String("nodes-ap-northeast-1c"),
					regions: []*string{
						aws.String("ap-northeast-1c"),
					},
					instances: []*autoscaling.Instance{
						{
							AvailabilityZone: aws.String("ap-northeast-1c"),
							InstanceId:       aws.String("test2"),
							InstanceType:     aws.String("t3.medium"),
						},
					},
					maxSize:         aws.Int64(2),
					minSize:         aws.Int64(0),
					desiredCapacity: aws.Int64(1),
					expectedDesired: aws.Int(1),
				},
				{
					asgName: aws.String("nodes-ap-northeast-1d"),
					regions: []*string{
						aws.String("ap-northeast-1d"),
					},
					instances: []*autoscaling.Instance{
						{
							AvailabilityZone: aws.String("ap-northeast-1d"),
							InstanceId:       aws.String("test3"),
							InstanceType:     aws.String("t3.medium"),
						},
					},
					maxSize:         aws.Int64(2),
					minSize:         aws.Int64(0),
					desiredCapacity: aws.Int64(1),
					expectedDesired: aws.Int(1),
				},
			},
			specDesiredTotal: 5,
			currentNodeCount: 2,
			expectedError:    NewInstanceNotYetJoinErrorf("not all instances join the cluster as nodes, all instances: %d, current nodes: %d", 3, 2),
		},
		// Multiple ASGs which include invalid launch template ASGs, and increase multiple nodes
		{
			title: "Multiple ASGs which include invalid launch template ASGs, and increase multiple nodes",
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
					expectedDesired: aws.Int(2),
				},
				{
					asgName: aws.String("nodes-ap-northeast-1c"),
					regions: []*string{
						aws.String("ap-northeast-1c"),
					},
					instances:       []*autoscaling.Instance{},
					maxSize:         aws.Int64(2),
					minSize:         aws.Int64(0),
					desiredCapacity: aws.Int64(1),
					expectedDesired: nil,
				},
				{
					asgName: aws.String("nodes-ap-northeast-1d"),
					regions: []*string{
						aws.String("ap-northeast-1d"),
					},
					instances: []*autoscaling.Instance{
						{
							AvailabilityZone: aws.String("ap-northeast-1d"),
							InstanceId:       aws.String("test3"),
							InstanceType:     aws.String("t3.medium"),
						},
					},
					maxSize:         aws.Int64(2),
					minSize:         aws.Int64(0),
					desiredCapacity: aws.Int64(1),
					expectedDesired: aws.Int(2),
				},
			},
			specDesiredTotal: 4,
			currentNodeCount: 2,
			expectedError:    nil,
		},
	}

CASE:
	for _, c := range cases {
		log.Printf("Running CASE: %s", c.title)
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
			Autoscaling: mocked,
		}

		err := a.AddInstancesToAutoScalingGroups(groups, c.specDesiredTotal, c.currentNodeCount)
		if c.expectedError != nil {
			if err != nil && errors.Is(err, c.expectedError) {
				continue CASE
			} else {
				t.Errorf("CASE: %s : error %v is not expectedError %v", c.title, err, c.expectedError)
				continue CASE
			}
		}
		if err != nil {
			t.Errorf("CASE: %s : %v", c.title, err)
			continue CASE
		}

		for _, ca := range c.asgs {
			if ca.expectedDesired != nil {
				if val, ok := mocked.RequestASGDesired[*ca.asgName]; !ok {
					t.Errorf("CASE: %s : %s does not exist in request", c.title, *ca.asgName)
					continue CASE
				} else if int(val) != *ca.expectedDesired {
					t.Errorf("CASE: %s : %s desired capacity is not matched, expected %d, but returned %d", c.title, *ca.asgName, *ca.expectedDesired, val)
					continue CASE
				}
			} else {
				if _, ok := mocked.RequestASGDesired[*ca.asgName]; ok {
					t.Errorf("CASE: %s : %s should not exist in request, but exists", c.title, *ca.asgName)
					continue CASE
				}
			}
		}
	}
}

func TestDeleteInstancesToAutoScalingGroups(t *testing.T) {
	cases := []struct {
		title            string
		asgs             []TestTargetASG
		specDesiredTotal int
		currentNodeCount int
		expectedError    error
	}{
		// Single ASG, and decrement 1 node
		{
			title: "Single ASG, and decrement 1 node",
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
						{
							AvailabilityZone: aws.String("ap-northeast-1a"),
							InstanceId:       aws.String("test2"),
							InstanceType:     aws.String("t3.medium"),
						},
					},
					maxSize:         aws.Int64(2),
					minSize:         aws.Int64(1),
					desiredCapacity: aws.Int64(2),
					expectedDesired: aws.Int(1),
				},
			},
			specDesiredTotal: 1,
			currentNodeCount: 2,
			expectedError:    nil,
		},
		// Single ASG, and decrease multiple nodes
		{
			title: "Single ASG, and decrement multiple nodes",
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
						{
							AvailabilityZone: aws.String("ap-northeast-1a"),
							InstanceId:       aws.String("test2"),
							InstanceType:     aws.String("t3.medium"),
						},
						{
							AvailabilityZone: aws.String("ap-northeast-1a"),
							InstanceId:       aws.String("test3"),
							InstanceType:     aws.String("t3.medium"),
						},
					},
					maxSize:         aws.Int64(4),
					minSize:         aws.Int64(0),
					desiredCapacity: aws.Int64(3),
					expectedDesired: aws.Int(1),
				},
			},
			specDesiredTotal: 1,
			currentNodeCount: 3,
			expectedError:    nil,
		},
		// Multiple ASGs, and decrease multiple nodes
		{
			title: "Multiple ASGs, and decrease multiple nodes",
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
						{
							AvailabilityZone: aws.String("ap-northeast-1a"),
							InstanceId:       aws.String("test2"),
							InstanceType:     aws.String("t3.medium"),
						},
						{
							AvailabilityZone: aws.String("ap-northeast-1a"),
							InstanceId:       aws.String("test3"),
							InstanceType:     aws.String("t3.medium"),
						},
					},
					maxSize:         aws.Int64(4),
					minSize:         aws.Int64(1),
					desiredCapacity: aws.Int64(3),
					expectedDesired: aws.Int(1),
				},
				{
					asgName: aws.String("nodes-ap-northeast-1c"),
					regions: []*string{
						aws.String("ap-northeast-1c"),
					},
					instances: []*autoscaling.Instance{
						{
							AvailabilityZone: aws.String("ap-northeast-1c"),
							InstanceId:       aws.String("test4"),
							InstanceType:     aws.String("t3.medium"),
						},
						{
							AvailabilityZone: aws.String("ap-northeast-1c"),
							InstanceId:       aws.String("test5"),
							InstanceType:     aws.String("t3.medium"),
						},
					},
					maxSize:         aws.Int64(4),
					minSize:         aws.Int64(0),
					desiredCapacity: aws.Int64(2),
					expectedDesired: aws.Int(1),
				},
			},
			specDesiredTotal: 2,
			currentNodeCount: 5,
			expectedError:    nil,
		},
		// Multiple ASGs which include minimized ASGs, and decrease multiple nodes
		{
			title: "Multiple ASGs which include minimized ASGs, and decrease multiple nodes",
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
						{
							AvailabilityZone: aws.String("ap-northeast-1a"),
							InstanceId:       aws.String("test2"),
							InstanceType:     aws.String("t3.medium"),
						},
					},
					maxSize:         aws.Int64(4),
					minSize:         aws.Int64(0),
					desiredCapacity: aws.Int64(2),
					expectedDesired: aws.Int(1),
				},
				{
					asgName: aws.String("nodes-ap-northeast-1c"),
					regions: []*string{
						aws.String("ap-northeast-1c"),
					},
					instances: []*autoscaling.Instance{
						{
							AvailabilityZone: aws.String("ap-northeast-1c"),
							InstanceId:       aws.String("test3"),
							InstanceType:     aws.String("t3.medium"),
						},
					},
					maxSize:         aws.Int64(2),
					minSize:         aws.Int64(1),
					desiredCapacity: aws.Int64(1),
					expectedDesired: aws.Int(1),
				},
				{
					asgName: aws.String("nodes-ap-northeast-1d"),
					regions: []*string{
						aws.String("ap-northeast-1d"),
					},
					instances: []*autoscaling.Instance{
						{
							AvailabilityZone: aws.String("ap-northeast-1d"),
							InstanceId:       aws.String("test4"),
							InstanceType:     aws.String("t3.medium"),
						},
					},
					maxSize:         aws.Int64(2),
					minSize:         aws.Int64(0),
					desiredCapacity: aws.Int64(1),
					expectedDesired: aws.Int(0),
				},
			},
			specDesiredTotal: 2,
			currentNodeCount: 4,
			expectedError:    nil,
		},
		// Multiple ASGs which are already minimized, and decrease multiple nodes
		{
			title: "Multpile ASGs which are already minimized, and decrease multiple nodes",
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
					maxSize:         aws.Int64(1),
					minSize:         aws.Int64(1),
					desiredCapacity: aws.Int64(1),
					expectedDesired: aws.Int(1),
				},
				{
					asgName: aws.String("nodes-ap-northeast-1c"),
					regions: []*string{
						aws.String("ap-northeast-1c"),
					},
					instances: []*autoscaling.Instance{
						{
							AvailabilityZone: aws.String("ap-northeast-1c"),
							InstanceId:       aws.String("test2"),
							InstanceType:     aws.String("t3.medium"),
						},
						{
							AvailabilityZone: aws.String("ap-northeast-1c"),
							InstanceId:       aws.String("test3"),
							InstanceType:     aws.String("t3.medium"),
						},
					},
					maxSize:         aws.Int64(2),
					minSize:         aws.Int64(2),
					desiredCapacity: aws.Int64(2),
					expectedDesired: aws.Int(2),
				},
				{
					asgName: aws.String("nodes-ap-northeast-1d"),
					regions: []*string{
						aws.String("ap-northeast-1d"),
					},
					instances:       []*autoscaling.Instance{},
					maxSize:         aws.Int64(1),
					minSize:         aws.Int64(0),
					desiredCapacity: aws.Int64(0),
					expectedDesired: aws.Int(0),
				},
			},
			specDesiredTotal: 1,
			currentNodeCount: 3,
			expectedError:    nil,
		},
		// Multiple ASGs which include updating ASG, and decrease multiple nodes
		{
			title: "Multiple ASGs which include updating ASG, and decrease multiple nodes",
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
					expectedDesired: aws.Int(1),
				},
				{
					asgName: aws.String("nodes-ap-northeast-1c"),
					regions: []*string{
						aws.String("ap-northeast-1c"),
					},
					instances: []*autoscaling.Instance{
						{
							AvailabilityZone: aws.String("ap-northeast-1c"),
							InstanceId:       aws.String("test2"),
							InstanceType:     aws.String("t3.medium"),
						},
					},
					maxSize:         aws.Int64(2),
					minSize:         aws.Int64(0),
					desiredCapacity: aws.Int64(1),
					expectedDesired: aws.Int(1),
				},
				{
					asgName: aws.String("nodes-ap-northeast-1d"),
					regions: []*string{
						aws.String("ap-northeast-1d"),
					},
					instances: []*autoscaling.Instance{
						{
							AvailabilityZone: aws.String("ap-northeast-1d"),
							InstanceId:       aws.String("test3"),
							InstanceType:     aws.String("t3.medium"),
						},
					},
					maxSize:         aws.Int64(2),
					minSize:         aws.Int64(0),
					desiredCapacity: aws.Int64(1),
					expectedDesired: aws.Int(1),
				},
			},
			specDesiredTotal: 5,
			currentNodeCount: 2,
			expectedError:    NewInstanceNotYetJoinErrorf("not all instances join the cluster as nodes, all instances: %d, current nodes: %d", 3, 2),
		},
		// Multiple ASGs which include invalid launch template ASGs, and decrease multiple nodes
		{
			title: "Multiple ASGs which include invalid launch template ASGs, and decrease multiple nodes",
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
						{
							AvailabilityZone: aws.String("ap-northeast-1d"),
							InstanceId:       aws.String("test2"),
							InstanceType:     aws.String("t3.medium"),
						},
					},
					maxSize:         aws.Int64(2),
					minSize:         aws.Int64(0),
					desiredCapacity: aws.Int64(2),
					expectedDesired: aws.Int(1),
				},
				{
					asgName: aws.String("nodes-ap-northeast-1c"),
					regions: []*string{
						aws.String("ap-northeast-1c"),
					},
					instances:       []*autoscaling.Instance{},
					maxSize:         aws.Int64(2),
					minSize:         aws.Int64(0),
					desiredCapacity: aws.Int64(1),
					expectedDesired: aws.Int(0),
				},
				{
					asgName: aws.String("nodes-ap-northeast-1d"),
					regions: []*string{
						aws.String("ap-northeast-1d"),
					},
					instances: []*autoscaling.Instance{
						{
							AvailabilityZone: aws.String("ap-northeast-1d"),
							InstanceId:       aws.String("test3"),
							InstanceType:     aws.String("t3.medium"),
						},
					},
					maxSize:         aws.Int64(2),
					minSize:         aws.Int64(0),
					desiredCapacity: aws.Int64(1),
					expectedDesired: aws.Int(0),
				},
			},
			specDesiredTotal: 1,
			currentNodeCount: 3,
			expectedError:    nil,
		},
	}

CASE:
	for _, c := range cases {
		log.Printf("Running CASE: %s", c.title)
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
			Autoscaling: mocked,
		}

		err := a.DeleteInstancesToAutoScalingGroups(groups, c.specDesiredTotal, c.currentNodeCount)
		if c.expectedError != nil {
			if err != nil && errors.Is(err, c.expectedError) {
				continue CASE
			} else {
				t.Errorf("CASE: %s : error %v is not expectedError %v", c.title, err, c.expectedError)
				continue CASE
			}
		}
		if err != nil {
			t.Errorf("CASE: %s : %v", c.title, err)
			continue CASE
		}

		for _, ca := range c.asgs {
			if ca.expectedDesired != nil {
				if val, ok := mocked.RequestASGDesired[*ca.asgName]; !ok {
					t.Errorf("CASE: %s : %s does not exist in request", c.title, *ca.asgName)
					continue CASE
				} else if int(val) != *ca.expectedDesired {
					t.Errorf("CASE: %s : %s desired capacity is not matched, expected %d, but returned %d", c.title, *ca.asgName, *ca.expectedDesired, val)
					continue CASE
				}
			} else {
				if _, ok := mocked.RequestASGDesired[*ca.asgName]; ok {
					t.Errorf("CASE: %s : %s should not exist in request, but exists", c.title, *ca.asgName)
					continue CASE
				}
			}
		}
	}
}
