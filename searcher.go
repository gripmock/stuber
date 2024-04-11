package stuber

import (
	"errors"
	"golang.org/x/exp/maps"
	"slices"

	"github.com/google/uuid"
)

var (
	ErrServiceNotFound = errors.New("service not found")
	ErrMethodNotFound  = errors.New("method not found")
	ErrStubNotFound    = errors.New("stub not found")
)

type searcher struct {
	stubUsed map[uuid.UUID]struct{}
	storage  *storage
}

func newSearcher() *searcher {
	return &searcher{
		storage:  newStorage(),
		stubUsed: make(map[uuid.UUID]struct{}),
	}
}

type Result struct {
	found   *Stub
	similar *Stub
}

func (r *Result) Found() *Stub {
	return r.found
}

func (r *Result) Similar() *Stub {
	return r.similar
}

func (s *searcher) upsert(values ...*Stub) []uuid.UUID {
	return s.storage.upsert(s.castToValue(values)...)
}

func (s *searcher) del(ids ...uuid.UUID) int {
	return s.storage.del(ids...)
}

func (s *searcher) findByID(id uuid.UUID) *Stub {
	if v, ok := s.storage.findByID(id).(*Stub); ok {
		return v
	}

	return nil
}

func (s *searcher) findBy(service, method string) ([]*Stub, error) {
	all, err := s.storage.findAll(service, method)
	if err != nil {
		return nil, s.wrap(err)
	}

	return s.castToStub(all), nil
}

func (s *searcher) all() []*Stub {
	return s.castToStub(s.storage.values())
}

func (s *searcher) used() []*Stub {
	return s.castToStub(s.storage.findByIDs(maps.Keys(s.stubUsed)...))
}

func (s *searcher) unused() []*Stub {
	results := make([]*Stub, 0, len(s.all()))
	for _, stub := range s.all() {
		if _, ok := s.stubUsed[stub.ID]; !ok {
			results = append(results, stub)
		}
	}

	return results
}

func (s *searcher) find(query Query) (*Result, error) {
	if query.ID != nil {
		return s.searchByID(query.Service, query.Method, *query.ID)
	}

	return s.search(query)
}

func (s *searcher) searchByID(service, method string, id uuid.UUID) (*Result, error) {
	_, err := s.storage.posByN(service, method)
	if err != nil {
		return nil, s.wrap(err)
	}

	if found := s.findByID(id); found != nil {
		s.stubUsed[id] = struct{}{}

		return &Result{found: found}, nil
	}

	return nil, ErrServiceNotFound
}

func (s *searcher) search(query Query) (*Result, error) {
	stubs, err := s.findBy(query.Service, query.Method)
	if err != nil {
		return nil, s.wrap(err)
	}

	slices.SortFunc(stubs, func(a, b *Stub) int {
		return a.Headers.Len() - a.Headers.Len()
	})

	var (
		similar *Stub   = nil
		rank    float64 = 0
	)

	for _, stub := range stubs {
		if match(query, stub) {
			s.stubUsed[stub.ID] = struct{}{}

			return &Result{found: stub}, nil
		}

		if current := rankMatch(query, stub); current > rank {
			similar = stub
			rank = current
		}
	}

	if similar == nil {
		return nil, ErrStubNotFound
	}

	return &Result{found: nil, similar: similar}, nil
}

func (s *searcher) castToValue(values []*Stub) []Value {
	result := make([]Value, len(values))
	for i, v := range values {
		result[i] = v
	}

	return result
}

func (s *searcher) castToStub(values []Value) []*Stub {
	ret := make([]*Stub, 0, len(values))
	for _, v := range values {
		if s, ok := v.(*Stub); ok {
			ret = append(ret, s)
		}
	}

	return ret
}

func (s *searcher) wrap(err error) error {
	if errors.Is(err, ErrLeftNotFound) {
		return ErrServiceNotFound
	}

	if errors.Is(err, ErrRightNotFound) {
		return ErrMethodNotFound
	}

	return err
}
