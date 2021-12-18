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
	"github.com/tkeel-io/security/apiserver/filters"
	"github.com/tkeel-io/security/models/oauth"
	"github.com/tkeel-io/tkeel/cmd"
	t_dapr "github.com/tkeel-io/tkeel/pkg/client/dapr"
	"github.com/tkeel-io/tkeel/pkg/config"
	"github.com/tkeel-io/tkeel/pkg/model/proute"
	proxy_v1 "github.com/tkeel-io/tkeel/pkg/proxy/v1"
	"github.com/tkeel-io/tkeel/pkg/server"
	"github.com/tkeel-io/tkeel/pkg/service"
)

var (
	configFile string

	conf    *config.Configuration
	keelApp *app.App
)

var rootCmd = &cobra.Command{
	Use: "keel is the main component in the tKeel.",
	Short: `keel is the proxy gateway component in the tkeel platform.
	Used to proxy internal and external request.`,
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
		httpSrv.Container.EnableContentEncoding(false)
		grpcSrv := server.NewGRPCServer(conf.GRPCAddr)

		keelApp = app.New("keel", &log.Conf{
			App:    "keel",
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
			// dapr http client.
			daprHTTPClient := t_dapr.NewHTTPClient(conf.Dapr.HTTPPort)

			// init operator.
			prOp := proute.NewDaprStateOperator(conf.Dapr.PublicStateName, daprGRPCClient)

			// init service.
			// proxy service.
			ProxySrvV1 := service.NewProxyServiceV1(conf.Tkeel.WatchPluginRouteInterval,
				conf.Proxy, daprHTTPClient, prOp)
			// init oauth oprator.
			if _, err = oauth.NewOperator(conf.SecurityConf.OAuth2); err != nil {
				log.Fatal("fatal new oauth oprator: %s", err)
				os.Exit(-1)
			}
			proxy_v1.RegisterPluginProxyHTTPServer(context.TODO(), httpSrv.Container, conf, filters.AuthFilter(conf.SecurityConf.OAuth2), ProxySrvV1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := keelApp.Run(context.TODO()); err != nil {
			log.Fatal("fatal keel app run: %s", err)
			os.Exit(-1)
		}

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
		<-stop

		if err := keelApp.Stop(context.TODO()); err != nil {
			log.Fatal("fatal keel app stop: %s", err)
			os.Exit(-2)
		}
	},
}

func init() {
	conf = config.NewDefaultConfiguration()
	conf.AttachCmdFlags(rootCmd.Flags().StringVar, rootCmd.Flags().BoolVar, rootCmd.Flags().IntVar)
	rootCmd.Flags().StringVar(&configFile, "config", getEnvStr("KEEL_CONFIG", ""), "keel config file path.")
	rootCmd.AddCommand(cmd.VersionCmd)
}

func getEnvStr(env string, defaultValue string) string {
	v := os.Getenv(env)
	if v == "" {
		return defaultValue
	}
	return v
}
