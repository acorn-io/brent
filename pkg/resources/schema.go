package resources

import (
	"github.com/acorn-io/brent/pkg/accesscontrol"
	"github.com/acorn-io/brent/pkg/client"
	"github.com/acorn-io/brent/pkg/clustercache"
	"github.com/acorn-io/brent/pkg/resources/apigroups"
	"github.com/acorn-io/brent/pkg/resources/common"
	"github.com/acorn-io/brent/pkg/resources/counts"
	"github.com/acorn-io/brent/pkg/resources/userpreferences"
	"github.com/acorn-io/brent/pkg/schema"
	brentschema "github.com/acorn-io/brent/pkg/schema"
	"github.com/acorn-io/brent/pkg/stores/apiroot"
	"github.com/acorn-io/brent/pkg/subscribe"
	"github.com/acorn-io/brent/pkg/summarycache"
	types2 "github.com/acorn-io/brent/pkg/types"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/client-go/discovery"
)

func DefaultSchemas(baseSchema *types2.APISchemas, ccache clustercache.ClusterCache,
	schemaFactory brentschema.Factory, serverVersion string) error {
	counts.Register(baseSchema, ccache)
	subscribe.Register(baseSchema, func(apiOp *types2.APIRequest) *types2.APISchemas {
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
	userpreferences.Register(baseSchema)
	return nil
}

func DefaultSchemaTemplates(cf *client.Factory,
	summaryCache *summarycache.SummaryCache,
	lookup accesscontrol.AccessSetLookup,
	discovery discovery.DiscoveryInterface) []schema.Template {
	return []schema.Template{
		common.DefaultTemplate(cf, summaryCache, lookup),
		apigroups.Template(discovery),
	}
}
