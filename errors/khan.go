package errors

import (
	"bytes"
	"fmt"
	"sort"

	simpler "github.com/StevenACoffman/simplerr/errors"
)

// khanError is our error implementation. `source` is a string that
// uniquely identifies the error source, such as "package.function". `kind`
// is an error category. `message` is an error message that will appear in
// the logs. `wrappedErr` is an optional wrapped error. `origin` is a
// string in the format "<filename>:<linenumber>". `extra` is an optional
// collection of key value pairs to log when logging the error.
type khanError struct {
	message    string
	kind       errorKind
	wrappedErr error
	extra      Fields
}

func (e *khanError) wrappedErrors() []Fields {
	if e == nil || e.wrappedErr == nil {
		return []Fields{}
	}
	// inner, ok := e.wrappedErr.(*khanError)
	innerFields := simpler.GetFields(e.wrappedErr)
	if len(innerFields) != 0 {
		return []Fields{Fields(innerFields)}
	}
	// unlikely to get past this but should work fine regardless
	var inner *khanError
	ok := As(e.wrappedErr, &inner)
	var wrapped []Fields
	if !ok {
		return append(wrapped, Fields{MessageKey: e.wrappedErr.Error()})
	}
	wrapped = append(wrapped, Fields{
		MessageKey: inner.message,
		KindKey:    string(getKind(inner)),
	})
	return append(wrapped, inner.wrappedErrors()...)
}

// Error returns a short error message. It constitutes the "error" interface.
// We expose all metadata (except some special empty fields) about the error
// here to ensure that when errors are sent to the requestlogs that all the
// data is captured. The error data is also exposed in a structured form in
// stackdriver logs using LogFields.
func (e *khanError) Error() string {
	if e == nil {
		return ""
	}

	var buf bytes.Buffer

	// TODO(csilvers): do we want to do something special if the
	// wrapped error is a sentinel?  Like maybe show our error text
	// above the sentinel, followed by a special "Wraps sentinel:"
	// line.  (We can tell if we're a sentinel based on `source`
	// having `"init"`.)
	// TODO(csilvers): similarly for non-khan errors: we may want to
	// show the first khan-error instead, that wraps the non-khan error.
	if e.wrappedErr != nil {
		buf.WriteString(e.wrappedErr.Error())
		buf.WriteString("\nWrapped by: ")
	}

	_, _ = fmt.Fprintf(&buf, "%s", string(getKind(e)))
	if e.message != "" {
		_, _ = fmt.Fprintf(&buf, " %s", e.message)
	}
	if e.extra != nil {
		keys := make([]string, len(e.extra))
		i := 0
		for k := range e.extra {
			keys[i] = k
			i++
		}
		sort.Strings(keys)
		for _, k := range keys {
			fieldValue := StringifyField(e.extra[k])
			// Ignore empty fields for special keys. These keys are set by the
			// graphql error handler with empty values to ensure that the
			// fields are present in the log schema and thus avoid log export
			// problems. But if they are empty we don't need to see them in the
			// message.
			if fieldValue == "" && (k == "handledGraphQLPanic" ||
				k == "panicErr.Kind" ||
				k == "panicErr.Message" ||
				k == "panicErr.Source" ||
				k == "panicValue") {
				continue
			}
			_, _ = fmt.Fprintf(&buf, ", %s = %s", k, fieldValue)
		}
	}
	return buf.String()
}

// Unwrap returns the wrapped error, if any. This function allows use of
// errors.Unwrap, errors.Is, and errors.As.
func (e *khanError) Unwrap() error {
	return e.wrappedErr
}

// Is implements the test that errors.Is uses to decide if two errors are
// "equal". errors.Is takes care of comparing the error and all wrapped
// errors using regular equality checks. So we only need to test for
// special cases here. The special case we support is matching with kinds.
// An error is "equal" to it's kind. This let's us use errors.Is find out
// which kind a khan error is.
func (e *khanError) Is(target error) bool {
	return e.kind == target
}

const (
	MessageKey        = "Message"
	KindKey           = "Kind"
	BadArgsKey        = "badargs"
	InvalidErrArgsKey = "Invalid error arguments"
)

