package helm

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/flock"
	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

func addRepo(name, url string) error {
	// Ensure the file directory exists as it is required for file locking.
	err := os.MkdirAll(filepath.Dir(env.RepositoryConfig), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		err = errors.Wrap(err, "mkdir "+env.RepositoryConfig+" err")
		return err
	}
	// Acquire a file lock for process synchronization.
	repoFileExt := filepath.Ext(env.RepositoryConfig)
	var lockPath string
	if len(repoFileExt) > 0 && len(repoFileExt) < len(env.RepositoryConfig) {
		lockPath = strings.Replace(env.RepositoryConfig, repoFileExt, ".lock", 1)
	} else {
		lockPath = env.RepositoryConfig + ".lock"
	}
	fileLock := flock.New(lockPath)
	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	locked, err := fileLock.TryLockContext(lockCtx, time.Second)
	if err == nil && locked {
		defer fileLock.Unlock()
	}
	if err != nil {
		err = errors.Wrap(err, "try to lock err")
		return err
	}

	b, err := ioutil.ReadFile(env.RepositoryConfig)
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

	if err := f.WriteFile(env.RepositoryConfig, 0644); err != nil {
		err = errors.Wrap(err, "write data to file err")
		return err
	}
	log.Infof("%q has been added to your repositories\n", name)
	return nil
}
