package helm

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/helmpath"
)

func deleteRepo(names ...string) error {
	rf, err := loadRepoFile()
	if err != nil {
		return errors.Wrap(err, "load repo file failed")
	}
	if len(rf.Repositories) == 0 {
		return errors.New("no repositories configured")
	}
	for _, name := range names {
		if !rf.Remove(name) {
			return errors.Errorf("no repo named %q found", name)
		}
		if err = removeRepoCache(env.RepositoryCache, name); err != nil {
			return err
		}
	}
	data, err := yaml.Marshal(rf)
	if err != nil {
		err = errors.Wrap(err, "marshal repo file err")
		return err
	}
	if err := setRepositoryToDapr(data); err != nil {
		err = errors.Wrap(err, "write repository to dapr err")
		return err
	}
	if err := rf.WriteFile(ownRepositoryConfigPath, 0644); err != nil {
		err = errors.Wrap(err, "write repository to local config file err")
		return err
	}
	return nil
}

func removeRepoCache(root, name string) error {
	idx := filepath.Join(root, helmpath.CacheChartsFile(name))
	if _, err := os.Stat(idx); err == nil {
		os.Remove(idx)
	}

	idx = filepath.Join(root, helmpath.CacheIndexFile(name))
	if _, err := os.Stat(idx); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return errors.Wrapf(err, "can't remove index file %s", idx)
	}
	if err := os.Remove(idx); err != nil {
		err = errors.Wrap(err, "call OS remove failed")
		return err
	}

	return nil
}
