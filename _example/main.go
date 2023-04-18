package main

import (
	"fmt"

	"github.com/StevenACoffman/khanerr/errors"
)

// ErrMyError is an error that can be returned from a public API.
type ErrMyError struct {
	Msg string
}

func (e ErrMyError) Error() string {
	return e.Msg
}

func foo() error {
	// Attach stack trace to the sentinel error.
	return errors.Internal("root",
		ErrMyError{Msg: "Something went wrong"},
	)
}

func bar() error {
	root := foo()
	return errors.Wrap(root, "bar", true)
}

func baz() error {
	e := bar()
	return errors.Wrap(e, "kind", errors.InternalKind)
}

func main() {
	fmt.Printf("%+v\n", baz())
}
