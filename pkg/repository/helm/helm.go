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
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tkeel/pkg/repository"
	helmAction "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	_indexFileName = "index.yaml"
	_repoDirName   = "/.tkeel/repo"
)

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
	driver       Driver
	namespace    string
	index        *Index
}

func NewHelmRepo(info *repository.Info, driver Driver, namespace string) (*Repo, error) {
	var index *Index
	if info != nil {
		// make repository directory.
		repoDirName := _repoDirName + "/" + info.Name + "/"
		_, err := os.Stat(repoDirName)
		if err != nil {
			if os.IsExist(err) {
				if err = os.RemoveAll(repoDirName); err != nil {
					return nil, errors.Wrapf(err, "remove repository directory %s", repoDirName)
				}
			}
			if !os.IsNotExist(err) {
				return nil, errors.Wrap(err, "get repository directory stat")
			}

			if err = os.MkdirAll(repoDirName, os.ModePerm); err != nil {
				return nil, errors.Wrapf(err, "make repository directory %s", repoDirName)
			}
		}
		i, err := NewIndex(info.URL, info.Name)
		if err != nil {
			return nil, errors.Wrapf(err, "new index %s", info.Name)
		}
		index = i
	}

	repo := &Repo{
		info:      info,
		namespace: namespace,
		driver:    driver,
		index:     index,
	}
	if err := repo.configSetup(); err != nil {
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
		return errors.Wrapf(err, "init helm action configuration")
	}
	r.actionConfig = config
	return nil
}

func (r *Repo) Info() *repository.Info {
	return r.info
}

// Search the word in repo, support "*" to get all installable in repo.
func (r *Repo) Search(word string) ([]*repository.InstallerBrief, error) {
	index := r.index

	res, err := index.Search(word, "")
	if err != nil {
		return nil, errors.Wrapf(err, "repo search %s/%s", word, "")
	}
	briefs := res.ToInstallerBrief()

	// modify briefs Installed status
	// 1. get this repo installed
	// 2. range briefs and change Installed status.
	rls, err := r.list()
	if err != nil {
		return nil, errors.Wrap(err, "get helm release")
	}

	installedMap := make(map[string]map[string]struct{}, len(rls))
	for _, v := range rls {
		vMap, ok := installedMap[v.Chart.Metadata.Name]
		if !ok {
			vMap = make(map[string]struct{})
			installedMap[v.Chart.Metadata.Name] = vMap
		}
		vMap[v.Chart.Metadata.Version] = struct{}{}
	}
	for i := 0; i < len(briefs); i++ {
		if vMap, ok := installedMap[briefs[i].Name]; ok {
			if _, ok := vMap[briefs[i].Version]; ok {
				briefs[i].Installed = true
			}
		}
	}

	return briefs, nil
}

// Get the Installer of the specified installable.
func (r *Repo) Get(name, version string) (repository.Installer, error) {
	index := r.index
	resList, err := index.Search(name, version)
	if err != nil {
		return nil, errors.Wrapf(err, "repo search %s/%s", name, version)
	}
	if len(resList) == 0 {
		return nil, ErrNotFound
	}
	res := resList[0]
	// check cache chart.
	chartFile := _repoDirName + "/" + r.info.Name + "/" + res.Name + "-" + res.Version + ".tgz"
	_, err = os.Stat(chartFile)
	if os.IsNotExist(err) {
		log.Debugf("stat err: %s", err)
		if err = downloadChart(chartFile, res.URLs...); err != nil {
			return nil, errors.Wrapf(err, "download chart %s", chartFile)
		}
	} else if err != nil {
		return nil, errors.Wrapf(err, "file %s stat", chartFile)
	}
	// load chart.
	body, err := os.ReadFile(chartFile)
	if err != nil {
		return nil, errors.Wrapf(err, "read chart %s", chartFile)
	}
	d := fmt.Sprintf("%x", sha256.Sum256(body))
	log.Debugf("check sha256: %s -- %s", res.ChartInfo.Digest, d)
	if res.ChartInfo.Digest != d {
		if err = updateChart(chartFile, res.URLs...); err != nil {
			return nil, errors.Wrapf(err, "update chart %s", chartFile)
		}
	}
	ch, err := loader.LoadArchive(bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.Wrapf(err, "load chart %s", chartFile)
	}
	brief := res.ToInstallerBrief()
	rls, err := r.list()
	if err != nil {
		return nil, errors.Wrap(err, "get helm release")
	}

	for _, v := range rls {
		if v.Chart.Metadata.Name == brief.Name && v.Chart.Metadata.Version == brief.Version {
			brief.Installed = true
		}
	}
	i := NewHelmInstaller(brief.Name, ch, *brief, r.namespace, r.actionConfig)
	return &i, nil
}

func (r *Repo) Installed() ([]repository.Installer, error) {
	return r.getInstalled()
}

func (r *Repo) Update() (bool, error) {
	ok, err := r.index.Update()
	if err != nil {
		return false, errors.Wrap(err, "index update")
	}
	if !ok {
		return false, nil
	}
	return true, nil
}

func (r *Repo) Len() int {
	if r.index != nil {
		return len(r.index.helmIndex.Entries)
	}
	return 0
}

func (r *Repo) Close() error {
	if r.info != nil {
		if err := os.RemoveAll(_repoDirName + "/" + r.info.Name); err != nil {
			return errors.Wrapf(err, "remove repository %s files", _repoDirName+"/"+r.info.Name)
		}
	}
	return nil
}

func updateChart(chartFile string, urls ...string) error {
	log.Debug("update chart")
	if err := os.Remove(chartFile); err != nil {
		return errors.Wrapf(err, "remove chart %s", chartFile)
	}
	if err := downloadChart(chartFile, urls...); err != nil {
		return errors.Wrapf(err, "download chart %s", chartFile)
	}
	return nil
}

func downloadChart(chartFile string, urls ...string) error {
	log.Debug("download chart")
	if len(urls) == 0 {
		return ErrNoValidURL
	}
	// download chart.
	var b *bytes.Buffer
	var err error
	for _, url := range urls {
		b, err = _getter.Get(url)
		if err != nil {
			continue
		}
		if err = os.WriteFile(chartFile, b.Bytes(), os.ModePerm); err != nil {
			return errors.Wrapf(err, "write file %s", chartFile)
		}
		break
	}
	if err != nil {
		return errors.Wrapf(err, "GET target file %v failed", urls)
	}
	return nil
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
	index := r.index
	res, err := index.Search("*", "")
	if err != nil {
		return nil, errors.Wrapf(err, "repo search %s/%s", "*", "")
	}
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
		log.Debugf(format, v...)
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
		return nil, errors.Wrap(err, "helmAction configuration init err")
	}
	return config, nil
}
