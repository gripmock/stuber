package stuber

import (
	"errors"
	"fmt"
	"iter"
	"maps"
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/google/uuid"
)

// ErrServiceNotFound is returned when the service is not found.
var ErrServiceNotFound = errors.New("service not found")

// ErrMethodNotFound is returned when the method is not found.
var ErrMethodNotFound = errors.New("method not found")

// ErrStubNotFound is returned when the stub is not found.
var ErrStubNotFound = errors.New("stub not found")

// PriorityMultiplier is used to boost priority in ranking calculations.
const PriorityMultiplier = 10000

// searcher is a struct that manages the storage of search results.
//
// It contains a mutex for concurrent access, a map to store and retrieve
// used stubs by their UUID, and a pointer to the storage struct.
type searcher struct {
	mu       sync.RWMutex // mutex for concurrent access
	stubUsed map[uuid.UUID]struct{}
	// map to store and retrieve used stubs by their UUID

	storage *storage // pointer to the storage struct
}

// newSearcher creates a new instance of the searcher struct.
//
// It initializes the stubUsed map and the storage pointer.
//
// Returns a pointer to the newly created searcher struct.
func newSearcher() *searcher {
	return &searcher{
		storage:  newStorage(),
		stubUsed: make(map[uuid.UUID]struct{}),
	}
}

// Result represents the result of a search operation.
//
// It contains two fields: found and similar. Found represents the exact
// match found in the search, while similar represents the most similar match
// found.
type Result struct {
	found   *Stub // The exact match found in the search
	similar *Stub // The most similar match found
}

// Found returns the exact match found in the search.
//
// Returns a pointer to the Stub struct representing the found match.
func (r *Result) Found() *Stub {
	return r.found
}

// Similar returns the most similar match found in the search.
//
// Returns a pointer to the Stub struct representing the similar match.
func (r *Result) Similar() *Stub {
	return r.similar
}

// BidiResult represents the result of a bidirectional streaming search operation.
// For bidirectional streaming, we need to maintain state and filter stubs as messages arrive.
type BidiResult struct {
	searcher       *searcher
	service        string
	method         string
	headers        map[string]any
	allStubs       []*Stub      // All available stubs for this service/method
	candidateStubs []*Stub      // Stubs that match the pattern so far
	messageIndex   int          // Current message index in the stream
	isFirstCall    bool         // Whether this is the first call to Next()
	mu             sync.RWMutex // Thread safety for concurrent access
}

// Next finds a matching stub for the given message data.
// Each call to Next filters the candidate stubs based on the new message.
//
//nolint:cyclop,funlen
func (br *BidiResult) Next(messageData map[string]any) (*Stub, error) {
	br.mu.Lock()
	defer br.mu.Unlock()

	// Validate input
	if messageData == nil {
		return nil, ErrStubNotFound
	}

	// Validate service and method
	if br.service == "" || br.method == "" {
		return nil, ErrStubNotFound
	}

	// Validate headers
	if br.headers == nil {
		br.headers = make(map[string]any)
	}

	// Validate allStubs
	if len(br.allStubs) == 0 {
		return nil, ErrStubNotFound
	}

	// On first call, initialize candidate stubs
	if br.isFirstCall {
		br.candidateStubs = make([]*Stub, 0, len(br.allStubs))
		br.messageIndex = 0
		br.isFirstCall = false

		// Find all stubs that could potentially match the pattern
		for _, stub := range br.allStubs {
			if br.canStubMatchPattern(stub, messageData) {
				br.candidateStubs = append(br.candidateStubs, stub)
			}
		}
	} else {
		// Filter existing candidate stubs based on new message
		br.messageIndex++

		var newCandidates []*Stub

		for _, stub := range br.candidateStubs {
			if br.canStubMatchPattern(stub, messageData) {
				newCandidates = append(newCandidates, stub)
			}
		}

		br.candidateStubs = newCandidates
	}

	// If no candidates remain, return error
	if len(br.candidateStubs) == 0 {
		return nil, ErrStubNotFound
	}

	// Find the best matching stub among candidates
	var (
		bestStub               *Stub
		bestRank               float64
		candidatesWithSameRank []*Stub
	)

	// Create query once and reuse
	query := QueryV2{
		Service: br.service,
		Method:  br.method,
		Headers: br.headers,
		Input:   []map[string]any{messageData},
	}

	for _, stub := range br.candidateStubs {
		if br.stubMatchesCurrentMessage(stub, messageData) {
			rank := br.rankStub(stub, query)
			// Add priority to ranking with higher multiplier
			priorityBonus := float64(stub.Priority) * PriorityMultiplier
			totalRank := rank + priorityBonus

			if totalRank > bestRank {
				bestStub = stub
				bestRank = totalRank
				candidatesWithSameRank = []*Stub{stub}
			} else if totalRank == bestRank {
				// Collect candidates with same rank for stable sorting
				candidatesWithSameRank = append(candidatesWithSameRank, stub)
			}
		}
	}

	// If we have multiple candidates with same rank, sort by ID for stability
	if len(candidatesWithSameRank) > 1 {
		sortStubsByID(candidatesWithSameRank)
		bestStub = candidatesWithSameRank[0]
	}

	if bestStub != nil {
		// Mark the stub as used
		br.searcher.markV2(query, bestStub.ID)

		return bestStub, nil
	}

	return nil, ErrStubNotFound
}

