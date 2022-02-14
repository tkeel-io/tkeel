package hub

import "github.com/pkg/errors"

var (
	ErrRepoNotFound  = errors.New("repo not found")
	ErrInternalError = errors.New("internal error")
	ErrRepoExist     = errors.New("repo exist")
)
