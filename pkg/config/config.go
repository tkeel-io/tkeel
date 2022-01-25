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

	"gopkg.in/yaml.v3"
)

// TkeelConf tkeel platform configuration.
type TkeelConf struct {
	// tkeel platform namespace.
	Namespace string `json:"namespace" yaml:"namespace"`
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
	// Proxy tkeel platform configuration.
	Proxy *ProxyConf `json:"proxy" yaml:"proxy"`
	// Dapr dapr configuration.
	Dapr *DaprConf `json:"dapr" yaml:"dapr"`
	// Log log configuration.
	Log *LogConf `json:"log" yaml:"log"`
	// SecurityConf security auth config.
	SecurityConf *SecurityConf `json:"security_conf" yaml:"securityConf"`
}

// SecurityConf.
type SecurityConf struct {
	// Mysql  mysql config of security.
	Mysql *MysqlConf `json:"mysql" yaml:"mysql"`
	// OAuth2Config oauth2 config of security.
	OAuth *OauthConfig `json:"oauth2" yaml:"oauth2"` // nolint
	// entity entity security config of auth.
	Entity *EntityConf `json:"entity" yaml:"entity"`
}

type EntityConf struct {
	SecurityKey string `json:"security_key" yaml:"securityKey"`
}

type MysqlConf struct {
	DBName   string `json:"dbname" yaml:"dbname"` //nolint
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	Host     string `json:"host" yaml:"host"`
	Port     string `json:"port" yaml:"port"`
}

type RedisConf struct {
	Addr     string `json:"addr" yaml:"addr"`
	DB       int    `json:"db" yaml:"db"`
	Password string `json:"password" yaml:"password"`
}

type OauthConfig struct {
	AuthType       string      `json:"auth_type" yaml:"authType"`
	Redis          *RedisConf  `json:"redis" yaml:"redis"`
	AccessGenerate *AccessConf `json:"access_generate" yaml:"accessGenerate"`
}

type AccessConf struct {
	AccessTokenExp string `json:"access_token_exp" yaml:"accessTokenExp"`
	SecurityKey    string `json:"security_key" yaml:"securityKey"`
}

// NewDefaultConfiguration returns the empty config.
func NewDefaultConfiguration() *Configuration {
	return &Configuration{
		Tkeel: &TkeelConf{},
		Proxy: &ProxyConf{},
		Dapr:  &DaprConf{},
		Log:   &LogConf{},
		SecurityConf: &SecurityConf{
			Mysql: &MysqlConf{},
			OAuth: &OauthConfig{
				Redis:          &RedisConf{},
				AccessGenerate: &AccessConf{},
			},
			Entity: &EntityConf{},
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
	strVar(&c.Proxy.Timeout, "proxy.timeout", getEnvStr("TKEEL_PROXY_TIMEOUT", "30s"), "proxy timeout(default 10s).")
	strVar(&c.Dapr.GRPCPort, "dapr.grpc.port", getEnvStr("DAPR_GRPC_PORT", "50001"), "dapr grpc listen address(default 50001).")
	strVar(&c.Dapr.HTTPPort, "dapr.http.port", getEnvStr("DAPR_HTTP_PORT", "3500"), "dapr grpc listen address(default 3500).")
	strVar(&c.Dapr.PrivateStateName, "dapr.private_state_name", getEnvStr("TKEEL_DAPR_PRIVATE_STATE_NAME", "tkeel-middleware-redis-private-store"), "dapr private store name(default keel-private-store).")
	strVar(&c.Dapr.PublicStateName, "dapr.public_state_name", getEnvStr("TKEEL_DAPR_PUBLIC_STATE_NAME", "tkeel-middleware-redis-public-store"), "dapr public store name(default keel-public-store).")
	strVar(&c.Tkeel.Namespace, "tkeel.namespace", getEnvStr("TKEEL_POD_NAMESPACE", "tkeel-system"), "tkeel pod namespace.(default tkeel-system)")
	strVar(&c.Tkeel.AdminPassword, "tkeel.admin_password", getEnvStr("TKEEL_ADMIN_PASSWD", "changeme"), "tkeel admin password.(default env TKEEL_ADMIN_PASSWD)")
	strVar(&c.Tkeel.WatchInterval, "tkeel.watch_interval", getEnvStr("TKEEL_WATCH_INTERVAL", "10s"), "tkeel watch change interval.(default 10s)")
	strVar(&c.SecurityConf.Mysql.DBName, "security.mysql.dbname", getEnvStr("TKEEL_SECURITY_MYSQL_DBNAME", "tkeelauth"), "database name of auth`s mysql config")
	strVar(&c.SecurityConf.Mysql.User, "security.mysql.user", getEnvStr("TKEEL_SECURITY_MYSQL_USER", "root"), "user name of auth`s mysql config")
	strVar(&c.SecurityConf.Mysql.Password, "security.mysql.password", getEnvStr("TKEEL_SECURITY_MYSQL_PASSWORD", "a3fks=ixmeb82a"), "password of auth`s mysql config")
	strVar(&c.SecurityConf.Mysql.Host, "security.mysql.host", getEnvStr("TKEEL_SECURITY_MYSQL_HOST", "tkeel-middleware-mysql"), "host of auth`s mysql config")
	strVar(&c.SecurityConf.Mysql.Port, "security.mysql.port", getEnvStr("TKEEL_SECURITY_MYSQL_PORT", "3306"), "port of auth`s mysql config")
	strVar(&c.SecurityConf.OAuth.Redis.Addr, "security.oauth2.redis.addr", getEnvStr("TKEEL_SECURITY_OAUTH2_REDIS_ADDR", "tkeel-middleware-redis-master:6379"), "address of auth`s redis config")
	strVar(&c.SecurityConf.OAuth.Redis.Password, "security.oauth2.redis.password", getEnvStr("TKEEL_SECURITY_OAUTH2_REDIS_PASSWORD", "Biz0P8Xoup"), "password of auth`s redis config")
	intVar(&c.SecurityConf.OAuth.Redis.DB, "security.oauth2.redis.db", getEnvInt("TKEEL_SECURITY_OAUTH2_REDIS_DB", 0), "db of auth`s redis")
	strVar(&c.SecurityConf.OAuth.AuthType, "security.oauth2.auth_type", getEnvStr("TKEEL_SECURITY_OAUTH2_AUTH_TYPE", ""), "security auth type of auth`s access type,if type == demo sikp auth filter.")
	strVar(&c.SecurityConf.OAuth.AccessGenerate.SecurityKey, "security.oauth2.access.sk", getEnvStr("TKEEL_SECURITY_ACCESS_SK", "eixn27adg3"), "security key of auth`s access generate")
	strVar(&c.SecurityConf.OAuth.AccessGenerate.AccessTokenExp, "security.oauth2.access.access_token_exp", getEnvStr("TKEEL_SECURITY_ACCESS_TOKEN_EXP", "30m"), "security token of auth`s access exp")
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
