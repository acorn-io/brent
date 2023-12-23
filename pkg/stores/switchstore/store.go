package switchstore

import (
	types2 "github.com/acorn-io/brent/pkg/types"
)

type StorePicker func(apiOp *types2.APIRequest, schema *types2.APISchema, verb, id string) (types2.Store, error)

type Store struct {
	Picker StorePicker
}

func (e *Store) Delete(apiOp *types2.APIRequest, schema *types2.APISchema, id string) (types2.APIObject, error) {
	s, err := e.Picker(apiOp, schema, "delete", id)
	if err != nil {
		return types2.APIObject{}, err
	}
	return s.Delete(apiOp, schema, id)
}

func (e *Store) ByID(apiOp *types2.APIRequest, schema *types2.APISchema, id string) (types2.APIObject, error) {
	s, err := e.Picker(apiOp, schema, "get", id)
	if err != nil {
		return types2.APIObject{}, err
	}
	return s.ByID(apiOp, schema, id)
}

func (e *Store) List(apiOp *types2.APIRequest, schema *types2.APISchema) (types2.APIObjectList, error) {
	s, err := e.Picker(apiOp, schema, "list", "")
	if err != nil {
		return types2.APIObjectList{}, err
	}
	return s.List(apiOp, schema)
}

func (e *Store) Create(apiOp *types2.APIRequest, schema *types2.APISchema, data types2.APIObject) (types2.APIObject, error) {
	s, err := e.Picker(apiOp, schema, "create", "")
	if err != nil {
		return types2.APIObject{}, err
	}
	return s.Create(apiOp, schema, data)
}

func (e *Store) Update(apiOp *types2.APIRequest, schema *types2.APISchema, data types2.APIObject, id string) (types2.APIObject, error) {
	s, err := e.Picker(apiOp, schema, "update", id)
	if err != nil {
		return types2.APIObject{}, err
	}
	return s.Update(apiOp, schema, data, id)
}

func (e *Store) Watch(apiOp *types2.APIRequest, schema *types2.APISchema, wr types2.WatchRequest) (chan types2.APIEvent, error) {
	s, err := e.Picker(apiOp, schema, "watch", "")
	if err != nil {
		return nil, err
	}
	return s.Watch(apiOp, schema, wr)
}
