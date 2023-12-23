package apiroot

import (
	"net/http"
	"path"
	"strings"

	"github.com/acorn-io/brent/pkg/stores/empty"
	types2 "github.com/acorn-io/brent/pkg/types"
	"github.com/acorn-io/schemer"
)

func Register(apiSchemas *types2.APISchemas, versions []string, roots ...string) {
	apiSchemas.MustAddSchema(types2.APISchema{
		Schema: &schemas.Schema{
			ID:                "apiRoot",
			CollectionMethods: []string{"GET"},
			ResourceMethods:   []string{"GET"},
			ResourceFields: map[string]schemas.Field{
				"apiVersion": {Type: "map[json]"},
				"path":       {Type: "string"},
			},
		},
		Formatter: Formatter,
		Store:     NewAPIRootStore(versions, roots),
	})
}

func Formatter(apiOp *types2.APIRequest, resource *types2.RawResource) {
	data := resource.APIObject.Data()
	path, _ := data["path"].(string)
	if path == "" {
		return
	}
	delete(data, "path")

	resource.Links["root"] = apiOp.URLBuilder.RelativeToRoot(path)

	if data, isAPIRoot := data["apiVersion"].(map[string]interface{}); isAPIRoot {
		apiVersion := apiVersionFromMap(apiOp.Schemas, data)
		for _, schema := range apiOp.Schemas.Schemas {
			addCollectionLink(apiOp, schema, apiVersion, resource.Links)
		}
		resource.Links["self"] = apiOp.URLBuilder.RelativeToRoot(apiVersion)
		resource.Links["schemas"] = apiOp.URLBuilder.RelativeToRoot(path)
	}

	return
}

func addCollectionLink(apiOp *types2.APIRequest, schema *types2.APISchema, apiVersion string, links map[string]string) {
	collectionLink := getSchemaCollectionLink(apiOp, schema)
	if collectionLink != "" {
		links[schema.PluralName] = apiOp.URLBuilder.RelativeToRoot(path.Join(apiVersion, path.Base(collectionLink)))
	}
}

func getSchemaCollectionLink(apiOp *types2.APIRequest, schema *types2.APISchema) string {
	if schema != nil && contains(schema.CollectionMethods, http.MethodGet) {
		return apiOp.URLBuilder.Collection(schema)
	}
	return ""
}

type Store struct {
	empty.Store
	roots    []string
	versions []string
}

func NewAPIRootStore(versions []string, roots []string) types2.Store {
	return &Store{
		roots:    roots,
		versions: versions,
	}
}

func (a *Store) ByID(apiOp *types2.APIRequest, schema *types2.APISchema, id string) (types2.APIObject, error) {
	return types2.DefaultByID(a, apiOp, schema, id)
}

func (a *Store) List(apiOp *types2.APIRequest, schema *types2.APISchema) (types2.APIObjectList, error) {
	var roots types2.APIObjectList

	versions := a.versions

	for _, version := range versions {
		roots.Objects = append(roots.Objects, types2.APIObject{
			Type:   "apiRoot",
			ID:     version,
			Object: apiVersionToAPIRootMap(version),
		})
	}

	for _, root := range a.roots {
		parts := strings.SplitN(root, ":", 2)
		if len(parts) == 2 {
			roots.Objects = append(roots.Objects, types2.APIObject{
				Type: "apiRoot",
				ID:   parts[0],
				Object: map[string]interface{}{
					"id":   parts[0],
					"path": parts[1],
				},
			})
		}
	}

	return roots, nil
}

func apiVersionToAPIRootMap(version string) map[string]interface{} {
	return map[string]interface{}{
		"id":   version,
		"type": "apiRoot",
		"apiVersion": map[string]interface{}{
			"version": version,
		},
		"path": "/" + version,
	}
}

func apiVersionFromMap(schemas *types2.APISchemas, apiVersion map[string]interface{}) string {
	version, _ := apiVersion["version"].(string)
	return version
}

func contains(list []string, needle string) bool {
	for _, v := range list {
		if v == needle {
			return true
		}
	}
	return false
}
