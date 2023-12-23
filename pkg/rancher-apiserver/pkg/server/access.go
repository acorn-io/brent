package server

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/acorn-io/brent/pkg/rancher-apiserver/pkg/apierror"
	"github.com/acorn-io/brent/pkg/rancher-apiserver/pkg/types"
	"github.com/acorn-io/brent/pkg/schemas/validation"
)

type SchemaBasedAccess struct {
}

func (*SchemaBasedAccess) CanCreate(apiOp *types.APIRequest, schema *types.APISchema) error {
	if slices.Contains(schema.CollectionMethods, http.MethodPost) {
		return nil
	}
	return apierror.NewAPIError(validation.PermissionDenied, "can not create "+schema.ID)
}

func (*SchemaBasedAccess) CanGet(apiOp *types.APIRequest, schema *types.APISchema) error {
	if slices.Contains(schema.ResourceMethods, http.MethodGet) {
		return nil
	}
	return apierror.NewAPIError(validation.PermissionDenied, "can not get "+schema.ID)
}

func (*SchemaBasedAccess) CanList(apiOp *types.APIRequest, schema *types.APISchema) error {
	if slices.Contains(schema.CollectionMethods, http.MethodGet) || slices.Contains(schema.CollectionMethods, http.MethodPost) {
		return nil
	}
	return apierror.NewAPIError(validation.PermissionDenied, "can not list "+schema.ID)
}

func (*SchemaBasedAccess) CanUpdate(apiOp *types.APIRequest, obj types.APIObject, schema *types.APISchema) error {
	if slices.Contains(schema.ResourceMethods, http.MethodPut) {
		return nil
	}
	return apierror.NewAPIError(validation.PermissionDenied, "can not update "+schema.ID)
}

func (*SchemaBasedAccess) CanDelete(apiOp *types.APIRequest, obj types.APIObject, schema *types.APISchema) error {
	if slices.Contains(schema.ResourceMethods, http.MethodDelete) {
		return nil
	}
	return apierror.NewAPIError(validation.PermissionDenied, "can not delete "+schema.ID)
}

func (a *SchemaBasedAccess) CanWatch(apiOp *types.APIRequest, schema *types.APISchema) error {
	return a.CanList(apiOp, schema)
}

func (a *SchemaBasedAccess) CanDo(apiOp *types.APIRequest, resource, verb, namespace, name string) error {
	schema := apiOp.Schemas.LookupSchema(resource)
	if schema == nil {
		return apierror.NewAPIError(validation.PermissionDenied, fmt.Sprintf("can not %s %s %s/%s"+verb, resource, namespace, name))
	}
	switch verb {
	case http.MethodGet:
		return a.CanList(apiOp, schema)
	case http.MethodDelete:
		return a.CanDelete(apiOp, types.APIObject{}, schema)
	case http.MethodPut:
		return a.CanUpdate(apiOp, types.APIObject{}, schema)
	case http.MethodPost:
		return a.CanCreate(apiOp, schema)
	default:
		return apierror.NewAPIError(validation.PermissionDenied, fmt.Sprintf("can not %s %s %s/%s"+verb, schema.ID, namespace, name))
	}
}

func (*SchemaBasedAccess) CanAction(apiOp *types.APIRequest, schema *types.APISchema, name string) error {
	if _, ok := schema.ActionHandlers[name]; !ok {
		return apierror.NewAPIError(validation.PermissionDenied, "no such action "+name)
	}
	return nil
}
