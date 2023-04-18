package errors

// errorKind is an error category like an exception class in Python. It's
// used to differentiate between different types of errors that a function
// can return when handling an error. It also is used when analyzing logs
// so that we can see what categories of errors we are experiencing the
// most. We would like to have a relatively small number of core error
// kinds defined, say 10-20 max. They should be general enough to be useful
// in many different contexts.
//
// Right now there isn't support for packages to define their own error kinds,
// since the function to create an error is private. In the future we should
// consider whether to allow this.
type errorKind string

// Error is a function that makes errorKind implement the error interface. This
// let's use use error.Is with kinds. We don't actually use the output of this
// function for anything.
func (e errorKind) Error() string {
	return string(e)
}

// NotFound creates an error of kind NotFoundKind.  args can be
// (1) an error to wrap
// (2) a string to use as the error message
// (3) an errors.Fields{} object of key/value pairs to associate with the error
// (4) an errors.Source("source-location") to override the default source-loc
// If you specify any of these multiple times, only the last one wins.
func NotFound(args ...any) error {
	return newError(NotFoundKind, args...)
}

// InvalidInput creates an error of kind InvalidKind.
func InvalidInput(args ...any) error {
	return newError(InvalidInputKind, args...)
}

// NotAllowed creates an error of kind NotAllowedKind.
func NotAllowed(args ...any) error {
	return newError(NotAllowedKind, args...)
}

// Unauthorized creates an error of kind UnauthorizedKind.
func Unauthorized(args ...any) error {
	return newError(UnauthorizedKind, args...)
}

// Internal creates an error of kind InternalKind.
func Internal(args ...any) error {
	return newError(InternalKind, args...)
}

// GraphqlResponse creates an error of kind GraphqlResponseKind.
func GraphqlResponse(args ...any) error {
	return newError(GraphqlResponseKind, args...)
}

// NotImplemented creates an error of kind NotImplementedKind.
func NotImplemented(args ...any) error {
	return newError(NotImplementedKind, args...)
}

// TransientKhanService creates an error of kind TransientKhanServiceKind.
func TransientKhanService(args ...any) error {
	return newError(TransientKhanServiceKind, args...)
}

// KhanService creates an error of kind KhanServiceKind.
func KhanService(args ...any) error {
	return newError(KhanServiceKind, args...)
}

// Service creates an error of kind ServiceKind.
func Service(args ...any) error {
	return newError(ServiceKind, args...)
}

// TransientService creates an error of kind TransientServiceKind.
func TransientService(args ...any) error {
	return newError(TransientServiceKind, args...)
}

// sync-start:error-kinds 1222935478 services/static/javascript/logging/internal/types.js
const (
	// NotFoundKind means that some requested resource wasn't found. If the
	// resource couldn't be retrieved due to access control use
	// UnauthorizedKind instead. If the resource couldn't be found because
	// the input was invalid use InvalidInputKind instead.
	NotFoundKind errorKind = "not found"

	// InvalidInputKind means that there was a problem with the provided input.
	// This kind indicates inputs that are problematic regardless of the state
	// of the system. Use NotAllowedKind when the input is valid but
	// conflicts with the state of the system.
	InvalidInputKind errorKind = "invalid input error"

	// NotAllowedKind means that there was a problem due to the state of
	// the system not matching the requested operation or input. For
	// example, trying to create a username that is valid, but is already
	// taken by another user. Use InvalidInputKind when the input isn't
	// valid regardless of the state of the system. Use NotFoundKind when
	// the failure is due to not being able to find a resource.
	NotAllowedKind errorKind = "not allowed"

	// UnauthorizedKind means that there was an access control problem.
	UnauthorizedKind errorKind = "unauthorized error"

	// InternalKind means that the function failed for a reason unrelated
	// to its input or problems working with a remote system. Use this kind
	// when other error kinds aren't appropriate.
	InternalKind errorKind = "internal error"

	// NotImplementedKind means that the function isn't implemented.
	NotImplementedKind errorKind = "not implemented error"

	// GraphqlResponseKind means that the graphql server returned an
	// error code as part of the graphql response.  This kind of error
	// is only ever returned by gqlclient calls.  It is set when the
	// graphql call successfully executes, but the graphql response struct
	// indicates the graphql request could not be executed due to an
	// error.  (e.g. mutation.MyMutation.Error.Code == "UNAUTHORIZED")
	GraphqlResponseKind errorKind = "graphql error response"

	// TransientKhanServiceKind means that there was a problem when contacting
	// another Khan service that might be resolvable by retrying.
	TransientKhanServiceKind errorKind = "transient khan service error"

	// KhanServiceKind means that there was a non-transient problem when
	// contacting another Khan service.
	KhanServiceKind errorKind = "khan service error"

	// TransientServiceKind means that there was a problem when making a
	// request to a non-Khan service, e.g. datastore that might be
	// resolvable by retrying.
	TransientServiceKind errorKind = "transient service error"

	// ServiceKind means that there was a non-transient problem when making a
	// request to a non-Khan service, e.g. datastore.
	ServiceKind errorKind = "service error"

	// UnspecifiedKind means that no error kind was specified. Note that there
	// isn't a constructor for this kind of error.
	UnspecifiedKind errorKind = "unspecified error"
)

// String presents the value of the string, like "Not Found"
// The fmt package (and many others) look for this to print values.
func (e errorKind) String() string {
	return string(e)
}

func (e errorKind) IsValidKind() bool {
	switch e {
	case GraphqlResponseKind,
		InternalKind,
		InvalidInputKind,
		KhanServiceKind,
		NotAllowedKind,
		NotFoundKind,
		NotImplementedKind,
		ServiceKind,
		TransientKhanServiceKind,
		TransientServiceKind,
		UnauthorizedKind,
		UnspecifiedKind:
		return true
	default:
		return false
	}
}

// GetKind returns the non-exported type, which can be annoying to use
// However, in tests, it can be handy.
func GetKind(err error) errorKind {
	var khanerr *khanError
	if As(err, &khanerr) {
		return getKind(khanerr)
	}
	var kind errorKind
	if As(err, &kind) {
		if kind.IsValidKind() {
			return kind
		}
	}
	return UnspecifiedKind
}

// kind returns the error's kind if defined. Otherwise it searches wrapped
// errors for a kind. If no kind is found it returns UnspecifiedKind.
func getKind(e *khanError) errorKind {
	if e == nil {
		return UnspecifiedKind
	}
	if e.kind.IsValidKind() {
		return e.kind
	}
	var khanErr *khanError
	if As(e.wrappedErr, &khanErr) {
		return getKind(khanErr)
	}

	return UnspecifiedKind
}
