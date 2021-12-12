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

package prepo

import (
	"errors"

	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/repository"
)

var (
	ErrPluginRepoExsist          = errors.New("error plugin repo existed")
	ErrPluginRepoNotExsist       = errors.New("error plugin repo not existed")
	ErrPluginRepoVersionMismatch = errors.New("error plugin repo version mismatch")
)

// Operator contains all operations to plugin repo.
type Operator interface {
	// Model2Info model.PluginRepo convert to repository.Info.
	Model2Info(*model.PluginRepo) *repository.Info
	// Info2Model repository.Info convert to model.PluginRepo.
	Info2Model(*repository.Info) *model.PluginRepo
	// GetChanges compare the old and the new one and get the new, update and delete.
	GetChanges(old, curr model.PluginRepoMap) (new, update, delete []*model.PluginRepo)
	// repository info.
	repository.InfoOperator
}
