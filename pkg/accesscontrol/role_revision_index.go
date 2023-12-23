package accesscontrol

import (
	"strings"
	"sync"

	"github.com/acorn-io/baaah/pkg/router"
	rbacv1 "k8s.io/api/rbac/v1"
)

type roleRevisionIndex struct {
	roleRevisions sync.Map
}

func newRoleRevision(router *router.Router) *roleRevisionIndex {
	r := &roleRevisionIndex{}
	router.Type(&rbacv1.Role{}).IncludeRemoved().HandlerFunc(r.onRoleChanged)
	router.Type(&rbacv1.ClusterRole{}).IncludeRemoved().HandlerFunc(r.onClusterRoleChanged)
	return r
}

func (r *roleRevisionIndex) roleRevision(namespace, name string) string {
	val, _ := r.roleRevisions.Load(roleKey{
		name:      name,
		namespace: namespace,
	})
	revision, _ := val.(string)
	return revision
}

func (r *roleRevisionIndex) onClusterRoleChanged(req router.Request, resp router.Response) error {
	if req.Object == nil {
		r.roleRevisions.Delete(roleKey{
			name: req.Key,
		})
	} else {
		r.roleRevisions.Store(roleKey{
			name: req.Key,
		}, req.Object.GetResourceVersion())
	}
	return nil
}

func (r *roleRevisionIndex) onRoleChanged(req router.Request, resp router.Response) error {
	if req.Object == nil {
		namespace, name, _ := strings.Cut(req.Key, "/")
		r.roleRevisions.Delete(roleKey{
			name:      name,
			namespace: namespace,
		})
	} else {
		r.roleRevisions.Store(roleKey{
			name:      req.Name,
			namespace: req.Namespace,
		}, req.Object.GetResourceVersion())
	}
	return nil
}
