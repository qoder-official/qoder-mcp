package gitlabtest

import (
	"errors"
)

// ErrNumericIDRequired indicates that the fake implementation only supports numeric group IDs.
var ErrNumericIDRequired = errors.New("the fake implementation only supports numeric group IDs")

// ErrNotFound indicates that the requested resource was not found.
var ErrNotFound = errors.New("not found")
