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
	_globalDB *db
	_dbOnce   sync.Once

	_dbLog = logger.NewLogger("keel.plugin auth")
)

type DB interface {
	Insert(ctx context.Context, key string, data []byte) error
	Select(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}

type db struct {
	dClient daprc.Client
}

func getDB() DB {
	c := keel.GetClient()
	_dbOnce.Do(
		func() {
			_globalDB = &db{c}
		})
	return _globalDB
}

func (db *db) Insert(ctx context.Context, key string, data []byte) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := db.dClient.SaveState(ctx, keel.PrivateStore, key, data); err != nil {
		return fmt.Errorf("error save state: %w", err)
	}
	return nil
}

func (db *db) Select(ctx context.Context, key string) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	item, err := db.dClient.GetState(ctx, keel.PrivateStore, key)
	if err != nil {
		return nil, fmt.Errorf("error get state: %w", err)
	}
	return item.Value, nil
}

func (db *db) Delete(ctx context.Context, key string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	err := db.dClient.DeleteState(ctx, keel.PrivateStore, key)
	if err != nil {
		return fmt.Errorf("error delete state: %w ", err)
	}
	return nil
}
