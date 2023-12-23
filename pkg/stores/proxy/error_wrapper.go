package proxy

import (
	"github.com/acorn-io/brent/pkg/apierror"
	types2 "github.com/acorn-io/brent/pkg/types"
	"github.com/acorn-io/schemer/validation"
	"k8s.io/apimachinery/pkg/api/errors"
)

type errorStore struct {
	types2.Store
}

func (e *errorStore) ByID(apiOp *types2.APIRequest, schema *types2.APISchema, id string) (types2.APIObject, error) {
	data, err := e.Store.ByID(apiOp, schema, id)
	return data, translateError(err)
}

func (e *errorStore) List(apiOp *types2.APIRequest, schema *types2.APISchema) (types2.APIObjectList, error) {
	data, err := e.Store.List(apiOp, schema)
	return data, translateError(err)
}

func (e *errorStore) Create(apiOp *types2.APIRequest, schema *types2.APISchema, data types2.APIObject) (types2.APIObject, error) {
	data, err := e.Store.Create(apiOp, schema, data)
	return data, translateError(err)

}

func (e *errorStore) Update(apiOp *types2.APIRequest, schema *types2.APISchema, data types2.APIObject, id string) (types2.APIObject, error) {
	data, err := e.Store.Update(apiOp, schema, data, id)
	return data, translateError(err)

}

func (e *errorStore) Delete(apiOp *types2.APIRequest, schema *types2.APISchema, id string) (types2.APIObject, error) {
	data, err := e.Store.Delete(apiOp, schema, id)
	return data, translateError(err)

}

func (e *errorStore) Watch(apiOp *types2.APIRequest, schema *types2.APISchema, wr types2.WatchRequest) (chan types2.APIEvent, error) {
	data, err := e.Store.Watch(apiOp, schema, wr)
	return data, translateError(err)
}

func translateError(err error) error {
	if apiError, ok := err.(errors.APIStatus); ok {
		status := apiError.Status()
		return apierror.NewAPIError(validation.ErrorCode{
			Status: int(status.Code),
			Code:   string(status.Reason),
		}, status.Message)
	}
	return err
}
