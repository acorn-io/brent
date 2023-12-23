package empty

import (
	types2 "github.com/acorn-io/brent/pkg/types"
	"github.com/acorn-io/schemer/validation"
)

type Store struct {
}

func (e *Store) Delete(apiOp *types2.APIRequest, schema *types2.APISchema, id string) (types2.APIObject, error) {
	return types2.APIObject{}, validation.NotFound
}

func (e *Store) ByID(apiOp *types2.APIRequest, schema *types2.APISchema, id string) (types2.APIObject, error) {
	return types2.APIObject{}, validation.NotFound
}

func (e *Store) List(apiOp *types2.APIRequest, schema *types2.APISchema) (types2.APIObjectList, error) {
	return types2.APIObjectList{}, validation.NotFound
}

func (e *Store) Create(apiOp *types2.APIRequest, schema *types2.APISchema, data types2.APIObject) (types2.APIObject, error) {
	return types2.APIObject{}, validation.NotFound
}

func (e *Store) Update(apiOp *types2.APIRequest, schema *types2.APISchema, data types2.APIObject, id string) (types2.APIObject, error) {
	return types2.APIObject{}, validation.NotFound
}

func (e *Store) Watch(apiOp *types2.APIRequest, schema *types2.APISchema, wr types2.WatchRequest) (chan types2.APIEvent, error) {
	return nil, nil
}
