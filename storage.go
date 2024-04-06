package stuber

import (
	"errors"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

var (
	ErrLeftNotFound  = errors.New("left not found")
	ErrRightNotFound = errors.New("right not found")
)

type Value interface {
	Key() uuid.UUID
	Left() string
	Right() string
}

type storage struct {
	mu         sync.RWMutex
	leftTotal  atomic.Uint64
	rightTotal atomic.Uint64
	lefts      map[string]uint64
	rights     map[string]uint64
	leftRights map[uint64][]uint64
	values     map[uuid.UUID][]Value
	valuesByID map[uuid.UUID]Value
}

func newStorage() *storage {
	return &storage{
		rights:     map[string]uint64{},
		lefts:      map[string]uint64{},
		leftRights: map[uint64][]uint64{},
		values:     map[uuid.UUID][]Value{},
		valuesByID: map[uuid.UUID]Value{},
	}
}

func (s *storage) findAll(left, right string) ([]Value, error) {
	pos, err := s.posByN(left, right)
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.values[pos], nil
}

func (s *storage) findByID(key uuid.UUID) Value {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if v, ok := s.valuesByID[key]; ok {
		return v
	}

	return nil
}

func (s *storage) upsert(values ...Value) {
	for _, v := range values {
		leftID := s.leftIdOrNew(v.Left())
		rightID := s.rightIdOrNew(v.Right())
		ind := s.pos(leftID, rightID)

		s.mu.Lock()
		s.leftRights[leftID] = append(s.leftRights[leftID], rightID)
		s.values[ind] = append(s.values[ind], v)
		s.valuesByID[v.Key()] = v
		s.mu.Unlock()
	}
}

func (s *storage) del(keys ...uuid.UUID) int {
	result := 0
	deleteIDs := make(map[uuid.UUID][]uuid.UUID, len(s.values))

	for _, key := range keys {
		v := s.findByID(key)
		if v == nil {
			continue
		}

		pos, err := s.posByN(v.Left(), v.Right())
		if err != nil {
			continue
		}

		deleteIDs[pos] = append(deleteIDs[pos], key)
		result++
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for pos, v := range deleteIDs {
		s.values[pos] = slices.DeleteFunc(s.values[pos], func(value Value) bool {
			return slices.Contains(v, value.Key())
		})
	}

	for _, key := range keys {
		delete(s.valuesByID, key)
	}

	return result
}

func (s *storage) leftId(name string) (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if id, ok := s.lefts[name]; ok {
		return id, nil
	}

	return 0, ErrLeftNotFound
}

func (s *storage) leftIdOrNew(name string) uint64 {
	if id, err := s.leftId(name); err == nil {
		return id
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.lefts[name] = s.leftTotal.Add(1)

	return s.lefts[name]
}

func (s *storage) rightId(name string) (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if id, ok := s.rights[name]; ok {
		return id, nil
	}

	return 0, ErrRightNotFound
}

func (s *storage) rightIdOrNew(name string) uint64 {
	if id, err := s.rightId(name); err == nil {
		return id
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.rights[name] = s.rightTotal.Add(1)

	return s.rights[name]
}

func (s *storage) posByN(left, right string) (uuid.UUID, error) {
	leftID, err := s.leftId(left)
	if err != nil {
		return uuid.UUID{}, err
	}

	rightID, err := s.rightId(right)
	if err != nil {
		return uuid.UUID{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	if !slices.Contains(s.leftRights[leftID], rightID) {
		return uuid.UUID{}, ErrRightNotFound
	}

	return s.pos(leftID, rightID), nil
}

func (s *storage) pos(left uint64, right uint64) uuid.UUID {
	return uuid.UUID{
		byte(left >> 56),
		byte(left >> 48),
		byte(left >> 40),
		byte(left >> 32),
		byte(left >> 24),
		byte(left >> 16),
		byte(left >> 8),
		byte(left),
		byte(right),
		byte(right >> 8),
		byte(right >> 16),
		byte(right >> 24),
		byte(right >> 32),
		byte(right >> 40),
		byte(right >> 48),
		byte(right >> 56),
	}
}
