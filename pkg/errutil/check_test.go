package errutil

import (
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type myError struct {
	msg string
}

func (e myError) Cause() error {
	return os.ErrNotExist
}

func (e myError) Error() string {
	return e.msg
}

func TestIsNotExist(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"test os.ErrNotExist", os.ErrNotExist, true},
		{"test other err", errors.New("other err"), false},
		{"test a impl Cause os.ErrNotExist", myError{msg: "my err"}, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := IsNotExist(test.err)
			assert.Equal(t, test.want, res)
		})
	}
}
