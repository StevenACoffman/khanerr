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
	return errors.Internal(
		ErrMyError{Msg: "Something went wrong"},
		errors.Fields{"internal": "inside"},
	)
}

func bar() error {
	withErr := errors.Unauthorized(foo(), errors.Fields{"bar": true})
	return errors.NotFound(withErr, errors.Fields{"found": false})
}

func main() {
	fmt.Printf("%+v\n", errors.Internal(errors.Fields{"bar": true}))
	// myErr := bar()
	// fmt.Println("Doing something")
	// err := errors.TransientKhanService(myErr)
	// fmt.Println("----")
	// fmt.Printf("%+v\n", err)
}