// canStubMatchPattern checks if a stub could potentially match the pattern
// based on the current message index and available stream data.
func (br *BidiResult) canStubMatchPattern(stub *Stub, _ map[string]any) bool {
	// For client streaming stubs, check if we have enough stream data
	if stub.IsClientStream() {
		return br.messageIndex < len(stub.Stream)
	}

	// For unary stubs, always consider them as fallback candidates
	if stub.IsUnary() {
		return true
	}

	// For server streaming stubs, consider them as fallback candidates
	if stub.IsServerStream() {
		return true
	}

	return false
}

// stubMatchesCurrentMessage checks if a stub matches the current message.
func (br *BidiResult) stubMatchesCurrentMessage(stub *Stub, messageData map[string]any) bool {
	// For client streaming stubs, use Stream matching at current index
	if stub.IsClientStream() && br.messageIndex < len(stub.Stream) {
		return br.matchInputData(stub.Stream[br.messageIndex], messageData)
	}

	// For unary stubs, use Input matching
	if stub.IsUnary() {
		return br.matchInputData(stub.Input, messageData)
	}

	// For server streaming stubs, use Input matching
	if stub.IsServerStream() {
		return br.matchInputData(stub.Input, messageData)
	}

	return false
}

// matchInputData checks if messageData matches the given InputData.
//
//nolint:cyclop
func (br *BidiResult) matchInputData(inputData InputData, messageData map[string]any) bool {
	// Early exit if InputData is empty
	if len(inputData.Equals) == 0 && len(inputData.Contains) == 0 && len(inputData.Matches) == 0 {
		return true
	}

	// Check Equals
	if len(inputData.Equals) > 0 {
		for key, expectedValue := range inputData.Equals {
			if actualValue, exists := br.findValueWithVariations(messageData, key); !exists || !deepEqual(actualValue, expectedValue) {
				return false
			}
		}
	}

	// Check Contains - avoid creating temporary maps
	if len(inputData.Contains) > 0 {
		for key, expectedValue := range inputData.Contains {
			actualValue, exists := messageData[key]
			if !exists {
				return false
			}
			// Create minimal map for contains check
			tempMap := map[string]any{key: expectedValue}
			if !contains(tempMap, actualValue, false) {
				return false
			}
		}
	}

	// Check Matches - avoid creating temporary maps
	if len(inputData.Matches) > 0 {
		for key, expectedValue := range inputData.Matches {
			actualValue, exists := messageData[key]
			if !exists {
				return false
			}
			// Create minimal map for matches check
			tempMap := map[string]any{key: expectedValue}
			if !matches(tempMap, actualValue, false) {
				return false
			}
		}
	}

	return true
}

// findValueWithVariations tries to find a value using different field name conventions.
func (br *BidiResult) findValueWithVariations(messageData map[string]any, key string) (any, bool) {
	// Try exact match first
	if value, exists := messageData[key]; exists {
		return value, true
	}

	// Try camelCase variations
	camelKey := toCamelCase(key)
	if value, exists := messageData[camelKey]; exists {
		return value, true
	}

	// Try snake_case variations
	snakeKey := toSnakeCase(key)
	if value, exists := messageData[snakeKey]; exists {
		return value, true
	}

	return nil, false
}

// toCamelCase converts snake_case to camelCase.
func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	if len(parts) == 1 {
		return s
	}

	result := parts[0]
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}

	return result
}

