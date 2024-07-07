package stuber

import (
	"errors"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"golang.org/x/exp/maps"
)

// ErrLeftNotFound is returned when the left value is not found.
var ErrLeftNotFound = errors.New("left not found")

// ErrRightNotFound is returned when the right value is not found.
var ErrRightNotFound = errors.New("right not found")

// Value is a type used to store the result of a search.
//
// This interface is used to represent the search results returned by the
// Find and Search methods.
type Value interface {
	Key() uuid.UUID // The UUID of the value.
	Left() string   // The left value of the value.
	Right() string  // The right value of the value.
}

// storage is a struct that manages the storage of search results.
//
// It contains a mutex for concurrent access, the total number of stored items,
// maps to store and retrieve values by their left and right values, a map to
// store values by their UUID, and a map to retrieve values by their UUID.
type storage struct {
	mu         sync.RWMutex          // Mutex for concurrent access.
	leftTotal  atomic.Uint64         // Total number of stored left values.
	rightTotal atomic.Uint64         // Total number of stored right values.
	lefts      map[string]uint64     // Map to store values by their left values.
	rights     map[string]uint64     // Map to store values by their right values.
	leftRights map[uint64][]uint64   // Map to store the right values associated with a left value.
	items      map[uuid.UUID][]Value // Map to store values by their UUID.
	itemsByID  map[uuid.UUID]Value   // Map to retrieve values by their UUID.
}

// newStorage creates a new storage instance.
//
// It creates a new instance of the storage struct with empty maps.
func newStorage() *storage {
	return &storage{
		rights:     map[string]uint64{},
		lefts:      map[string]uint64{},
		leftRights: map[uint64][]uint64{},
		items:      map[uuid.UUID][]Value{},
		itemsByID:  map[uuid.UUID]Value{},
	}
}

// clear resets the storage.
//
// It resets all the internal maps and counters to their initial state.
func (s *storage) clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Reset the total number of stored left values.
	s.leftTotal = atomic.Uint64{}

	// Reset the total number of stored right values.
	s.rightTotal = atomic.Uint64{}

	// Reset the map that stores values by their left values.
	s.lefts = map[string]uint64{}

	// Reset the map that stores values by their right values.
	s.rights = make(map[string]uint64)

	// Reset the map that stores the right values associated with a left value.
	s.leftRights = map[uint64][]uint64{}

	// Reset the map that stores values by their UUID.
	s.items = map[uuid.UUID][]Value{}

	// Reset the map that retrieves values by their UUID.
	s.itemsByID = map[uuid.UUID]Value{}
}

func (s *storage) values() []Value {
	// values returns all the values stored in the storage.
	//
	// This function returns a slice of Value objects containing all the values
	// stored in the storage. The values are returned in an arbitrary order.
	return maps.Values(s.itemsByID)
}

// findAll retrieves all the values associated with a given left and right values.
//
// This function takes a left and right value as parameters and returns a slice of
// Value objects containing all the values associated with those values. If no
// values are found, it returns an empty slice and a nil error.
//
// Parameters:
// - left: The left value to search for.
// - right: The right value to search for.
//
// Returns:
//   - []Value: A slice containing all the values associated with the given left
//     and right values.
//   - error: A nil error if the values are found, otherwise an error indicating
//     that the values were not found.
func (s *storage) findAll(left, right string) ([]Value, error) {
	// Find the position of the given left and right values.
	pos, err := s.posByN(left, right)
	if err != nil {
		return nil, err
	}

	// Lock the storage for reading.
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Retrieve the values associated with the given position.
	return s.items[pos], nil
}

// findByID retrieves the value associated with the given ID.
//
// This function takes a key as a parameter and returns the value associated with
// that key. If no value is found, it returns nil.
//
// Parameters:
//   - key: The ID of the value to search for.
//
// Returns:
//   - Value: The value associated with the given ID, or nil if no value is
//     found.
func (s *storage) findByID(key uuid.UUID) Value { //nolint:ireturn
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if the value exists in the storage.
	if v, ok := s.itemsByID[key]; ok {
		return v
	}

	return nil
}

