package config

import (
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

// Open api only supports HTTP protocol
var (
	DefaultHTTPPort     = 8080
	DefaultDaprHTTPPort = os.Getenv("DAPR_HTTP_PORT")
	DefaultSecretKey    = "zhu88jie"
)

// Configuration is required by all plug-ins in keel.
type Configuration struct {
	Plugin PluginSpec `json:"plugin,omitempty" yaml:"plugin,omitempty"`
}

// PluginSpec describes plugin information
type PluginSpec struct {
	ID      string `json:"id" yaml:"id"`
	Version string `json:"version" yaml:"version"`
	Port    int    `json:"port" yaml:"port"`
}

// LoadDefaultConfiguration returns the default config.
func LoadDefaultConfiguration() *Configuration {
	return &Configuration{
		Plugin: PluginSpec{
			Port: DefaultHTTPPort,
		},
	}
}

// LoadStandaloneConfiguration gets the path to a config file and loads it into a configuration.
func LoadStandaloneConfiguration(config string) (*Configuration, error) {
	_, err := os.Stat(config)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadFile(config)
	if err != nil {
		return nil, err
	}

	// Parse environment variables from yaml
	b = []byte(os.ExpandEnv(string(b)))

	conf := LoadDefaultConfiguration()
	err = yaml.Unmarshal(b, conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}
