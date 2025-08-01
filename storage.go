package stuber

import (
	"errors"
	"iter"
	"slices"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/zeebo/xxh3"
)

// ErrLeftNotFound is returned when the left value is not found.
var ErrLeftNotFound = errors.New("left not found")

// ErrRightNotFound is returned when the right value is not found.
var ErrRightNotFound = errors.New("right not found")

// Value is a type used to store the result of a search.
type Value interface {
	Key() uuid.UUID
	Left() string
	Right() string
	Score() int // Score determines the order of values when sorting
}

// storage is responsible for managing search results with enhanced
// performance and memory efficiency. It supports concurrent access
// through the use of a read-write mutex.
//
// Fields:
// - mu: Ensures safe concurrent access to the storage.
// - lefts: A map that tracks unique left values by their hashed IDs.
// - items: Stores items by a composite key of hashed left and right IDs.
// - itemsByID: Provides quick access to items by their unique UUIDs.
type storage struct {
	mu        sync.RWMutex
	lefts     map[uint32]struct{}
	items     map[uint64]map[uuid.UUID]Value
	itemsByID map[uuid.UUID]Value
}

// newStorage creates a new instance of the storage struct.
func newStorage() *storage {
	return &storage{
		lefts:     make(map[uint32]struct{}),
		items:     make(map[uint64]map[uuid.UUID]Value),
		itemsByID: make(map[uuid.UUID]Value),
	}
}

// clear resets the storage.
func (s *storage) clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lefts = make(map[uint32]struct{})
	s.items = make(map[uint64]map[uuid.UUID]Value)
	s.itemsByID = make(map[uuid.UUID]Value)
}

// values returns an iterator sequence of all Value items stored in the
// storage.
func (s *storage) values() iter.Seq[Value] {
	return func(yield func(Value) bool) {
		s.mu.RLock()
		defer s.mu.RUnlock()

		for _, v := range s.itemsByID {
			if !yield(v) {
				return
			}
		}
	}
}

// findAll retrieves all Value items that match the given left and right names,
// sorted by score in descending order.
func (s *storage) findAll(left, right string) (iter.Seq[Value], error) {
	indexes, err := s.posByPN(left, right)
	if err != nil {
		return nil, err
	}

	return func(yield func(Value) bool) {
		s.yieldSortedValues(indexes, yield)
	}, nil
}

// yieldSortedValues yields values sorted by score in descending order,
// minimizing memory allocations and maximizing iterator usage.
func (s *storage) yieldSortedValues(indexes []uint64, yield func(Value) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Collect all values and sort them by score in descending order.
	// This approach is memory efficient for large datasets.
	type sortItem struct {
		value Value
		score int
	}

	// Collect values with scores for sorting
	var items []sortItem

	// First pass: collect all values with scores
	for _, index := range indexes {
		if m, exists := s.items[index]; exists {
			for _, v := range m {
				items = append(items, sortItem{value: v, score: v.Score()})
			}
		}
	}

	// Sort by score in descending order
	if len(items) > 1 {
		slices.SortFunc(items, func(a, b sortItem) int {
			return b.score - a.score
		})
	}

	// Yield sorted values
	for _, item := range items {
		if !yield(item.value) {
			return
		}
	}
}

