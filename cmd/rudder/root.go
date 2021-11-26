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

	"github.com/spf13/cobra"
	"github.com/tkeel-io/kit/app"
	"github.com/tkeel-io/kit/log"
	tenant_v1 "github.com/tkeel-io/security/pkg/apirouter/tenant/v1"
	auth_dao "github.com/tkeel-io/security/pkg/models/dao"
	oauth2_v1 "github.com/tkeel-io/tkeel/api/oauth2/v1"
	plugin_v1 "github.com/tkeel-io/tkeel/api/plugin/v1"
	repo "github.com/tkeel-io/tkeel/api/repo/v1"
	"github.com/tkeel-io/tkeel/cmd"
	"github.com/tkeel-io/tkeel/pkg/client/openapi"
	"github.com/tkeel-io/tkeel/pkg/config"
	"github.com/tkeel-io/tkeel/pkg/helm"
	"github.com/tkeel-io/tkeel/pkg/model/plugin"
	"github.com/tkeel-io/tkeel/pkg/model/proute"
	"github.com/tkeel-io/tkeel/pkg/server"
	"github.com/tkeel-io/tkeel/pkg/service"

	dapr "github.com/dapr/go-sdk/client"
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
				panic(err)
			}
			conf = c
		}
		httpSrv := server.NewHTTPServer(conf.HTTPAddr)
		grpcSrv := server.NewGRPCServer(conf.GRPCAddr)

		{
			// init client.
			// dapr grpc client.
			daprGRPCClient, err := dapr.NewClientWithPort(conf.Dapr.GRPCPort)
			if err != nil {
				panic(err)
			}
			openapiCli := openapi.NewDaprClient("rudder", daprGRPCClient)

			helm.SetDaprConfig(&daprGRPCClient, conf.Dapr.PrivateStateName)

			// init operator.
			pOp := plugin.NewDaprStateOperator(conf.Dapr.PrivateStateName, daprGRPCClient)
			prOp := proute.NewDaprStateOperator(conf.Dapr.PublicStateName, daprGRPCClient)

			// init service.
			// plugin service.
			PluginSrvV1 := service.NewPluginServiceV1(conf.Tkeel, pOp, prOp, openapiCli)
			plugin_v1.RegisterPluginHTTPServer(httpSrv.Container, PluginSrvV1)
			plugin_v1.RegisterPluginServer(grpcSrv.GetServe(), PluginSrvV1)
			// oauth2 service.
			Oauth2SrvV1 := service.NewOauth2Service(conf.Tkeel.Secret, pOp)
			oauth2_v1.RegisterOauth2HTTPServer(httpSrv.Container, Oauth2SrvV1)
			oauth2_v1.RegisterOauth2Server(grpcSrv.GetServe(), Oauth2SrvV1)
			// tenant.
			auth_dao.SetUp(conf.SecurityConf.Mysql)
			tenant_v1.RegisterToRestContainer(httpSrv.Container)

			// repo service.
			repoSrv := service.NewRepoService()
			repo.RegisterRepoHTTPServer(httpSrv.Container, repoSrv)
			repo.RegisterRepoServer(grpcSrv.GetServe(), repoSrv)
		}

		rudderApp = app.New("rudder", &log.Conf{
			App:    "rudder",
			Level:  conf.Log.Level,
			Dev:    conf.Log.Dev,
			Output: conf.Log.Output,
		}, httpSrv, grpcSrv)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := rudderApp.Run(context.TODO()); err != nil {
			log.Fatal("fatal rudder app run: %s", err)
			os.Exit(-1)
		}

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
		<-stop

		if err := rudderApp.Stop(context.TODO()); err != nil {
			log.Fatal("fatal rudder app stop: %s", err)
			os.Exit(-2)
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
