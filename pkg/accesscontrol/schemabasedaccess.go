package accesscontrol

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/acorn-io/brent/pkg/apierror"
	types2 "github.com/acorn-io/brent/pkg/types"
	"github.com/acorn-io/schemer/validation"
)

type SchemaBasedAccess struct {
}

func (*SchemaBasedAccess) CanCreate(apiOp *types2.APIRequest, schema *types2.APISchema) error {
	if slices.Contains(schema.CollectionMethods, http.MethodPost) {
		return nil
	}
	return apierror.NewAPIError(validation.PermissionDenied, "can not create "+schema.ID)
}

func (*SchemaBasedAccess) CanGet(apiOp *types2.APIRequest, schema *types2.APISchema) error {
	if slices.Contains(schema.ResourceMethods, http.MethodGet) {
		return nil
	}
	return apierror.NewAPIError(validation.PermissionDenied, "can not get "+schema.ID)
}

func (*SchemaBasedAccess) CanList(apiOp *types2.APIRequest, schema *types2.APISchema) error {
	if slices.Contains(schema.CollectionMethods, http.MethodGet) || slices.Contains(schema.CollectionMethods, http.MethodPost) {
		return nil
	}
	return apierror.NewAPIError(validation.PermissionDenied, "can not list "+schema.ID)
}

func (*SchemaBasedAccess) CanUpdate(apiOp *types2.APIRequest, obj types2.APIObject, schema *types2.APISchema) error {
	if slices.Contains(schema.ResourceMethods, http.MethodPut) {
		return nil
	}
	return apierror.NewAPIError(validation.PermissionDenied, "can not update "+schema.ID)
}

func (*SchemaBasedAccess) CanDelete(apiOp *types2.APIRequest, obj types2.APIObject, schema *types2.APISchema) error {
	if slices.Contains(schema.ResourceMethods, http.MethodDelete) {
		return nil
	}
	return apierror.NewAPIError(validation.PermissionDenied, "can not delete "+schema.ID)
}

func (a *SchemaBasedAccess) CanWatch(apiOp *types2.APIRequest, schema *types2.APISchema) error {
	return a.CanList(apiOp, schema)
}

func (a *SchemaBasedAccess) CanDo(apiOp *types2.APIRequest, resource, verb, namespace, name string) error {
	schema := apiOp.Schemas.LookupSchema(resource)
	if schema == nil {
		return apierror.NewAPIError(validation.PermissionDenied, fmt.Sprintf("can not %s %s %s/%s"+verb, resource, namespace, name))
	}
	switch verb {
	case http.MethodGet:
		return a.CanList(apiOp, schema)
	case http.MethodDelete:
		return a.CanDelete(apiOp, types2.APIObject{}, schema)
	case http.MethodPut:
		return a.CanUpdate(apiOp, types2.APIObject{}, schema)
	case http.MethodPost:
		return a.CanCreate(apiOp, schema)
	default:
		return apierror.NewAPIError(validation.PermissionDenied, fmt.Sprintf("can not %s %s %s/%s"+verb, schema.ID, namespace, name))
	}
}

func (*SchemaBasedAccess) CanAction(apiOp *types2.APIRequest, schema *types2.APISchema, name string) error {
	if _, ok := schema.ActionHandlers[name]; !ok {
		return apierror.NewAPIError(validation.PermissionDenied, "no such action "+name)
	}
	return nil
}
