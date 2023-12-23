package subscribe

import (
	"net/http"

	types2 "github.com/acorn-io/brent/pkg/types"
)

type SchemasGetter func(apiOp *types2.APIRequest) *types2.APISchemas

func DefaultGetter(apiOp *types2.APIRequest) *types2.APISchemas {
	return apiOp.Schemas
}

func Register(schemas *types2.APISchemas, getter SchemasGetter, serverVersion string) {
	if getter == nil {
		getter = DefaultGetter
	}
	schemas.MustImportAndCustomize(Subscribe{}, func(schema *types2.APISchema) {
		schema.CollectionMethods = []string{http.MethodGet}
		schema.ResourceMethods = []string{}
		schema.ListHandler = NewHandler(getter, serverVersion)
		schema.PluralName = "subscribe"
	})
}
