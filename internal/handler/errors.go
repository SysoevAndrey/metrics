package handlers

import "errors"

var (
	ErrUnknownType = errors.New("unknown metrics type")
	ErrNotFound    = errors.New("metrics not found")
	ErrParseValue  = errors.New("unable to parse metrics value")
)
