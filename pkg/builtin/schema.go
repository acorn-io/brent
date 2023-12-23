package builtin

import (
	"net/http"
	"slices"

	"github.com/acorn-io/brent/pkg/stores/schema"
	types2 "github.com/acorn-io/brent/pkg/types"
	"github.com/acorn-io/schemer"
)

var (
	Schema = types2.APISchema{
		Schema: &schemas.Schema{
			ID:                "schema",
			PluralName:        "schemas",
			CollectionMethods: []string{"GET"},
			ResourceMethods:   []string{"GET"},
			ResourceFields: map[string]schemas.Field{
				"collectionActions": {Type: "map[json]"},
				"collectionFields":  {Type: "map[json]"},
				"collectionFilters": {Type: "map[json]"},
				"collectionMethods": {Type: "array[string]"},
				"pluralName":        {Type: "string"},
				"resourceActions":   {Type: "map[json]"},
				"attributes":        {Type: "map[json]"},
				"resourceFields":    {Type: "map[json]"},
				"resourceMethods":   {Type: "array[string]"},
				"version":           {Type: "map[json]"},
			},
		},
		Formatter: SchemaFormatter,
		Store:     schema.NewSchemaStore(),
	}

	Error = types2.APISchema{
		Schema: &schemas.Schema{
			ID:                "error",
			ResourceMethods:   []string{},
			CollectionMethods: []string{},
			ResourceFields: map[string]schemas.Field{
				"code":      {Type: "string"},
				"detail":    {Type: "string", Nullable: true},
				"message":   {Type: "string", Nullable: true},
				"fieldName": {Type: "string", Nullable: true},
				"status":    {Type: "int"},
			},
		},
	}

	Collection = types2.APISchema{
		Schema: &schemas.Schema{
			ID:                "collection",
			ResourceMethods:   []string{},
			CollectionMethods: []string{},
			ResourceFields: map[string]schemas.Field{
				"data":       {Type: "array[json]"},
				"pagination": {Type: "map[json]"},
				"sort":       {Type: "map[json]"},
				"filters":    {Type: "map[json]"},
			},
		},
	}

	Schemas = types2.EmptyAPISchemas().
		MustAddSchema(Schema).
		MustAddSchema(Error).
		MustAddSchema(Collection)
)

func SchemaFormatter(apiOp *types2.APIRequest, resource *types2.RawResource) {
	schema, ok := resource.APIObject.Object.(*types2.APISchema)
	if !ok {
		return
	}

	collectionLink := getSchemaCollectionLink(apiOp, schema)
	if collectionLink != "" {
		resource.Links["collection"] = collectionLink
	}
}

func getSchemaCollectionLink(apiOp *types2.APIRequest, schema *types2.APISchema) string {
	if schema != nil && (slices.Contains(schema.CollectionMethods, http.MethodGet) || slices.Contains(schema.CollectionMethods, http.MethodPost)) {
		return apiOp.URLBuilder.Collection(schema)
	}
	return ""
}
