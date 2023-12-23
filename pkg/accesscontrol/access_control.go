package accesscontrol

import (
	"strings"

	"github.com/acorn-io/brent/pkg/attributes"
	types2 "github.com/acorn-io/brent/pkg/types"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type AccessControl struct {
	SchemaBasedAccess
}

func NewAccessControl() *AccessControl {
	return &AccessControl{}
}

func (a *AccessControl) CanDo(apiOp *types2.APIRequest, resource, verb, namespace, name string) error {
	apiSchema := apiOp.Schemas.LookupSchema(resource)
	if apiSchema != nil && attributes.GVK(apiSchema).Kind != "" {
		access := GetAccessListMap(apiSchema)
		if access[verb].Grants(namespace, name) {
			return nil
		}
	}
	group, resource, _ := strings.Cut(resource, "/")
	accessSet := apiOp.Schemas.Attributes["accessSet"].(*AccessSet)
	if accessSet.Grants(verb, schema.GroupResource{
		Group:    group,
		Resource: resource,
	}, namespace, name) {
		return nil
	}
	return a.SchemaBasedAccess.CanDo(apiOp, resource, verb, namespace, name)
}

func (a *AccessControl) CanWatch(apiOp *types2.APIRequest, schema *types2.APISchema) error {
	if attributes.GVK(schema).Kind != "" {
		access := GetAccessListMap(schema)
		if _, ok := access["watch"]; ok {
			return nil
		}
	}
	return a.SchemaBasedAccess.CanWatch(apiOp, schema)
}
