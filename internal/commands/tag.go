package commands

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/auto"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/server"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/internal/workers"
)

// TagCommand registers the start cli command.
var TagCommand = cli.Command{
	Name:   "tag",
	Usage:  "Starts the tag server",
	Flags:  tagFlags,
	Action: tagAction,
}

var tagFlags = []cli.Flag{
	cli.BoolFlag{
		Name:   "detach-server, d",
		Usage:  "detach from the console (daemon mode)",
		EnvVar: "PHOTOPRISM_DETACH_SERVER",
	},
	cli.BoolFlag{
		Name:  "config, c",
		Usage: "show config",
	},
}

// tagAction start the web server and initializes the daemon
func tagAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)
	service.SetConfig(conf)

	if ctx.IsSet("config") {
		fmt.Printf("NAME                  VALUE\n")
		fmt.Printf("detach-server         %t\n", conf.DetachServer())

		fmt.Printf("http-host             %s\n", conf.HttpHost())
		fmt.Printf("http-port             %d\n", conf.HttpPort())
		fmt.Printf("http-mode             %s\n", conf.HttpMode())

		return nil
	}

	if conf.HttpPort() < 1 || conf.HttpPort() > 65535 {
		log.Fatal("server port must be a number between 1 and 65535")
	}

	// start web server
	go server.Start(cctx, conf)

	if count, err := photoprism.RestoreAlbums(conf.AlbumsPath(), false); err != nil {
		log.Errorf("restore: %s", err)
	} else if count > 0 {
		log.Infof("%d albums restored", count)
	}

	// start share & sync workers
	workers.Start(conf)
	auto.Start(conf)

	// set up proper shutdown of daemon and web server
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	// stop share & sync workers
	workers.Stop()
	auto.Stop()

	log.Info("shutting down...")
	conf.Shutdown()
	cancel()
	err := dctx.Release()

	if err != nil {
		log.Error(err)
	}

	time.Sleep(3 * time.Second)

	return nil
}
