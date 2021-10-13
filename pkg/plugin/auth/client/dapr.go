package client

import (
	"fmt"
	"sync"

	daprc "github.com/dapr/go-sdk/client"
)

var (
	dOnce      sync.Once
	DaprClient daprc.Client
)

func GetGrpcClient() (daprc.Client, error) {
	var err error
	dOnce.Do(func() {
		DaprClient, err = daprc.NewClient()
		if err != nil {
			panic(err)
		}
	})

	return DaprClient, fmt.Errorf("error get grpc client: %w", err)
}
