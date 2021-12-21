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
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/emicklei/go-restful"
	"github.com/spf13/cobra"
	"github.com/tkeel-io/kit/app"
	"github.com/tkeel-io/kit/log"
	entity_v1 "github.com/tkeel-io/security/apirouter/entity/v1"
	oauth_v1 "github.com/tkeel-io/security/apirouter/oauth/v1"
	rbac_v1 "github.com/tkeel-io/security/apirouter/rbac/v1"
	tenant_v1 "github.com/tkeel-io/security/apirouter/tenant/v1"
	"github.com/tkeel-io/security/apiserver/filters"
	security_dao "github.com/tkeel-io/security/models/dao"
	"github.com/tkeel-io/security/models/entity"
	oauth2_v1 "github.com/tkeel-io/tkeel/api/oauth2/v1"
	plugin_v1 "github.com/tkeel-io/tkeel/api/plugin/v1"
	repo "github.com/tkeel-io/tkeel/api/repo/v1"
	"github.com/tkeel-io/tkeel/cmd"
	t_dapr "github.com/tkeel-io/tkeel/pkg/client/dapr"
	"github.com/tkeel-io/tkeel/pkg/client/openapi"
	"github.com/tkeel-io/tkeel/pkg/config"
	"github.com/tkeel-io/tkeel/pkg/hub"
	"github.com/tkeel-io/tkeel/pkg/model/plugin"
	"github.com/tkeel-io/tkeel/pkg/model/prepo"
	"github.com/tkeel-io/tkeel/pkg/model/proute"
	"github.com/tkeel-io/tkeel/pkg/repository"
	"github.com/tkeel-io/tkeel/pkg/repository/helm"
	"github.com/tkeel-io/tkeel/pkg/server"
	"github.com/tkeel-io/tkeel/pkg/service"
)

var (
	configFile string

	conf      *config.Configuration
	rudderApp *app.App
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
			daprGRPCClient, err := t_dapr.NewGPRCClient(10, "5s", conf.Dapr.GRPCPort)
			if err != nil {
				log.Fatal("fatal new dapr client: %s", err)
				os.Exit(-1)
			}
			openapiCli := openapi.NewDaprClient("rudder", daprGRPCClient)

			// init operator.
			pOp := plugin.NewDaprStateOperator(conf.Dapr.PrivateStateName, daprGRPCClient)
			prOp := proute.NewDaprStateOperator(conf.Dapr.PublicStateName, daprGRPCClient)
			riOp := prepo.NewDaprStateOperator(conf.Dapr.PrivateStateName, daprGRPCClient)

			// init repo hub.
			hub.Init(conf.Tkeel.WatchPluginRouteInterval, riOp,
				func(connectInfo *repository.Info,
					args ...interface{}) (repository.Repository, error) {
					if len(args) != 2 {
						return nil, errors.New("invalid arguments")
					}
					drive, ok := args[0].(string)
					if !ok {
						return nil, errors.New("invaild argument type")
					}
					namespace, ok := args[1].(string)
					if !ok {
						return nil, errors.New("invaild argument type")
					}
					repo, err := helm.NewHelmRepo(*connectInfo, helm.Driver(drive), namespace)
					if err != nil {
						return nil, fmt.Errorf("error new helm repo: %w", err)
					}
					return repo, nil
				},
				func(pluginID string) error {
					repo, err := helm.NewHelmRepo(repository.Info{}, helm.Mem, conf.Tkeel.Namespace)
					if err != nil {
						return fmt.Errorf("error new helm repo: %w", err)
					}
					installer := helm.NewHelmInstallerQuick(pluginID, conf.Tkeel.Namespace, repo.Config())
					if err = installer.Uninstall(); err != nil {
						return fmt.Errorf("error uninstall(%s) err: %w", pluginID, err)
					}
					return nil
				}, helm.Mem, conf.Tkeel.Namespace)

			// init service.
			// plugin service.
			PluginSrvV1 := service.NewPluginServiceV1(conf.Tkeel, pOp, prOp, openapiCli)
			plugin_v1.RegisterPluginHTTPServer(httpSrv.Container, PluginSrvV1)
			plugin_v1.RegisterPluginServer(grpcSrv.GetServe(), PluginSrvV1)
			// oauth2 service.
			Oauth2SrvV1 := service.NewOauth2ServiceV1(conf.Tkeel.Secret, pOp)
			oauth2_v1.RegisterOauth2HTTPServer(httpSrv.Container, Oauth2SrvV1)
			oauth2_v1.RegisterOauth2Server(grpcSrv.GetServe(), Oauth2SrvV1)
			// repo service.
			repoSrv := service.NewRepoService()
			repo.RegisterRepoHTTPServer(httpSrv.Container, repoSrv)
			repo.RegisterRepoServer(grpcSrv.GetServe(), repoSrv)
			{
				// copy mysql configuration.
				conf.SecurityConf.RBAC.Adapter = conf.SecurityConf.Mysql
				// init security service.
				security_dao.SetUp(conf.SecurityConf.Mysql)
				// tenant.
				tenant_v1.RegisterToRestContainer(httpSrv.Container)
				// oauth2.
				oauth_v1.RegisterToRestContainer(httpSrv.Container, conf.SecurityConf.OAuth2)
				// rbac.
				rbac_v1.RegisterToRestContainer(httpSrv.Container, conf.SecurityConf.RBAC, conf.SecurityConf.OAuth2)
				// entity token.
				entityTokenOperator := entity.NewEntityTokenOperator(conf.Dapr.PrivateStateName, daprGRPCClient)
				if entityTokenOperator == nil {
					os.Exit(-1)
				}
				entity_v1.RegisterToRestContainer(httpSrv.Container, conf.SecurityConf.Entity, entityTokenOperator)
				// add auth role filter.
				tenantAdminRoleFilter := filters.AuthFilter(conf.SecurityConf.OAuth2, "admin")
				for _, ws := range httpSrv.Container.RegisteredWebServices() {
					if ws.RootPath() == "/v1/tenants" {
						ws.Filter(func(r1 *restful.Request, r2 *restful.Response, fc *restful.FilterChain) {
							if strings.HasPrefix(r1.Request.URL.Path, "/v1/tenants/users") {
								tenantAdminRoleFilter(r1, r2, fc)
								return
							}
							fc.ProcessFilter(r1, r2)
						})
					}
				}
			}
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
