package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/tkeel-io/tkeel/pkg/logger"
	"github.com/tkeel-io/tkeel/pkg/plugin"
	"github.com/tkeel-io/tkeel/pkg/plugin/plugins"
	"github.com/tkeel-io/tkeel/pkg/version"
)

var (
	log = logger.NewLogger("tKeel.plugins")
)

func main() {
	logger.PluginVersion = version.Version()
	log.Infof("starting tKeel plugins -- version %s -- commit %s", version.Version(), version.Commit())
	plugin, err := plugin.FromFlags()
	if err != nil {
		log.Fatalf("error init plugin: %s", err)
		return
	}
	m, err := plugins.New(plugin)
	if err != nil {
		log.Fatalf("error new plugins: %s", err)
		return
	}

	m.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
	<-stop
}
