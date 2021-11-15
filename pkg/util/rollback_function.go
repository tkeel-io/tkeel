package util

import (
	"github.com/tkeel-io/kit/log"
)

type RollbackFunc func() error

type RollBackStack []RollbackFunc

func NewRollbackStack() RollBackStack {
	return make(RollBackStack, 0)
}

func (rbs *RollBackStack) Run() {
	for _, v := range *rbs {
		err := v()
		if err != nil {
			log.Errorf("error run rollback func: %s", err)
			return
		}
	}
}
