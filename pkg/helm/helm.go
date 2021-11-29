package helm

import (
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"helm.sh/helm/v3/pkg/chart/loader"
	"github.com/Masterminds/semver/v3"
	dapr "github.com/dapr/go-sdk/client"
	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/log"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/cmd/helm/search"
	helmAction "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
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
	tkeelRepo            = "https://tkeel-io.github.io/helm-charts"

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
	componentChart          = loadComponentChart()

	componentChartName = "tkeel-plugin-components"
	driver             = "secret"
	namespace          = "tkeel"
	daprStoreName      = "keel-private-store"

	errNoRepositories              = errors.New("no repositories found. You must add one before updating")
	errNoDaprClientInit            = errors.New("no dapr client init")
	errNoRepositoryConfigFileExist = errors.New("no repository config file exist")

	daprClient *dapr.Client
)

func SetDaprConfig(client *dapr.Client, storeName string) {
	daprClient = client
	daprStoreName = storeName
}

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

	configFormDapr, err := getRepositoryFromDapr()
	if err != nil {
		log.Fatal(err)
	}

	if _, err := f.Write(configFormDapr); err != nil {
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
	repoConf, err := getRepositoryFromDapr()
	if err != nil {
		err = errors.Wrap(err, "failed try to get repository.yaml config")
		return nil, err
	}
	rf, err := newHelmRepoFile(repoConf)
	if err != nil {
		return nil, errors.Wrap(err, "new repository.yaml as repo.File failed")
	}
	if len(rf.Repositories) == 0 {
		return nil, errNoRepositories
	}
	return rf, nil
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

func getRepositoryFromDapr() ([]byte, error) {
	if daprClient == nil {
		return nil, errNoDaprClientInit
	}

	item, err := (*daprClient).GetState(context.Background(), daprStoreName, configFilename)
	if err != nil {
		err = errors.Wrap(err, "get state form dapr err")
		return nil, err
	}
	if len(item.Value) == 0 {
		if err := setRepositoryToDapr([]byte(repositoryConfig)); err != nil {
			err = errors.Wrap(err, "set repository config to dapr status err")
			return nil, err
		}
		return []byte(repositoryConfig), nil
	}
	return item.Value, nil
}

func setRepositoryToDapr(content []byte) error {
	if daprClient == nil {
		return errNoDaprClientInit
	}
	if err := (*daprClient).SaveState(context.Background(), daprStoreName, configFilename, content); err != nil {
		err = errors.Wrap(err, "save state to dapr err")
		return err
	}
	return nil
}

func newHelmRepoFile(content []byte) (*repo.File, error) {
	r := new(repo.File)

	if err := yaml.Unmarshal(content, r); err != nil {
		err = errors.Wrap(err, "yaml unmarshal err")
		return nil, err
	}
	if err := syncRepositoriesConfig(content); err != nil {
		log.Warn("sync repository config to local file from dapr err", err)
	}
	return r, nil
}

func syncRepositoriesConfig(content []byte) error {
	if _, err := os.Stat(ownRepositoryConfigPath); os.IsNotExist(err) {
		return errNoRepositoryConfigFileExist
	}

	if err := ioutil.WriteFile(ownRepositoryConfigPath, content, fs.ModePerm); err != nil {
		err = errors.Wrap(err, "try write file")
		return err
	}
	return nil
}

func loadComponentChart() *chart.Chart {
	pullAction := helmAction.NewPull()
	pullAction.ChartPathOptions.RepoURL = tkeelRepo
	cp, err := pullAction.ChartPathOptions.LocateChart(componentChartName, env)
	if err != nil {
		log.Warn("can't get the chart: %s", componentChartName)
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
