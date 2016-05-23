package errorutils_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/cerana/cerana/pkg/errorutils"
	"github.com/stretchr/testify/suite"
)

type errorUtils struct {
	suite.Suite
}

func TestErrorUtils(t *testing.T) {
	suite.Run(t, new(errorUtils))
}

func (s *errorUtils) TestFirst() {
	tests := []struct {
		in  []error
		out error
	}{
		{[]error{errors.New("foo")}, errors.New("foo")},
		{[]error{errors.New("foo"), errors.New("bar"), errors.New("baz")}, errors.New("foo")},
		{[]error{nil, errors.New("bar"), errors.New("baz")}, errors.New("bar")},
		{[]error{nil, nil, errors.New("baz")}, errors.New("baz")},
		{[]error{nil, errors.New("bar"), nil}, errors.New("bar")},
		{[]error{errors.New("foo"), nil, errors.New("baz")}, errors.New("foo")},
		{[]error{}, nil},
	}

	for _, test := range tests {
		s.Equal(test.out, errorutils.First(test.in...), fmt.Sprintf("%+v", test.in))
	}
}
