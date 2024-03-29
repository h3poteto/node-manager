package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type AWS struct {
	EC2         ec2iface.EC2API
	Autoscaling autoscalingiface.AutoScalingAPI
}

func New(sess *session.Session, region string) *AWS {
	e := ec2.New(sess, aws.NewConfig().WithRegion(region))
	asg := autoscaling.New(sess, aws.NewConfig().WithRegion(region))
	return &AWS{
		Autoscaling: asg,
		EC2:         e,
	}
}