// toSnakeCase converts camelCase to snake_case.
func toSnakeCase(s string) string {
	if s == "" {
		return ""
	}

	var result strings.Builder

	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result.WriteByte('_')
		}

		result.WriteRune(unicode.ToLower(r))
	}

	return result.String()
}

// deepEqual performs deep equality check with better implementation.
//
//nolint:cyclop,gocognit,nestif
func deepEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	// Try direct comparison first (only for comparable types)
	switch a.(type) {
	case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return a == b
	}

	// For maps, compare keys and values
	if aMap, aOk := a.(map[string]any); aOk {
		if bMap, bOk := b.(map[string]any); bOk {
			if len(aMap) != len(bMap) {
				return false
			}

			for k, v := range aMap {
				if bv, exists := bMap[k]; !exists || !deepEqual(v, bv) {
					return false
				}
			}

			return true
		}
	}

	// For slices, compare elements
	if aSlice, aOk := a.([]any); aOk {
		if bSlice, bOk := b.([]any); bOk {
			if len(aSlice) != len(bSlice) {
				return false
			}

			for i, v := range aSlice {
				if !deepEqual(v, bSlice[i]) {
					return false
				}
			}

			return true
		}
	}

	// Fallback to string comparison
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// sortStubsByID sorts stubs by ID for stable ordering when ranks are equal
// This ensures consistent results across multiple runs.
func sortStubsByID(stubs []*Stub) {
	sort.Slice(stubs, func(i, j int) bool {
		// Compare UUIDs directly for better performance
		return stubs[i].ID.String() < stubs[j].ID.String()
	})
}

// rankStub calculates the ranking score for a stub.
func (br *BidiResult) rankStub(stub *Stub, query QueryV2) float64 {
	// Use the existing V2 ranking logic
	return rankMatchV2(query, stub)
}

// upsert inserts the given stub values into the searcher. If a stub value
// already exists with the same key, it is updated.
//
// The function returns a slice of UUIDs representing the keys of the
// inserted or updated values.
func (s *searcher) upsert(values ...*Stub) []uuid.UUID {
	return s.storage.upsert(s.castToValue(values)...)
}

// del deletes the stub values with the given UUIDs from the searcher.
//
// Returns the number of stub values that were successfully deleted.
func (s *searcher) del(ids ...uuid.UUID) int {
	return s.storage.del(ids...)
}

// findByID retrieves the stub value associated with the given ID from the
// searcher.
//
// Returns a pointer to the Stub struct associated with the given ID, or nil
// if not found.
func (s *searcher) findByID(id uuid.UUID) *Stub {
	if v, ok := s.storage.findByID(id).(*Stub); ok {
		return v
	}

	return nil
}

// findBy retrieves all Stub values that match the given service and method
// from the searcher, sorted by score in descending order.
func (s *searcher) findBy(service, method string) ([]*Stub, error) {
	seq, err := s.storage.findAll(service, method)
	if err != nil {
		return nil, s.wrap(err)
	}

	return collectStubs(seq), nil
}

// clear resets the searcher.
//
// It clears the stubUsed map and calls the storage clear method.
func (s *searcher) clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear the stubUsed map.
	s.stubUsed = make(map[uuid.UUID]struct{})

	// Clear the storage.
	s.storage.clear()
}

// all returns all Stub values stored in the searcher.
//
// Returns:
// - []*Stub: The Stub values stored in the searcher.
func (s *searcher) all() []*Stub {
	return collectStubs(s.storage.values())
}

// used returns all Stub values that have been used by the searcher.
//
// Returns:
// - []*Stub: The Stub values that have been used by the searcher.
func (s *searcher) used() []*Stub {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return collectStubs(s.storage.findByIDs(maps.Keys(s.stubUsed)))
}

// unused returns all Stub values that have not been used by the searcher.
//
// Returns:
// - []*Stub: The Stub values that have not been used by the searcher.
func (s *searcher) unused() []*Stub {
	s.mu.RLock()
	defer s.mu.RUnlock()

	unused := make([]*Stub, 0)

	for stub := range s.iterAll() {
		if _, exists := s.stubUsed[stub.ID]; !exists {
			unused = append(unused, stub)
		}
	}

	return unused
}

