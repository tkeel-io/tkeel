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
	"encoding/json"
	"errors"
)

var (
	ErrAnnotationsInvaild = errors.New("annotations invaild")
	ErrOptionsInvaild     = errors.New("options invaild")
)

// Annotations json object.
type Annotations map[string]interface{}

func (a *Annotations) Check() error {
	_, err := json.Marshal(a)
	if err != nil {
		return ErrAnnotationsInvaild
	}
	return nil
}

// Options key and value.
type Option struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func (o *Option) Check() error {
	_, err := json.Marshal(o)
	if err != nil {
		return ErrAnnotationsInvaild
	}
	return nil
}

// ConfigurationItem installer configutation item,json format.
type ConfigurationItem struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	Desc  string      `json:"desc,omitempty"`
}

func (c *ConfigurationItem) String() string {
	b, err := json.Marshal(c)
	if err != nil {
		return "<" + err.Error() + ">"
	}
	return string(b)
}

// InstallerBrief installer brief information.
type InstallerBrief struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Installed bool   `json:"installed"`
}

func (ib *InstallerBrief) String() string {
	b, err := json.Marshal(ib)
	if err != nil {
		return "<" + err.Error() + ">"
	}
	return string(b)
}

// Installer plugin installer.
type Installer interface {
	// SetPluginID set plugin id after installing to tKeel.
	SetPluginID(pluginID string)
	// Annotations get annotations.custom datas.
	Annotations() Annotations
	// Options get installer options.
	Options() []Option
	// SetOption set installer option.
	SetOption(...*Option) error
	// Install install plugin.
	Install(...Option) error
	// Uninstall Uninstall plugin.
	Uninstall(pluginID string) error
	// Brief get installer brief information.
	Brief() *InstallerBrief
}

// Info repository information.
type Info struct {
	Name        string      `json:"name"`                  // repository name.
	URL         string      `json:"url"`                   // repository url.
	Annotations Annotations `json:"annotations,omitempty"` // repository annotations.
}

func (i *Info) String() string {
	b, err := json.Marshal(i)
	if err != nil {
		return "<" + err.Error() + ">"
	}
	return string(b)
}

// InfoOperator manager repository information operator.
type InfoOperator interface {
	// create plugin repo.
	Create(context.Context, *Info) error
	// Get plugin repo info with the repo name.
	Get(ctx context.Context, repoName string) (*Info, error)
	// Delete plugin repo with the repo name.
	Delete(ctx context.Context, repoName string) (*Info, error)
	// List get all plugin repo.
	List(ctx context.Context) ([]*Info, error)
	// Watch plugin repo map change. parameter is the changed data.
	Watch(ctx context.Context, interval string, callback func(news, updates, deletes []*Info) error) error
}

// Repository plugin installer repository.
type Repository interface {
	// Info get repository information.
	Info() *Info
	// Search for installers whose names match words.
	Search(word string) ([]*InstallerBrief, error)
	// Get the installer with matching name and version.
	Get(name, version string) (Installer, error)
	// Installed find installed installer(contains installation packages that have been deleted in the repository).
	Installed() []Installer
	// Close close this repository.
	Close() error
}

// Constructor return new repository.
type Constructor func(connectInfo *Info, args ...interface{}) (Repository, error)

// DestoryPlugin destory model.Plugin.
type DestoryPlugin func(pluginID string) error
