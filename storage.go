package stuber

import (
	"errors"
	"iter"
	"strings"
	"sync"

	"github.com/google/uuid"
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
}

// storage is a struct that manages the storage of search results with optimized performance and memory usage.
//
// Fields:
// - mu: A mutex that is used to protect the storage from concurrent access.
// - leftTotal and rightTotal: The total number of left and right values in the storage.
// - lefts and rights: Maps of left and right values to their respective IDs.
// - leftRights: A map that is used to store the pairing of left and right values.
// - items: A map of left-right pair IDs to their respective items.
// - itemsByID: A map of item IDs to their respective items.
type storage struct {
	mu         sync.RWMutex
	leftTotal  uint64
	rightTotal uint64
	lefts      map[string]uint64
	rights     map[string]uint64
	leftRights map[[2]uint64]struct{}
	items      map[uuid.UUID]map[uuid.UUID]Value
	itemsByID  map[uuid.UUID]Value
}

// itemMapPool is a sync.Pool that is used to store and retrieve maps that
// are used to store items in the storage.
//
// The New method of the sync.Pool is used to create a new map when the pool
// is empty. The method returns a pointer to the newly created map.
//
//nolint:gochecknoglobals
var itemMapPool = sync.Pool{
	New: func() any {
		return make(map[uuid.UUID]Value)
	},
}

// newStorage creates a new instance of the storage struct.
func newStorage() *storage {
	return &storage{
		lefts:      make(map[string]uint64),
		rights:     make(map[string]uint64),
		leftRights: make(map[[2]uint64]struct{}),
		items:      make(map[uuid.UUID]map[uuid.UUID]Value),
		itemsByID:  make(map[uuid.UUID]Value),
	}
}

// clear resets the storage.
func (s *storage) clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.leftTotal = 0
	s.rightTotal = 0
	s.lefts = make(map[string]uint64)
	s.rights = make(map[string]uint64)
	s.leftRights = make(map[[2]uint64]struct{})
	s.items = make(map[uuid.UUID]map[uuid.UUID]Value)
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

// findAll retrieves all Value items that match the given left and right names.
// It returns an iterator sequence of the matched Value items or an error if the
// names are not found.
//
// Parameters:
// - left: The left name for matching.
// - right: The right name for matching.
//
// Returns:
// - iter.Seq[Value]: A sequence of matched Value items.
// - error: An error if the left or right name is not found.
func (s *storage) findAll(left, right string) (iter.Seq[Value], error) {
	indexes, err := s.posByPN(left, right)
	if err != nil {
		return nil, err
	}

	return func(yield func(Value) bool) {
		s.mu.RLock()
		defer s.mu.RUnlock()

		for _, index := range indexes {
			for _, v := range s.items[index] {
				if !yield(v) {
					return
				}
			}
		}
	}, nil
}

