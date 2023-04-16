package errors

// this file is for compatibility with both:
// + stdlib errors package after Go 1.13
// + pkg/errors
// + golang.org/x/xerrors
//
// this allows this package to be a drop in replacement
// for all three
// EXCEPT there is no errors.New
// As Khan webapp does not allow std lib New Errors, preferring root sentinel error types

import (
	simpler "github.com/StevenACoffman/simplerr/errors"
)

// Cause aliases UnwrapAll() for compatibility with github.com/pkg/errors.
func Cause(err error) error { return simpler.UnwrapAll(err) }

// Unwrap aliases UnwrapOnce() for compatibility with xerrors.
func Unwrap(err error) error { return simpler.UnwrapOnce(err) }

// As finds the first error in err's chain that matches the type to which
// target points, and if so, sets the target to its value and returns true.
// An error matches a type if it is assignable to the target type, or if it
// has a method As(interface{}) bool such that As(target) returns true. As
// will panic if target is not a non-nil pointer to a type which implements
// error or is of interface type.
//
// The As method should set the target to its value and return true if err
// matches the type to which target points.
//
// Note: this implementation differs from that of xerrors as follows:
// - it also supports recursing through causes with Cause().
// - if it detects an API use error, its panic object is a valid error.

// As finds the first error in err's chain that matches the type to which
// target points, and if so, sets the target to its value and returns true.
// An error matches a type if it is assignable to the target type, or if it
// has a method As(interface{}) bool such that As(target) returns true. As
// will panic if target is not a non-nil pointer to a type which implements
// error or is of interface type.
//
// The As method should set the target to its value and return true if err
// matches the type to which target points.
//
// Note: this implementation differs from that of xerrors as follows:
// - it also supports recursing through causes with Cause().
// - if it detects an API use error, its panic object is a valid error.
func As(err error, target interface{}) bool {
	return simpler.As(err, target)
}

// Is determines whether one of the causes of the given error or any
// of its causes is equivalent to some reference error.
//
// As in the Go standard library, an error is considered to match a
// reference error if it is equal to that target or if it implements a
// method Is(error) bool such that Is(reference) returns true.
//
// Note: the inverse is not true - making an Is(reference) method
// return false does not imply that errors.Is() also returns
// false. Errors can be equal because their network equality marker is
// the same. To force errors to appear different to Is(), use
// errors.Mark().
//
// Note: if any of the error types has been migrated from a previous
// package location or a different type, ensure that
// RegisterTypeMigration() was called prior to Is().
// Is determines whether one of the causes of the given error or any
// of its causes is equivalent to some reference error.
//
// As in the Go standard library, an error is considered to match a
// reference error if it is equal to that target or if it implements a
// method Is(error) bool such that Is(reference) returns true.
//
// Note: the inverse is not true - making an Is(reference) method
// return false does not imply that errors.Is() also returns
// false. Errors can be equal because their network equality marker is
// the same. To force errors to appear different to Is(), use
// errors.Mark().
//
// Note: if any of the error types has been migrated from a previous
// package location or a different type, ensure that
// RegisterTypeMigration() was called prior to Is().
func Is(err, reference error) bool {
	return simpler.Is(err, reference)
}
