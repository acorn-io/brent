package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/acorn-io/brent/pkg/debug"
	brentcli "github.com/acorn-io/brent/pkg/server/cli"
	"github.com/acorn-io/brent/pkg/version"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"k8s.io/apiserver/pkg/server"
)

var (
	config      brentcli.Config
	debugconfig debug.Config
)

func main() {
	app := cli.NewApp()
	app.Name = "brent"
	app.Version = version.FriendlyVersion()
	app.Usage = ""
	app.Flags = append(
		brentcli.Flags(&config),
		debug.Flags(&debugconfig)...)
	app.Action = run

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(_ *cli.Context) error {
	ctx := server.SetupSignalContext()
	debugconfig.MustSetupDebug()
	s, err := config.ToServer(ctx)
	if err != nil {
		return err
	}
	addr := fmt.Sprintf(":%d", config.HTTPListenPort)
	logrus.Info("Listening on " + addr)
	return http.ListenAndServe(addr, s)
}
