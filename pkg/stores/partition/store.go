package partition

import (
	"context"
	"net/http"
	"strconv"

	types2 "github.com/acorn-io/brent/pkg/types"
	"golang.org/x/sync/errgroup"
)

type Partitioner interface {
	Lookup(apiOp *types2.APIRequest, schema *types2.APISchema, verb, id string) (Partition, error)
	All(apiOp *types2.APIRequest, schema *types2.APISchema, verb, id string) ([]Partition, error)
	Store(apiOp *types2.APIRequest, partition Partition) (types2.Store, error)
}

type Store struct {
	Partitioner Partitioner
}

func (s *Store) getStore(apiOp *types2.APIRequest, schema *types2.APISchema, verb, id string) (types2.Store, error) {
	p, err := s.Partitioner.Lookup(apiOp, schema, verb, id)
	if err != nil {
		return nil, err
	}

	return s.Partitioner.Store(apiOp, p)
}

func (s *Store) Delete(apiOp *types2.APIRequest, schema *types2.APISchema, id string) (types2.APIObject, error) {
	target, err := s.getStore(apiOp, schema, "delete", id)
	if err != nil {
		return types2.APIObject{}, err
	}

	return target.Delete(apiOp, schema, id)
}

func (s *Store) ByID(apiOp *types2.APIRequest, schema *types2.APISchema, id string) (types2.APIObject, error) {
	target, err := s.getStore(apiOp, schema, "get", id)
	if err != nil {
		return types2.APIObject{}, err
	}

	return target.ByID(apiOp, schema, id)
}

func (s *Store) listPartition(ctx context.Context, apiOp *types2.APIRequest, schema *types2.APISchema, partition Partition,
	cont string, revision string, limit int) (types2.APIObjectList, error) {
	store, err := s.Partitioner.Store(apiOp, partition)
	if err != nil {
		return types2.APIObjectList{}, err
	}

	req := apiOp.Clone()
	req.Request = req.Request.Clone(ctx)

	values := req.Request.URL.Query()
	values.Set("continue", cont)
	values.Set("revision", revision)
	if limit > 0 {
		values.Set("limit", strconv.Itoa(limit))
	} else {
		values.Del("limit")
	}
	req.Request.URL.RawQuery = values.Encode()

	return store.List(req, schema)
}

func (s *Store) List(apiOp *types2.APIRequest, schema *types2.APISchema) (types2.APIObjectList, error) {
	var (
		result types2.APIObjectList
	)

	paritions, err := s.Partitioner.All(apiOp, schema, "list", "")
	if err != nil {
		return result, err
	}

	lister := ParallelPartitionLister{
		Lister: func(ctx context.Context, partition Partition, cont string, revision string, limit int) (types2.APIObjectList, error) {
			return s.listPartition(ctx, apiOp, schema, partition, cont, revision, limit)
		},
		Concurrency: 3,
		Partitions:  paritions,
	}

	resume := apiOp.Request.URL.Query().Get("continue")
	limit := getLimit(apiOp.Request)

	list, err := lister.List(apiOp.Context(), limit, resume)
	if err != nil {
		return result, err
	}

	for items := range list {
		result.Objects = append(result.Objects, items...)
	}

	result.Revision = lister.Revision()
	result.Continue = lister.Continue()
	return result, lister.Err()
}

func (s *Store) Create(apiOp *types2.APIRequest, schema *types2.APISchema, data types2.APIObject) (types2.APIObject, error) {
	target, err := s.getStore(apiOp, schema, "create", "")
	if err != nil {
		return types2.APIObject{}, err
	}

	return target.Create(apiOp, schema, data)
}

func (s *Store) Update(apiOp *types2.APIRequest, schema *types2.APISchema, data types2.APIObject, id string) (types2.APIObject, error) {
	target, err := s.getStore(apiOp, schema, "update", id)
	if err != nil {
		return types2.APIObject{}, err
	}

	return target.Update(apiOp, schema, data, id)
}

func (s *Store) Watch(apiOp *types2.APIRequest, schema *types2.APISchema, wr types2.WatchRequest) (chan types2.APIEvent, error) {
	partitions, err := s.Partitioner.All(apiOp, schema, "watch", wr.ID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(apiOp.Context())
	apiOp = apiOp.Clone().WithContext(ctx)

	eg := errgroup.Group{}
	response := make(chan types2.APIEvent)

	for _, partition := range partitions {
		store, err := s.Partitioner.Store(apiOp, partition)
		if err != nil {
			cancel()
			return nil, err
		}

		eg.Go(func() error {
			defer cancel()
			c, err := store.Watch(apiOp, schema, wr)
			if err != nil {
				return err
			}
			for i := range c {
				response <- i
			}
			return nil
		})
	}

	go func() {
		defer close(response)
		<-ctx.Done()
		eg.Wait()
		cancel()
	}()

	return response, nil
}

func getLimit(req *http.Request) int {
	limitString := req.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitString)
	if err != nil {
		limit = 0
	}
	if limit <= 0 {
		limit = 100000
	}
	return limit
}
