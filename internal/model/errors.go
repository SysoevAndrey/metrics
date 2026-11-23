package models

import "errors"

var (
	ErrNotFound     = errors.New("metrics not found")
	ErrMissingValue = errors.New("no value found")
	ErrUnknownType  = errors.New("unknown metrics type")
)
