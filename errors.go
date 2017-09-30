package oss_addons

type InvalidArgumentError struct {
	msg string
}

func (e *InvalidArgumentError) Error() string {
	return e.msg
}

func NewInvalidArgumentError(message string) error {
	return &InvalidArgumentError{message}
}
