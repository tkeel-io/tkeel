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
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"

	authentication_v1 "github.com/tkeel-io/tkeel/api/authentication/v1"
	config_v1 "github.com/tkeel-io/tkeel/api/config/v1"
	entity_token_v1 "github.com/tkeel-io/tkeel/api/entity/v1"
	entry_v1 "github.com/tkeel-io/tkeel/api/entry/v1"
	oauth2_v1 "github.com/tkeel-io/tkeel/api/oauth2/v1"
	plugin_v1 "github.com/tkeel-io/tkeel/api/plugin/v1"
	rbac_v1 "github.com/tkeel-io/tkeel/api/rbac/v1"
	repo "github.com/tkeel-io/tkeel/api/repo/v1"
	oauth_v1 "github.com/tkeel-io/tkeel/api/security_oauth/v1"
	tenant_v1 "github.com/tkeel-io/tkeel/api/tenant/v1"
	"github.com/tkeel-io/tkeel/cmd"
	t_dapr "github.com/tkeel-io/tkeel/pkg/client/dapr"
	"github.com/tkeel-io/tkeel/pkg/client/kubernetes"
	"github.com/tkeel-io/tkeel/pkg/client/openapi"
	"github.com/tkeel-io/tkeel/pkg/config"
	"github.com/tkeel-io/tkeel/pkg/hub"
	"github.com/tkeel-io/tkeel/pkg/model"
	"github.com/tkeel-io/tkeel/pkg/model/kv"
	"github.com/tkeel-io/tkeel/pkg/model/plugin"
	"github.com/tkeel-io/tkeel/pkg/model/prepo"
	"github.com/tkeel-io/tkeel/pkg/model/proute"
	"github.com/tkeel-io/tkeel/pkg/repository"
	"github.com/tkeel-io/tkeel/pkg/repository/helm"
	"github.com/tkeel-io/tkeel/pkg/server"
	"github.com/tkeel-io/tkeel/pkg/service"

	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/store"
	oredis "github.com/go-oauth2/redis/v4"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"
	"github.com/spf13/cobra"
	"github.com/tkeel-io/kit/app"
	"github.com/tkeel-io/kit/log"
	security_casbin "github.com/tkeel-io/security/authz/casbin"
	"github.com/tkeel-io/security/authz/rbac"
	"github.com/tkeel-io/security/gormdb"
)

var (
	configFile string
	conf       *config.Configuration
	rudderApp  *app.App
)

