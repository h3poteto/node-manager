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

type DesiredInvalidError struct {
	Msg string
}

func NewDesiredInvalidErrorf(format string, a ...interface{}) *DesiredInvalidError {
	return &DesiredInvalidError{
		Msg: fmt.Sprintf(format, a...),
	}
}

func (err *DesiredInvalidError) Error() string {
	return fmt.Sprintf(err.Msg)
}

func (err *DesiredInvalidError) Is(target error) bool {
	return err.Error() == target.Error()
}

type CouldNotFoundNameTagError struct {
	Msg string
}

func NewCouldNotFoundNameTagError(format string, a ...interface{}) *CouldNotFoundNameTagError {
	return &CouldNotFoundNameTagError{
		Msg: fmt.Sprintf(format, a...),
	}
}

func (err *CouldNotFoundNameTagError) Error() string {
	return fmt.Sprintf(err.Msg)
}

func (err *CouldNotFoundNameTagError) Is(target error) bool {
	return err.Error() == target.Error()
}
