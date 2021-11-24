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

	security_conf "github.com/tkeel-io/security/pkg/apiserver/config"

	"gopkg.in/yaml.v2"
)

// TkeelConf tkeel platform configuration.
type TkeelConf struct {
	// tkeel platform secret. set up when installing the platform.
	Secret string `json:"secret" yaml:"secret"`
	// tkeel platform version.
	Version string `json:"version" yaml:"version"`
	// watch plugin route interval.
	WatchPluginRouteInterval string `json:"watch_plugin_route_interval" yaml:"watchPluginRouteInterval"`
}

// DaprConf dapr sidecar configuration.
type DaprConf struct {
	// dapr sidecar grpc listen port.
	GRPCPort string `json:"grpc_port" yaml:"grpcPort"`
	// dapr sidecar http listen port.
	HTTPPort string `json:"http_port" yaml:"httpPort"`
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

// SecurityConf.
type SecurityConf struct {
	// Mysql  mysql config of security.
	Mysql *security_conf.MysqlConf `json:"mysql" yaml:"mysql"`
	// RBAC rbac config of security.
	RBAC *security_conf.RBACConfig `json:"rbac" yaml:"rbac"`
	// OAuth2Config oauth2 config of security.
	OAuth2 *security_conf.OAuth2Config `json:"oauth2" yaml:"oauth2"` // nolint
	// entity entity security config of auth.
	Entity *security_conf.EntityConfig `json:"entity" yaml:"entity"`
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
	// SecurityConf security auth config.
	SecurityConf *SecurityConf `json:"security_conf" yaml:"securityConf"`
}

// NewDefaultConfiguration returns the empty config.
func NewDefaultConfiguration() *Configuration {
	return &Configuration{
		Tkeel: &TkeelConf{},
		Dapr:  &DaprConf{},
		Log:   &LogConf{},
		SecurityConf: &SecurityConf{
			Mysql:  &security_conf.MysqlConf{},
			RBAC:   &security_conf.RBACConfig{Adapter: &security_conf.MysqlConf{}},
			OAuth2: &security_conf.OAuth2Config{Redis: &security_conf.RedisConf{}, AccessGenerate: &security_conf.AccessConf{}},
			Entity: &security_conf.EntityConfig{},
		},
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

func (c *Configuration) AttachCmdFlags(strVar func(p *string, name string, value string, usage string),
	boolVar func(p *bool, name string, value bool, usage string), intVar func(p *int, name string, value int, usage string)) {
	boolVar(&c.Log.Dev, "debug", getEnvBool("TMANAGER_DEBUG", false), "enable debug mod.")
	strVar(&c.Log.Level, "log.level", getEnvStr("TMANAGER_LOG_LEVEL", "debug"), "log level(default debug).")
	strVar(&c.HTTPAddr, "http.addr", getEnvStr("TMANAGER_HTTP_ADDR", ":31234"), "http listen address(default :31234).")
	strVar(&c.GRPCAddr, "grpc.addr", getEnvStr("TMANAGER_GRPC_ADDR", ":31233"), "grpc listen address(default :31233).")
	strVar(&c.Dapr.GRPCPort, "dapr.grpc.port", getEnvStr("DAPR_GRPC_PORT", "50001"), "dapr grpc listen address(default 50001).")
	strVar(&c.Dapr.GRPCPort, "dapr.http.port", getEnvStr("DAPR_HTTP_PORT", "3500"), "dapr grpc listen address(default 3500).")
	strVar(&c.Dapr.PrivateStateName, "dapr.private_state_name", getEnvStr("DAPR_PRIVATE_STATE_NAME", "keel-private-store"), "dapr private store name(default keel-private-store).")
	strVar(&c.Dapr.PublicStateName, "dapr.public_state_name", getEnvStr("DAPR_PUBLIC_STATE_NAME", "keel-public-store"), "dapr public store name(default keel-public-store).")
	strVar(&c.Tkeel.Secret, "tkeel.secret", getEnvStr("TKEEL_SECRET", "changeme"), "tkeel secret.(default changeme)")
	strVar(&c.Tkeel.Version, "tkeel.version", getEnvStr("TKEEL_VERSION", "v0.2.0"), "tkeel version.(default v0.2.0)")
	strVar(&c.Tkeel.WatchPluginRouteInterval, "tkeel.watch_plugin_route_interval", getEnvStr("TKEEL_WATCH_PLUGIN_ROUTE_INTERVAL", "10s"), "tkeel watch plugin route change interval.(default 10s)")
	strVar(&c.SecurityConf.Mysql.DBName, "security.mysql.dbname", getEnvStr("SECURITY_MYSQL_DBNAME", "tkeelauth"), "database name of auth`s mysql config")
	strVar(&c.SecurityConf.Mysql.User, "security.mysql.user", getEnvStr("SECURITY_MYSQL_USER", "root"), "user name of auth`s mysql config")
	strVar(&c.SecurityConf.Mysql.Password, "security.mysql.password", getEnvStr("SECURITY_MYSQL_PASSWORD", "123456"), "password of auth`s mysql config")
	strVar(&c.SecurityConf.Mysql.Host, "security.mysql.host", getEnvStr("SECURITY_MYSQL_HOST", "127.0.0.1"), "host of auth`s mysql config")
	strVar(&c.SecurityConf.Mysql.Port, "security.mysql.port", getEnvStr("SECURITY_MYSQL_PORT", "3306"), "port of auth`s mysql config")
	strVar(&c.SecurityConf.OAuth2.Redis.Addr, "security.redis.addr", getEnvStr("SECURITY_REDIS_ADDR", "127.0.0.1:6379"), "address of auth`s redis config")
	intVar(&c.SecurityConf.OAuth2.Redis.DB, "security.redis.db", getEnvInt("SECURITY_REDIS_DB", 0), "db of auth`s redis")
	strVar(&c.SecurityConf.OAuth2.AccessGenerate.SecurityKey, "security.access.sk", getEnvStr("SECURITY_ACCESS_SK", "00000000"), "security key of auth`s access generate")
	strVar(&c.SecurityConf.Entity.SecurityKey, "security.entity.sk", getEnvStr("SECURITY_ENTITY_SK", "99999999"), "security  key auth`s entity token access")
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

func getEnvInt(env string, defaultValue int) int {
	v := os.Getenv(env)
	if v == "" {
		return defaultValue
	}
	ret, err := strconv.Atoi(v)
	if err != nil {
		panic(fmt.Errorf("error get env(%s) int: %w", env, err))
	}
	return ret
}
