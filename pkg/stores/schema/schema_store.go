package schema

import (
	"github.com/acorn-io/brent/pkg/apierror"
	"github.com/acorn-io/brent/pkg/stores/empty"
	types2 "github.com/acorn-io/brent/pkg/types"
	"github.com/acorn-io/schemer/definition"
	"github.com/acorn-io/schemer/validation"
)

type Store struct {
	empty.Store
}

func NewSchemaStore() types2.Store {
	return &Store{}
}

func toAPIObject(schema *types2.APISchema) types2.APIObject {
	s := schema.DeepCopy()
	delete(s.Schema.Attributes, "access")
	return types2.APIObject{
		Type:   "schema",
		ID:     schema.ID,
		Object: s,
	}
}

func (s *Store) ByID(apiOp *types2.APIRequest, schema *types2.APISchema, id string) (types2.APIObject, error) {
	schema = apiOp.Schemas.LookupSchema(id)
	if schema == nil {
		return types2.APIObject{}, apierror.NewAPIError(validation.NotFound, "no such schema")
	}
	return toAPIObject(schema), nil
}

func (s *Store) List(apiOp *types2.APIRequest, schema *types2.APISchema) (types2.APIObjectList, error) {
	return FilterSchemas(apiOp, apiOp.Schemas.Schemas), nil
}

func FilterSchemas(apiOp *types2.APIRequest, schemaMap map[string]*types2.APISchema) types2.APIObjectList {
	schemas := types2.APIObjectList{}

	included := map[string]bool{}
	for _, schema := range schemaMap {
		if included[schema.ID] {
			continue
		}

		if len(schema.CollectionMethods) > 0 || len(schema.ResourceMethods) > 0 {
			schemas = addSchema(apiOp, schema, schemaMap, schemas, included)
		}
	}

	return schemas
}

func addSchema(apiOp *types2.APIRequest, schema *types2.APISchema, schemaMap map[string]*types2.APISchema, schemas types2.APIObjectList, included map[string]bool) types2.APIObjectList {
	included[schema.ID] = true
	schemas = traverseAndAdd(apiOp, schema, schemaMap, schemas, included)
	schemas.Objects = append(schemas.Objects, toAPIObject(schema))
	return schemas
}

func traverseAndAdd(apiOp *types2.APIRequest, schema *types2.APISchema, schemaMap map[string]*types2.APISchema, schemas types2.APIObjectList, included map[string]bool) types2.APIObjectList {
	for _, field := range schema.ResourceFields {
		t := ""
		subType := field.Type
		for subType != t {
			t = subType
			subType = definition.SubType(t)
		}

		if refSchema, ok := schemaMap[t]; ok && !included[t] {
			schemas = addSchema(apiOp, refSchema, schemaMap, schemas, included)
		}
	}

	for _, action := range schema.ResourceActions {
		for _, t := range []string{action.Output, action.Input} {
			if t == "" {
				continue
			}

			if refSchema, ok := schemaMap[t]; ok && !included[t] {
				schemas = addSchema(apiOp, refSchema, schemaMap, schemas, included)
			}
		}
	}

	for _, action := range schema.CollectionActions {
		for _, t := range []string{action.Output, action.Input} {
			if t == "" {
				continue
			}

			if refSchema, ok := schemaMap[t]; ok && !included[t] {
				schemas = addSchema(apiOp, refSchema, schemaMap, schemas, included)
			}
		}
	}

	return schemas
}
