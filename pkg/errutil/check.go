package errutil

import (
	"os"

	"github.com/pkg/errors"
)

func IsNotExist(err error) bool {
	return os.IsExist(err) || os.IsExist(errors.Cause(err))
}
