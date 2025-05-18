package cli

import (
	"errors"
	"strings"
)

const (
	errorModeDefault   errorMode = `default`
	errorModeNoContent errorMode = `no-content`
	errorModeOnContent errorMode = `on-content`
)

type (
	stringSliceFlag []string

	errorMode string
)

func (s *stringSliceFlag) String() string {
	return strings.Join(*s, ", ")
}

func (s *stringSliceFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func (x *errorMode) String() string {
	if x.Valid() {
		return string(*x)
	}
	return "invalid (" + string(*x) + ")"
}

func (x *errorMode) Set(value string) error {
	if !(*errorMode)(&value).Valid() {
		return errors.New("invalid error mode")
	}
	*x = errorMode(value)
	return nil
}

func (x *errorMode) Valid() bool {
	switch *x {
	case errorModeDefault, errorModeNoContent, errorModeOnContent:
		return true
	}
	return false
}