var rootCmd = &cobra.Command{
	Use: "rudder is the main component in the tKeel.",
	Short: `rudder is the main control component in the tkeel platform.
	Used to manage plugins and tenants.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if configFile != "" {
			c, err := config.LoadStandaloneConfiguration(configFile)
			if err != nil {
				log.Fatal("fatal config load(%s): %s", configFile, err)
				os.Exit(-1)
			}
			conf = c
		}
		conf.Init()
		httpSrv := server.NewHTTPServer(conf.HTTPAddr)
		grpcSrv := server.NewGRPCServer(conf.GRPCAddr)

		rudderApp = app.New("rudder", &log.Conf{
			App:    "rudder",
			Level:  conf.Log.Level,
			Dev:    conf.Log.Dev,
			Output: conf.Log.Output,
		}, httpSrv, grpcSrv)

		{
			// init client.
			// dapr grpc client.
			daprGRPCClient, err := t_dapr.NewGPRCClient(10, "1s", conf.Dapr.GRPCPort)
			if err != nil {
				log.Fatal("fatal new dapr client: %s", err)
				os.Exit(-1)
			}
			openapiCli := openapi.NewDaprClient(conf.Dapr.HTTPPort)

			// init k8s client.
			k8sClient := kubernetes.NewClient(conf.DeploymentConfigmap, conf.Tkeel.Namespace)

			// init operator.
			pOp := plugin.NewDaprStateOperator(conf.Dapr.PrivateStateName, daprGRPCClient)
			prOp := proute.NewDaprStateOperator(conf.Dapr.PublicStateName, daprGRPCClient)
			riOp := prepo.NewDaprStateOperator(conf.Dapr.PrivateStateName, daprGRPCClient)
			kvOp := kv.NewDaprStateOperator(conf.Tkeel.WatchInterval, conf.Dapr.PrivateStateName, daprGRPCClient)
			kvOp.Watch(context.TODO(), model.KeyPermissionSet, func(value []byte, version string) error {
				log.Debugf("update %s %s", model.KeyPermissionSet, string(value))
				if err1 := model.GetPermissionSet().Unmarshal(value); err1 != nil {
					return errors.Wrapf(err1, "unmarshal %s %s", model.KeyPermissionSet, value)
				}
				return nil
			})

			// init security operator.
			tokenConf := &service.TokenConf{TokenType: service.TokenTypeBearer, AllowedGrantTypes: service.DefaultGrantType}
			tokenStore := oredis.NewRedisStore(&redis.Options{
				Addr:     conf.SecurityConf.OAuth.Redis.Addr,
				DB:       conf.SecurityConf.OAuth.Redis.DB,
				Password: conf.SecurityConf.OAuth.Redis.Password,
			})
			tokenGenerator := generates.NewJWTAccessGenerate("", []byte(conf.SecurityConf.OAuth.AccessGenerate.SecurityKey), jwt.SigningMethodHS512)
			gormdb, err := gormdb.SetUp(gormdb.DBConfig{
				Type: "mysql", Host: conf.SecurityConf.Mysql.Host, Port: conf.SecurityConf.Mysql.Port,
				Dbname: conf.SecurityConf.Mysql.DBName, Username: conf.SecurityConf.Mysql.User, Password: conf.SecurityConf.Mysql.Password,
			})
			if err != nil {
				log.Fatal(err)
				os.Exit(-1)
			}
			rbacOp, err := security_casbin.NewRBACOperator(&security_casbin.MysqlConf{
				DBName: conf.SecurityConf.Mysql.DBName,
				User:   conf.SecurityConf.Mysql.User, Password: conf.SecurityConf.Mysql.Password,
				Host: conf.SecurityConf.Mysql.Host, Port: conf.SecurityConf.Mysql.Port,
			})
			if err != nil {
				log.Fatal("fatal new rbac operator", err)
				os.Exit(-1)
			}
			tenantPluginOp := rbac.NewTenantPluginOperator(rbacOp)
			m := manage.NewDefaultManager()
			clientStore := store.NewClientStore()
			client := &models.Client{ID: "tkeel", Secret: "tkeel", Domain: "tkeel.io"}
			clientStore.Set(client.GetID(), client)
			mConf := manage.DefaultAuthorizeCodeTokenCfg
			if tokenConf.AccessTokenExp != 0 && tokenConf.RefreshTokenExp != 0 {
				mConf.AccessTokenExp = tokenConf.AccessTokenExp
				mConf.RefreshTokenExp = tokenConf.RefreshTokenExp
			}
			m.SetPasswordTokenCfg(mConf)
			m.MapClientStorage(clientStore)
			m.MapTokenStorage(tokenStore)
			m.MapAccessGenerate(tokenGenerator)

			// init repo hub.
			hub.Init(conf.Tkeel.WatchInterval, riOp,
				func(connectInfo *repository.Info,
					args ...interface{}) (repository.Repository, error) {
					if len(args) != 2 {
						return nil, errors.New("invalid arguments")
					}
					drive, ok := args[0].(helm.Driver)
					if !ok {
						return nil, errors.New("invalid argument type")
					}
					namespace, ok := args[1].(string)
					if !ok {
						return nil, errors.New("invalid argument type")
					}
					repo, err := helm.NewHelmRepo(connectInfo, drive, namespace)
					if err != nil {
						return nil, errors.Wrap(err, "new helm repo")
					}
					return repo, nil
				},
				func(pluginID string) error {
					repo, err := helm.NewHelmRepo(nil, helm.Secret, conf.Tkeel.Namespace)
					if err != nil {
						return errors.Wrap(err, "new helm repo")
					}
					installer := helm.NewHelmInstallerQuick(pluginID, conf.Tkeel.Namespace, repo.Config())
					if err = installer.Uninstall(); err != nil {
						return errors.Wrapf(err, "uninstall(%s)", pluginID)
					}
					return nil
				}, helm.Secret, conf.Tkeel.Namespace)

			// init service.
			// plugin service.
			pluginSrvV1 := service.NewPluginServiceV1(rbacOp, gormdb, conf.Tkeel,
				kvOp, pOp, prOp, tenantPluginOp, openapiCli)
			plugin_v1.RegisterPluginHTTPServer(httpSrv.Container, pluginSrvV1)
			plugin_v1.RegisterPluginServer(grpcSrv.GetServe(), pluginSrvV1)
			// oauth2 service.
			oauth2SrvV1 := service.NewOauth2ServiceV1(conf.Tkeel.AdminPassword, kvOp, pOp)
			oauth2_v1.RegisterOauth2HTTPServer(httpSrv.Container, oauth2SrvV1)
			oauth2_v1.RegisterOauth2Server(grpcSrv.GetServe(), oauth2SrvV1)
			// repo service.
			repoSrv := service.NewRepoService()
			repo.RegisterRepoHTTPServer(httpSrv.Container, repoSrv)
			repo.RegisterRepoServer(grpcSrv.GetServe(), repoSrv)
			// entries service.
			entriesSrvV1 := service.NewEntryService(pOp, tenantPluginOp, rbacOp)
			entry_v1.RegisterEntryHTTPServer(httpSrv.Container, entriesSrvV1)
			entry_v1.RegisterEntryServer(grpcSrv.GetServe(), entriesSrvV1)

			// tenant service.
			tenantSrv := service.NewTenantService(gormdb, tenantPluginOp, rbacOp)
			tenant_v1.RegisterTenantHTTPServer(httpSrv.Container, tenantSrv)
			tenant_v1.RegisterTenantServer(grpcSrv.GetServe(), tenantSrv)
			// oauth server.
			oauthSrv := service.NewOauthService(m, gormdb, tokenConf, daprGRPCClient, conf.Dapr.PrivateStateName, k8sClient)
			oauth_v1.RegisterOauthHTTPServer(httpSrv.Container, oauthSrv)
			oauth_v1.RegisterOauthServer(grpcSrv.GetServe(), oauthSrv)

			// entity token.
			tokenOp := service.NewEntityTokenOperator(conf.Dapr.PrivateStateName, daprGRPCClient)
			EntityTokenSrv := service.NewEntityTokenService(tokenOp)
			entity_token_v1.RegisterEntityTokenHTTPServer(httpSrv.Container, EntityTokenSrv)
			entity_token_v1.RegisterEntityTokenServer(grpcSrv.GetServe(), EntityTokenSrv)

			// rbac service.
			rbacSrv := service.NewRBACService(gormdb, rbacOp, tenantPluginOp)
			rbac_v1.RegisterRBACHTTPServer(httpSrv.Container, rbacSrv)
			rbac_v1.RegisterRBACServer(grpcSrv.GetServe(), rbacSrv)

			// authentication service.
			authenticationSrv := service.NewAuthenticationService(m, gormdb, tokenConf, rbacOp, prOp, tenantPluginOp)
			authentication_v1.RegisterAuthenticationHTTPServer(httpSrv.Container, authenticationSrv)
			authentication_v1.RegisterAuthenticationServer(grpcSrv.GetServe(), authenticationSrv)

			// config service.
			configSrv := service.NewConfigService(k8sClient)
			config_v1.RegisterConfigHTTPServer(httpSrv.Container, configSrv)
			config_v1.RegisterConfigServer(grpcSrv.GetServe(), configSrv)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := rudderApp.Run(context.TODO()); err != nil {
			log.Fatal("fatal rudder app run: %s", err)
			os.Exit(-2)
		}

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
		<-stop

		if err := rudderApp.Stop(context.TODO()); err != nil {
			log.Fatal("fatal rudder app stop: %s", err)
			os.Exit(-3)
		}
	},
}

func init() {
	conf = config.NewDefaultConfiguration()
	conf.AttachCmdFlags(rootCmd.Flags().StringVar, rootCmd.Flags().BoolVar, rootCmd.Flags().IntVar)
	rootCmd.Flags().StringVar(&configFile, "config", getEnvStr("RUDDER_CONFIG", ""), "rudder config file path.")
	rootCmd.AddCommand(cmd.VersionCmd)
}

func getEnvStr(env string, defaultValue string) string {
	v := os.Getenv(env)
	if v == "" {
		return defaultValue
	}
	return v
}
