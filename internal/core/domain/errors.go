package domain

import "errors"

var (
	ErrPermissionDenied  = errors.New("permission denied")
	ErrNotFound          = errors.New("port not found")
	ErrOwnershipConflict = errors.New("process ownership changed between check and action")
	ErrCriticalProcess   = errors.New("target is a critical system process")
	ErrProcessExited     = errors.New("target process has already exited")
	ErrUnsupported       = errors.New("operation not supported on this platform")
	ErrTimeout           = errors.New("operation timed out")
	ErrNoFreePort        = errors.New("no free port available in the specified range")
	ErrInvalidPort       = errors.New("invalid port number")
	ErrInvalidRange      = errors.New("invalid port range")
)

// PartialResult wraps a result that may be incomplete due to permission restrictions.
type PartialResult[T any] struct {
	Data        T
	Partial     bool
	DeniedCount int
	Warnings    []string
}
