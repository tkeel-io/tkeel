/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tkeel/pkg/util"
)

var (
	once sync.Once
	h    *Hub
)

// Hub repository manager hub.
type Hub struct {
	repoOperator    InfoOperator
	repoSet         *sync.Map
	constructor     Constructor
	destory         DestoryPlugin
	constructorArgs []interface{}
}

// InitHub initial hub.
func InitHub(interval string, op InfoOperator, c Constructor, d DestoryPlugin, initRepoArgs ...interface{}) {
	once.Do(func() {
		h = &Hub{
			repoOperator:    op,
			repoSet:         new(sync.Map),
			constructor:     c,
			destory:         d,
			constructorArgs: initRepoArgs,
		}
		if err := h.Init(interval); err != nil {
			log.Fatalf("error init hub: %s", err)
			os.Exit(-1)
		}
	})
}

// GetHub get hub instance.
func GetHub() *Hub {
	return h
}

// Init init hub, read and watch model repo.
func (h *Hub) Init(interval string) error {
	if h.repoSet == nil {
		return errors.New("need initial")
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	modelRepos, err := h.repoOperator.List(ctx)
	if err != nil {
		return fmt.Errorf("error list repo: %w", err)
	}
	for _, v := range modelRepos {
		repo, err := h.constructor(v, h.constructorArgs...)
		if err != nil {
			return fmt.Errorf("error constructor repo(%s): %w", v, err)
		}
		h.repoSet.Store(v.Name, repo)
	}

	go func() {
		if err := h.repoOperator.Watch(context.Background(),
			interval, h.updateRepoSet); err != nil {
			log.Panicf("error watch repo: %s", err)
		}
	}()
	return nil
}

// updateRepoSet watch call back func.
func (h *Hub) updateRepoSet(newInfo, updateInfo, deleteInfo []*Info) error {
	log.Debugf("update hub repo set, news: %s, updates: %s, deletes: %s",
		infoSliStr(newInfo), infoSliStr(updateInfo), infoSliStr(deleteInfo))
	// create new repo.
	for _, v := range newInfo {
		newRepo, err := h.constructor(v, h.constructorArgs...)
		if err != nil {
			err = fmt.Errorf("error constructor(%s): %w", v, err)
			log.Error(err)
			return err
		}
		h.repoSet.Store(v.Name, newRepo)
	}
	// delete old repo.
	for _, v := range deleteInfo {
		h.repoSet.Delete(v.Name)
	}
	// update new repo.
	for _, v := range updateInfo {
		changeRepo, err := h.constructor(v, h.constructorArgs...)
		if err != nil {
			err = fmt.Errorf("error constructor(%s): %w", v, err)
			log.Error(err)
			return err
		}
		h.repoSet.Store(v.Name, changeRepo)
	}
	return nil
}

// SetConstructor overflow constructor
func (h *Hub) SetConstructor(c Constructor, args ...interface{}) {
	h.constructor = c
	h.constructorArgs = args
}

// Add add new repo into hub.
func (h *Hub) Add(i *Info) error {
	repo, err := h.constructor(i, h.constructorArgs...)
	if err != nil {
		return fmt.Errorf("error hub constructor repo(%s): %w", i, err)
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	if err = h.repoOperator.Create(ctx, i); err != nil {
		return fmt.Errorf("error repo operator create(%s): %w", i, err)
	}
	h.repoSet.Store(i.Name, repo)
	return nil
}

// Delete delete repo.
func (h *Hub) Delete(name string) (ret Repository, err error) {
	rbStack := util.NewRollbackStack()
	defer func() {
		if err != nil {
			rbStack.Run()
		}
	}()
	repoIn, ok := h.repoSet.LoadAndDelete(name)
	if !ok {
		return nil, errors.New("repo not exsist")
	}
	repo, ok := repoIn.(Repository)
	if !ok {
		return nil, errors.New("invaild repo type")
	}
	rbStack = append(rbStack, func() error {
		rbRepo, err := h.constructor(repo.Info(), h.constructorArgs...)
		if err != nil {
			return fmt.Errorf("error delete roll back constructor(%s): %w",
				repo.Info().Name, err)
		}
		h.repoSet.Store(name, rbRepo)
		return nil
	})
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	_, err = h.repoOperator.Delete(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("error repo operator delete repo(%s): %w", name, err)
	}
	return repo, nil
}

// Get get repo.
func (h *Hub) Get(name string) (Repository, error) {
	repoIn, ok := h.repoSet.Load(name)
	if !ok {
		// find model repo.
		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()
		modelInfo, err := h.repoOperator.Get(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("error repo operator find %s: %w", name, err)
		}
		repo, err := h.constructor(modelInfo, h.constructorArgs...)
		if err != nil {
			return nil, fmt.Errorf("error constructor(%s) repo: %w", modelInfo, err)
		}
		h.repoSet.Store(name, repo)
		return repo, nil
	}
	repo, ok := repoIn.(Repository)
	if !ok {
		return nil, errors.New("invaild repo type")
	}
	return repo, nil
}

// Uninstall plugin.
func (h *Hub) Uninstall(pluginID string, i Installer) error {
	if i == nil {
		return errors.New("invaild plugin installer info.")
	}
	breif := i.Brief()
	repoIn, ok := h.repoSet.Load(breif)
	if ok {
		repo, ok := repoIn.(Repository)
		if !ok {
			return errors.New("invaild repo type")
		}
		for _, v := range repo.Installed() {
			if v.Brief().Name == breif.Name && v.Brief().Version == breif.Version {
				if err := v.Uninstall(pluginID); err != nil {
					return fmt.Errorf("error uninstall plugin(%s): %w", breif, err)
				}
				return nil
			}
		}
	}
	if err := h.destory(pluginID); err != nil {
		return fmt.Errorf("error destory plugin(%s): %w", pluginID, err)
	}
	return nil
}

func infoSliStr(is []*Info) string {
	ret := make([]string, 0, len(is)+2)
	ret = append(ret, "[")
	for _, v := range is {
		ret = append(ret, v.String())
	}
	ret = append(ret, "]")
	return strings.Join(ret, ", ")
}
