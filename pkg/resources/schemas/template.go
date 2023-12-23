package schemas

import (
	"context"
	"sync"
	"time"

	"github.com/acorn-io/brent/pkg/builtin"
	schemastore "github.com/acorn-io/brent/pkg/stores/schema"
	types2 "github.com/acorn-io/brent/pkg/types"
	"github.com/acorn-io/broadcaster"
	"k8s.io/apimachinery/pkg/api/equality"

	"github.com/acorn-io/brent/pkg/accesscontrol"
	"github.com/acorn-io/brent/pkg/schema"
	"github.com/acorn-io/schemer/validation"
	"github.com/sirupsen/logrus"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/endpoints/request"
)

func SetupWatcher(ctx context.Context, schemas *types2.APISchemas, asl accesscontrol.AccessSetLookup, factory schema.Factory) {
	// one instance shared with all stores
	notifier := schemaChangeNotifier(ctx, factory)

	schema := builtin.Schema
	schema.Store = &Store{
		Store:              schema.Store,
		asl:                asl,
		sf:                 factory,
		schemaChangeNotify: notifier,
	}

	schemas.AddSchema(schema)
}

type Store struct {
	types2.Store

	asl                accesscontrol.AccessSetLookup
	sf                 schema.Factory
	schemaChangeNotify func(context.Context) (chan interface{}, error)
}

func (s *Store) Watch(apiOp *types2.APIRequest, schema *types2.APISchema, w types2.WatchRequest) (chan types2.APIEvent, error) {
	user, ok := request.UserFrom(apiOp.Request.Context())
	if !ok {
		return nil, validation.Unauthorized
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	result := make(chan types2.APIEvent)

	go func() {
		wg.Wait()
		close(result)
	}()

	go func() {
		defer wg.Done()
		c, err := s.schemaChangeNotify(apiOp.Context())
		if err != nil {
			return
		}
		schemas, err := s.sf.Schemas(user)
		if err != nil {
			logrus.Errorf("failed to generate schemas for user %v: %v", user, err)
			return
		}
		for range c {
			schemas = s.sendSchemas(result, apiOp, user, schemas)
		}
	}()

	go func() {
		defer wg.Done()
		schemas, err := s.sf.Schemas(user)
		if err != nil {
			logrus.Errorf("failed to generate schemas for notify user %v: %v", user, err)
			return
		}
		for range s.userChangeNotify(apiOp.Context(), user) {
			schemas = s.sendSchemas(result, apiOp, user, schemas)
		}
	}()

	return result, nil
}

func (s *Store) sendSchemas(result chan types2.APIEvent, apiOp *types2.APIRequest, user user.Info, oldSchemas *types2.APISchemas) *types2.APISchemas {
	schemas, err := s.sf.Schemas(user)
	if err != nil {
		logrus.Errorf("failed to get schemas for %v: %v", user, err)
		return oldSchemas
	}

	inNewSchemas := map[string]bool{}
	for _, apiObject := range schemastore.FilterSchemas(apiOp, schemas.Schemas).Objects {
		inNewSchemas[apiObject.ID] = true
		eventName := types2.ChangeAPIEvent
		if oldSchema := oldSchemas.LookupSchema(apiObject.ID); oldSchema == nil {
			eventName = types2.CreateAPIEvent
		} else {
			newSchemaCopy := apiObject.Object.(*types2.APISchema).Schema.DeepCopy()
			oldSchemaCopy := oldSchema.Schema.DeepCopy()
			newSchemaCopy.Mapper = nil
			oldSchemaCopy.Mapper = nil
			if equality.Semantic.DeepEqual(newSchemaCopy, oldSchemaCopy) {
				continue
			}
		}
		result <- types2.APIEvent{
			Name:         eventName,
			ResourceType: "schema",
			Object:       apiObject,
		}
	}

	for _, oldSchema := range schemastore.FilterSchemas(apiOp, oldSchemas.Schemas).Objects {
		if inNewSchemas[oldSchema.ID] {
			continue
		}
		result <- types2.APIEvent{
			Name:         types2.RemoveAPIEvent,
			ResourceType: "schema",
			Object:       oldSchema,
		}
	}

	return schemas
}

func (s *Store) userChangeNotify(ctx context.Context, user user.Info) chan interface{} {
	as := s.asl.AccessFor(user)
	result := make(chan interface{})
	go func() {
		defer close(result)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}

			newAS := s.asl.AccessFor(user)
			if newAS.ID != as.ID {
				result <- struct{}{}
				as = newAS
			}
		}
	}()

	return result
}

func schemaChangeNotifier(ctx context.Context, factory schema.Factory) func(ctx context.Context) (chan interface{}, error) {
	bcast := broadcaster.New[any]()
	factory.OnChange(ctx, func() {
		select {
		case bcast.C <- struct{}{}:
		default:
		}
	})
	return func(ctx context.Context) (chan interface{}, error) {
		sub := bcast.Subscribe()
		context.AfterFunc(ctx, sub.Close)
		return sub.C, nil
	}
}