// findByIDs retrieves the values associated with the given IDs.
//
// This function takes a slice of keys as a parameter and returns the values
// associated with those keys. If a value is not found, it is not included in
// the results.
//
// Parameters:
// - keys: A slice of IDs to search for.
//
// Returns:
//   - []Value: A slice of values associated with the given IDs.
func (s *storage) findByIDs(keys ...uuid.UUID) []Value {
	// Lock the storage for reading.
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Initialize a slice to store the results.
	results := make([]Value, 0, len(keys))

	// Iterate over each key.
	for _, key := range keys {
		// Check if the value exists in the storage.
		if v, ok := s.itemsByID[key]; ok {
			// Append the value to the results if it exists.
			results = append(results, v)
		}
	}

	// Return the results.
	return results
}

func (s *storage) upsert(values ...Value) []uuid.UUID {
	// upsert inserts the given values into the storage. If a value already exists
	// with the same key, it is updated.
	//
	// The function returns a slice of UUIDs representing the keys of the inserted
	// or updated values.
	results := make([]uuid.UUID, len(values))

	for i, v := range values {
		// Get the ID of the left value. If it does not exist, create a new ID.
		leftID := s.leftIDOrNew(v.Left())

		// Get the ID of the right value. If it does not exist, create a new ID.
		rightID := s.rightIDOrNew(v.Right())

		// Calculate the index of the value based on the left and right IDs.
		ind := s.pos(leftID, rightID)

		// Lock the storage for writing.
		s.mu.Lock()

		// Store the key and value in the storage.
		results[i] = v.Key()

		s.leftRights[leftID] = append(s.leftRights[leftID], rightID)
		s.items[ind] = append(s.items[ind], v)
		s.itemsByID[v.Key()] = v

		// Unlock the storage.
		s.mu.Unlock()
	}

	// Return the keys of the inserted or updated values.
	return results
}

// del deletes the values with the given keys from the storage.
//
// The function returns the number of values that were successfully deleted.
func (s *storage) del(keys ...uuid.UUID) int {
	result := 0
	// Map to store the keys to be deleted for each position.
	deleteIDs := make(map[uuid.UUID][]uuid.UUID, len(keys))

	// Iterate over the keys to be deleted.
	for _, key := range keys {
		// Get the value associated with the key.
		v := s.findByID(key)
		// Skip if the value doesn't exist.
		if v == nil {
			continue
		}

		// Get the position of the value in the storage.
		pos, err := s.posByN(v.Left(), v.Right())
		// Skip if the position couldn't be determined.
		if err != nil {
			continue
		}

		// Add the key to the list of keys to be deleted for the position.
		deleteIDs[pos] = append(deleteIDs[pos], key)
		result++
	}

	// Lock the storage for writing.
	s.mu.Lock()
	defer s.mu.Unlock()

	// Delete the values with the keys from the storage.
	for pos, v := range deleteIDs {
		s.items[pos] = slices.DeleteFunc(s.items[pos], func(value Value) bool {
			// Check if the key of the value is in the list of keys to be deleted.
			return slices.Contains(v, value.Key())
		})
	}

	// Delete the values from the itemsByID map.
	for _, key := range keys {
		delete(s.itemsByID, key)
	}

	// Return the number of values that were successfully deleted.
	return result
}

func (s *storage) leftID(name string) (uint64, error) {
	// leftId returns the ID associated with the given left name.
	//
	// This function takes a left name as a parameter and returns the ID associated
	// with that name. If no ID is found, it returns 0 and an ErrLeftNotFound
	// error.
	//
	// Parameters:
	// - name: The name of the left to search for.
	//
	// Returns:
	//   - uint64: The ID associated with the given left name, or 0 if no ID is
	//     found.
	//   - error: An error if the left name is not found.
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if the ID exists in the lefts map.
	if id, ok := s.lefts[name]; ok {
		// Return the ID if it exists.
		return id, nil
	}

	// Return 0 and an error if the ID does not exist.
	return 0, ErrLeftNotFound
}

