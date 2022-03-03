package kubernetes

import (
	"context"

	"github.com/pkg/errors"
	"github.com/tkeel-io/kit/config"
	"github.com/tkeel-io/kit/log"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k_kubernetes "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Client struct {
	namespace            string
	deploymentConfigName string
	c                    *k_kubernetes.Clientset
}

func NewClient(configName, namespace string) *Client {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("error rest in cluster config: %s", err)
	}
	clientset, err := k_kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("error new k8s clientset: %s", err)
	}
	return &Client{
		c:                    clientset,
		deploymentConfigName: configName,
		namespace:            namespace,
	}
}

func (c *Client) GetDeploymentConfig(ctx context.Context) (*config.InstallConfig, error) {
	cm, err := c.c.CoreV1().ConfigMaps(c.namespace).Get(ctx, c.deploymentConfigName, v1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "get %s configmap", c.deploymentConfigName)
	}
	conf := &config.InstallConfig{}
	raw := cm.Data["config"]
	if raw != "" {
		if err = yaml.Unmarshal([]byte(raw), conf); err != nil {
			return nil, errors.Wrapf(err, "unmarshal %s", raw)
		}
	}
	return conf, nil
}
