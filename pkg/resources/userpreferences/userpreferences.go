package userpreferences

import (
	"net/http"

	types2 "github.com/acorn-io/brent/pkg/types"
)

type UserPreference struct {
	Data map[string]string `json:"data"`
}

func Register(schemas *types2.APISchemas) {
	schemas.InternalSchemas.TypeName("userpreference", UserPreference{})
	schemas.MustImportAndCustomize(UserPreference{}, func(schema *types2.APISchema) {
		schema.CollectionMethods = []string{http.MethodGet}
		schema.ResourceMethods = []string{http.MethodGet, http.MethodPut, http.MethodDelete}
		schema.Store = &localStore{}
	})
}
