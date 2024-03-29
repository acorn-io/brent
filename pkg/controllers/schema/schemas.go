package schema

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/acorn-io/baaah/pkg/router"
	"github.com/acorn-io/brent/pkg/attributes"
	"github.com/acorn-io/brent/pkg/resources/common"
	schema2 "github.com/acorn-io/brent/pkg/schema"
	"github.com/acorn-io/brent/pkg/schema/converter"
	"github.com/acorn-io/brent/pkg/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	authorizationv1 "k8s.io/api/authorization/v1"
	"k8s.io/client-go/discovery"
	apiv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	listPool        = semaphore.NewWeighted(10)
	typeNameChanges = map[string]string{
		"extensions.v1beta1.ingress": "networking.k8s.io.v1beta1.ingress",
	}
)

type SchemasHandler interface {
	OnSchemas(schemas *schema2.Collection) error
}

type handler struct {
	sync.Mutex

	ctx                context.Context
	toSync             int32
	schemas            *schema2.Collection
	discoveryInterface discovery.DiscoveryInterface
	client             kclient.Client
	cols               *common.DynamicColumns
}

func Register(ctx context.Context,
	cols *common.DynamicColumns,
	discovery discovery.DiscoveryInterface,
	router *router.Router,
	schemas *schema2.Collection) {

	h := &handler{
		ctx:                ctx,
		cols:               cols,
		discoveryInterface: discovery,
		schemas:            schemas,
		client:             router.Backend(),
	}

	router.HandleFunc(&apiv1.APIService{}, h.OnChangeAPIService)
}

func (h *handler) OnChangeAPIService(req router.Request, resp router.Response) error {
	h.queueRefresh()
	return nil
}

func (h *handler) queueRefresh() {
	atomic.StoreInt32(&h.toSync, 1)

	go func() {
		time.Sleep(500 * time.Millisecond)
		if err := h.refreshAll(h.ctx); err != nil {
			logrus.Errorf("failed to sync schemas: %v", err)
			atomic.StoreInt32(&h.toSync, 1)
		}
	}()
}

func isListOrGetable(schema *types.APISchema) bool {
	for _, verb := range attributes.Verbs(schema) {
		switch verb {
		case "list":
			return true
		case "get":
			return true
		}
	}

	return false
}

func isListWatchable(schema *types.APISchema) bool {
	var (
		canList  bool
		canWatch bool
	)

	for _, verb := range attributes.Verbs(schema) {
		switch verb {
		case "list":
			canList = true
		case "watch":
			canWatch = true
		}
	}

	return canList && canWatch
}

func (h *handler) getColumns(ctx context.Context, schemas map[string]*types.APISchema) error {
	eg := errgroup.Group{}

	for _, schema := range schemas {
		if !isListOrGetable(schema) {
			continue
		}

		if err := listPool.Acquire(ctx, 1); err != nil {
			return err
		}

		s := schema
		eg.Go(func() error {
			defer listPool.Release(1)
			return h.cols.SetColumns(ctx, s)
		})
	}

	return eg.Wait()
}

func (h *handler) refreshAll(ctx context.Context) error {
	h.Lock()
	defer h.Unlock()

	if !h.needToSync() {
		return nil
	}

	schemas, err := converter.ToSchemas(h.discoveryInterface)
	if err != nil {
		return err
	}

	filteredSchemas := map[string]*types.APISchema{}
	for _, schema := range schemas {
		if isListWatchable(schema) {
			if preferredTypeExists(schema, schemas) {
				continue
			}
			if ok, err := h.allowed(ctx, schema); err != nil {
				return err
			} else if !ok {
				continue
			}
		}

		gvk := attributes.GVK(schema)
		if gvk.Kind != "" {
			gvr := attributes.GVR(schema)
			schema.ID = converter.GVKToSchemaID(gvk)
			schema.PluralName = converter.GVRToPluralName(gvr)
		}
		filteredSchemas[schema.ID] = schema
	}

	if err := h.getColumns(h.ctx, filteredSchemas); err != nil {
		return err
	}

	h.schemas.Reset(filteredSchemas)

	return nil
}

func preferredTypeExists(schema *types.APISchema, schemas map[string]*types.APISchema) bool {
	if replacement, ok := typeNameChanges[schema.ID]; ok && schemas[replacement] != nil {
		return true
	}
	pg := attributes.PreferredGroup(schema)
	pv := attributes.PreferredVersion(schema)
	if pg == "" && pv == "" {
		return false
	}

	gvk := attributes.GVK(schema)
	if pg != "" {
		gvk.Group = pg
	}
	if pv != "" {
		gvk.Version = pv
	}

	_, ok := schemas[converter.GVKToVersionedSchemaID(gvk)]
	return ok
}

func (h *handler) allowed(ctx context.Context, schema *types.APISchema) (bool, error) {
	gvr := attributes.GVR(schema)
	ssar := &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Verb:     "list",
				Group:    gvr.Group,
				Version:  gvr.Version,
				Resource: gvr.Resource,
			},
		},
	}
	err := h.client.Create(ctx, ssar)
	if err != nil {
		return false, err
	}
	return ssar.Status.Allowed && !ssar.Status.Denied, nil
}

func (h *handler) needToSync() bool {
	old := atomic.SwapInt32(&h.toSync, 0)
	return old == 1
}
