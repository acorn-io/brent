package switchschema

import (
	types2 "github.com/acorn-io/brent/pkg/types"
)

type Store struct {
	Schema *types2.APISchema
}

func (e *Store) Delete(apiOp *types2.APIRequest, oldSchema *types2.APISchema, id string) (types2.APIObject, error) {
	obj, err := e.Schema.Store.Delete(apiOp, e.Schema, id)
	obj.Type = oldSchema.ID
	return obj, err
}

func (e *Store) ByID(apiOp *types2.APIRequest, oldSchema *types2.APISchema, id string) (types2.APIObject, error) {
	obj, err := e.Schema.Store.ByID(apiOp, e.Schema, id)
	obj.Type = oldSchema.ID
	return obj, err
}

func (e *Store) List(apiOp *types2.APIRequest, oldSchema *types2.APISchema) (types2.APIObjectList, error) {
	obj, err := e.Schema.Store.List(apiOp, e.Schema)
	for i := range obj.Objects {
		obj.Objects[i].Type = oldSchema.ID
	}
	return obj, err
}

func (e *Store) Create(apiOp *types2.APIRequest, oldSchema *types2.APISchema, data types2.APIObject) (types2.APIObject, error) {
	obj, err := e.Schema.Store.Create(apiOp, e.Schema, data)
	obj.Type = oldSchema.ID
	return obj, err
}

func (e *Store) Update(apiOp *types2.APIRequest, oldSchema *types2.APISchema, data types2.APIObject, id string) (types2.APIObject, error) {
	obj, err := e.Schema.Store.Update(apiOp, e.Schema, data, id)
	obj.Type = oldSchema.ID
	return obj, err
}

func (e *Store) Watch(apiOp *types2.APIRequest, oldSchema *types2.APISchema, wr types2.WatchRequest) (chan types2.APIEvent, error) {
	c, err := e.Schema.Store.Watch(apiOp, e.Schema, wr)
	if err != nil || c == nil {
		return c, err
	}

	result := make(chan types2.APIEvent)
	go func() {
		defer close(result)
		for obj := range c {
			if obj.Object.Type == e.Schema.ID {
				obj.Object.Type = oldSchema.ID
			}
			result <- obj
		}
	}()

	return result, nil
}
