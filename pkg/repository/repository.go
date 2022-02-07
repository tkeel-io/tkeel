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

const (
	ConfigurationKey       = "configuration"
	ConfigurationSchemaKey = "configuration_schema"
)

var (
	ErrInvalidAnnotations = errors.New("invalid annotations")
	ErrInvalidOptions     = errors.New("invalid options")
)

// Annotations is a json object. Any data you want it attach on.
type Annotations map[string]interface{}

// Option key and value.
type Option struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// InstallerBrief installer brief information.
type InstallerBrief struct {
	Name        string            `json:"name"`
	Repo        string            `json:"repo"`
	Version     string            `json:"version"`
	Installed   bool              `json:"installed"`
	Annotations map[string]string `json:"annotations"`
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
	// Annotations get annotations. custom data. Return a copy of the Annotations.
	Annotations() Annotations
	// Options get installer options.
	Options() []*Option
	// SetOption set option to Installer.
	SetOption(...*Option) error
	// Install plugin.
	Install(...*Option) error
	// Uninstall plugin.
	Uninstall() error
	// Brief get installer brief information.
	Brief() *InstallerBrief
}

// Info repository information.
type Info struct {
	Name        string      `json:"name"`                  // repository name.
	URL         string      `json:"url"`                   // repository url.
	Annotations Annotations `json:"annotations,omitempty"` // repository annotations.
}

func NewInfo(name, url string, annotations Annotations) *Info {
	return &Info{
		Name:        name,
		URL:         url,
		Annotations: annotations,
	}
}

func (i *Info) String() string {
	// TODO: [Warning!] if err is a valid data then this programming is well.
	b, err := json.Marshal(i)
	if err != nil {
		return "<" + err.Error() + ">"
	}
	return string(b)
}

// InfoOperator manager repository information operator.
type InfoOperator interface {
	// Create plugin repo.
	Create(context.Context, *Info) error
	// Get plugin repo info with the repo name.
	Get(ctx context.Context, repoName string) (*Info, error)
	// Delete plugin repo with the repo name.
	Delete(ctx context.Context, repoName string) (*Info, error)
	// List all plugin repo.
	List(ctx context.Context) ([]*Info, error)
	// Watch plugin repo map change. parameter is the changed data.
	Watch(ctx context.Context, interval string, callback func(news, updates, deletes []*Info) error) error
}

// Repository plugin installer repository.
type Repository interface {
	// Info get repository information.
	Info() *Info
	// Search for installers whose names match words. * match all installers.
	Search(word string) ([]*InstallerBrief, error)
	// Get the installer with matching name and version.
	Get(name, version string) (Installer, error)
	// Installed find installed installer(contains installation packages that have been deleted in the repository).
	Installed() ([]Installer, error)
	// Update update repository installer.return whether updated.
	Update() (bool, error)
	// Close this repository.
	Close() error
}

// Constructor return new repository.
type Constructor func(connectInfo *Info, args ...interface{}) (Repository, error)

// DestroyPlugin destroy model.Plugin.
type DestroyPlugin func(pluginID string) error
