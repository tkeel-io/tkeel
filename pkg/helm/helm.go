package helm

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tkeel/pkg/errutil"
	"helm.sh/helm/v3/cmd/helm/search"
	helmAction "helm.sh/helm/v3/pkg/action"
	helmCLI "helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	defaultSelectVersion = ">0.0.0"
	privateRepoName      = "own"
	configFilename       = "repositories.yaml"
	tkeelDir             = ".tkeel"

	repositoryConfig = `apiVersion: ""
generated: "0001-01-01T00:00:00Z"
repositories:
- caFile: ""
  certFile: ""
  insecure_skip_tls_verify: false
  keyFile: ""
  name: tkeel
  pass_credentials_all: false
  password: ""
  url: https://tkeel-io.github.io/helm-charts
  username: ""`
)

var (
	env                     = helmCLI.New()
	defaultCfg, _           = getConfiguration()
	ownRepositoryConfigPath = checkRepositoryConfigPath()

	driver    = "secret"
	namespace = "tkeel"

	errNoRepositories = errors.New("no repositories found. You must add one before updating")
)

func checkRepositoryConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	repoConfigPath := filepath.Join(home, tkeelDir, configFilename)
	if _, err = os.Stat(repoConfigPath); !os.IsNotExist(err) {
		return repoConfigPath
	}

	if err = os.MkdirAll(filepath.Join(home, tkeelDir), os.ModePerm); err != nil {
		if !os.IsExist(err) {
			log.Fatal(err)
		}
	}

	f, err := os.Create(repoConfigPath)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := f.WriteString(repositoryConfig); err != nil {
		log.Fatal(err)
	}

	return repoConfigPath
}

func SetDriver(name string) error {
	var err error
	driver = name
	defaultCfg, err = getConfiguration()
	return err
}

func GetUsingDriver() string {
	return driver
}

func SetNamespace(name string) error {
	var err error
	namespace = name
	defaultCfg, err = getConfiguration()
	return err
}

func GetUsingNamespace() string {
	return namespace
}

func loadRepoFile() (*repo.File, error) {
	rf, err := repo.LoadFile(env.RepositoryConfig)
	switch {
	case errutil.IsNotExist(err):
		return nil, errNoRepositories
	case err != nil:
		return nil, errors.Wrapf(err, "failed loading file: %s", env.RepositoryConfig)
	case len(rf.Repositories) == 0:
		return nil, errNoRepositories
	}
	if err != nil {
		err = errors.Wrap(err, "load repo config err")
	}
	return rf, err
}

func buildIndex() (*search.Index, error) {
	rf, err := loadRepoFile()
	if err != nil {
		return nil, errors.Wrap(err, "load helm repo config file failed")
	}

	i := search.NewIndex()
	for _, re := range rf.Repositories {
		n := re.Name
		f := filepath.Join(env.RepositoryCache, helmpath.CacheIndexFile(n))
		ind, err := repo.LoadIndexFile(f)
		if err != nil {
			log.Warn("Repo %q is corrupt or missing. Try 'helm repo update'.", n)
			log.Warn("%s", err)
			continue
		}

		i.AddRepo(n, ind, true)
	}
	return i, nil
}

func applyConstraint(version string, res []*search.Result) ([]*search.Result, error) {
	if version == "" {
		return res, nil
	}

	constraint, err := semver.NewConstraint(version)
	if err != nil {
		return res, errors.Wrap(err, "an invalid version/constraint format")
	}

	data := res[:0]
	foundNames := map[string]bool{}
	for _, r := range res {
		// if not returning all versions and already have found a result,
		// you're done!
		if foundNames[r.Name] {
			continue
		}
		v, err := semver.NewVersion(r.Chart.Version)
		if err != nil {
			continue
		}
		if constraint.Check(v) {
			data = append(data, r)
			foundNames[r.Name] = true
		}
	}

	return data, nil
}

func getConfiguration() (*helmAction.Configuration, error) {
	helmConf := new(helmAction.Configuration)
	flags := &genericclioptions.ConfigFlags{
		Namespace: &namespace,
	}
	err := helmConf.Init(flags, namespace, driver, getLog())
	if err != nil {
		err = fmt.Errorf("helmAction configuration init err:%w", err)
	}
	env.RepositoryConfig = ownRepositoryConfigPath
	return helmConf, err
}

func getLog() helmAction.DebugLog {
	return func(format string, v ...interface{}) {
		log.Infof(format, v...)
	}
}

func isNotExist(err error) bool {
	return os.IsNotExist(errors.Cause(err))
}
