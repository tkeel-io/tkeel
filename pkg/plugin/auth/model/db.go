package model

import (
	"context"
	"fmt"
	"sync"

	"github.com/tkeel-io/tkeel/pkg/keel"
	"github.com/tkeel-io/tkeel/pkg/logger"

	daprc "github.com/dapr/go-sdk/client"
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
	if err := g.dClient.SaveState(ctx, keel.PrivateStore, key, data); err != nil {
		return fmt.Errorf("error save state: %w", err)
	}
	return nil
}

func (g *gdb) Select(ctx context.Context, key string) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	item, err := g.dClient.GetState(ctx, keel.PrivateStore, key)
	if err != nil {
		return nil, fmt.Errorf("error get state: %w", err)
	}
	return item.Value, nil
}
