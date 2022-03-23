package helm

import (
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const srcYamlName = "src.yaml"

const daprKustomizeFormat = `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

commonAnnotations:
  dapr.io/enabled: "true"
  dapr.io/app-id: "%s"
  dapr.io/app-port: "%s"
  dapr.io/config: %s
resources:
  - %s
`

const pluginLabelKustomizeFormat = `
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

commonLabels:
  tkeel-plugin: %s

resources:
  - %s
`

func kustomizationRenderer(src map[string]interface{}, formatStr string, args ...interface{}) (map[string]interface{}, error) {
	fSys := filesys.MakeEmptyDirInMemory()
	if err := generateKustomizeFiles(fSys, src, formatStr, args...); err != nil {
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

func generateKustomizeFiles(fSys filesys.FileSystem, src map[string]interface{}, formatStr string, args ...interface{}) error {
	o, err := yaml.Marshal(src)
	if err != nil {
		return errors.Wrapf(err, "marshal yaml(%v)", src)
	}
	fSys.WriteFile("kustomization.yaml", []byte(fmt.Sprintf(formatStr, args...)))
	fSys.WriteFile(srcYamlName, o)
	return nil
}
