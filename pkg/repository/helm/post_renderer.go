package helm

import (
	"bytes"
	"io"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	kindField         = "kind"
	deploymentKind    = "Deployment"
	metadataField     = "metadata"
	metadataNameField = "name"
	specField         = "spec"
	templateField     = "template"
	annotationsField  = "annotations"
	daprEnableKey     = "dapr.io/enabled"
)

type kustomizeRender struct {
	InjectDeploymentName string
	InjectAppID          string
	InjectAppPort        string
}

func newKustomizeRender(deploymentName, appID, appPort string) *kustomizeRender {
	return &kustomizeRender{
		deploymentName,
		appID,
		appPort,
	}
}

func (kr *kustomizeRender) Run(renderedManifests *bytes.Buffer) (*bytes.Buffer, error) {
	dec := yaml.NewDecoder(renderedManifests)
	out := bytes.NewBuffer(make([]byte, 0))
	enc := yaml.NewEncoder(out)
	defer enc.Close()
	for {
		data := make(map[string]interface{})
		err := dec.Decode(&data)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return nil, errors.Wrap(err, "transform data err")
			}
			break
		}
		if kr.isTargetDeployment(data) {
			newData, err := kustomizationRenderer(data, daprKustomizeFormat,
				kr.InjectAppID, kr.InjectAppPort, kr.InjectAppID, srcYamlName)
			if err != nil {
				return nil, errors.Wrap(err, "kustomization renderer")
			}
			data = newData
		}
		if kr.isDaprPod(data) {
			newData, err := kustomizationRenderer(data, pluginLabelKustomizeFormat,
				kr.InjectAppID, srcYamlName)
			if err != nil {
				return nil, errors.Wrap(err, "kustomization renderer")
			}
			data = newData
		}
		enc.Encode(data)
	}

	return out, nil
}

func (kr *kustomizeRender) isTargetDeployment(in map[string]interface{}) bool {
	ki, ok := in[kindField]
	if !ok {
		return false
	}
	k, ok := ki.(string)
	if !ok {
		return false
	}

	mi, ok := in[metadataField]
	if !ok {
		return false
	}
	m, ok := mi.(map[string]interface{})
	if !ok {
		return false
	}
	ni, ok := m[metadataNameField]
	if !ok {
		return false
	}
	n, ok := ni.(string)
	if !ok {
		return false
	}

	return k == deploymentKind && n == kr.InjectDeploymentName
}

func (kr *kustomizeRender) isDaprPod(in map[string]interface{}) bool {
	si, ok := in[specField]
	if !ok {
		return false
	}
	s, ok := si.(map[string]interface{})
	if !ok {
		return false
	}
	ti, ok := s[templateField]
	if !ok {
		return false
	}
	t, ok := ti.(map[string]interface{})
	if !ok {
		return false
	}
	ai, ok := t[annotationsField]
	if !ok {
		return false
	}
	a, ok := ai.(map[string]interface{})
	if !ok {
		return false
	}
	_, ok = a[daprEnableKey]
	return ok
}
