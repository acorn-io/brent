package resources

import (
	"context"

	"github.com/acorn-io/brent/pkg/accesscontrol"
	"github.com/acorn-io/brent/pkg/client"
	"github.com/acorn-io/brent/pkg/clustercache"
	"github.com/acorn-io/brent/pkg/rancher-apiserver/pkg/store/apiroot"
	"github.com/acorn-io/brent/pkg/rancher-apiserver/pkg/subscribe"
	"github.com/acorn-io/brent/pkg/rancher-apiserver/pkg/types"
	"github.com/acorn-io/brent/pkg/resources/apigroups"
	"github.com/acorn-io/brent/pkg/resources/cluster"
	"github.com/acorn-io/brent/pkg/resources/common"
	"github.com/acorn-io/brent/pkg/resources/counts"
	"github.com/acorn-io/brent/pkg/resources/formatters"
	"github.com/acorn-io/brent/pkg/resources/userpreferences"
	"github.com/acorn-io/brent/pkg/schema"
	brentschema "github.com/acorn-io/brent/pkg/schema"
	"github.com/acorn-io/brent/pkg/stores/proxy"
	"github.com/acorn-io/brent/pkg/summarycache"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/client-go/discovery"
)

func DefaultSchemas(ctx context.Context, baseSchema *types.APISchemas, ccache clustercache.ClusterCache,
	cg proxy.ClientGetter, schemaFactory brentschema.Factory, serverVersion string) error {
	counts.Register(baseSchema, ccache)
	subscribe.Register(baseSchema, func(apiOp *types.APIRequest) *types.APISchemas {
		user, ok := request.UserFrom(apiOp.Context())
		if ok {
			schemas, err := schemaFactory.Schemas(user)
			if err == nil {
				return schemas
			}
		}
		return apiOp.Schemas
	}, serverVersion)
	apiroot.Register(baseSchema, []string{"v1"}, "proxy:/apis")
	cluster.Register(ctx, baseSchema, cg, schemaFactory)
	userpreferences.Register(baseSchema)
	return nil
}

func DefaultSchemaTemplates(cf *client.Factory,
	baseSchemas *types.APISchemas,
	summaryCache *summarycache.SummaryCache,
	lookup accesscontrol.AccessSetLookup,
	discovery discovery.DiscoveryInterface) []schema.Template {
	return []schema.Template{
		common.DefaultTemplate(cf, summaryCache, lookup),
		apigroups.Template(discovery),
		{
			ID:        "configmap",
			Formatter: formatters.DropHelmData,
		},
		{
			ID:        "secret",
			Formatter: formatters.DropHelmData,
		},
		{
			ID:        "pod",
			Formatter: formatters.Pod,
		},
		{
			ID: "management.cattle.io.cluster",
			Customize: func(apiSchema *types.APISchema) {
				cluster.AddApply(baseSchemas, apiSchema)
			},
		},
	}
}
