package handlers

import (
	"net/http"

	"github.com/acorn-io/brent/pkg/apierror"
	"github.com/acorn-io/brent/pkg/parse"
	"github.com/acorn-io/brent/pkg/types"
	"github.com/acorn-io/schemer/validation"
)

func UpdateHandler(apiOp *types.APIRequest) (types.APIObject, error) {
	if err := apiOp.AccessControl.CanUpdate(apiOp, types.APIObject{}, apiOp.Schema); err != nil {
		return types.APIObject{}, err
	}

	var (
		data types.APIObject
		err  error
	)
	if apiOp.Method != http.MethodPatch {
		data, err = parse.Body(apiOp.Request)
		if err != nil {
			return types.APIObject{}, err
		}
	}

	store := apiOp.Schema.Store
	if store == nil {
		return types.APIObject{}, apierror.NewAPIError(validation.NotFound, "no store found")
	}

	data, err = store.Update(apiOp, apiOp.Schema, data, apiOp.Name)
	if err != nil {
		return types.APIObject{}, err
	}

	return data, nil
}
