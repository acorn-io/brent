package cli

import (
	"os"
	"time"

	"github.com/acorn-io/baaah/pkg/ratelimit"
	"github.com/acorn-io/brent/pkg/auth"
	"k8s.io/client-go/tools/clientcmd"
)

type WebhookConfig struct {
	WebhookAuthentication  bool
	WebhookKubeconfig      string
	WebhookURL             string
	WebhookCacheTTLSeconds int
}

func (w *WebhookConfig) MustWebhookMiddleware() auth.Middleware {
	m, err := w.WebhookMiddleware()
	if err != nil {
		panic("failed to create webhook middleware: " + err.Error())
	}
	return m
}

func (w *WebhookConfig) WebhookMiddleware() (auth.Middleware, error) {
	if !w.WebhookAuthentication {
		return nil, nil
	}

	config := w.WebhookKubeconfig
	if config == "" && w.WebhookURL != "" {
		tempFile, err := auth.WebhookConfigForURL(w.WebhookURL)
		if err != nil {
			return nil, err
		}
		defer os.Remove(tempFile)
		config = tempFile
	}

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", config)
	if err != nil {
		return nil, err
	}

	kubeConfig.RateLimiter = ratelimit.None
	return auth.NewWebhookMiddleware(time.Duration(w.WebhookCacheTTLSeconds)*time.Second, kubeConfig)
}
