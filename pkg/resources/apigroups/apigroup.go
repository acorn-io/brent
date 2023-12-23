package apigroups

import (
	"net/http"

	"github.com/acorn-io/brent/pkg/schema"
	"github.com/acorn-io/brent/pkg/stores/empty"
	types2 "github.com/acorn-io/brent/pkg/types"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
)

func Template(discovery discovery.DiscoveryInterface) schema.Template {
	return schema.Template{
		ID: "apigroup",
		Customize: func(apiSchema *types2.APISchema) {
			apiSchema.CollectionMethods = []string{http.MethodGet}
			apiSchema.ResourceMethods = []string{http.MethodGet}
		},
		Formatter: func(request *types2.APIRequest, resource *types2.RawResource) {
			resource.ID = resource.APIObject.Data().String("name")
		},
		Store: NewStore(discovery),
	}
}

type Store struct {
	empty.Store

	discovery discovery.DiscoveryInterface
}

func NewStore(discovery discovery.DiscoveryInterface) types2.Store {
	return &Store{
		Store:     empty.Store{},
		discovery: discovery,
	}
}

func (e *Store) ByID(apiOp *types2.APIRequest, schema *types2.APISchema, id string) (types2.APIObject, error) {
	return types2.DefaultByID(e, apiOp, schema, id)
}

func toAPIObject(schema *types2.APISchema, group v1.APIGroup) types2.APIObject {
	if group.Name == "" {
		group.Name = "core"
	}
	return types2.APIObject{
		Type:   schema.ID,
		ID:     group.Name,
		Object: group,
	}

}

func (e *Store) List(apiOp *types2.APIRequest, schema *types2.APISchema) (types2.APIObjectList, error) {
	groupList, err := e.discovery.ServerGroups()
	if err != nil {
		return types2.APIObjectList{}, err
	}

	var result types2.APIObjectList
	for _, item := range groupList.Groups {
		result.Objects = append(result.Objects, toAPIObject(schema, item))
	}

	return result, nil
}
