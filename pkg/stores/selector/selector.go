package selector

import (
	types2 "github.com/acorn-io/brent/pkg/types"
	"k8s.io/apimachinery/pkg/labels"
)

type Store struct {
	types2.Store
	Selector labels.Selector
}

func (s *Store) List(apiOp *types2.APIRequest, schema *types2.APISchema) (types2.APIObjectList, error) {
	return s.Store.List(s.addSelector(apiOp), schema)
}

func (s *Store) addSelector(apiOp *types2.APIRequest) *types2.APIRequest {

	apiOp = apiOp.Clone()
	apiOp.Request = apiOp.Request.Clone(apiOp.Context())
	q := apiOp.Request.URL.Query()
	q.Add("labelSelector", s.Selector.String())
	apiOp.Request.URL.RawQuery = q.Encode()
	return apiOp
}

func (s *Store) Watch(apiOp *types2.APIRequest, schema *types2.APISchema, w types2.WatchRequest) (chan types2.APIEvent, error) {
	return s.Store.Watch(s.addSelector(apiOp), schema, w)
}
