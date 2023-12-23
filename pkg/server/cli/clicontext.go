package cli

import (
	"context"

	"github.com/acorn-io/baaah/pkg/ratelimit"
	"github.com/acorn-io/baaah/pkg/restconfig"
	brentauth "github.com/acorn-io/brent/pkg/auth"
	authcli "github.com/acorn-io/brent/pkg/auth/cli"
	"github.com/acorn-io/brent/pkg/server"
	"github.com/urfave/cli"
)

type Config struct {
	KubeConfig     string
	Context        string
	HTTPListenPort int
	UIPath         string

	WebhookConfig authcli.WebhookConfig
}

func (c *Config) MustServer(ctx context.Context) *server.Server {
	cc, err := c.ToServer(ctx)
	if err != nil {
		panic(err)
	}
	return cc
}

func (c *Config) ToServer(ctx context.Context) (*server.Server, error) {
	var (
		auth brentauth.Middleware
	)

	restConfig, err := restconfig.FromFile(c.KubeConfig, c.Context)
	if err != nil {
		return nil, err
	}
	restConfig.RateLimiter = ratelimit.None

	if c.WebhookConfig.WebhookAuthentication {
		auth, err = c.WebhookConfig.WebhookMiddleware()
		if err != nil {
			return nil, err
		}
	}

	return server.New(ctx, restConfig, &server.Options{
		AuthMiddleware: auth,
	})
}

func Flags(config *Config) []cli.Flag {
	flags := []cli.Flag{
		cli.StringFlag{
			Name:        "kubeconfig",
			EnvVar:      "KUBECONFIG",
			Destination: &config.KubeConfig,
		},
		cli.StringFlag{
			Name:        "context",
			EnvVar:      "CONTEXT",
			Destination: &config.Context,
		},
		cli.StringFlag{
			Name:        "ui-path",
			Destination: &config.UIPath,
		},
		cli.IntFlag{
			Name:        "http-listen-port",
			Value:       9080,
			Destination: &config.HTTPListenPort,
		},
	}

	return append(flags, authcli.Flags(&config.WebhookConfig)...)
}
