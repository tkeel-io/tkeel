package helm

import (
	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"os"
)

func addRepo(name, url string) error {
	b, err := getRepositoryFormDapr()
	if err != nil && !os.IsNotExist(err) {
		err = errors.Wrap(err, "open a file err")
		return err
	}

	var f repo.File
	if err = yaml.Unmarshal(b, &f); err != nil {
		err = errors.Wrap(err, "unmarshal yaml err")
		return err
	}

	c := repo.Entry{
		Name: name,
		URL:  url,
	}

	// always force update repo file.
	r, err := repo.NewChartRepository(&c, getter.All(env))
	if err != nil {
		err = errors.Wrap(err, "new chart repository err")
		return err
	}

	if env.RepositoryCache != "" {
		r.CachePath = env.RepositoryCache
	}
	if _, err := r.DownloadIndexFile(); err != nil {
		return errors.Wrapf(err, "looks like %q is not a valid chart repository or cannot be reached", url)
	}

	f.Update(&c)

	data, err := yaml.Marshal(f)
	if err != nil {
		err = errors.Wrap(err, "yaml marshal repo.File err")
		return err
	}
	if err = setRepositoryToDapr(data); err != nil {
		err = errors.Wrap(err, "write repository to dapr err")
		return err
	}

	if err := syncRepositoriesConfig(data); err != nil {
		return err
	}
	log.Infof("%q has been added to your repositories\n", name)
	return nil
}
