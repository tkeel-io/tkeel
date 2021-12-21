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

	security_conf "github.com/tkeel-io/security/apiserver/config"

	"gopkg.in/yaml.v2"
)

// TkeelConf tkeel platform configuration.
type TkeelConf struct {
	// tkeel platform namespace.
	Namespace string `json:"namespace" yaml:"namespace"`
	// tkeel platform version.
	Version string `json:"version" yaml:"version"`
	// AdminPassword admin password.
	AdminPassword string `json:"admin_password" yaml:"adminPassword"`
	// watch interval.
	WatchInterval string `json:"watch_interval" yaml:"watchInterval"`
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

// ProxyConf proxy service configuration.
type ProxyConf struct {
	// proxy timeout.
	Timeout string `json:"timeout" yamlL:"timeout"`
	// core address.
	CoreAddr string `json:"core_addr" yaml:"coreAddr"`
	// rudder address.
	RudderAddr string `json:"rudder_addr" yaml:"rudderAddr"`
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
	// Proxy tkeel platform configuration.
	Proxy *ProxyConf `json:"proxy" yaml:"proxy"`
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
		Proxy: &ProxyConf{},
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
	boolVar(&c.Log.Dev, "debug", getEnvBool("TKEEL_DEBUG", false), "enable debug mod.")
	strVar(&c.Log.Level, "log.level", getEnvStr("TKEEL_LOG_LEVEL", "debug"), "log level(default debug).")
	strVar(&c.HTTPAddr, "http.addr", getEnvStr("TKEEL_HTTP_ADDR", ":31234"), "http listen address(default :31234).")
	strVar(&c.GRPCAddr, "grpc.addr", getEnvStr("TKEEL_GRPC_ADDR", ":31233"), "grpc listen address(default :31233).")
	strVar(&c.Proxy.Timeout, "proxy.timeout", getEnvStr("TKEEL_PROXY_TIMEOUT", "10s"), "proxy timeout(default 10s).")
	strVar(&c.Proxy.CoreAddr, "proxy.core_addr", getEnvStr("TKEEL_PROXY_CORE_ADDR", "core:6789"), "core listen address(default core:6789).")
	strVar(&c.Proxy.RudderAddr, "proxy.rudder_addr", getEnvStr("TKEEL_PROXY_RUDDER_ADDR", "rudder:31234"), "rudder listen address(default rudder:31234).")
	strVar(&c.Dapr.GRPCPort, "dapr.grpc.port", getEnvStr("DAPR_GRPC_PORT", "50001"), "dapr grpc listen address(default 50001).")
	strVar(&c.Dapr.HTTPPort, "dapr.http.port", getEnvStr("DAPR_HTTP_PORT", "3500"), "dapr grpc listen address(default 3500).")
	strVar(&c.Dapr.PrivateStateName, "dapr.private_state_name", getEnvStr("TKEEL_DAPR_PRIVATE_STATE_NAME", "tkeel-middleware-redis-private-store"), "dapr private store name(default keel-private-store).")
	strVar(&c.Dapr.PublicStateName, "dapr.public_state_name", getEnvStr("TKEEL_DAPR_PUBLIC_STATE_NAME", "tkeel-middleware-redis-public-store"), "dapr public store name(default keel-public-store).")
	strVar(&c.Tkeel.Namespace, "tkeel.namespace", getEnvStr("TKEEL_POD_NAMESPACE", "tkeel-system"), "tkeel pod namespace.(default tkeel-system)")
	strVar(&c.Tkeel.AdminPassword, "tkeel.admin_password", getEnvStr("TKEEL_ADMIN_PASSWD", "v0.2.0"), "tkeel version.(default v0.2.0)")
	strVar(&c.Tkeel.Version, "tkeel.version", getEnvStr("TKEEL_VERSION", "v0.2.0"), "tkeel version.(default v0.2.0)")
	strVar(&c.Tkeel.WatchInterval, "tkeel.watch_interval", getEnvStr("TKEEL_WATCH_INTERVAL", "10s"), "tkeel watch change interval.(default 10s)")
	strVar(&c.SecurityConf.Mysql.DBName, "security.mysql.dbname", getEnvStr("TKEEL_SECURITY_MYSQL_DBNAME", "tkeelauth"), "database name of auth`s mysql config")
	strVar(&c.SecurityConf.Mysql.User, "security.mysql.user", getEnvStr("TKEEL_SECURITY_MYSQL_USER", "root"), "user name of auth`s mysql config")
	strVar(&c.SecurityConf.Mysql.Password, "security.mysql.password", getEnvStr("TKEEL_SECURITY_MYSQL_PASSWORD", "a3fks=ixmeb82a"), "password of auth`s mysql config")
	strVar(&c.SecurityConf.Mysql.Host, "security.mysql.host", getEnvStr("TKEEL_SECURITY_MYSQL_HOST", "tkeel-middleware-mysql"), "host of auth`s mysql config")
	strVar(&c.SecurityConf.Mysql.Port, "security.mysql.port", getEnvStr("TKEEL_SECURITY_MYSQL_PORT", "3306"), "port of auth`s mysql config")
	strVar(&c.SecurityConf.OAuth2.Redis.Addr, "security.oauth2.redis.addr", getEnvStr("TKEEL_SECURITY_OAUTH2_REDIS_ADDR", "tkeel-middleware-redis-master:6379"), "address of auth`s redis config")
	strVar(&c.SecurityConf.OAuth2.Redis.Password, "security.oauth2.redis.password", getEnvStr("TKEEL_SECURITY_OAUTH2_REDIS_PASSWORD", "Biz0P8Xoup"), "password of auth`s redis config")
	intVar(&c.SecurityConf.OAuth2.Redis.DB, "security.oauth2.redis.db", getEnvInt("TKEEL_SECURITY_OAUTH2_REDIS_DB", 0), "db of auth`s redis")
	strVar(&c.SecurityConf.OAuth2.AuthType, "security.oauth2.auth_type", getEnvStr("TKEEL_SECURITY_OAUTH2_AUTH_TYPE", ""), "security auth type of auth`s access type,if type == demo sikp auth filter.")
	strVar(&c.SecurityConf.OAuth2.AccessGenerate.SecurityKey, "security.oauth2.access.sk", getEnvStr("TKEEL_SECURITY_ACCESS_SK", "eixn27adg3"), "security key of auth`s access generate")
	strVar(&c.SecurityConf.OAuth2.AccessGenerate.AccessTokenExp, "security.oauth2.access.access_token_exp", getEnvStr("TKEEL_SECURITY_ACCESS_TOKEN_EXP", "30m"), "security token of auth`s access exp")
	strVar(&c.SecurityConf.Entity.SecurityKey, "security.entity.sk", getEnvStr("TKEEL_SECURITY_ENTITY_SK", "i5s2x3nov894"), "security  key auth`s entity token access")
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
