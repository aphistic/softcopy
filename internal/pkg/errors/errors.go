package errors

import "fmt"

var (
	ErrExists      = fmt.Errorf("exists")
	ErrNotFound    = fmt.Errorf("not found")
	ErrAlreadyOpen = fmt.Errorf("already open")

	ErrInvalidModeAction = fmt.Errorf("invalid mode action")
	ErrNotPermitted      = fmt.Errorf("not permitted")
)
