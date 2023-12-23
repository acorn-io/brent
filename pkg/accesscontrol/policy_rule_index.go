package accesscontrol

import (
	"context"
	"fmt"
	"hash"
	"sort"

	"github.com/acorn-io/baaah/pkg/router"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	rbacGroup = "rbac.authorization.k8s.io"
	All       = "*"
)

type policyRuleIndex struct {
	ctx                 context.Context
	client              kclient.Reader
	revisions           *roleRevisionIndex
	kind                string
	roleIndexKey        string
	clusterRoleIndexKey string
}

func newPolicyRuleIndex(ctx context.Context, user bool, revisions *roleRevisionIndex, router *router.Router) (*policyRuleIndex, error) {
	key := "Group"
	if user {
		key = "User"
	}
	pi := &policyRuleIndex{
		ctx:                 ctx,
		kind:                key,
		client:              router.Backend(),
		clusterRoleIndexKey: "crb" + key,
		roleIndexKey:        "rb" + key,
		revisions:           revisions,
	}

	if err := router.Backend().IndexField(ctx, &rbacv1.ClusterRoleBinding{}, pi.clusterRoleIndexKey, pi.clusterRoleBindingBySubjectIndexer); err != nil {
		return nil, err
	}
	if err := router.Backend().IndexField(ctx, &rbacv1.RoleBinding{}, pi.clusterRoleIndexKey, pi.roleBindingBySubject); err != nil {
		return nil, err
	}

	return pi, nil
}

func (p *policyRuleIndex) clusterRoleBindingBySubjectIndexer(obj kclient.Object) (result []string) {
	crb := obj.(*rbacv1.ClusterRoleBinding)
	for _, subject := range crb.Subjects {
		if subject.APIGroup == rbacGroup && subject.Kind == p.kind && crb.RoleRef.Kind == "ClusterRole" {
			result = append(result, subject.Name)
		} else if subject.APIGroup == "" && p.kind == "User" && subject.Kind == "ServiceAccount" && subject.Namespace != "" && crb.RoleRef.Kind == "ClusterRole" {
			// Index is for Users and this references a service account
			result = append(result, fmt.Sprintf("serviceaccount:%s:%s", subject.Namespace, subject.Name))
		}
	}
	return
}

func (p *policyRuleIndex) roleBindingBySubject(obj kclient.Object) (result []string) {
	rb := obj.(*rbacv1.RoleBinding)
	for _, subject := range rb.Subjects {
		if subject.APIGroup == rbacGroup && subject.Kind == p.kind {
			result = append(result, subject.Name)
		} else if subject.APIGroup == "" && p.kind == "User" && subject.Kind == "ServiceAccount" && subject.Namespace != "" {
			// Index is for Users and this references a service account
			result = append(result, fmt.Sprintf("serviceaccount:%s:%s", subject.Namespace, subject.Name))
		}
	}
	return
}

var null = []byte{'\x00'}

func (p *policyRuleIndex) addRolesToHash(digest hash.Hash, subjectName string) {
	for _, crb := range p.getClusterRoleBindings(subjectName) {
		digest.Write([]byte(crb.RoleRef.Name))
		digest.Write(null)
		digest.Write([]byte(p.revisions.roleRevision("", crb.RoleRef.Name)))
		digest.Write(null)
	}

	for _, rb := range p.getRoleBindings(subjectName) {
		switch rb.RoleRef.Kind {
		case "Role":
			digest.Write([]byte(rb.RoleRef.Name))
			digest.Write(null)
			digest.Write([]byte(rb.Namespace))
			digest.Write(null)
			digest.Write([]byte(p.revisions.roleRevision(rb.Namespace, rb.RoleRef.Name)))
			digest.Write(null)
		case "ClusterRole":
			digest.Write([]byte(rb.RoleRef.Name))
			digest.Write(null)
			digest.Write([]byte(rb.Namespace))
			digest.Write(null)
			digest.Write([]byte(p.revisions.roleRevision("", rb.RoleRef.Name)))
			digest.Write(null)
		}
	}
}

func (p *policyRuleIndex) get(subjectName string) *AccessSet {
	result := &AccessSet{}

	for _, binding := range p.getRoleBindings(subjectName) {
		p.addAccess(result, binding.Namespace, binding.RoleRef)
	}

	for _, binding := range p.getClusterRoleBindings(subjectName) {
		p.addAccess(result, All, binding.RoleRef)
	}

	return result
}

func (p *policyRuleIndex) addAccess(accessSet *AccessSet, namespace string, roleRef rbacv1.RoleRef) {
	for _, rule := range p.getRules(namespace, roleRef) {
		for _, group := range rule.APIGroups {
			for _, resource := range rule.Resources {
				names := rule.ResourceNames
				if len(names) == 0 {
					names = []string{All}
				}
				for _, resourceName := range names {
					for _, verb := range rule.Verbs {
						accessSet.Add(verb,
							schema.GroupResource{
								Group:    group,
								Resource: resource,
							}, Access{
								Namespace:    namespace,
								ResourceName: resourceName,
							})
					}
				}
			}
		}
	}
}

func (p *policyRuleIndex) getRules(namespace string, roleRef rbacv1.RoleRef) []rbacv1.PolicyRule {
	switch roleRef.Kind {
	case "ClusterRole":
		var role rbacv1.ClusterRole
		if err := p.client.Get(p.ctx, router.Key("", roleRef.Name), &role); err != nil {
			// ignore error
			return nil
		}
		return role.Rules
	case "Role":
		var role rbacv1.Role
		if err := p.client.Get(p.ctx, router.Key(namespace, roleRef.Name), &role); err != nil {
			return nil
		}
		return role.Rules
	}

	return nil
}

func (p *policyRuleIndex) getClusterRoleBindings(subjectName string) []rbacv1.ClusterRoleBinding {
	var list rbacv1.ClusterRoleBindingList
	err := p.client.List(p.ctx, &list, &kclient.ListOptions{
		FieldSelector: fields.SelectorFromSet(map[string]string{
			p.clusterRoleIndexKey: subjectName,
		}),
	})
	if err != nil {
		return nil
	}
	sort.Slice(list.Items, func(i, j int) bool {
		return list.Items[i].Name < list.Items[j].Name
	})
	return list.Items
}

func (p *policyRuleIndex) getRoleBindings(subjectName string) []rbacv1.RoleBinding {
	var list rbacv1.RoleBindingList
	err := p.client.List(p.ctx, &list, &kclient.ListOptions{
		FieldSelector: fields.SelectorFromSet(map[string]string{
			p.roleIndexKey: subjectName,
		}),
	})
	if err != nil {
		return nil
	}
	sort.Slice(list.Items, func(i, j int) bool {
		return string(list.Items[i].UID) < string(list.Items[j].UID)
	})
	return list.Items
}
