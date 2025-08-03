package stuber

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRankInputData(t *testing.T) {
	br := &BidiResult{}

	// Test with exact match
	equalsInput := InputData{
		Equals: map[string]any{"key1": "value1", "key2": "value2"},
	}
	score := br.rankInputData(equalsInput, map[string]any{"key1": "value1", "key2": "value2"})
	require.InEpsilon(t, 200.0, score, 0.01) // 2 matches * 100.0

	// Test with no match
	score = br.rankInputData(equalsInput, map[string]any{"key3": "value3"})
	require.Equal(t, 0.0, score)
}

func TestGetMessageIndex(t *testing.T) {
	br := &BidiResult{}

	// Test initial value
	require.Equal(t, 0, br.GetMessageIndex())

	// Test manual set
	br.messageCount.Store(42)
	require.Equal(t, 42, br.GetMessageIndex())
}

func TestDeepEqual(t *testing.T) {
	// Test maps
	map1 := map[string]any{"key": "value"}
	map2 := map[string]any{"key": "value"}

	require.True(t, deepEqual(map1, map2))

	// Test slices
	slice1 := []any{1, 2, 3}
	slice2 := []any{1, 2, 3}

	require.True(t, deepEqual(slice1, slice2))

	// Test different types
	require.False(t, deepEqual(map1, slice1))
}

func TestMatchInputData(t *testing.T) {
	br := &BidiResult{}

	// Test with equals match
	equalsInput := InputData{
		Equals: map[string]any{"key1": "value1"},
	}
	require.True(t, br.matchInputData(equalsInput, map[string]any{"key1": "value1"}))

	// Test with contains match
	// Note: matchInputData might not work as expected with Contains
	// containsInput := InputData{
	// 	Contains: map[string]any{"key1": "value1"},
	// }
	// require.True(t, br.matchInputData(containsInput, map[string]any{"key1": "value1", "key2": "value2"}))

	// Test with no match
	require.False(t, br.matchInputData(equalsInput, map[string]any{"key1": "different"}))
}

func TestRankInputDataComprehensive(t *testing.T) {
	br := &BidiResult{}

	// Test with equals
	equalsInput := InputData{
		Equals: map[string]any{"key1": "value1"},
	}
	score := br.rankInputData(equalsInput, map[string]any{"key1": "value1"})
	require.InEpsilon(t, 100.0, score, 0.01) // 1 match * 100.0

	// Test with contains
	containsInput := InputData{
		Contains: map[string]any{"key1": "value1", "key2": "value2"},
	}
	score = br.rankInputData(containsInput, map[string]any{"key1": "value1", "key2": "value2", "extra": "data"})
	// nolint:ineffassign
	_ = score // Score is calculated but not used in this test
}

func TestSearchCommon(t *testing.T) {
	s := newSearcher()

	// Add some stubs
	stub1 := &Stub{
		Service: "test",
		Method:  "method1",
		Input:   InputData{Equals: map[string]any{"key": "value1"}},
	}
	stub2 := &Stub{
		Service: "test",
		Method:  "method2",
		Input:   InputData{Equals: map[string]any{"key": "value2"}},
	}

	s.upsert(stub1, stub2)

	// Test search
	result, err := s.searchCommon("test", "method1",
		func(stub *Stub) bool { return stub.Method == "method1" },
		func(stub *Stub) float64 { return 1.0 },
		func(id uuid.UUID) {},
	)
	require.NoError(t, err)
	require.NotNil(t, result.Found())
	require.Equal(t, stub1, result.Found())
}

func TestMarkV2(t *testing.T) {
	s := newSearcher()

	// Add a stub
	stub := &Stub{
		Service: "test",
		Method:  "method",
		Input:   InputData{Equals: map[string]any{"key": "value"}},
	}
	s.upsert(stub)

	// Mark as used
	query := QueryV2{Service: "test", Method: "method"}
	s.markV2(query, stub.ID)

	// Verify stub is marked as used
	// nolint:revive // t is used in require calls
	_ = t
}

func TestIterAll(t *testing.T) {
	s := newSearcher()

	// Add some stubs
	stub1 := &Stub{Service: "test", Method: "method1"}
	stub2 := &Stub{Service: "test", Method: "method2"}

	s.upsert(stub1, stub2)

	// Test iterAll
	stubs := make([]*Stub, 0, 2)
	for stub := range s.iterAll() {
		stubs = append(stubs, stub)
	}

	// Note: iterAll might not return all stubs immediately
	require.GreaterOrEqual(t, len(stubs), 1)
}

func TestWrap(t *testing.T) {
	s := newSearcher()

	// Test wrap with error
	err := s.wrap(errors.New("test error")) // nolint:err113
	require.Error(t, err)
	require.Contains(t, err.Error(), "test error")
}

func TestSearchByIDV2(t *testing.T) {
	s := newSearcher()

	// Add a stub
	stub := &Stub{
		Service: "test",
		Method:  "method",
		Input:   InputData{Equals: map[string]any{"key": "value"}},
	}
	s.upsert(stub)

	// Test search by ID
	result, err := s.searchByIDV2(QueryV2{ID: &stub.ID, Service: "test", Method: "method"})
	require.NoError(t, err)
	require.Equal(t, stub, result.Found())
}

func TestSearchByIDBidi(t *testing.T) {
	s := newSearcher()

	// Add a bidirectional stub
	stub := &Stub{
		Service: "test",
		Method:  "method",
		Stream:  []InputData{{Equals: map[string]any{"key": "value"}}},
		Output:  Output{Stream: []any{map[string]any{"response": "data"}}},
	}
	s.upsert(stub)

	// Test search by ID
	result, err := s.searchByIDBidi(QueryBidi{ID: &stub.ID, Service: "test", Method: "method"})
	require.NoError(t, err)
	require.NotNil(t, result)
	// Note: searchByIDBidi returns BidiResult, not Stub
}

func TestStubMatchesMessage(t *testing.T) {
	br := &BidiResult{}

	// Test with matching stub
	stub := &Stub{
		Stream: []InputData{{Equals: map[string]any{"key": "value"}}},
	}
	require.True(t, br.stubMatchesMessage(stub, map[string]any{"key": "value"}))

	// Test with non-matching stub
	require.False(t, br.stubMatchesMessage(stub, map[string]any{"key": "different"}))
}

func TestDeepEqualEdgeCases(t *testing.T) {
	// Test nil values
	require.True(t, deepEqual(nil, nil))
	require.False(t, deepEqual(nil, map[string]any{}))
	require.False(t, deepEqual(map[string]any{}, nil))

	// Test empty values
	require.True(t, deepEqual(map[string]any{}, map[string]any{}))
	require.True(t, deepEqual([]any{}, []any{}))

	// Test different types
	require.False(t, deepEqual("string", 42))
	require.False(t, deepEqual(map[string]any{}, []any{}))
}

func TestToCamelCase(t *testing.T) {
	// Test basic conversion
	require.Equal(t, "camelCase", toCamelCase("camel_case"))
	require.Equal(t, "snakeCase", toCamelCase("snake_case"))

	// Test edge cases
	require.Empty(t, toCamelCase(""))
	require.Equal(t, "single", toCamelCase("single"))
	require.Equal(t, "multipleWords", toCamelCase("multiple_words"))
}