// posByPN attempts to resolve UUIDs for a given left and right name pair.
// It first tries to resolve the full left name with the right name, and then
// attempts to resolve using a truncated version of the left name if necessary.
//
// Parameters:
// - left: The left name for matching.
// - right: The right name for matching.
//
// Returns:
// - []uuid.UUID: A slice of resolved UUIDs.
// - error: An error if no UUIDs were resolved.
func (s *storage) posByPN(left, right string) ([]uuid.UUID, error) {
	// Initialize a slice to store the resolved UUIDs.
	var resolvedUUIDs []uuid.UUID

	// Attempt to resolve the full left name with the right name.
	uuid, err := s.posByN(left, right)
	if err == nil {
		// Append the resolved UUID to the slice.
		resolvedUUIDs = append(resolvedUUIDs, uuid)
	}

	// Check for a potential truncation point in the left name.
	if dotIndex := strings.LastIndex(left, "."); dotIndex != -1 {
		truncatedLeft := left[dotIndex+1:]

		// Attempt to resolve the truncated left name with the right name.
		if uuid, err := s.posByN(truncatedLeft, right); err == nil {
			// Append the resolved UUID to the slice.
			resolvedUUIDs = append(resolvedUUIDs, uuid)
		} else if errors.Is(err, ErrRightNotFound) && len(resolvedUUIDs) == 0 {
			// Return an error if the right name was not found
			// and no UUIDs were resolved.
			return nil, err
		}
	}

	// Return an error if no UUIDs were resolved.
	if len(resolvedUUIDs) == 0 {
		// Return the original error if we have it.
		return nil, err
	}

	// Return the resolved UUIDs.
	return resolvedUUIDs, nil
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
// It returns the UUIDs of the inserted or updated items.
func (s *storage) upsert(values ...Value) []uuid.UUID {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Prepare a slice to store the result UUIDs.
	results := make([]uuid.UUID, len(values))

	// Process each value to insert or update.
	for i, v := range values {
		// Retrieve or create IDs for the left and right names.
		leftID := s.leftIDOrNew(v.Left())
		rightID := s.rightIDOrNew(v.Right())

		// Compute the composite index from the left and right IDs.
		index := s.pos(leftID, rightID)

		// Register the left-right pairing.
		s.leftRights[[2]uint64{leftID, rightID}] = struct{}{}

		// Initialize the map at the index if it doesn't exist.
		if s.items[index] == nil {
			if p := itemMapPool.Get(); p != nil {
				s.items[index], _ = p.(map[uuid.UUID]Value)
				// Clear any pre-existing entries.
				for k := range s.items[index] {
					delete(s.items[index], k)
				}
			} else {
				s.items[index] = make(map[uuid.UUID]Value)
			}
		}

		// Insert or update the value in the storage.
		s.items[index][v.Key()] = v
		s.itemsByID[v.Key()] = v

		// Record the UUID of the processed value.
		results[i] = v.Key()
	}

	// Return the UUIDs of the inserted or updated values.
	return results
}

// del deletes the Stub values with the given UUIDs from the storage.
// It returns the number of Stub values that were successfully deleted.
func (s *storage) del(keys ...uuid.UUID) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	type posKey struct {
		pos uuid.UUID
		key uuid.UUID
	}

	toDelete := make([]posKey, 0, len(keys))

	for _, key := range keys {
		if v, ok := s.itemsByID[key]; ok {
			leftID := s.lefts[v.Left()]
			rightID := s.rights[v.Right()]
			pos := s.pos(leftID, rightID)
			toDelete = append(toDelete, posKey{pos, key})
		}
	}

	deleted := 0
	posMap := make(map[uuid.UUID][]uuid.UUID, len(toDelete))

	for _, pk := range toDelete {
		posMap[pk.pos] = append(posMap[pk.pos], pk.key)
	}

	for pos, ids := range posMap {
		m := s.items[pos]
		for _, id := range ids {
			delete(m, id)
			delete(s.itemsByID, id)

			deleted++
		}

		if len(m) == 0 {
			itemMapPool.Put(m)
			delete(s.items, pos)
		}
	}

	return deleted
}

// leftID retrieves the ID associated with the given left name.
// It returns the ID if found, or an error if the left name does not exist.
func (s *storage) leftID(leftName string) (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.lefts[leftName]
	if !ok {
		return 0, ErrLeftNotFound
	}

	return id, nil
}

// leftIDOrNew returns the ID of the given left name.
// If the ID does not exist, it will be created.
func (s *storage) leftIDOrNew(name string) uint64 {
	if id, exists := s.lefts[name]; exists {
		return id
	}

	s.leftTotal++
	s.lefts[name] = s.leftTotal

	return s.leftTotal
}

// rightID returns the ID of the given right name.
// If the ID does not exist, ErrRightNotFound will be returned.
func (s *storage) rightID(rightName string) (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.rights[rightName]
	if !ok {
		return 0, ErrRightNotFound
	}

	return id, nil
}

// rightIDOrNew returns the ID of the given right name.
// If the ID does not exist, it will be created.
func (s *storage) rightIDOrNew(name string) uint64 {
	if id, exists := s.rights[name]; exists {
		return id
	}

	s.rightTotal++
	s.rights[name] = s.rightTotal

	return s.rightTotal
}

// posByN creates a UUID from the given left and right names.
// The resulting UUID is 128 bits long, with the left ID in the first 64 bits,
// and the right ID in the second 64 bits.
func (s *storage) posByN(leftName, rightName string) (uuid.UUID, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	leftID, leftExists := s.lefts[leftName]
	if !leftExists {
		return uuid.Nil, ErrLeftNotFound
	}

	rightID, rightExists := s.rights[rightName]
	if !rightExists {
		return uuid.Nil, ErrRightNotFound
	}

	key := [2]uint64{leftID, rightID}
	if _, exists := s.leftRights[key]; !exists {
		return uuid.Nil, ErrRightNotFound
	}

	return s.pos(leftID, rightID), nil
}

// pos creates a UUID from the given left and right IDs.
// The resulting UUID is 128 bits long, with the left ID in the first 64 bits,
// and the right ID in the second 64 bits.
func (s *storage) pos(leftID, rightID uint64) uuid.UUID {
	//nolint:mnd
	return uuid.UUID{
		byte(leftID >> 56),
		byte(leftID >> 48),
		byte(leftID >> 40),
		byte(leftID >> 32),
		byte(leftID >> 24),
		byte(leftID >> 16),
		byte(leftID >> 8),
		byte(leftID),
		byte(rightID >> 56),
		byte(rightID >> 48),
		byte(rightID >> 40),
		byte(rightID >> 32),
		byte(rightID >> 24),
		byte(rightID >> 16),
		byte(rightID >> 8),
		byte(rightID),
	}
}
