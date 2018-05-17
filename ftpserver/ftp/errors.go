package ftp

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidCommand = errors.New("invalid command")
)

type FTPError struct {
	Code    int
	Message string
}

func NewFTPError(code int, message string) *FTPError {
	return &FTPError{
		Code:    code,
		Message: message,
	}
}

func (fe *FTPError) Error() string {
	msg := fmt.Sprintf("%d", fe.Code)
	if fe.Message != "" {
		msg += fmt.Sprintf(" %s", fe.Message)
	}
	return msg
}
