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

package helm

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tkeel/pkg/repository"
	helmAction "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const _indexFileName = "index.yaml"

var (
	ErrNotFound       = errors.New("not found")
	ErrNoValidURL     = errors.New("no valid url")
	ErrNoChartInfoSet = errors.New("no chart info set in installer")
)

// Driver is a short way define for Helm Store Status.
type Driver string

func (d Driver) String() string {
	return string(d)
}

const (
	Secret    Driver = "secret"
	ConfigMap Driver = "configmap"
	Mem       Driver = "memory"
	SQL       Driver = "sql"
)

const (
	_tkeelRepo          = "https://tkeel-io.github.io/helm-charts"
	_componentChartName = "tkeel-plugin-components"
)

var (
	_                repository.Repository = &Repo{}
	_componentSecret                       = "changeme"
)

func SetSecret(s string) {
	_componentSecret = s
}

func GetSecret() string {
	return _componentSecret
}

// Repo is the impl repository.Repository.
type Repo struct {
	info         *repository.Info
	actionConfig *helmAction.Configuration
	httpGetter   getter.Getter
	driver       Driver
	namespace    string
}

func NewHelmRepo(info repository.Info, driver Driver, namespace string) (*Repo, error) {
	httpGetter, err := getter.NewHTTPGetter()
	if err != nil {
		log.Warn("init helm action configuration err", err)
		return nil, errors.Wrap(err, "init http getter failed")
	}
	repo := &Repo{
		info:       &info,
		namespace:  namespace,
		driver:     driver,
		httpGetter: httpGetter,
	}
	if err = repo.configSetup(); err != nil {
		return nil, errors.Wrap(err, "setup helm action configuration failed")
	}
	return repo, nil
}

func (r *Repo) SetInfo(info repository.Info) {
	r.info = &info
}

func (r Repo) Config() *helmAction.Configuration {
	return r.actionConfig
}

func (r *Repo) Namespace() string {
	return r.namespace
}

func (r *Repo) SetNamespace(namespace string) error {
	r.namespace = namespace
	return r.configSetup()
}

func (r *Repo) SetDriver(driver Driver) error {
	r.driver = driver
	return r.configSetup()
}

func (r Repo) GetDriver() Driver {
	return r.driver
}

func (r *Repo) configSetup() error {
	config, err := initActionConfig(r.namespace, r.driver)
	if err != nil {
		log.Warn("init helm action configuration err", err)
		return err
	}
	r.actionConfig = config
	return nil
}

func (r *Repo) Info() *repository.Info {
	return r.info
}

// Search the word in repo, support "*" to get all installable in repo.
func (r *Repo) Search(word string) ([]*repository.InstallerBrief, error) {
	index, err := r.BuildIndex()
	if err != nil {
		return nil, errors.Wrap(err, "can't build helm index configSetup")
	}

	res := index.Search(word, "")
	briefs := res.ToInstallerBrief()

	// modify briefs Installed status
	// 1. get this repo installed
	// 2. range briefs and change Installed status.
	installedList, err := r.getInstalled()
	if err != nil {
		return nil, err
	}

	installedMap := make(map[string]string, len(installedList))
	for i := range installedList {
		installedMap[installedList[i].Brief().Name] = installedList[i].Brief().Version
	}
	for i := 0; i < len(briefs); i++ {
		if version, ok := installedMap[briefs[i].Name]; ok {
			if version == briefs[i].Version {
				briefs[i].Installed = true
			}
		}
	}

	return briefs, nil
}

// Get the Installer of the specified installable.
func (r *Repo) Get(name, version string) (repository.Installer, error) {
	index, err := r.BuildIndex()
	if err != nil {
		return nil, errors.Wrap(err, "can't build helm index configSetup")
	}

	res := index.Search(name, version)
	if len(res) != 1 {
		return nil, ErrNotFound
	}

	if len(res[0].URLs) == 0 {
		return nil, ErrNoValidURL
	}

	var buf *bytes.Buffer
	err = nil
	for i := range res[0].URLs {
		buf, err = r.httpGetter.Get(res[0].URLs[i])
		if err != nil {
			continue
		}
		break
	}
	if err != nil {
		return nil, errors.Wrap(err, "GET target file failed")
	}

	ch, err := loader.LoadArchive(buf)
	if err != nil {
		return nil, errors.Wrap(err, "Load archive to struct Chart failed")
	}

	brief := res[0].ToInstallerBrief()
	i := NewHelmInstaller(brief.Name, ch, *brief, r.namespace, r.actionConfig)
	return &i, nil
}

func (r *Repo) Installed() ([]repository.Installer, error) {
	return r.getInstalled()
}

func (r *Repo) Close() error {
	return nil
}

func (r *Repo) BuildIndex() (*Index, error) {
	fileContent, err := r.GetIndex()
	if err != nil {
		return nil, err
	}
	return NewIndex(r.info.Name, fileContent)
}

// GetIndex get the repo index.yaml file content.
func (r *Repo) GetIndex() ([]byte, error) {
	url := strings.TrimSuffix(r.info.URL, "/") + "/" + _indexFileName

	buf, err := r.httpGetter.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "HTTP GET error")
	}

	return buf.Bytes(), nil
}

// list the installed release plugin by helm .
func (r *Repo) list() ([]*release.Release, error) {
	listAction := helmAction.NewList(r.actionConfig)
	releases, err := listAction.Run()
	if err != nil {
		return nil, errors.Wrap(err, "run helm list action failed")
	}
	return releases, nil
}

func (r *Repo) getInstalled() ([]repository.Installer, error) {
	index, err := r.BuildIndex()
	if err != nil {
		return nil, err
	}

	res := index.Search("*", "")

	rls, err := r.list()
	if err != nil {
		return nil, err
	}

	cache := make(map[string]*PluginRes)
	for i := 0; i < len(res); i++ {
		cache[res[i].Name] = res[i]
	}

	list := make([]repository.Installer, 0)
	for i := 0; i < len(rls); i++ {
		if plugin, ok := cache[rls[i].Chart.Name()]; ok {
			installer := NewHelmInstaller(
				rls[i].Name,                /* Installed Plugin ID. */
				rls[i].Chart,               /* Plugin Chart. */
				*plugin.ToInstallerBrief(), /* Brief. */
				r.namespace,                /* Namespace. */
				r.actionConfig,             /* Action Config. */
			)
			list = append(list, &installer)
		}
	}

	return list, nil
}

func getDebugLogFunc() helmAction.DebugLog {
	return func(format string, v ...interface{}) {
		log.Infof(format, v...)
	}
}

// initActionConfig Initialize a usable helm action.Configuration.
func initActionConfig(namespace string, driver Driver) (*helmAction.Configuration, error) {
	config := new(helmAction.Configuration)
	k8sFlags := &genericclioptions.ConfigFlags{
		Namespace: &namespace,
	}
	err := config.Init(k8sFlags, namespace, driver.String(), getDebugLogFunc())
	if err != nil {
		return nil, fmt.Errorf("helmAction configuration init err:%w", err)
	}
	return config, nil
}
