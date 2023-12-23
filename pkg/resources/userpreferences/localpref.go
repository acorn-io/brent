package userpreferences

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/acorn-io/brent/pkg/stores/empty"
	types2 "github.com/acorn-io/brent/pkg/types"
	"github.com/adrg/xdg"
	"k8s.io/apiserver/pkg/endpoints/request"
)

var (
	rancherSchema = "management.cattle.io.preference"
)

type localStore struct {
	empty.Store
}

func confDir() string {
	return filepath.Join(xdg.ConfigHome, "brent")
}

func confFile() string {
	return filepath.Join(confDir(), "prefs.json")
}

func set(data map[string]interface{}) error {
	if err := os.MkdirAll(confDir(), 0700); err != nil {
		return err
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(confFile(), bytes, 0600)
}

func get() (map[string]string, error) {
	data := UserPreference{}
	f, err := os.Open(confFile())
	if os.IsNotExist(err) {
		return map[string]string{}, nil
	} else if err != nil {
		return nil, err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&data); err != nil {
		return nil, err
	}
	return data.Data, nil
}

func getUserName(apiOp *types2.APIRequest) string {
	user, ok := request.UserFrom(apiOp.Context())
	if !ok {
		return "local"
	}
	return user.GetName()
}

func (l *localStore) ByID(apiOp *types2.APIRequest, schema *types2.APISchema, id string) (types2.APIObject, error) {
	data, err := get()
	if err != nil {
		return types2.APIObject{}, err
	}

	return types2.APIObject{
		Type: "userpreference",
		ID:   getUserName(apiOp),
		Object: UserPreference{
			Data: data,
		},
	}, nil
}

func (l *localStore) List(apiOp *types2.APIRequest, schema *types2.APISchema) (types2.APIObjectList, error) {
	obj, err := l.ByID(apiOp, schema, "")
	if err != nil {
		return types2.APIObjectList{}, err
	}
	return types2.APIObjectList{
		Objects: []types2.APIObject{
			obj,
		},
	}, nil
}

func (l *localStore) Update(apiOp *types2.APIRequest, schema *types2.APISchema, data types2.APIObject, id string) (types2.APIObject, error) {
	err := set(data.Data())
	if err != nil {
		return types2.APIObject{}, err
	}
	return l.ByID(apiOp, schema, "")
}

func (l *localStore) Delete(apiOp *types2.APIRequest, schema *types2.APISchema, id string) (types2.APIObject, error) {
	return l.Update(apiOp, schema, types2.APIObject{
		Object: map[string]interface{}{},
	}, "")
}