// searchCommon is a common search function that can be used by both search and searchV2.
func (s *searcher) searchCommon(
	service, method string,
	matchFunc func(*Stub) bool,
	rankFunc func(*Stub) float64,
	markFunc func(uuid.UUID),
) (*Result, error) {
	var (
		found       *Stub
		foundRank   float64
		similar     *Stub
		similarRank float64
	)

	seq, err := s.storage.findAll(service, method)
	if err != nil {
		return nil, s.wrap(err)
	}

	// Collect all stubs first for stable sorting
	stubs := make([]*Stub, 0)
	for v := range seq {
		stub, ok := v.(*Stub)
		if !ok {
			continue
		}

		stubs = append(stubs, stub)
	}

	// Sort stubs by ID for stable ordering when ranks are equal
	sortStubsByID(stubs)

	// Process stubs in sorted order
	for _, stub := range stubs {
		current := rankFunc(stub)
		// Add priority to ranking with higher multiplier
		priorityBonus := float64(stub.Priority) * PriorityMultiplier
		totalRank := current + priorityBonus

		if totalRank > similarRank {
			similar, similarRank = stub, totalRank
		}

		if matchFunc(stub) && totalRank > foundRank {
			found, foundRank = stub, totalRank
		}
	}

	if found != nil {
		markFunc(found.ID)

		return &Result{found: found}, nil
	}

	if similar != nil {
		return &Result{similar: similar}, nil
	}

	return nil, ErrStubNotFound
}

// find retrieves the Stub value associated with the given Query from the searcher.
//
// Parameters:
// - query: The Query used to search for a Stub value.
//
// Returns:
// - *Result: The Result containing the found Stub value (if any), or nil.
// - error: An error if the search fails.
func (s *searcher) find(query Query) (*Result, error) {
	// Check if the Query has an ID field.
	if query.ID != nil {
		// Search for the Stub value with the given ID.
		return s.searchByID(query)
	}

	// Search for the Stub value with the given service and method.
	return s.search(query)
}

// searchByID retrieves the Stub value associated with the given ID from the searcher.
//
// Parameters:
// - query: The Query used to search for a Stub value.
//
// Returns:
// - *Result: The Result containing the found Stub value (if any), or nil.
// - error: An error if the search fails.
func (s *searcher) searchByID(query Query) (*Result, error) {
	// Check if the given service and method are valid.
	_, err := s.storage.posByPN(query.Service, query.Method)
	if err != nil {
		return nil, s.wrap(err)
	}

	// Search for the Stub value with the given ID.
	if found := s.findByID(*query.ID); found != nil {
		// Mark the Stub value as used.
		s.mark(query, *query.ID)

		// Return the found Stub value.
		return &Result{found: found}, nil
	}

	// Return an error if the Stub value is not found.
	return nil, ErrServiceNotFound
}

// search retrieves the Stub value associated with the given Query from the searcher.
//
// Parameters:
// - query: The Query used to search for a Stub value.
//
// Returns:
// - *Result: The Result containing the found Stub value (if any), or nil.
// - error: An error if the search fails.
func (s *searcher) search(query Query) (*Result, error) {
	return s.searchCommon(query.Service, query.Method,
		func(stub *Stub) bool { return match(query, stub) },
		func(stub *Stub) float64 { return rankMatch(query, stub) },
		func(id uuid.UUID) { s.mark(query, id) })
}

// mark marks the given Stub value as used in the searcher.
//
// If the query's RequestInternal flag is set, the mark is skipped.
//
// Parameters:
// - query: The query used to mark the Stub value.
// - id: The UUID of the Stub value to mark.
func (s *searcher) mark(query Query, id uuid.UUID) {
	// If the query's RequestInternal flag is set, skip the mark.
	if query.RequestInternal() {
		return
	}

	// Lock the mutex to ensure concurrent access.
	s.mu.Lock()
	defer s.mu.Unlock()

	// Mark the Stub value as used by adding it to the stubUsed map.
	s.stubUsed[id] = struct{}{}
}

// findV2 retrieves the Stub value associated with the given QueryV2 from the searcher.
func (s *searcher) findV2(query QueryV2) (*Result, error) {
	// Check if the QueryV2 has an ID field
	if query.ID != nil {
		// Search for the Stub value with the given ID
		return s.searchByIDV2(query)
	}

	// Search for the Stub value with the given service and method
	return s.searchV2(query)
}

