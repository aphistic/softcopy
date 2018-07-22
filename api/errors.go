package api

import (
	"errors"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrHashCollision = errors.New("hash collision")
)
