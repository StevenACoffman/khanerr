package errors_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/StevenACoffman/khanerr/errors"
)

type errorSuite struct{ suite.Suite }

func (es *errorSuite) TestSimpleError() {
	e := errors.Internal("Testing one two")
	es.Require().Equal(
		"Fields: [Kind:internal error,Message:Testing one two], Cause: internal error", e.Error())
}

func (es *errorSuite) TestExtra() {
	e := errors.Internal("Testing",
		errors.Fields{"kaid": "123", "tags": []string{"one", "two"}, "empty": "", "panicValue": ""})
	es.Require().Equal(
		"Fields: [Kind:internal error,Message:Testing,empty:,kaid:123,panicValue:,tags:[one two]], Cause: internal error",
		e.Error())
	outer := errors.Internal(e)
	es.Require().Equal(
		"Fields: [Kind:internal error,Message:Testing,empty:,kaid:123,panicValue:,tags:[one two]], Cause: internal error: Fields: [Kind:internal error,Message:Testing,empty:,kaid:123,panicValue:,tags:[one two]], Cause: internal error",
		outer.Error())
}

func (es *errorSuite) TestEmpty() {
	e := errors.Internal()
	es.Require().Equal(
		"Fields: [Kind:internal error], Cause: internal error", e.Error())
}

func (es *errorSuite) TestWrappedError() {
	innerError := fmt.Errorf("This is not OK")
	e := errors.Internal(innerError)
	es.Require().Equal(
		"Fields: [Kind:internal error,Message:This is not OK], Cause: internal error: This is not OK",
		e.Error())
	// non-kind, non-khan errors are wrapped twice
	// inner wrapper front is kind, back is non-khan, non-kind root
	// then withstackandfields holds the khan err
	es.Require().Equal(innerError, errors.Unwrap(errors.Unwrap(e)))
}

func (es *errorSuite) TestNew() {
	e := errors.Internal("yikes", errors.Fields{"a": "b"})
	fields := errors.GetFields(e)
	es.FieldContainsValue(fields, errors.MessageKey, "yikes")

	es.Require().Equal(errors.InternalKind, errors.GetKind(e))
	es.Require().
		Equal(errors.Fields{"Kind": "internal error", "a": "b", "Message": "yikes"}, errors.GetFields(e))
}

func (es *errorSuite) TestInvalidParameter() {
	e := errors.Internal("Message", 42)
	fields := errors.GetFields(e)
	es.Require().Equal([]string{"42"}, fields["Invalid error arguments"])
	es.Require().Equal("Invalid error constructor argument(s): Message", fields["Message"])
}

func (es *errorSuite) TestKind() {
	inner := fmt.Errorf("more inner")
	es.Require().Equal(errors.UnspecifiedKind, errors.GetKind(inner))
	e := errors.Wrap(inner, errors.Fields{})
	es.Require().Equal(errors.InternalKind, errors.GetKind(e))
	e = errors.Wrap(errors.UnauthorizedKind)
	es.Require().Equal(errors.UnauthorizedKind, errors.GetKind(e))
	e = errors.Wrap(errors.InternalKind)
	es.Require().Equal(errors.InternalKind, errors.GetKind(e))
}

func (es *errorSuite) TestIs() {
	var empty error
	e := errors.Unauthorized()
	e2 := errors.Internal(e)
	e3 := fmt.Errorf("sentinal")
	e4 := errors.NotFound(e3)
	es.Require().True(errors.Is(e, e))
	es.Require().True(errors.Is(e, errors.UnauthorizedKind))
	es.Require().False(errors.Is(e, errors.InternalKind))
	es.Require().False(errors.Is(e, errors.NotFoundKind))
	es.Require().False(errors.Is(e, e3))
	es.Require().False(errors.Is(e, empty))

	es.Require().True(errors.Is(e2, e2))
	es.Require().True(errors.Is(e2, errors.InternalKind))
	es.Require().True(errors.Is(e2, errors.UnauthorizedKind))
	es.Require().False(errors.Is(e2, errors.NotFoundKind))
	es.Require().False(errors.Is(e2, e3))

	es.Require().False(errors.Is(e3, e2))
	es.Require().False(errors.Is(e3, errors.InternalKind))
	es.Require().False(errors.Is(e3, errors.UnauthorizedKind))
	es.Require().False(errors.Is(e3, errors.NotFoundKind))
	es.Require().True(errors.Is(e3, e3))

	es.Require().True(errors.Is(e4, e4))
	es.Require().False(errors.Is(e4, errors.InternalKind))
	es.Require().False(errors.Is(e4, errors.UnauthorizedKind))
	es.Require().True(errors.Is(e4, errors.NotFoundKind))
	es.Require().True(errors.Is(e4, e3))
}

