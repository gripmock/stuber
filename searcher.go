package stuber

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/exp/maps"
)

var (
	ErrServiceNotFound = errors.New("service not found")
	ErrMethodNotFound  = errors.New("method not found")
	ErrStubNotFound    = errors.New("stub not found")
)

type searcher struct {
	mu       sync.RWMutex
	stubUsed map[uuid.UUID]struct{}

	storage *storage
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

func (s *searcher) clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stubUsed = make(map[uuid.UUID]struct{})
	s.storage.clear()
}

func (s *searcher) all() []*Stub {
	return s.castToStub(s.storage.values())
}

func (s *searcher) used() []*Stub {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.castToStub(s.storage.findByIDs(maps.Keys(s.stubUsed)...))
}

func (s *searcher) unused() []*Stub {
	s.mu.RLock()
	defer s.mu.RUnlock()

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
		return s.searchByID(query.Service, query.Method, query)
	}

	return s.search(query)
}

func (s *searcher) searchByID(service, method string, query Query) (*Result, error) {
	_, err := s.storage.posByN(service, method)
	if err != nil {
		return nil, s.wrap(err)
	}

	if found := s.findByID(*query.ID); found != nil {
		s.mark(query, *query.ID)

		return &Result{found: found}, nil
	}

	return nil, ErrServiceNotFound
}

func (s *searcher) search(query Query) (*Result, error) {
	stubs, err := s.findBy(query.Service, query.Method)
	if err != nil {
		return nil, s.wrap(err)
	}

	var (
		found       *Stub   = nil
		foundRank   float64 = 0
		similar     *Stub   = nil
		similarRank float64 = 0
	)

	for _, stub := range stubs {
		current := rankMatch(query, stub)
		if current > similarRank {
			similar = stub
			similarRank = current
		}

		if match(query, stub) && current > foundRank {
			found = stub
			foundRank = current
		}
	}

	if found != nil {
		s.mark(query, found.ID)

		return &Result{found: found}, nil
	}

	if similar == nil {
		return nil, ErrStubNotFound
	}

	return &Result{found: nil, similar: similar}, nil
}

func (s *searcher) mark(query Query, id uuid.UUID) {
	if query.RequestInternal() {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.stubUsed[id] = struct{}{}
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
