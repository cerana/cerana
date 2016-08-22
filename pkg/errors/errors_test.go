package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type Errors struct {
	suite.Suite
}

func TestErrors(t *testing.T) {
	suite.Run(t, new(Errors))
}

func (s *Errors) TestStack() {
	// getPCs
	pcs := getPCs(0)
	if !s.True(len(pcs) > 0, "wrong pcs length") {
		return
	}

	// Skip one frame
	pcsSkip1 := getPCs(1)
	s.Equal(len(pcs)-1, len(pcsSkip1), "wrong pcs length with skip 1")

	// callstack
	stack := callstack(pcs)
	s.Equal(len(pcs), len(stack), "wrong stack length")
}

func (s *Errors) TestFromError() {
	err := errors.New("error message")
	newErr := fromError(err)
	s.Equal(err, newErr.cause, "cause should be original error")
	s.NotNil(newErr.context, "context should be initialized")
	s.NotNil(newErr.data, "data should be initialized")
	s.True(len(newErr.pcs) > 0, "callstack should be generated")

	newErr2 := fromError(newErr)
	s.Equal(newErr, newErr2, "should not change an errorExt")
}

func (s *Errors) TestNew() {
	msg := "error message"
	err := New(msg)
	eExt, ok := err.(*errorExt)
	if !s.True(ok, "wrong error return type") {
		return
	}
	if s.NotNil(eExt.cause, "cause should be created") {
		s.Equal(msg, eExt.cause.Error(), "cause should use supplied message")
	}
	s.NotNil(eExt.context, "context should be initialized")
	s.NotNil(eExt.data, "data should be initialized")
	s.True(len(eExt.pcs) > 0, "callstack should be generated")
}

func (s *Errors) TestNewf() {
	format := "%s:%d"
	args := []interface{}{"foo", uint64(10)}
	msg := fmt.Sprintf(format, args...)

	err := Newf(format, args...)
	eExt, ok := err.(*errorExt)
	if !s.True(ok, "wrong error return type") {
		return
	}
	if s.NotNil(eExt.cause, "cause should be created") {
		s.Equal(msg, eExt.cause.Error(), "cause should use correct message")
	}
	s.NotNil(eExt.context, "context should be initialized")
	s.NotNil(eExt.data, "data should be initialized")
	s.True(len(eExt.pcs) > 0, "callstack should be generated")
}

func (s *Errors) TestWrap() {
	// Wrap a regular error
	msg := "error message"
	err := errors.New(msg)
	wrappedErr := Wrap(err)
	eExt, ok := wrappedErr.(*errorExt)
	if !s.True(ok, "wrong error return type") {
		return
	}
	s.Equal(err, eExt.cause, "cause should be original error")
	s.NotNil(eExt.context, "context should be initialized")
	s.NotNil(eExt.data, "data should be initialized")
	s.True(len(eExt.pcs) > 0, "callstack should be generated")

	// Wrap a regular error with context
	ctx := "some context"
	wrappedErrWithContext := Wrap(err, ctx)
	eExt, ok = wrappedErrWithContext.(*errorExt)
	if !s.True(ok, "wrong error return type") {
		return
	}
	s.Equal(err, eExt.cause, "cause should be original error")
	if s.Len(eExt.context, 1, "should have added context") {
		s.Equal(ctx, eExt.context[0], "should be correct context")
	}

	// Wrap an errorExt
	rewrappedErr := Wrap(wrappedErrWithContext)
	s.Equal(wrappedErrWithContext, rewrappedErr, "should not change an errorExt")

	// Wrap an errorExt with context
	ctx2 := "more context"
	rewrappedErr = Wrap(wrappedErrWithContext, ctx2)
	eExt, ok = wrappedErrWithContext.(*errorExt)
	if !s.True(ok, "wrong error return type") {
		return
	}
	s.Equal(err, eExt.cause, "cause should be original error")
	if s.Len(eExt.context, 2, "should have added context") {
		s.Equal(ctx, eExt.context[0], "should be correct context")
		s.Equal(ctx2, eExt.context[1], "should be correct context")
	}
}

func (s *Errors) TestWrapf() {
	// Wrap a regular error
	format := "%s:%d"
	args := []interface{}{"foo", uint64(10)}
	ctx := fmt.Sprintf(format, args...)
	msg := "error message"
	err := errors.New(msg)
	wrappedErr := Wrapf(err, format, args...)
	eExt, ok := wrappedErr.(*errorExt)
	if !s.True(ok, "wrong error return type") {
		return
	}
	s.Equal(err, eExt.cause, "cause should be original error")
	if s.Len(eExt.context, 1, "should have added context") {
		s.Equal(ctx, eExt.context[0], "should be correct context")
	}
	s.NotNil(eExt.data, "data should be initialized")
	s.True(len(eExt.pcs) > 0, "callstack should be generated")

	// Wrap an errorExt
	format = "%s:%d"
	args = []interface{}{"bar", uint64(20)}
	ctx2 := fmt.Sprintf(format, args...)
	rewrappedErr := Wrap(wrappedErr, ctx2)
	eExt, ok = rewrappedErr.(*errorExt)
	if !s.True(ok, "wrong error return type") {
		return
	}
	s.Equal(err, eExt.cause, "cause should be original error")
	if s.Len(eExt.context, 2, "should have added context") {
		s.Equal(ctx, eExt.context[0], "should be correct context")
		s.Equal(ctx2, eExt.context[1], "should be correct context")
	}
}

