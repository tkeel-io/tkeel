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
	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
)

var _ Installer = &HelmInstaller{}

type HelmInstaller struct {
	chart       *chart.Chart
	helmConfig  *action.Configuration
	options     []*Option
	id          string
	brief       InstallerBrief
	annotations Annotations
	namespace   string
}

func NewHelmInstaller(id string, ch *chart.Chart, brief InstallerBrief, namespace string, helmConfig *action.Configuration, options ...*Option) HelmInstaller {
	return HelmInstaller{
		chart:       ch,
		helmConfig:  helmConfig,
		id:          id,
		namespace:   namespace,
		annotations: make(Annotations),
		brief:       brief,
		options:     options,
	}
}

func NewHelmInstallerQuick(id, namespace string, config *action.Configuration, options ...*Option) HelmInstaller {
	return HelmInstaller{
		namespace:   namespace,
		id:          id,
		helmConfig:  config,
		options:     options,
		annotations: make(Annotations),
	}
}

func (h *HelmInstaller) SetChart(ch *chart.Chart) {
	h.chart = ch
}

func (h HelmInstaller) GetChart() *chart.Chart {
	return h.chart
}

func (h *HelmInstaller) SetPluginID(id string) {
	h.id = id
}

func (h HelmInstaller) Annotations() Annotations {
	return h.annotations
}

func (h HelmInstaller) Options() []*Option {
	return h.options
}

func (h *HelmInstaller) SetOption(options ...*Option) error {
	h.options = append(h.options, options...)
	return nil
}

func (h HelmInstaller) Install(options ...*Option) error {
	ops := h.options[:]
	ops = append(ops, options...)

	installer := action.NewInstall(h.helmConfig)

	installer.Version = h.brief.Version

	if err := checkIfInstallable(h.chart); err != nil {
		return err
	}

	if h.chart.Metadata.Deprecated {
		log.Warn("This chart is deprecated")
	}

	// Add inject dependencies.
	inject := loadComponentChart()
	failInject(inject, h.id)
	h.chart.AddDependency(inject)

	installer.Namespace = h.namespace
	if _, err := installer.Run(h.chart, nil); err != nil {
		return errors.Wrap(err, "INSTALLATION FAILED")
	}
	return nil
}

func (h HelmInstaller) Uninstall(id string) error {
	uninstallClint := action.NewUninstall(h.helmConfig)
	_, err := uninstallClint.Run(id)
	if err != nil {
		err = errors.Wrap(err, "call uninstall err")
		return err
	}

	return nil
}

func (h HelmInstaller) Brief() *InstallerBrief {
	return &h.brief
}

func loadComponentChart() *chart.Chart {
	pullAction := action.NewPull()
	pullAction.RepoURL = _tkeelRepo
	cp, err := pullAction.LocateChart(_componentChartName, &cli.EnvSettings{})
	if err != nil {
		log.Warn("can't get the chart: %s", _componentChartName)
		return nil
	}
	log.Debugf("CHART PATH: %s\n", cp)
	c, err := loader.Load(cp)
	if err != nil {
		log.Warn("can't parse the file %q", cp, err)
		return nil
	}
	if err := checkIfInstallable(c); err != nil {
		log.Warn("uninstallable chart request")
		return nil
	}

	if c.Metadata.Deprecated {
		log.Warn("%q: This chart is deprecated", cp)
	}

	return c
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
