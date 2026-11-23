package handlers

import "errors"

var (
	ErrUnknownType  = errors.New("unknown metrics type")
	ErrParseValue   = errors.New("unable to parse metrics value")
	ErrMissingValue = errors.New("no value found")
)
