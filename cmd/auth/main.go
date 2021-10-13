package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/tkeel-io/tkeel/pkg/logger"
	plugin2 "github.com/tkeel-io/tkeel/pkg/plugin"
	"github.com/tkeel-io/tkeel/pkg/plugin/auth"
	"github.com/tkeel-io/tkeel/pkg/plugin/auth/api"
	"github.com/tkeel-io/tkeel/pkg/version"
)

var (
	log = logger.NewLogger("tKeel.auth")
)

func main() {
	logger.SetPluginVersion(version.Version())
	log.Infof("starting tKeel auth -- version %s -- commit %s", version.Version(), version.Commit())
	plugin, err := plugin2.FromFlags()
	if err != nil {
		log.Fatalf("error init plugin: %s", err)
		return
	}

	pluginAuth := auth.NewPluginAuth(plugin)
	api.InitEntityIdp("./id_rsa", "./id_rsa.pem")
	pluginAuth.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
	<-stop
}
