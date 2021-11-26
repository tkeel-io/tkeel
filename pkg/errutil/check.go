package errutil

import (
	"os"

	"github.com/pkg/errors"
)

func IsNotExist(err error) bool {
	return os.IsNotExist(err) || os.IsNotExist(errors.Cause(err))
}
