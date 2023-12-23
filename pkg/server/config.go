package server

import (
	"context"

	"github.com/acorn-io/baaah"
	"github.com/acorn-io/baaah/pkg/router"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	apiv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
)

type Controllers struct {
	K8s    kubernetes.Interface
	Router *router.Router
}

func (c *Controllers) Start(ctx context.Context) error {
	return c.Router.Start(ctx)
}

func NewController(cfg *rest.Config) (*Controllers, error) {
	k8s, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	s := runtime.NewScheme()
	if err := scheme.AddToScheme(s); err != nil {
		return nil, err
	}
	if err := apiv1.AddToScheme(s); err != nil {
		return nil, err
	}

	router, err := baaah.NewRouter("brent", &baaah.Options{
		DefaultRESTConfig: cfg,
		Scheme:            s,
	})

	return &Controllers{
		K8s:    k8s,
		Router: router,
	}, nil
}
