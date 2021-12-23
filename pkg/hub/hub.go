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

package hub

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tkeel/pkg/repository"
	"github.com/tkeel-io/tkeel/pkg/util"
)

var (
	once sync.Once
	h    *Hub
)

// Hub repository manager hub.
type Hub struct {
	infoOperator    repository.InfoOperator
	repoSet         *sync.Map
	constructor     repository.Constructor
	destroy         repository.DestroyPlugin
	constructorArgs []interface{}
}

// Init use Singleton pattern design, generating a new Hub that is globally one assigned to the h variable.
func Init(interval string, op repository.InfoOperator, c repository.Constructor, d repository.DestroyPlugin, initRepoArgs ...interface{}) {
	once.Do(func() {
		h = &Hub{
			infoOperator:    op,
			repoSet:         new(sync.Map),
			constructor:     c,
			destroy:         d,
			constructorArgs: initRepoArgs,
		}
		if err := h.Init(interval); err != nil {
			log.Fatalf("error init hub: %s", err)
			os.Exit(-1)
		}
	})
}

// GetInstance get hub instance.
func GetInstance() *Hub {
	return h
}

// Init hub, read and watch repo info.
func (h *Hub) Init(interval string) error {
	if h.repoSet == nil {
		return errors.New("need initial")
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	modelRepos, err := h.infoOperator.List(ctx)
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
		if err := h.infoOperator.Watch(context.Background(),
			interval, h.updateRepoSet); err != nil {
			log.Panicf("error watch repo: %s", err)
		}
	}()
	return nil
}

// updateRepoSet watch call back func.
func (h *Hub) updateRepoSet(newInfo, updateInfo, deleteInfo []*repository.Info) error {
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

// SetConstructor overflow constructor.
func (h *Hub) SetConstructor(c repository.Constructor, args ...interface{}) {
	h.constructor = c
	h.constructorArgs = args
}

// Add new repo into hub.
func (h *Hub) Add(i *repository.Info) error {
	repo, err := h.constructor(i, h.constructorArgs...)
	if err != nil {
		return fmt.Errorf("error hub constructor repo(%s): %w", i, err)
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	if err = h.infoOperator.Create(ctx, i); err != nil {
		return fmt.Errorf("error repo operator create(%s): %w", i, err)
	}
	h.repoSet.Store(i.Name, repo)
	return nil
}

// Delete delete repo.
func (h *Hub) Delete(name string) (repository.Repository, error) {
	rbStack := util.NewRollbackStack()
	defer rbStack.Run()
	repoIn, ok := h.repoSet.LoadAndDelete(name)
	if !ok {
		return nil, errors.New("repo not exsist")
	}
	repo, ok := repoIn.(repository.Repository)
	if !ok {
		return nil, errors.New("invaild repo type")
	}
	rbStack = append(rbStack, func() error {
		rbRepo, err1 := h.constructor(repo.Info(), h.constructorArgs...)
		if err1 != nil {
			return fmt.Errorf("error delete roll back constructor(%s): %w",
				repo.Info().Name, err1)
		}
		h.repoSet.Store(name, rbRepo)
		return nil
	})
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	if _, err := h.infoOperator.Delete(ctx, name); err != nil {
		return nil, fmt.Errorf("error repo operator delete repo(%s): %w", name, err)
	}
	rbStack = util.NewRollbackStack()
	return repo, nil
}

// Get repo.
func (h *Hub) Get(name string) (repository.Repository, error) {
	repoIn, ok := h.repoSet.Load(name)
	if !ok {
		// find repo info.
		ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
		defer cancel()
		modelInfo, err := h.infoOperator.Get(ctx, name)
		if err != nil {
			log.Errorf("error repo operator find %s: %s", name, err)
			return nil, ErrRepoNotFound
		}
		repo, err := h.constructor(modelInfo, h.constructorArgs...)
		if err != nil {
			log.Errorf("error constructor(%s) repo: %s", modelInfo, err)
			return nil, ErrInternalError
		}
		h.repoSet.Store(name, repo)
		return repo, nil
	}
	repo, ok := repoIn.(repository.Repository)
	if !ok {
		log.Errorf("invalid repo(%s) type.", name)
		return nil, ErrInternalError
	}
	return repo, nil
}

// List get all repo.
func (h *Hub) List() []repository.Repository {
	ret := make([]repository.Repository, 0)
	h.repoSet.Range(func(key, value interface{}) bool {
		r, ok := value.(repository.Repository)
		if !ok {
			log.Errorf("error invalid repo type(%v)", key)
			return false
		}
		ret = append(ret, r)
		return true
	})
	return ret
}

// Uninstall plugin.
func (h *Hub) Uninstall(pluginID string, brief *repository.InstallerBrief) error {
	if brief == nil {
		return errors.New("invalid plugin installer info")
	}
	repoIn, ok := h.repoSet.Load(brief)
	if ok {
		repo, ok := repoIn.(repository.Repository)
		if !ok {
			return errors.New("invaild repo type")
		}
		installedList, err := repo.Installed()
		if err != nil {
			return errors.Wrap(err, "get repo installed error")
		}
		for _, installer := range installedList {
			if installer.Brief().Name == brief.Name && installer.Brief().Version == brief.Version {
				// Here can be call Uninstall() immediate
				// SetPluginID just for make sure the ID is pass by args.
				installer.SetPluginID(pluginID)
				if err := installer.Uninstall(); err != nil {
					return fmt.Errorf("error uninstall plugin(%s): %w", brief, err)
				}
				return nil
			}
		}
	}
	if err := h.destroy(pluginID); err != nil {
		return fmt.Errorf("error destroy plugin(%s): %w", pluginID, err)
	}
	return nil
}
