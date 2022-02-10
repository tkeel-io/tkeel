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
	"regexp"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tkeel/pkg/repository"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"
)

var _getter getter.Getter

func init() {
	g, err := getter.NewHTTPGetter()
	if err != nil {
		log.Fatal(err)
	}
	_getter = g
}

type PluginRes struct {
	Name        string             `json:"name"`
	Version     string             `json:"version"`
	Repo        string             `json:"repository"` // nolint
	URLs        []string           `json:"urls"`       //nolint
	Description string             `json:"description"`
	ChartInfo   *repo.ChartVersion `json:"chart_info"`
}

func (r PluginRes) ToInstallerBrief() *repository.InstallerBrief {
	return &repository.InstallerBrief{
		Name:        r.Name,
		Repo:        r.Repo,
		Version:     r.Version,
		Installed:   false,
		Desc:        r.ChartInfo.Description,
		Annotations: r.ChartInfo.Annotations,
		Maintainers: func() []*repository.Maintainer {
			ret := make([]*repository.Maintainer, 0, len(r.ChartInfo.Maintainers))
			for _, v := range r.ChartInfo.Maintainers {
				ret = append(ret, &repository.Maintainer{
					Name:  v.Name,
					URL:   v.URL,
					Email: v.Email,
				})
			}
			return ret
		}(),
		CreateTimestamp: r.ChartInfo.Created.Unix(),
		Icon:            r.ChartInfo.Icon,
	}
}

type PluginResList []*PluginRes

func (r *PluginResList) ToInstallerBrief() []*repository.InstallerBrief {
	list := make([]*repository.InstallerBrief, 0, len(*r))
	for _, res := range *r {
		list = append(list, res.ToInstallerBrief())
	}
	return list
}

type Index struct {
	URL       string
	RepoName  string
	helmIndex *repo.IndexFile
	charts    map[string]map[string]*repo.ChartVersion
	lock      *sync.RWMutex
}

// NewIndex creates a new Index.
func NewIndex(url, repoName string) (*Index, error) {
	i, err := getIndex(url, _getter)
	if err != nil {
		return nil, errors.Wrapf(err, "get repository(%s) index", url)
	}
	index := &Index{
		URL:       url,
		RepoName:  repoName,
		helmIndex: i,
		charts:    make(map[string]map[string]*repo.ChartVersion),
		lock:      new(sync.RWMutex),
	}
	for name, ref := range i.Entries {
		if len(ref) == 0 {
			continue
		}
		for _, rr := range ref {
			versionMap, ok := index.charts[name]
			if !ok {
				versionMap = make(map[string]*repo.ChartVersion)
				index.charts[name] = versionMap
			}
			versionMap[rr.Version] = rr
		}
	}
	return index, nil
}

func (r *Index) Search(word string, version string) (PluginResList, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	list := make(PluginResList, 0, len(r.helmIndex.Entries))
	if word == "*" {
		for _, vMap := range r.charts {
			for _, ch := range vMap {
				if _, ok := ch.Metadata.Annotations[tKeelPluginEnableKey]; ok {
					res := PluginRes{
						Name:        ch.Name,
						Version:     ch.Version,
						Repo:        r.RepoName,
						URLs:        ch.URLs,
						Description: ch.Description,
						ChartInfo:   ch,
					}
					list = append(list, &res)
				}
			}
		}
		return list, nil
	}
	exp, err := regexp.Compile(word)
	if err != nil {
		return nil, errors.Wrapf(err, "%s is not a valid regular expression", word)
	}
	for chartName, vMap := range r.charts {
		if exp.MatchString(chartName) {
			for _, ch := range vMap {
				if version == "" || version == ch.Version {
					if _, ok := ch.Metadata.Annotations[tKeelPluginEnableKey]; ok {
						res := PluginRes{
							Name:        ch.Name,
							Version:     ch.Version,
							Repo:        r.RepoName,
							URLs:        ch.URLs,
							Description: ch.Description,
							ChartInfo:   ch,
						}
						list = append(list, &res)
					}
				}
			}
		}
	}
	return list, nil
}

func (r *Index) Update() (bool, error) {
	iFile, err := getIndex(r.URL, _getter)
	if err != nil {
		return false, errors.Wrapf(err, "get repository(%s) index", r.URL)
	}
	if iFile.Generated.After(r.helmIndex.Generated) {
		r.helmIndex = iFile
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	for name, ref := range iFile.Entries {
		if len(ref) == 0 {
			continue
		}
		for _, rr := range ref {
			versionMap, ok := r.charts[name]
			if !ok {
				versionMap = make(map[string]*repo.ChartVersion)
				r.charts[name] = versionMap
			}
			versionMap[rr.Version] = rr
		}
	}
	for k, vMap := range r.charts {
		vList, ok := iFile.Entries[k]
		if !ok {
			delete(r.charts, k)
		} else {
			for ver := range vMap {
				exist := false
				for _, v := range vList {
					if v.Version == ver {
						exist = true
						break
					}
				}
				if !exist {
					delete(vMap, ver)
				}
			}
		}
	}
	return true, nil
}

// getIndex get the repo index.yaml file content.
func getIndex(url string, g getter.Getter) (*repo.IndexFile, error) {
	url = strings.TrimSuffix(url, "/") + "/" + _indexFileName
	buf, err := g.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "HTTP GET %s error", url)
	}
	i := &repo.IndexFile{}
	if buf.Len() == 0 {
		return nil, repo.ErrEmptyIndexYaml
	}
	if err := yaml.UnmarshalStrict(buf.Bytes(), i); err != nil {
		return nil, errors.Wrap(err, "unmarshal data to IndexFile failed")
	}
	for name, cvs := range i.Entries {
		for idx := len(cvs) - 1; idx >= 0; idx-- {
			if cvs[idx].APIVersion == "" {
				cvs[idx].APIVersion = chart.APIVersionV1
			}
			if err := cvs[idx].Validate(); err != nil {
				log.Infof("skipping loading invalid entry for chart %q %q : %s", name, cvs[idx].Version, err)
				cvs = append(cvs[:idx], cvs[idx+1:]...)
			}
		}
	}
	i.SortEntries()
	if i.APIVersion == "" {
		return nil, repo.ErrNoAPIVersion
	}
	return i, nil
}
