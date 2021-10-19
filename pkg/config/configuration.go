package config

import (
	"fmt"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

// Open api only supports HTTP protocol.
var (
	DefaultHTTPPort     = 8080
	DefaultDaprHTTPPort = os.Getenv("DAPR_HTTP_PORT")
	DefaultSecretKey    = "zhu88jie"
)

// Configuration is required by all plug-ins in keel.
type Configuration struct {
	Plugin *PluginSpec `json:"plugin,omitempty" yaml:"plugin,omitempty"`
	Tkeel  *TkeelSpec  `json:"tkeel,omitempty" yaml:"tkeel,omitempty"`
}

// PluginSpec describes plugin information.
type PluginSpec struct {
	ID      string `json:"id" yaml:"id"`
	Version string `json:"version" yaml:"version"`
	Port    int    `json:"port" yaml:"port"`
}

type TkeelSpec struct {
	Version string `json:"version" yaml:"version"`
}

// LoadDefaultConfiguration returns the default config.
func LoadDefaultConfiguration() *Configuration {
	return &Configuration{
		Plugin: &PluginSpec{
			Port: DefaultHTTPPort,
		},
		Tkeel: &TkeelSpec{},
	}
}

// LoadStandaloneConfiguration gets the path to a config file and loads it into a configuration.
func LoadStandaloneConfiguration(config string) (*Configuration, error) {
	_, err := os.Stat(config)
	if err != nil {
		return nil, fmt.Errorf("error os Stat: %w", err)
	}

	b, err := ioutil.ReadFile(config)
	if err != nil {
		return nil, fmt.Errorf("error ioutil readfile: %w", err)
	}

	// Parse environment variables from yaml.
	b = []byte(os.ExpandEnv(string(b)))

	conf := LoadDefaultConfiguration()
	err = yaml.Unmarshal(b, conf)
	if err != nil {
		return nil, fmt.Errorf("error yaml unmarshal: %w", err)
	}

	return conf, nil
}
