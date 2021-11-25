package errutil

import (
	"github.com/pkg/errors"
	"os"
)

func IsNotExist(err error) bool {
	return os.IsExist(err) || os.IsExist(errors.Cause(err))
}

