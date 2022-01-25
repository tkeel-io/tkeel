package hub

import "errors"

var (
	ErrRepoNotFound  = errors.New("repo not found")
	ErrInternalError = errors.New("internal error")
	ErrRepoExist     = errors.New("repo exist")
)
