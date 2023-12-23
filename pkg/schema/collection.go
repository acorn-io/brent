package schema

import (
	"context"
	"strings"
	"sync"

	"github.com/acorn-io/brent/pkg/accesscontrol"
	"github.com/acorn-io/brent/pkg/attributes"
	types2 "github.com/acorn-io/brent/pkg/types"
	"github.com/acorn-io/schemer/name"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/cache"
	"k8s.io/apiserver/pkg/authentication/user"
)

type Factory interface {
	Schemas(user user.Info) (*types2.APISchemas, error)
	ByGVR(gvr schema.GroupVersionResource) string
	ByGVK(gvr schema.GroupVersionKind) string
	OnChange(ctx context.Context, cb func())
	AddTemplate(template ...Template)
}

type Collection struct {
	toSync     int32
	baseSchema *types2.APISchemas
	schemas    map[string]*types2.APISchema
	templates  map[string][]*Template
	notifiers  map[int]func()
	notifierID int
	byGVR      map[schema.GroupVersionResource]string
	byGVK      map[schema.GroupVersionKind]string
	cache      *cache.LRUExpireCache
	lock       sync.RWMutex

	ctx     context.Context
	running map[string]func()
	as      accesscontrol.AccessSetLookup
}

type Template struct {
	Group        string
	Kind         string
	ID           string
	Customize    func(*types2.APISchema)
	Formatter    types2.Formatter
	Store        types2.Store
	Start        func(ctx context.Context) error
	StoreFactory func(types2.Store) types2.Store
}

func NewCollection(ctx context.Context, baseSchema *types2.APISchemas, access accesscontrol.AccessSetLookup) *Collection {
	return &Collection{
		baseSchema: baseSchema,
		schemas:    map[string]*types2.APISchema{},
		templates:  map[string][]*Template{},
		byGVR:      map[schema.GroupVersionResource]string{},
		byGVK:      map[schema.GroupVersionKind]string{},
		cache:      cache.NewLRUExpireCache(1000),
		notifiers:  map[int]func(){},
		ctx:        ctx,
		as:         access,
		running:    map[string]func(){},
	}
}

func (c *Collection) OnChange(ctx context.Context, cb func()) {
	c.lock.Lock()
	id := c.notifierID
	c.notifierID++
	c.notifiers[id] = cb
	c.lock.Unlock()

	go func() {
		<-ctx.Done()
		c.lock.Lock()
		delete(c.notifiers, id)
		c.lock.Unlock()
	}()
}

func (c *Collection) Reset(schemas map[string]*types2.APISchema) {
	byGVK := map[schema.GroupVersionKind]string{}
	byGVR := map[schema.GroupVersionResource]string{}

	for _, s := range schemas {
		gvr := attributes.GVR(s)
		if gvr.Resource != "" {
			byGVR[gvr] = s.ID
		}
		gvk := attributes.GVK(s)
		if gvk.Kind != "" {
			byGVK[gvk] = s.ID
		}

		c.applyTemplates(s)
	}

	c.lock.Lock()
	c.startStopTemplate(schemas)
	c.schemas = schemas
	c.byGVR = byGVR
	c.byGVK = byGVK
	for _, k := range c.cache.Keys() {
		c.cache.Remove(k)
	}
	c.lock.Unlock()
	c.lock.RLock()
	for _, f := range c.notifiers {
		f()
	}
	c.lock.RUnlock()
}

func start(ctx context.Context, templates []*Template) error {
	for _, template := range templates {
		if template.Start == nil {
			continue
		}
		if err := template.Start(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (c *Collection) startStopTemplate(schemas map[string]*types2.APISchema) {
	for id := range schemas {
		if _, ok := c.running[id]; ok {
			continue
		}
		templates := c.templates[id]
		if len(templates) == 0 {
			continue
		}

		subCtx, cancel := context.WithCancel(c.ctx)
		if err := start(subCtx, templates); err != nil {
			cancel()
			logrus.Errorf("failed to start schema template: %s", id)
			continue
		}
		c.running[id] = cancel
	}

	for id, cancel := range c.running {
		if _, ok := schemas[id]; !ok {
			cancel()
			delete(c.running, id)
		}
	}
}

func (c *Collection) Schema(id string) *types2.APISchema {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.schemas[id]
}

func (c *Collection) IDs() (result []string) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	seen := map[string]bool{}
	for _, id := range c.byGVR {
		if seen[id] {
			continue
		}
		seen[id] = true
		result = append(result, id)
	}
	return
}

func (c *Collection) ByGVR(gvr schema.GroupVersionResource) string {
	c.lock.RLock()
	defer c.lock.RUnlock()

	id, ok := c.byGVR[gvr]
	if ok {
		return id
	}
	gvr.Resource = name.GuessPluralName(strings.ToLower(gvr.Resource))
	return c.byGVK[schema.GroupVersionKind{
		Group:   gvr.Group,
		Version: gvr.Version,
		Kind:    gvr.Resource,
	}]
}

func (c *Collection) ByGVK(gvk schema.GroupVersionKind) string {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.byGVK[gvk]
}

func (c *Collection) AddTemplate(templates ...Template) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	for i, template := range templates {
		if template.Kind != "" {
			c.templates[template.Group+"/"+template.Kind] = append(c.templates[template.Group+"/"+template.Kind], &templates[i])
		} else if template.ID != "" {
			c.templates[template.ID] = append(c.templates[template.ID], &templates[i])
		}
		if template.Kind == "" && template.Group == "" && template.ID == "" {
			c.templates[""] = append(c.templates[""], &templates[i])
		}
	}
}
