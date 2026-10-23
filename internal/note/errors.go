package note

import "errors"

var (
	ErrNotFound   = errors.New("note not found")
	ErrForbidden  = errors.New("access to note is forbidden")
	ErrEmptyTitle = errors.New("title must not be empty")
)