func newError(kind errorKind, args ...any) error {
	e := &khanError{kind: kind}
	badArgs := make([]any, 0)
	for _, arg := range args {
		switch v := arg.(type) {
		case error:
			e.wrappedErr = v
		case string:
			e.message = v
		case Fields:
			e.extra = v
		case map[string]any:
			e.extra = v
		default:
			badArgs = append(badArgs, v)
		}
	}
	if len(badArgs) > 0 {
		e.message = "Invalid error constructor argument(s): " + e.message
		details := make([]string, len(badArgs))
		for i, arg := range badArgs {
			details[i] = fmt.Sprintf("%#v", arg)
		}
		if e.extra == nil {
			e.extra = Fields{}
		}
		e.extra[InvalidErrArgsKey] = details
	}

	fields := Fields{
		KindKey: string(getKind(e)),
	}
	for _, f := range e.wrappedErrors() {
		for s, a := range f {
			if _, ok := fields[s]; !ok && s != "Source" && s != "Origin" {
				fields[s] = a
			}
		}
	}
	if e.extra != nil {
		for k, v := range e.extra {
			fields[k] = v
		}
	}

	if e.message != "" {
		if len(fields) == 0 {
			fields = Fields{MessageKey: e.message}
		} else {
			fields[MessageKey] = e.message
		}
	}
	// if no other wrapped error, use kind
	if e.wrappedErr == nil || e.wrappedErr == kind {
		return simpler.WrapWithFieldsAndDepth(kind, simpler.Fields(fields), 2)
	}
	// we double wrap to ensure errors.Is true for both kind and original
	tmpErr := simpler.With(e.wrappedErr, kind)
	return simpler.WrapWithFieldsAndDepth(tmpErr, simpler.Fields(fields), 2)
}

// Fail if Wrap() has the wrong args.  All the errors here are
// programming errors, so we fail in tests (and on dev) but just note
// the error in prod.
func _fail(args ...any) error {
	return Internal(args...)
}

// Wrap takes a khanError as input and some new field key/value pairs,
// and returns a new error that has the same "kind" as the existing
// error, plus the specified key/value pairs.  For convenience, rather
// than using errors.Fields{} to specify the key/value pairs, they
// are specified as alternating string/any objects.
// Also for convenience, if nil is passed in then nil is returned.
//
// If there is an error in wrapping -- the input is not a khanError,
// a non-string key is specified -- then the wrapped error is actually
// an error.Internal() that indicates the problem with wrapping.
// .
// Wrap here is NOT github.com/pkg/errors Wrap compatible
func Wrap(err error, args ...any) error {
	if err == nil {
		return nil
	}

	if len(args)%2 != 0 {
		return _fail("Passed an odd number of field-args to errors.Wrap()",
			err, Fields{BadArgsKey: args})
	}

	fields := Fields{}
	for i := 0; i < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			return _fail("Passed a non-string key-field to errors.Wrap()",
				err, Fields{"key": args[i]})
		}
		fields[key] = args[i+1]
	}
	var khanKind errorKind
	if As(err, &khanKind) {
		if khanKind.IsValidKind() {
			return newError(khanKind, fields)
		}
	}

	// khanErr, ok := err.(*khanError)
	var khanErr *khanError
	ok := As(err, &khanErr)
	if !ok {
		// "Internal" is the best default, but not always right.
		// e.g. for client.GCS() errors, "Service" would be better.
		// The solution is to change our GCS wrapper to return khanErrors,
		// like we do for our Datastore wrapper.
		return Internal(err, fields)
	}
	errKind := getKind(khanErr)
	if errKind == UnspecifiedKind {
		// This probably can't happen, but just in case...
		return _fail("Cannot determine kind of error-to-wrap", err)
	}
	return newError(errKind, khanErr, fields)
}

//
// func (ke *khanError) MarshalLogObject(enc zapcore.ObjectEncoder) error {
//	enc.AddString("kind", string(ke.kind))
//	enc.AddString("message", ke.Error())
//	enc.AddString("stacktrace", fmt.Sprintf("%+v", ke.StackTrace()))
//	err := enc.AddReflected("fields", ke.fields)
//	if err != nil {
//		return errors.Wrapf(err, "Unable to AddReflected fields to log: %+v", ke.fields)
//	}
//	err = enc.AddReflected("cause", ke.cause)
//	if err != nil {
//		return errors.Wrapf(err, "Unable to AddReflected cause to log %+v", ke.cause)
//	}
//
//	return nil
//}

func GetFields(err error) Fields {
	return Fields(simpler.GetFields(err))
}

// IsKhanError returns true if the error is a khan error. Note we don't
// check wrapped errors - this is a check of the outer error only. This
// check isn't like errors.As which is used to get access to error details
// for a particular type of errors. Instead use this function to see if the
// outer error is a khan error so that you can tell whether you might want
// to wrap the error in a khan error before logging.
func IsKhanError(err error) bool {
	var kind errorKind
	if As(err, &kind) {
		return kind.IsValidKind()
	}
	return false
}

// StringifyField turns a field value into a string for logging.
func StringifyField(value any) string {
	switch value.(type) {
	case []string:
		return fmt.Sprintf("%q", value)
	default:
		return fmt.Sprintf("%v", value)
	}
}

// Fields is re-exported here to avoid leaking direct import implementation details
type Fields simpler.Fields
