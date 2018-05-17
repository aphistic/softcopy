package ftp

import (
	"fmt"
)

type Response struct {
	Code    int
	Message string
}

func NewResponse(code int, msg string) *Response {
	return &Response{
		Code:    code,
		Message: msg,
	}
}
func (r *Response) MarshalText() ([]byte, error) {
	msg := fmt.Sprintf("%d %s", r.Code, r.Message)
	return []byte(msg), nil
}

func newErrorResponse(err error, code int, msg string) *Response {
	if ftpErr, ok := err.(*FTPError); ok {
		return NewResponse(ftpErr.Code, ftpErr.Message)
	}

	return NewResponse(code, msg)
}
