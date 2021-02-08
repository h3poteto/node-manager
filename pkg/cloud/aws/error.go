package aws

import "fmt"

type InstanceNotYetJoinError struct {
	Msg string
}

func NewInstanceNotYetJoinErrorf(format string, a ...interface{}) *InstanceNotYetJoinError {
	return &InstanceNotYetJoinError{
		Msg: fmt.Sprintf(format, a...),
	}
}

func (err *InstanceNotYetJoinError) Error() string {
	return fmt.Sprintf(err.Msg)
}

func (err *InstanceNotYetJoinError) Is(target error) bool {
	return err.Error() == target.Error()
}
