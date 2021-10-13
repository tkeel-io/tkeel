package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/tkeel-io/tkeel/pkg/logger"
	"github.com/tkeel-io/tkeel/pkg/plugin"
	"github.com/tkeel-io/tkeel/pkg/plugin/keel"
	"github.com/tkeel-io/tkeel/pkg/version"
)

var (
	log = logger.NewLogger("tKeel.keel")
)

func main() {
	logger.SetPluginVersion(version.Version())
	log.Infof("starting tKeel keel -- version %s -- commit %s", version.Version(), version.Commit())
	plugin, err := plugin.FromFlags()
	if err != nil {
		log.Fatalf("error init plugin: %s", err)
		return
	}
	g, err := keel.New(plugin)
	if err != nil {
		log.Fatalf("error new keel: %s", err)
		return
	}

	g.Run()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
	<-stop
}
