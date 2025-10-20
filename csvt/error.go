package csvt

import (
	"errors"
	"fmt"
)

func IsMissingField(err error) *ErrorMissingField {
	var e *ErrorMissingField
	if errors.As(err, &e) {
		return e
	}
	return nil
}

type ErrorMissingField struct {
	Field string
}

func MissingField(field string) *ErrorMissingField {
	return &ErrorMissingField{
		Field: field,
	}
}

func (e *ErrorMissingField) Error() string {
	return fmt.Sprintf("missing required field: %s", e.Field)
}

func IsTypeMismatch(err error) *ErrorTypeMismatch {
	var e *ErrorTypeMismatch
	if errors.As(err, &e) {
		return e
	}
	return nil
}

func TypeMismatch(expected, found any, element string) *ErrorTypeMismatch {
	return &ErrorTypeMismatch{
		Element:  element,
		Expected: expected,
		Found:    found,
	}
}

func TypeMismatchf(expected, found any, format string, args ...any) *ErrorTypeMismatch {
	return TypeMismatch(expected, found, fmt.Sprintf(format, args...))
}

type ErrorTypeMismatch struct {
	Element  string
	Expected any
	Found    any
}

func (e *ErrorTypeMismatch) Error() string {
	return fmt.Sprintf("\"%s\" must be \"%v\", but \"%v\" found", e.Element, e.Expected, e.Found)
}
