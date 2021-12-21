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
	"strings"

	"helm.sh/helm/v3/pkg/cli"

	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"

	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tkeel/pkg/repository"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

const (
	ReadmeFileNameKey = "README"
	ValuesSchemaKey   = "VALUES.SCHEMA"
)

var _ repository.Installer = &Installer{}

type Installer struct {
	chart       *chart.Chart
	helmConfig  *action.Configuration
	options     map[string]interface{}
	id          string
	brief       repository.InstallerBrief
	annotations repository.Annotations
	namespace   string
}

func NewHelmInstaller(id string, ch *chart.Chart, brief repository.InstallerBrief, namespace string, helmConfig *action.Configuration) Installer {
	return Installer{
		chart:      ch,
		helmConfig: helmConfig,
		id:         id,
		namespace:  namespace,
		annotations: func() repository.Annotations {
			a := make(repository.Annotations)
			for _, v := range ch.Files {
				if strings.HasPrefix(v.Name, ReadmeFileNameKey) {
					a[ReadmeFileNameKey] = v.Data
				}
				if ch.Schema != nil {
					a[ValuesSchemaKey] = ch.Schema
				}
			}
			return a
		}(),
		brief:   brief,
		options: ch.Values,
	}
}

func NewHelmInstallerQuick(id, namespace string, config *action.Configuration) Installer {
	return Installer{
		namespace:   namespace,
		id:          id,
		helmConfig:  config,
		annotations: make(repository.Annotations),
	}
}

func (h *Installer) SetChart(ch *chart.Chart) {
	h.chart = ch
}

func (h Installer) GetChart() *chart.Chart {
	return h.chart
}

func (h *Installer) SetPluginID(id string) {
	h.id = id
}

func (h Installer) Annotations() repository.Annotations {
	return h.annotations
}

func (h Installer) Options() []*repository.Option {
	return func() []*repository.Option {
		ret := make([]*repository.Option, 0, len(h.options))
		for k, v := range h.options {
			ret = append(ret, &repository.Option{
				Key:   k,
				Value: v,
			})
		}
		return ret
	}()
}

func (h *Installer) SetOption(ops ...*repository.Option) error {
	for _, v := range ops {
		_, ok := h.options[v.Key]
		if !ok {
			return errors.New("option(" + v.Key + ") not found")
		}
		h.options[v.Key] = v.Value
	}
	return nil
}

func (h Installer) Install(ops ...*repository.Option) error {
	for _, v := range ops {
		_, ok := h.options[v.Key]
		if ok {
			h.options[v.Key] = v.Value
		}
	}

	for i := 0; i < len(ops); i++ {
		if err := ops[i].Check(); err != nil {
			return errors.Wrap(err, "check option failed")
		}
	}

	installer := action.NewInstall(h.helmConfig)

	installer.Version = h.brief.Version

	if err := checkIfInstallable(h.chart); err != nil {
		return err
	}

	if h.chart.Metadata.Deprecated {
		log.Warn("This chart is deprecated")
	}

	// Add inject dependencies.
	inject, err := loadComponentChart()
	if err != nil {
		log.Error(err)
		return errors.Wrap(err, "load component chart err")
	}
	failInject(inject, h.id)
	h.chart.AddDependency(inject)

	installer.Namespace = h.namespace
	installer.ReleaseName = h.id
	if _, err := installer.Run(h.chart, nil); err != nil {
		return errors.Wrap(err, "INSTALLATION FAILED")
	}
	return nil
}

func (h Installer) Uninstall() error {
	uninstallClint := action.NewUninstall(h.helmConfig)
	_, err := uninstallClint.Run(h.id)
	if err != nil {
		err = errors.Wrap(err, "call uninstall err")
		return err
	}

	return nil
}

func (h Installer) Brief() *repository.InstallerBrief {
	return &h.brief
}

func loadComponentChart() (*chart.Chart, error) {
	chartURL, err := repo.FindChartInRepoURL(_tkeelRepo, _componentChartName, "", "", "", "", getter.All(new(cli.EnvSettings)))
	if err != nil {
		return nil, errors.Wrap(err, "get component chart url err")
	}

	httpGetter, err := getter.NewHTTPGetter()
	if err != nil {
		return nil, errors.Wrap(err, "init http getter err")
	}

	buf, err := httpGetter.Get(chartURL)
	if err != nil {
		return nil, errors.Wrap(err, "HTTP GET component chart err")
	}

	c, err := loader.LoadArchive(buf)
	if err != nil {
		log.Warn("can't parse the file %q", chartURL, err)
		return nil, errors.Wrap(err, "load helm chart failed")
	}

	if err = checkIfInstallable(c); err != nil {
		log.Warn("uninstallable chart request")
		return nil, err
	}

	if c.Metadata.Deprecated {
		log.Warn("%q: This chart is deprecated", chartURL)
	}

	return c, nil
}

func checkIfInstallable(ch *chart.Chart) error {
	if ch == nil {
		return ErrNoChartInfoSet
	}
	switch ch.Metadata.Type {
	case "", "application":
		return nil
	}
	return errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}

func failInject(inject *chart.Chart, pluginName string) {
	inject.Values["pluginID"] = pluginName
}
