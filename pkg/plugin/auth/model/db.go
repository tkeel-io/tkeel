package model

import (
	"context"
	"sync"

	daprc "github.com/dapr/go-sdk/client"
	"github.com/tkeel-io/tkeel/pkg/keel"
	"github.com/tkeel-io/tkeel/pkg/logger"
)

var (
	globalDB *gdb
	dbonce   sync.Once

	dblog = logger.NewLogger("Keel.PluginAuth")
)

type DB interface {
	Insert(ctx context.Context, key string, data []byte) error
	Select(ctx context.Context, key string) ([]byte, error)
}

type gdb struct {
	dClient daprc.Client
}

func getDB() DB {
	c := keel.GetClient()
	dbonce.Do(
		func() {
			globalDB = &gdb{c}
		})
	return globalDB
}

func (g *gdb) Insert(ctx context.Context, key string, data []byte) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return g.dClient.SaveState(ctx, keel.PrivateStore, key, data)
}

func (g *gdb) Select(ctx context.Context, key string) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	item, err := g.dClient.GetState(ctx, keel.PrivateStore, key)
	if err != nil {
		return nil, err
	}

	return item.Value, nil
}
