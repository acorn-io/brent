package handlers

import (
	"github.com/acorn-io/brent/pkg/rancher-apiserver/pkg/apierror"
	"github.com/acorn-io/brent/pkg/rancher-apiserver/pkg/parse"
	"github.com/acorn-io/brent/pkg/rancher-apiserver/pkg/types"
	"github.com/acorn-io/brent/pkg/schemas/validation"
)

func CreateHandler(apiOp *types.APIRequest) (types.APIObject, error) {
	var err error

	if err := apiOp.AccessControl.CanCreate(apiOp, apiOp.Schema); err != nil {
		return types.APIObject{}, err
	}

	data, err := parse.Body(apiOp.Request)
	if err != nil {
		return types.APIObject{}, err
	}

	store := apiOp.Schema.Store
	if store == nil {
		return types.APIObject{}, apierror.NewAPIError(validation.NotFound, "no store found")
	}

	data, err = store.Create(apiOp, apiOp.Schema, data)
	if err != nil {
		return types.APIObject{}, err
	}

	return data, nil
}
