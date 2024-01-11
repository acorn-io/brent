package cli

import (
	"fmt"
	"net/http"

	"github.com/acorn-io/baaah/pkg/ratelimit"
	"github.com/acorn-io/baaah/pkg/restconfig"
	brentauth "github.com/acorn-io/brent/pkg/auth"
	authcli "github.com/acorn-io/brent/pkg/auth/cli"
	"github.com/acorn-io/brent/pkg/server"
	"github.com/acorn-io/cmd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type Brent struct {
	cmd.DebugLogging

	Kubeconfig     string `env:"KUBECONFIG"`
	Context        string `env:"CONTEXT"`
	HttpListenPort int    `default:"9080"`

	authcli.WebhookConfig
}

func NewBrent() *cobra.Command {
	return cmd.Command(&Brent{}, cobra.Command{})
}

func (c *Brent) Run(cmd *cobra.Command, args []string) error {
	var (
		auth brentauth.Middleware
	)

	if err := c.DebugLogging.InitLogging(); err != nil {
		return err
	}

	restConfig, err := restconfig.FromFile(c.Kubeconfig, c.Context)
	if err != nil {
		return err
	}
	restConfig.RateLimiter = ratelimit.None

	if c.WebhookConfig.WebhookAuthentication {
		auth, err = c.WebhookConfig.WebhookMiddleware()
		if err != nil {
			return err
		}
	}

	s, err := server.New(cmd.Context(), restConfig, &server.Options{
		AuthMiddleware: auth,
	})
	if err != nil {
		return err
	}

	addr := fmt.Sprintf(":%d", c.HttpListenPort)
	logrus.Info("Listening on " + addr)
	return http.ListenAndServe(addr, s)
}