// posByPN attempts to resolve IDs for a given left and right name pair.
// It first tries to resolve the full left name with the right name, and then
// attempts to resolve using a truncated version of the left name if necessary.
//
// Parameters:
// - left: The left name for matching.
// - right: The right name for matching.
//
// Returns:
// - [][2]uint64: A slice of resolved ID pairs.
// - error: An error if no IDs were resolved.
func (s *storage) posByPN(left, right string) ([]uint64, error) {
	// Initialize a slice to store the resolved IDs.
	var resolvedIDs []uint64

	// Attempt to resolve the full left name with the right name.
	id, err := s.posByN(left, right)
	if err == nil {
		// Append the resolved ID to the slice.
		resolvedIDs = append(resolvedIDs, id)
	}

	// Check for a potential truncation point in the left name.
	if dotIndex := strings.LastIndex(left, "."); dotIndex != -1 {
		truncatedLeft := left[dotIndex+1:]

		// Attempt to resolve the truncated left name with the right name.
		id, err := s.posByN(truncatedLeft, right)
		if err == nil {
			// Append the resolved ID to the slice.
			resolvedIDs = append(resolvedIDs, id)
		} else if errors.Is(err, ErrRightNotFound) && len(resolvedIDs) == 0 {
			// Return an error if the right name was not found
			// and no IDs were resolved.
			return nil, err
		}
	}

	// Return an error if no IDs were resolved.
	if len(resolvedIDs) == 0 {
		// Return the original error if we have it.
		return nil, err
	}

	// Return the resolved IDs.
	return resolvedIDs, nil
}

// findByID retrieves the Stub value associated with the given UUID from the
// storage.
//
// Parameters:
// - key: The UUID of the Stub value to retrieve.
//
// Returns:
// - Value: The Stub value associated with the given UUID, or nil if not found.
func (s *storage) findByID(key uuid.UUID) Value { //nolint:ireturn
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.itemsByID[key]
}

// findByIDs retrieves the Stub values associated with the given UUIDs from the
// storage.
//
// Returns:
//   - iter.Seq[Value]: The Stub values associated with the given UUIDs, or nil if
//     not found.
func (s *storage) findByIDs(ids iter.Seq[uuid.UUID]) iter.Seq[Value] {
	return func(yield func(Value) bool) {
		s.mu.RLock()
		defer s.mu.RUnlock()

		for id := range ids {
			if v, ok := s.itemsByID[id]; ok {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// upsert inserts or updates the given Value items in storage.
// Optimized for minimal allocations and maximum performance.
func (s *storage) upsert(values ...Value) []uuid.UUID {
	if len(values) == 0 {
		return nil
	}

	// Pre-allocate with exact size to minimize allocations
	results := make([]uuid.UUID, len(values))

	s.mu.Lock()
	defer s.mu.Unlock()

	// Process all values in a single pass
	for i, v := range values {
		results[i] = v.Key()

		// Calculate IDs directly without string interning
		leftID := s.id(v.Left())
		rightID := s.id(v.Right())
		index := s.pos(leftID, rightID)

		// Initialize the map at the index if it doesn't exist.
		if s.items[index] == nil {
			s.items[index] = make(map[uuid.UUID]Value, 1)
		}

		// Insert or update the value in the storage.
		s.items[index][v.Key()] = v
		s.itemsByID[v.Key()] = v
		s.lefts[leftID] = struct{}{}
	}

	return results
}

// del deletes the Stub values with the given UUIDs from the storage.
// It returns the number of Stub values that were successfully deleted.
func (s *storage) del(keys ...uuid.UUID) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	deleted := 0

	for _, key := range keys {
		if v, ok := s.itemsByID[key]; ok {
			pos := s.pos(s.id(v.Left()), s.id(v.Right()))

			if m, exists := s.items[pos]; exists {
				delete(m, key)
				delete(s.itemsByID, key)

				deleted++

				if len(m) == 0 {
					delete(s.items, pos)
				}
			}
		}
	}

	return deleted
}

func (s *storage) id(value string) uint32 {
	return uint32(xxh3.HashString(value)) //nolint:gosec
}

func (s *storage) pos(a, b uint32) uint64 {
	return uint64(a)<<32 | uint64(b)
}

func (s *storage) posByN(leftName, rightName string) (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	leftID := s.id(leftName)
	if _, exists := s.lefts[leftID]; !exists {
		return 0, ErrLeftNotFound
	}

	key := s.pos(leftID, s.id(rightName))

	if _, exists := s.items[key]; !exists {
		return 0, ErrRightNotFound
	}

	return key, nil
}