// searchByIDV2 retrieves the Stub value associated with the given ID from the searcher.
func (s *searcher) searchByIDV2(query QueryV2) (*Result, error) {
	// Check if the given service and method are valid
	_, err := s.storage.posByPN(query.Service, query.Method)
	if err != nil {
		return nil, s.wrap(err)
	}

	// Search for the Stub value with the given ID
	if found := s.findByID(*query.ID); found != nil {
		// Mark the Stub value as used
		s.markV2(query, *query.ID)

		// Return the found Stub value
		return &Result{found: found}, nil
	}

	// Return an error if the Stub value is not found
	return nil, ErrServiceNotFound
}

// findBidi retrieves a BidiResult for bidirectional streaming with the given QueryBidi.
// For bidirectional streaming, each message is treated as a separate unary request.
func (s *searcher) findBidi(query QueryBidi) (*BidiResult, error) {
	// Check if the QueryBidi has an ID field
	if query.ID != nil {
		// For ID-based queries, we can't use bidirectional streaming - fallback to regular search
		return s.searchByIDBidi(query)
	}

	// Check if the given service and method are valid
	_, err := s.storage.posByPN(query.Service, query.Method)
	if err != nil {
		return nil, s.wrap(err)
	}

	// Fetch all stubs for this service/method
	seq, err := s.storage.findAll(query.Service, query.Method)
	if err != nil {
		return nil, s.wrap(err)
	}

	var allStubs []*Stub
	for v := range seq {
		if stub, ok := v.(*Stub); ok {
			allStubs = append(allStubs, stub)
		}
	}

	return &BidiResult{
		searcher:       s,
		service:        query.Service,
		method:         query.Method,
		headers:        query.Headers,
		allStubs:       allStubs,
		candidateStubs: make([]*Stub, 0, len(allStubs)),
		messageIndex:   0,
		isFirstCall:    true,
	}, nil
}

// searchByIDBidi handles ID-based queries for bidirectional streaming.
// Since we can't use bidirectional streaming for ID-based queries, we fallback to regular search.
func (s *searcher) searchByIDBidi(query QueryBidi) (*BidiResult, error) {
	// Check if the given service and method are valid
	_, err := s.storage.posByPN(query.Service, query.Method)
	if err != nil {
		return nil, s.wrap(err)
	}

	// Search for the Stub value with the given ID
	if found := s.findByID(*query.ID); found != nil {
		return &BidiResult{
			searcher:       s,
			service:        query.Service,
			method:         query.Method,
			headers:        query.Headers,
			allStubs:       []*Stub{found},
			candidateStubs: []*Stub{found},
			messageIndex:   0,
			isFirstCall:    true,
		}, nil
	}

	// Return an error if the Stub value is not found
	return nil, ErrServiceNotFound
}

// searchV2 retrieves the Stub value associated with the given QueryV2 from the searcher.
func (s *searcher) searchV2(query QueryV2) (*Result, error) {
	return s.searchCommon(query.Service, query.Method,
		func(stub *Stub) bool { return matchV2(query, stub) },
		func(stub *Stub) float64 { return rankMatchV2(query, stub) },
		func(id uuid.UUID) { s.markV2(query, id) })
}

// markV2 marks the given Stub value as used in the searcher.
func (s *searcher) markV2(query QueryV2, id uuid.UUID) {
	// If the query's RequestInternal flag is set, skip the mark
	if query.RequestInternal() {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.stubUsed[id] = struct{}{}
}

func collectStubs(seq iter.Seq[Value]) []*Stub {
	result := make([]*Stub, 0)

	for v := range seq {
		if stub, ok := v.(*Stub); ok {
			result = append(result, stub)
		}
	}

	return result
}

func (s *searcher) iterAll() iter.Seq[*Stub] {
	return func(yield func(*Stub) bool) {
		for v := range s.storage.values() {
			if stub, ok := v.(*Stub); ok {
				if !yield(stub) {
					return
				}
			}
		}
	}
}

// castToValue converts a slice of *Stub values to a slice of Value any.
//
// Parameters:
// - values: A slice of *Stub values to convert.
//
// Returns:
// - A slice of Value any containing the converted values.
func (s *searcher) castToValue(values []*Stub) []Value {
	result := make([]Value, len(values))
	for i, v := range values {
		result[i] = v
	}

	return result
}

// wrap wraps an error with specific error types.
//
// Parameters:
// - err: The error to wrap.
//
// Returns:
// - The wrapped error.
func (s *searcher) wrap(err error) error {
	if errors.Is(err, ErrLeftNotFound) {
		return ErrServiceNotFound
	}

	if errors.Is(err, ErrRightNotFound) {
		return ErrMethodNotFound
	}

	return err
}
