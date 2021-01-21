package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type AWS struct {
	ec2         *ec2.EC2
	autoscaling *autoscaling.AutoScaling
}

func New(sess *session.Session, region string) *AWS {
	e := ec2.New(sess, aws.NewConfig().WithRegion(region))
	asg := autoscaling.New(sess, aws.NewConfig().WithRegion(region))
	return &AWS{
		autoscaling: asg,
		ec2:         e,
	}
}
