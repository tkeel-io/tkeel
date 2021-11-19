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
	plugin_v1 "github.com/tkeel-io/tkeel/api/plugin/v1"
	"github.com/tkeel-io/tkeel/cmd"
	"github.com/tkeel-io/tkeel/pkg/client/openapi"
	"github.com/tkeel-io/tkeel/pkg/config"
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
			// http srv add Filter.
			// httpSrv.Container.Filter().

			// init client.
			// dapr client.
			daprCli, err := dapr.NewClientWithPort(conf.Dapr.GRPCPort)
			if err != nil {
				panic(err)
			}
			openapiCli := openapi.NewDaprClient("rudder", daprCli)

			// init operator.
			pOp := plugin.NewDaprStateOperator(conf.Dapr.PrivateStateName, daprCli)
			prOp := proute.NewDaprStateOperator(conf.Dapr.PublicStateName, daprCli)

			// init service.
			// plugin service.
			PluginSrvV1 := service.NewPluginServiceV1(conf.Tkeel, pOp, prOp, openapiCli)
			plugin_v1.RegisterPluginHTTPServer(httpSrv.Container, PluginSrvV1)
			plugin_v1.RegisterPluginServer(grpcSrv.GetServe(), PluginSrvV1)
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
			panic(err)
		}

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
		<-stop

		if err := rudderApp.Stop(context.TODO()); err != nil {
			panic(err)
		}
	},
}

func init() {
	conf = config.NewDefaultConfiguration()
	conf.AttachCmdFlags(rootCmd.Flags().StringVar, rootCmd.Flags().BoolVar)
	rootCmd.Flags().StringVar(&configFile, "config", getEnvStr("TMANAGER_CONFIG", ""), "tmanager config file path.")
	rootCmd.AddCommand(cmd.VersionCmd)
}

func getEnvStr(env string, defaultValue string) string {
	v := os.Getenv(env)
	if v == "" {
		return defaultValue
	}
	return v
}