func (es *errorSuite) TestConstructors() {
	es.Require().Equal(errors.NotFoundKind, errors.GetKind(errors.NotFound()))
	es.Require().Equal(errors.InvalidInputKind, errors.GetKind(errors.InvalidInput()))
	es.Require().Equal(errors.NotAllowedKind, errors.GetKind(errors.NotAllowed()))
	es.Require().Equal(errors.UnauthorizedKind, errors.GetKind(errors.Unauthorized()))
	es.Require().Equal(errors.InternalKind, errors.GetKind(errors.Internal()))
	es.Require().Equal(errors.NotImplementedKind, errors.GetKind(errors.NotImplemented()))
	es.Require().Equal(errors.GraphqlResponseKind, errors.GetKind(errors.GraphqlResponse()))
	es.Require().
		Equal(errors.TransientKhanServiceKind, errors.GetKind(errors.TransientKhanService()))
	es.Require().Equal(errors.KhanServiceKind, errors.GetKind(errors.KhanService()))
	es.Require().Equal(errors.ServiceKind, errors.GetKind(errors.Service()))
	es.Require().Equal(errors.TransientServiceKind, errors.GetKind(errors.TransientService()))
}

type _error struct{}

func (e *_error) Error() string {
	return ""
}

func (es *errorSuite) TestAs() {
	inner := &_error{}
	var err *_error
	e := errors.Internal()
	es.Require().False(errors.As(e, &err))

	e2 := errors.Internal(inner)
	es.Require().True(errors.As(e2, &err))
	es.Require().Equal(inner, err)
}

func (es *errorSuite) TestWrap() {
	e := errors.NotFound("Testing one two", errors.Fields{"three": "not yet"})

	e2 := errors.Wrap(e, "three", 4)
	es.Require().Equal(
		"Fields: [Kind:not found,Message:Testing one two,three:4], Cause: not found: Fields: [Kind:not found,Message:Testing one two,three:not yet], Cause: not found",
		e2.Error())

	e2 = errors.Wrap(fmt.Errorf("sentinal"), "three", 4)
	es.Require().Equal(
		"Fields: [Kind:internal error,Message:sentinal,three:4], Cause: internal error: sentinal",
		e2.Error())

	e2 = errors.Wrap(e, "three", 4, "five")
	es.Require().Equal(
		"Fields: [Kind:internal error,Message:Passed an odd number of field-args to errors.Wrap(),badargs:[three 4 five],three:not yet], Cause: internal error: Fields: [Kind:not found,Message:Testing one two,three:not yet], Cause: not found",
		e2.Error())

	e2 = errors.Wrap(e, "three", 4, 5, 6)
	es.Require().Equal(
		"Fields: [Kind:internal error,Message:Passed a non-string key-field to errors.Wrap(),key:5,three:not yet], Cause: internal error: Fields: [Kind:not found,Message:Testing one two,three:not yet], Cause: not found",
		e2.Error())

	e2 = errors.Wrap(nil, "three", 4)
	es.Require().Equal(nil, e2)
}

func (es *errorSuite) TestWrapInDev() {
	e := errors.NotFound("Testing one two", errors.Fields{"three": "not yet"})
	es.Require().Equal("not yet", errors.GetFields(e)["three"])
	e2 := errors.Wrap(e, "three", 4)
	es.Require().Equal(
		"Fields: [Kind:not found,Message:Testing one two,three:4], Cause: not found: Fields: [Kind:not found,Message:Testing one two,three:not yet], Cause: not found",
		e2.Error())

	e2 = errors.Wrap(nil, "three", 4)
	es.Require().Equal(nil, e2)
}

func (es *errorSuite) TestStringifyField() {
	es.Require().Equal("42", errors.StringifyField(42))
	es.Require().Equal("a", errors.StringifyField("a"))
	es.Require().Equal("[1 2]", errors.StringifyField([]int{1, 2}))
	es.Require().Equal(`["a" "b"]`, errors.StringifyField([]string{"a", "b"}))
}

func TestError(t *testing.T) {
	suite.Run(t, new(errorSuite))
}

func (es *errorSuite) FieldContainsValue(
	s errors.Fields,
	key string,
	value interface{},
	msgAndArgs ...interface{},
) {
	es.T().Helper()
	es.Contains(s, key, msgAndArgs...)
	es.Require().Equal(s[key], value, msgAndArgs...)
}
