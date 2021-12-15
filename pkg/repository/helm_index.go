package repository

import (
	"path"
	"strings"

	"github.com/tkeel-io/kit/log"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
)

// verSep is a separator for version fields in map keys.
const verSep = "$$"

type PluginRes struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Repo        string   `json:"repository"`
	URLs        []string `json:"urls"`
	Description string   `json:"description"`
}

func (r PluginRes) ToInstallerBrief() *InstallerBrief {
	return &InstallerBrief{
		Name:      r.Name,
		Repo:      r.Repo,
		Version:   r.Version,
		Installed: false,
	}
}

type PluginResList []*PluginRes

func (r *PluginResList) ToInstallerBrief() []*InstallerBrief {
	list := make([]*InstallerBrief, 0, len(*r))
	for _, res := range *r {
		list = append(list, res.ToInstallerBrief())
	}
	return list
}

type Index struct {
	RepoName  string
	helmIndex *repo.IndexFile
	charts    map[string]*repo.ChartVersion
}

// NewIndex creates a new Index.
func NewIndex(repoName string, data []byte) (*Index, error) {
	i := &repo.IndexFile{}

	if len(data) == 0 {
		return nil, repo.ErrEmptyIndexYaml
	}

	if err := yaml.UnmarshalStrict(data, i); err != nil {
		return nil, err
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

	index := &Index{
		helmIndex: i,
		charts:    make(map[string]*repo.ChartVersion),
	}

	for name, ref := range i.Entries {
		if len(ref) == 0 {
			continue
		}

		fname := path.Join(repoName, name)
		for _, rr := range ref {
			versionedName := fname + verSep + rr.Version
			index.charts[versionedName] = rr
		}
	}

	return index, nil
}

func (r *Index) Search(word string, version string) PluginResList {
	list := make(PluginResList, 0, len(r.helmIndex.Entries))
	if word == "*" {
		for name, ch := range r.charts {
			res := PluginRes{
				Name:        strings.Split(name, verSep)[0],
				Version:     ch.Version,
				Repo:        r.RepoName,
				URLs:        ch.URLs,
				Description: ch.Description,
			}
			list = append(list, &res)
		}
		return list
	}

	for name, ch := range r.charts {
		if name == word {
			if version == "" || version == ch.Version {
				res := PluginRes{
					Name:        ch.Name,
					Version:     ch.Version,
					Repo:        r.RepoName,
					URLs:        ch.URLs,
					Description: ch.Description,
				}
				list = append(list, &res)
			}
		}
	}

	return list
}
