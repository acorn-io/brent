package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/acorn-io/brent/pkg/accesscontrol"
	"github.com/acorn-io/brent/pkg/auth"
	"github.com/acorn-io/brent/pkg/client"
	"github.com/acorn-io/brent/pkg/clustercache"
	schemacontroller "github.com/acorn-io/brent/pkg/controllers/schema"
	apiserver "github.com/acorn-io/brent/pkg/rancher-apiserver/pkg/server"
	"github.com/acorn-io/brent/pkg/rancher-apiserver/pkg/types"
	"github.com/acorn-io/brent/pkg/resources"
	"github.com/acorn-io/brent/pkg/resources/common"
	"github.com/acorn-io/brent/pkg/resources/schemas"
	"github.com/acorn-io/brent/pkg/schema"
	"github.com/acorn-io/brent/pkg/server/handler"
	"github.com/acorn-io/brent/pkg/server/router"
	"github.com/acorn-io/brent/pkg/summarycache"
	"github.com/rancher/dynamiclistener/server"
	"k8s.io/client-go/rest"
)

var ErrConfigRequired = errors.New("rest config is required")

type Server struct {
	http.Handler

	ClientFactory   *client.Factory
	ClusterCache    clustercache.ClusterCache
	SchemaFactory   schema.Factory
	RESTConfig      *rest.Config
	BaseSchemas     *types.APISchemas
	AccessSetLookup accesscontrol.AccessSetLookup
	APIServer       *apiserver.Server
	Version         string

	authMiddleware      auth.Middleware
	controllers         *Controllers
	needControllerStart bool
	next                http.Handler
	router              router.RouterFunc
}

type Options struct {
	// Controllers If the controllers are passed in the caller must also start the controllers
	Controllers     *Controllers
	ClientFactory   *client.Factory
	AccessSetLookup accesscontrol.AccessSetLookup
	AuthMiddleware  auth.Middleware
	Next            http.Handler
	Router          router.RouterFunc
	ServerVersion   string
}

func New(ctx context.Context, restConfig *rest.Config, opts *Options) (*Server, error) {
	if opts == nil {
		opts = &Options{}
	}

	server := &Server{
		RESTConfig:      restConfig,
		ClientFactory:   opts.ClientFactory,
		AccessSetLookup: opts.AccessSetLookup,
		authMiddleware:  opts.AuthMiddleware,
		controllers:     opts.Controllers,
		next:            opts.Next,
		router:          opts.Router,
		Version:         opts.ServerVersion,
	}

	if err := setup(ctx, server); err != nil {
		return nil, err
	}

	return server, server.start(ctx)
}

func setDefaults(server *Server) error {
	if server.RESTConfig == nil {
		return ErrConfigRequired
	}

	if server.controllers == nil {
		var err error
		server.controllers, err = NewController(server.RESTConfig, nil)
		server.needControllerStart = true
		if err != nil {
			return err
		}
	}

	if server.next == nil {
		server.next = http.NotFoundHandler()
	}

	if server.BaseSchemas == nil {
		server.BaseSchemas = types.EmptyAPISchemas()
	}

	return nil
}

func setup(ctx context.Context, server *Server) error {
	err := setDefaults(server)
	if err != nil {
		return err
	}

	cf := server.ClientFactory
	if cf == nil {
		cf, err = client.NewFactory(server.RESTConfig, server.authMiddleware != nil)
		if err != nil {
			return err
		}
		server.ClientFactory = cf
	}

	asl := server.AccessSetLookup
	if asl == nil {
		asl = accesscontrol.NewAccessStore(ctx, true, server.controllers.RBAC)
	}

	ccache := clustercache.NewClusterCache(ctx, cf.AdminDynamicClient())
	server.ClusterCache = ccache
	sf := schema.NewCollection(ctx, server.BaseSchemas, asl)

	if err = resources.DefaultSchemas(ctx, server.BaseSchemas, ccache, server.ClientFactory, sf, server.Version); err != nil {
		return err
	}

	summaryCache := summarycache.New(sf, ccache)
	summaryCache.Start(ctx)

	for _, template := range resources.DefaultSchemaTemplates(cf, server.BaseSchemas, summaryCache, asl, server.controllers.K8s.Discovery()) {
		sf.AddTemplate(template)
	}

	cols, err := common.NewDynamicColumns(server.RESTConfig)
	if err != nil {
		return err
	}

	schemas.SetupWatcher(ctx, server.BaseSchemas, asl, sf)

	schemacontroller.Register(ctx,
		cols,
		server.controllers.K8s.Discovery(),
		server.controllers.CRD.CustomResourceDefinition(),
		server.controllers.API.APIService(),
		server.controllers.K8s.AuthorizationV1().SelfSubjectAccessReviews(),
		ccache,
		sf)

	apiServer, handler, err := handler.New(server.RESTConfig, sf, server.authMiddleware, server.next, server.router)
	if err != nil {
		return err
	}

	server.APIServer = apiServer
	server.Handler = handler
	server.SchemaFactory = sf
	return nil
}

func (c *Server) start(ctx context.Context) error {
	if c.needControllerStart {
		if err := c.controllers.Start(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (c *Server) ListenAndServe(ctx context.Context, httpsPort, httpPort int, opts *server.ListenOpts) error {
	if opts == nil {
		opts = &server.ListenOpts{}
	}
	if opts.Storage == nil && opts.Secrets == nil {
		opts.Secrets = c.controllers.Core.Secret()
	}

	if err := server.ListenAndServe(ctx, httpsPort, httpPort, c, opts); err != nil {
		return err
	}

	<-ctx.Done()
	return ctx.Err()
}
