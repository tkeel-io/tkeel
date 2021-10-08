package client

import (
	daprc "github.com/dapr/go-sdk/client"
	"sync"
)

var (
	dOnce      sync.Once
	DaprClient daprc.Client
)

func GetGrpcClient() (daprc.Client, error) {
	var err error
	dOnce.Do(func() {
		DaprClient, err = daprc.NewClient()
	})

	return DaprClient, err
}
