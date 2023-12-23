package accesscontrol

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"time"

	"github.com/acorn-io/baaah/pkg/router"
	"k8s.io/apimachinery/pkg/util/cache"
	"k8s.io/apiserver/pkg/authentication/user"
)

type AccessSetLookup interface {
	AccessFor(user user.Info) *AccessSet
}

type AccessStore struct {
	users  *policyRuleIndex
	groups *policyRuleIndex
	cache  *cache.LRUExpireCache
}

type roleKey struct {
	namespace string
	name      string
}

func NewAccessStore(ctx context.Context, cacheResults bool, router *router.Router) (*AccessStore, error) {
	revisions := newRoleRevision(router)
	users, err := newPolicyRuleIndex(ctx, true, revisions, router)
	if err != nil {
		return nil, err
	}
	groups, err := newPolicyRuleIndex(ctx, false, revisions, router)
	if err != nil {
		return nil, err
	}
	as := &AccessStore{
		users:  users,
		groups: groups,
	}
	if cacheResults {
		as.cache = cache.NewLRUExpireCache(50)
	}
	return as, nil
}

func (l *AccessStore) AccessFor(user user.Info) *AccessSet {
	var cacheKey string
	if l.cache != nil {
		cacheKey = l.CacheKey(user)
		val, ok := l.cache.Get(cacheKey)
		if ok {
			as, _ := val.(*AccessSet)
			return as
		}
	}

	result := l.users.get(user.GetName())
	for _, group := range user.GetGroups() {
		result.Merge(l.groups.get(group))
	}

	if l.cache != nil {
		result.ID = cacheKey
		l.cache.Add(cacheKey, result, 24*time.Hour)
	}

	return result
}

func (l *AccessStore) CacheKey(user user.Info) string {
	d := sha256.New()

	l.users.addRolesToHash(d, user.GetName())

	groupBase := user.GetGroups()
	groups := make([]string, 0, len(groupBase))
	copy(groups, groupBase)

	sort.Strings(groups)
	for _, group := range user.GetGroups() {
		l.groups.addRolesToHash(d, group)
	}

	return hex.EncodeToString(d.Sum(nil))
}
