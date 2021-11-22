/*
Copyright 2021 The tKeel Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
	http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

// TkeelConf tkeel platform configuration.
type TkeelConf struct {
	// tkeel platform secret. set up when installing the platform.
	Secret string `json:"secret" yaml:"secret"`
	// tkeel platform version.
	Version string `json:"version" yaml:"version"`
}

// DaprConf dapr sidecar configuration.
type DaprConf struct {
	// dapr sidecar grpc listion port.
	GRPCPort string `json:"grpc_port" yaml:"grpcPort"`
	// private state name.
	PrivateStateName string `json:"private_state_name" yaml:"privateStateName"`
	// public state name.
	PublicStateName string `json:"public_state_name" yaml:"publicStateName"`
}

// LogConf log configuration.
type LogConf struct {
	// log level.
	Level  string   `json:"level" yaml:"level"`
	Dev    bool     `json:"dev" yaml:"dev"`
	Output []string `json:"output" yaml:"output"`
}

// Configuration.
type Configuration struct {
	// HTTPAddr http server listen address.
	HTTPAddr string `json:"http_addr" yaml:"httpAddr"`
	// GRPCAddr grpc server listen address.
	GRPCAddr string `json:"grpc_addr" yaml:"grpcAddr"`
	// Tkeel tkeel configuration.
	Tkeel *TkeelConf `json:"tkeel" yaml:"tkeel"`
	// Dapr dapr configuration.
	Dapr *DaprConf `json:"dapr" yaml:"dapr"`
	// Log log configuration.
	Log *LogConf `json:"log" yaml:"log"`
}

// NewDefaultConfiguration returns the empty config.
func NewDefaultConfiguration() *Configuration {
	return &Configuration{
		Tkeel: &TkeelConf{},
		Dapr:  &DaprConf{},
		Log:   &LogConf{},
	}
}

// LoadStandaloneConfiguration gets the path to a config file and loads it into a configuration.
func LoadStandaloneConfiguration(configPath string) (*Configuration, error) {
	_, err := os.Stat(configPath)
	if err != nil {
		return nil, fmt.Errorf("error os Stat: %w", err)
	}

	b, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error ioutil readfile: %w", err)
	}

	// Parse environment variables from yaml.
	b = []byte(os.ExpandEnv(string(b)))

	conf := NewDefaultConfiguration()
	err = yaml.Unmarshal(b, conf)
	if err != nil {
		return nil, fmt.Errorf("error yaml unmarshal: %w", err)
	}

	return conf, nil
}

func (c *Configuration) AttachFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&c.Log.Dev, "debug", getEnvBool("RUDDER_DEBUG", false), "enable debug mod.")
	fs.StringVar(&c.Log.Level, "log.level", getEnvStr("RUDDER_LOG_LEVEL", "debug"), "log level(default debug).")
	fs.StringVar(&c.HTTPAddr, "http.addr", getEnvStr("RUDDER_HTTP_ADDR", ":31234"), "http listen address(default :31234).")
	fs.StringVar(&c.GRPCAddr, "grpc.addr", getEnvStr("RUDDER_GRPC_ADDR", ":31233"), "grpc listen address(default :31233).")
	fs.StringVar(&c.Dapr.GRPCPort, "dapr.port", getEnvStr("DAPR_GRPC_PORT", "50001"), "dapr grpc listen address(default 50001).")
	fs.StringVar(&c.Dapr.PrivateStateName, "dapr.private_state_name", getEnvStr("DAPR_PRIVATE_STATE_NAME", "keel-private-store"), "dapr private store name(default keel-private-store).")
	fs.StringVar(&c.Dapr.PublicStateName, "dapr.public_state_name", getEnvStr("DAPR_PUBLIC_STATE_NAME", "keel-public-store"), "dapr public store name(default keel-public-store).")
	fs.StringVar(&c.Tkeel.Secret, "tkeel.secret", getEnvStr("TKEEL_SECRET", "changeme"), "tkeel secret.(default changeme)")
	fs.StringVar(&c.Tkeel.Version, "tkeel.version", getEnvStr("TKEEL_VERSION", "v0.2.0"), "tkeel version.(default v0.2.0)")
}

func getEnvStr(env string, defaultValue string) string {
	v := os.Getenv(env)
	if v == "" {
		return defaultValue
	}
	return v
}

func getEnvBool(env string, defaultValue bool) bool {
	v := os.Getenv(env)
	if v == "" {
		return defaultValue
	}
	ret, err := strconv.ParseBool(v)
	if err != nil {
		panic(fmt.Errorf("error get env(%s) bool: %w", env, err))
	}
	return ret
}
