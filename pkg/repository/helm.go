package repository

import (
	"fmt"
	"strings"

	"helm.sh/helm/v3/pkg/getter"

	"github.com/pkg/errors"

	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/tkeel/pkg/client/dapr"
	helmAction "helm.sh/helm/v3/pkg/action"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const _indexFileName = "index.yaml"

type Driver string

func (d Driver) String() string {
	return string(d)
}

var _ Repository = &HelmRepo{}

const (
	Secret    Driver = "secret"
	ConfigMap Driver = "configmap"
	Mem       Driver = "memory"
	SQL       Driver = "sql"
)

type HelmRepo struct {
	info *Info

	actionConfig *helmAction.Configuration
	daprClient   *dapr.Client
	httpGetter   getter.Getter
	driver       Driver
	namespace    string
}

func NewHelmRepo(info Info, driver Driver, namespace string) (*HelmRepo, error) {
	httpGetter, err := getter.NewHTTPGetter()
	if err != nil {
		log.Warn("init helm action configuration err", err)
		return nil, err
	}
	repo := &HelmRepo{
		info:       &info,
		namespace:  namespace,
		driver:     driver,
		httpGetter: httpGetter,
	}
	if err := repo.setActionConfig(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *HelmRepo) SetInfo(info Info) {
	r.info = &info
}

func (r *HelmRepo) DaprClient() *dapr.Client {
	return r.daprClient
}

func (r *HelmRepo) SetDaprClient(daprClient *dapr.Client) {
	r.daprClient = daprClient
}

func (r *HelmRepo) Namespace() string {
	return r.namespace
}

func (r *HelmRepo) SetNamespace(namespace string) error {
	r.namespace = namespace
	return r.setActionConfig()
}

func (r *HelmRepo) SetDriver(driver Driver) error {
	r.driver = driver
	return r.setActionConfig()
}

func (r HelmRepo) GetDriver() Driver {
	return r.driver
}

func (r *HelmRepo) setActionConfig() error {
	config, err := initActionConfig(r.namespace, r.driver)
	if err != nil {
		log.Warn("init helm action configuration err", err)
		return err
	}
	r.actionConfig = config
	return nil
}

func (r *HelmRepo) Info() *Info {
	return r.info
}

func (r *HelmRepo) Search(word string) ([]*InstallerBrief, error) {
	index, err := r.buildIndex()
	if err != nil {
		return nil, errors.Wrap(err, "can't build helm index config")
	}

	res := index.Search(word)
	return res.ToInstallerBrief(), nil
}

func (r *HelmRepo) Get(name, version string) (Installer, error) {
	// TODO implement me
	panic("implement me")
}

func (r *HelmRepo) Installed() []Installer {
	// TODO implement me
	panic("implement me")
}

func (r *HelmRepo) Close() error {
	// TODO implement me
	panic("implement me")
}

func (r *HelmRepo) buildIndex() (*Index, error) {
	fileContent, err := r.QueryIndex()
	if err != nil {
		return nil, err
	}
	return NewIndex(r.info.Name, fileContent)
}

func (r *HelmRepo) QueryIndex() ([]byte, error) {
	url := strings.TrimSuffix(r.info.URL, "/")
	url += "/" + _indexFileName

	buf, err := r.httpGetter.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "HTTP GET error")
	}

	return buf.Bytes(), nil
}

func getDebugLogFunc() helmAction.DebugLog {
	return func(format string, v ...interface{}) {
		log.Infof(format, v...)
	}
}

func initActionConfig(namespace string, driver Driver) (*helmAction.Configuration, error) {
	config := new(helmAction.Configuration)
	k8sFlags := &genericclioptions.ConfigFlags{
		Namespace: &namespace,
	}
	err := config.Init(k8sFlags, namespace, driver.String(), getDebugLogFunc())
	if err != nil {
		return nil, fmt.Errorf("helmAction configuration init err:%w", err)
	}
	return config, nil
}
