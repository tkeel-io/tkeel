package helm

import (
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const kustomizeFormat = `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

commonAnnotations:
  dapr.io/enabled: "true"
  dapr.io/app-id: "%s"
  dapr.io/app-port: "%s"
  dapr.io/config: %s
resources:
  - deployment.yaml
`

func kustomizationRenderer(deployment map[string]interface{}, appID, appPort string) (map[string]interface{}, error) {
	fSys := filesys.MakeEmptyDirInMemory()
	if err := generateKustomizeFiles(fSys, deployment, appID, appPort); err != nil {
		return nil, errors.Wrap(err, "generate kustomize files")
	}
	k := krusty.MakeKustomizer(
		krusty.MakeDefaultOptions(),
	)
	m, err := k.Run(fSys, ".")
	if err != nil {
		return nil, errors.Wrap(err, "kustomize run")
	}
	b, err := m.AsYaml()
	if err != nil {
		return nil, errors.Wrap(err, "res map as yaml")
	}
	out := make(map[string]interface{})
	if err = yaml.Unmarshal(b, &out); err != nil {
		return nil, errors.Wrapf(err, "yaml unmarshal deployment(%s)", b)
	}
	return out, nil
}

func generateKustomizeFiles(fSys filesys.FileSystem, deployment map[string]interface{}, appID, appPort string) error {
	o, err := yaml.Marshal(deployment)
	if err != nil {
		return errors.Wrapf(err, "marshal yaml(%v)", deployment)
	}
	fSys.WriteFile("kustomization.yaml", []byte(fmt.Sprintf(kustomizeFormat, appID, appPort, appID)))
	fSys.WriteFile("deployment.yaml", o)
	return nil
}
