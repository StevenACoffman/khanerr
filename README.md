# khanerr - gRPC status code Sentinel Errors with Fields and StackTraces

Package errors is a reimplementation of Khan's error package, but here, it is based on
the minimalist `github.com/StevenACoffman/simplerr`.

`khanerr` provides errors with key value Fields, StackTraces, and Sentinel errors
that correspond to the gRPC status codes.

You can also use this package as a drop in replacement for the `Unwrap`, `As`, and `Is` functions from the errors packages.

This package is influenced
by https://middlemost.com/failure-is-your-domain/.  

### Sentinel Error values

Error kinds are sentinel error values
influenced by gRPC status codes, documented at
https://github.com/grpc/grpc/blob/master/doc/statuscodes.md.

An error kind ("InternalKind", "InvalidKind", etc) is used for matching
errors as well as giving us information in the logs about what kind
of thing went wrong.

### New Error Creation

There are functions for each error kind (e.g. `NotFoundKind`) to create errors, e.g. `NotFound`,
`Internal`, etc. All arguments are optional:

	e1 := errors.Internal("Message")
	e2 := errors.Invalid(aWrappedErr)
	e3 := errors.NotFound("Message", errors.Fields{{"kaid": "123"})

Each of these functions takes an arbitrary number of args, in an
arbitrary order. The args can be:
1. an error object to wrap
2. a string to use as the error message
3. an errors.Fields{} object of key/value pairs to associate with the error
4. an errors.Source("source-location") to override the default source-loc

You should always provide one of (1) and (2); you can provide both
if it's helpful.  (3) is used to detail things like the name of the
item that couldn't be found, or the affected kaid, etc. Don't
embed these details in the error message - by putting them in
fields you make it much easier to search logs for them.

If you specify any one type multiple times, only the last one wins.

### --- IS / AS / ETC ---

This package exposes the `Is`, `As` and `Unwrap` functions from
the stdlib errors package.

You can use the `Is` function to test against sentinel errors and wrapped
regular errors. Then when handling the error you can test it against the
sentinel using the `Is` function:

	err := SomeFunction()
	if errors.Is(err, UnauthorizedKind)  {
	    // The error or an error it wraps has the UnauthorizedKind kind.
	}

In general you should prefer to test against
sentinel errors to know what action to
take based on error kind.

The `Unwrap` function works normally and returns wrapped errors.

The `As` function works normally, but isn't very useful with Khan errors.
Since the `errorKind` type is private you can't use `As` to search for it.
Moreover since error kinds aren't types you can't use them with `As`.
However you can use `As` to find wrapped errors of other public error
types.

### --- WRAP ---

This package also provides a utility to "wrap" an existing khanError
to add more fields:

	errors.Wrap(err, "newfield", "newvalue", "did i just wrap?", true)

This is equivalent to

	errors.<SameKindAsErr>("", err,
	    errors.Fields{"newfield": "newvalue", "did i just wrap?": true})

### GetFields
When logging, it is recommended to use GetFields(err) to collect all the Fields
of nested errors but ensure that the last key value pair wins.

For instance, if a `Field{"message":"oh no!"}` is set on an error that is wrapped inside a new
error that has `Field{"message":"nevermind"}`, then the value for `message` key is `nevermind`.