func (s *Errors) TestWrapv() {
	// Wrap a regular error with nil values
	msg := "error message"
	err := errors.New(msg)
	wrappedErr := Wrapv(err, nil)
	eExt, ok := wrappedErr.(*errorExt)
	if !s.True(ok, "wrong error return type") {
		return
	}
	if s.NotNil(eExt.cause, "cause should be created") {
		s.Equal(err, eExt.cause, "cause should be original error")
	}
	s.NotNil(eExt.data, "data should be initialized")

	// Wrap a regular error with values
	values := map[string]interface{}{"foo": "bar"}
	wrappedErr = Wrapv(err, values)
	eExt, ok = wrappedErr.(*errorExt)
	if !s.True(ok, "wrong error return type") {
		return
	}
	if s.NotNil(eExt.cause, "cause should be created") {
		s.Equal(err, eExt.cause, "cause should be original error")
	}
	s.Equal(values, eExt.data, "data should be values")

	// Wrap an errorExt with nil values
	rewrappedErr := Wrapv(wrappedErr, nil)
	s.Equal(wrappedErr, rewrappedErr, "should not reset data or change error")

	// Wrap an errorExt with values
	values2 := map[string]interface{}{"baz": "bang"}
	rewrappedErr = Wrapv(wrappedErr, values2)
	eExt, ok = rewrappedErr.(*errorExt)
	if !s.True(ok, "wrong error return type") {
		return
	}
	combinedValues := make(map[string]interface{})
	for k, v := range values {
		combinedValues[k] = v
	}
	for k, v := range values2 {
		combinedValues[k] = v
	}
	s.Equal(combinedValues, eExt.data, "data should be combined values")
}

type testErr struct {
	SomeValue int `json:"someValue"`
	msg       string
}

func (t *testErr) Error() string {
	return t.msg
}

func (s *Errors) TestMarshalJSON() {
	msg := "error message"
	ctx1 := "some context"
	ctx2 := "more context"
	values := map[string]interface{}{"foo": "bar"}

	origErr := &testErr{SomeValue: 5, msg: msg}
	err := Wrap(origErr, ctx1)
	err = Wrap(err, ctx2)
	err = Wrapv(err, values)
	err = Wrapv(err, map[string]interface{}{"self": err})

	j, jmErr := json.Marshal(err)
	if !s.NoError(jmErr, "failed to marshal error") {
		return
	}

	output := make(map[string]interface{})
	if !s.NoError(json.Unmarshal(j, &output), "failed to unmarshal output") {
		return
	}

	eExt := err.(*errorExt)

	// error message should be present
	causeI, ok := output["cause"]
	if s.True(ok, "output missing cause") {
		var cause string
		cause, ok = causeI.(string)
		if s.True(ok, "cause should be a string") {
			s.Equal(ctx2+": "+ctx1+": "+msg, cause, "unexpected cause string")
		}
	}

	// stack with strings should be present
	stackI, ok := output["stack"]
	if s.True(ok, "output missing stack") {
		var stackAI []interface{}
		stackAI, ok = stackI.([]interface{})
		if s.True(ok, "stack should be an array of interface{}") {
			stack := make([]string, len(stackAI))
			for i, v := range stackAI {
				stack[i], ok = v.(string)
				s.True(ok, "stack array value should be a string")
			}
			s.Equal(callstack(eExt.pcs), stack, "wrong stack")
		}
	}

	// data and cause fields should be top level
	values["someValue"] = origErr.SomeValue
	for k, v := range values {
		valueI, ok := output[k]
		if s.True(ok, "missing value:"+k) {
			s.EqualValues(v, valueI, "wrong value:"+k)
		}
	}

	// self error should be omitted
	s.Nil(output["self"])
}

func (s *Errors) TestCause() {
	s.Nil(Cause(nil))
	err := errors.New("an error")
	if !s.NotNil(Cause(err)) {
		return
	}
	s.Equal(err, Cause(err))
	wrapped := Wrap(err, "context")
	s.Equal(err, Cause(wrapped))
	s.NotEqual(wrapped, Cause(wrapped))
}