// leftIDOrNew returns the ID associated with the given left name.
//
// This function takes a left name as a parameter and returns the ID associated
// with that name. If no ID is found, it creates a new ID and returns it.
//
// Parameters:
// - name: The name of the left to search for or create.
//
// Returns:
//   - uint64: The ID associated with the given left name, either found or
//     created.
func (s *storage) leftIDOrNew(name string) uint64 {
	// Check if the ID exists in the lefts map.
	if id, err := s.leftID(name); err == nil {
		// Return the ID if it exists.
		return id
	}

	// Acquire a write lock to ensure atomicity.
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a new ID by incrementing the total count of lefts.
	s.lefts[name] = s.leftTotal.Add(1)

	// Return the newly created ID.
	return s.lefts[name]
}

// rightID returns the ID associated with the given right name.
//
// This function takes a right name as a parameter and returns the ID associated
// with that name. If no ID is found, it returns an error.
//
// Parameters:
// - name: The name of the right to search for.
//
// Returns:
//   - uint64: The ID associated with the given right name.
//   - error: An error if the ID is not found.
func (s *storage) rightID(name string) (uint64, error) {
	// Acquire a read lock to ensure read consistency.
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if the ID exists in the rights map.
	if id, ok := s.rights[name]; ok {
		// Return the ID if it exists.
		return id, nil
	}

	// Return 0 and an error if the ID does not exist.
	return 0, ErrRightNotFound
}

func (s *storage) rightIDOrNew(name string) uint64 {
	// Get the ID associated with the given right name.
	// If the ID exists, return it.
	if id, err := s.rightID(name); err == nil {
		return id
	}

	// Acquire a write lock to ensure atomicity.
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a new ID by incrementing the total count of rights.
	s.rights[name] = s.rightTotal.Add(1)

	// Return the newly created ID.
	return s.rights[name]
}

// posByN retrieves the position associated with the given left and right values.
//
// This function takes a left and right value as parameters and returns a UUID
// representing the position of those values. If the left or right ID is not
// found, it returns uuid.Nil and an error. It also checks if the left-right
// combination exists in the leftRights map and returns an error if it does not.
//
// Parameters:
// - left: The left value to search for.
// - right: The right value to search for.
//
// Returns:
//   - uuid.UUID: A UUID representing the position of the given left and right values.
//   - error: An error if the ID is not found or the left-right combination does not exist.
func (s *storage) posByN(left, right string) (uuid.UUID, error) {
	// Get the ID associated with the given left value.
	// If the ID exists, continue.
	leftID, err := s.leftID(left)
	if err != nil {
		return uuid.Nil, err
	}

	// Get the ID associated with the given right value.
	// If the ID exists, continue.
	rightID, err := s.rightID(right)
	if err != nil {
		return uuid.Nil, err
	}

	// Acquire a read lock to ensure read consistency.
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if the left-right combination exists in the leftRights map.
	if !slices.Contains(s.leftRights[leftID], rightID) {
		return uuid.Nil, ErrRightNotFound
	}

	// Calculate the position based on the left and right IDs.
	return s.pos(leftID, rightID), nil
}

// pos calculates the UUID based on the given left and right values.
//
// Parameters:
// - left: The left value.
// - right: The right value.
//
// Returns:
//   - uuid.UUID: The calculated UUID.
//
//nolint:mnd
func (s *storage) pos(left, right uint64) uuid.UUID {
	return uuid.UUID{
		byte(left >> 56),
		byte(left >> 48),
		byte(left >> 40),
		byte(left >> 32),
		byte(left >> 24),
		byte(left >> 16),
		byte(left >> 8),
		byte(left),
		byte(right >> 56),
		byte(right >> 48),
		byte(right >> 40),
		byte(right >> 32),
		byte(right >> 24),
		byte(right >> 16),
		byte(right >> 8),
		byte(right),
	}
}
