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
	"fmt"
	"os"
	"os/signal"
	"syscall"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/spf13/cobra"
	"github.com/tkeel-io/kit/app"
	"github.com/tkeel-io/kit/log"
	pluginAPI "github.com/tkeel-io/tkeel/api/plugin/v1"
	"github.com/tkeel-io/tkeel/pkg/client/openapi"
	"github.com/tkeel-io/tkeel/pkg/config"
	"github.com/tkeel-io/tkeel/pkg/model/plugin"
	"github.com/tkeel-io/tkeel/pkg/model/proute"
	"github.com/tkeel-io/tkeel/pkg/server"
	"github.com/tkeel-io/tkeel/pkg/service"
)

const longUsage = `rudder is the main control plugin in the tkeel platform.
	Used to manage plugins and tenants.`

var (
	configFilePath = ""
	conf           = config.NewDefaultConfiguration()
)

func main() {
	cmd, err := newRootCmd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v", err)
		os.Exit(1)
	}

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%+v", err)
		os.Exit(1)
	}
}

func serverSetup() *app.App {
	httpSrv := server.NewHTTPServer(conf.HTTPAddr)
	grpcSrv := server.NewGRPCServer(conf.GRPCAddr)
	daprClient, err := dapr.NewClientWithPort(conf.Dapr.GRPCPort)
	if err != nil {
		panic(err)
	}

	pluginSrv := service.NewPluginServiceV1(conf.Tkeel,
		plugin.NewDaprStateOperator(conf.Dapr.PrivateStateName, daprClient),
		proute.NewDaprStateOperator(conf.Dapr.PublicStateName, daprClient),
		openapi.NewDaprClient("rudder", daprClient))
	pluginAPI.RegisterPluginHTTPServer(httpSrv.Container, pluginSrv)
	pluginAPI.RegisterPluginServer(grpcSrv.GetServe(), pluginSrv)

	return app.New("rudder", &log.Conf{
		App:    "rudder",
		Level:  conf.Log.Level,
		Dev:    conf.Log.Dev,
		Output: conf.Log.Output,
	}, httpSrv, grpcSrv)
}

func newRootCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "rudder",
		Short: "The main control plugin in the tkeel platform.",
		Long:  longUsage,
		Run: func(cmd *cobra.Command, args []string) {
			rudder := serverSetup()
			if err := rudder.Run(context.TODO()); err != nil {
				panic(err)
			}
			stop := make(chan os.Signal, 1)
			signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
			<-stop

			if err := rudder.Stop(context.TODO()); err != nil {
				panic(err)
			}
		},
	}

	flags := cmd.PersistentFlags()
	conf.AttachFlags(flags)

	flags.StringVar(&configFilePath, "config", getEnvStr("RUDDER_CONFIG", ""), "rubber config file path.")
	if configFilePath != "" {
		c, err := config.LoadStandaloneConfiguration(configFilePath)
		if err != nil {
			panic(err)
		}
		conf = c
	}

	// Add subcommands.
	cmd.AddCommand(
		NewVersionCmd(),
	)

	return cmd, nil
}

func getEnvStr(env string, defaultValue string) string {
	v := os.Getenv(env)
	if v == "" {
		return defaultValue
	}
	return v
}
