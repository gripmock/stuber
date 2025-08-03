package stuber_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bavix/features"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/gripmock/stuber"
)

// V2 equivalents of V1 tests

func TestFindByNotFoundV2(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	s.PutMany(&stuber.Stub{ID: uuid.New(), Service: "Greeter1", Method: "SayHello1"})

	tests := []struct {
		service string
		method  string
		err     error
	}{
		{"hello", "SayHello1", stuber.ErrServiceNotFound},
		{"Greeter", "SayHello1", stuber.ErrServiceNotFound},
		{"Greeter1", "world", stuber.ErrMethodNotFound},
		{"helloworld.Greeter1", "world", stuber.ErrMethodNotFound},
		{"helloworld.v1.Greeter1", "world", stuber.ErrMethodNotFound},
		{"Greeter1", "SayHello1", nil},
		{"helloworld.Greeter1", "SayHello1", nil},
		{"helloworld.v1.Greeter1", "SayHello1", nil},
	}

	for _, tt := range tests {
		_, err := s.FindBy(tt.service, tt.method)
		require.ErrorIs(t, err, tt.err)
	}
}

func TestStubNilV2(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	require.Nil(t, s.FindByID(uuid.New()))
}

func TestFindByV2(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	require.Empty(t, s.All())

	s.PutMany(
		&stuber.Stub{ID: uuid.New(), Service: "Greeter1", Method: "SayHello1"},
		&stuber.Stub{ID: uuid.New(), Service: "Greeter1", Method: "SayHello1"},
		&stuber.Stub{ID: uuid.New(), Service: "Greeter2", Method: "SayHello2"},
		&stuber.Stub{ID: uuid.New(), Service: "Greeter3", Method: "SayHello2"},
		&stuber.Stub{ID: uuid.New(), Service: "Greeter4", Method: "SayHello3"},
		&stuber.Stub{ID: uuid.New(), Service: "Greeter5", Method: "SayHello3"},
		&stuber.Stub{ID: uuid.New(), Service: "Greeter1", Method: "SayHello3"},
	)

	require.Len(t, s.All(), 7)
}

func TestFindBySortedV2(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Verify that FindBy returns stubs sorted by priority in descending order
	stub1 := &stuber.Stub{ID: uuid.New(), Service: "Greeter1", Method: "SayHello1", Priority: 10}
	stub2 := &stuber.Stub{ID: uuid.New(), Service: "Greeter1", Method: "SayHello1", Priority: 30}
	stub3 := &stuber.Stub{ID: uuid.New(), Service: "Greeter1", Method: "SayHello1", Priority: 20}
	stub4 := &stuber.Stub{ID: uuid.New(), Service: "Greeter2", Method: "SayHello2", Priority: 50}

	s.PutMany(stub1, stub2, stub3, stub4)

	// Test sorted results
	results, err := s.FindBy("Greeter1", "SayHello1")
	require.NoError(t, err)
	require.Len(t, results, 3)

	// Should be sorted by priority in descending order: 30, 20, 10
	require.Equal(t, 30, results[0].Priority)
	require.Equal(t, 20, results[1].Priority)
	require.Equal(t, 10, results[2].Priority)

	// Test single result
	results, err = s.FindBy("Greeter2", "SayHello2")
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, 50, results[0].Priority)

	// Test not found
	_, err = s.FindBy("Greeter3", "SayHello3")
	require.ErrorIs(t, err, stuber.ErrServiceNotFound)
}

func TestPutMany_FixIDV2(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Verify that PutMany assigns IDs to stubs that are created without them
	stub1 := &stuber.Stub{Service: "Greeter1", Method: "SayHello1"}
	stub2 := &stuber.Stub{Service: "Greeter2", Method: "SayHello2"}

	// PutMany should assign IDs
	s.PutMany(stub1, stub2)

	// Check that IDs were assigned
	require.NotEqual(t, uuid.Nil, stub1.ID)
	require.NotEqual(t, uuid.Nil, stub2.ID)
	require.NotEqual(t, stub1.ID, stub2.ID)

	// Check that stubs are stored
	all := s.All()
	require.Len(t, all, 2)
}

func TestUpdateManyV2(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	stub1 := &stuber.Stub{ID: uuid.New(), Service: "Greeter1", Method: "SayHello1"}
	stub2 := &stuber.Stub{ID: uuid.New(), Service: "Greeter2", Method: "SayHello2"}

	s.PutMany(stub1, stub2)

	// Update stubs
	stub1Updated := &stuber.Stub{ID: stub1.ID, Service: "Greeter1Updated", Method: "SayHello1Updated"}
	stub2Updated := &stuber.Stub{ID: stub2.ID, Service: "Greeter2Updated", Method: "SayHello2Updated"}

	s.UpdateMany(stub1Updated, stub2Updated)

	// Check that stubs were updated
	all := s.All()
	require.Len(t, all, 2)

	// Find updated stubs
	found1 := s.FindByID(stub1.ID)
	require.NotNil(t, found1)
	require.Equal(t, "Greeter1Updated", found1.Service)
	require.Equal(t, "SayHello1Updated", found1.Method)

	found2 := s.FindByID(stub2.ID)
	require.NotNil(t, found2)
	require.Equal(t, "Greeter2Updated", found2.Service)
	require.Equal(t, "SayHello2Updated", found2.Method)
}

func TestRelationshipV2(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Create two independent stubs to verify that multiple stubs can coexist and be retrieved separately
	stub1 := &stuber.Stub{ID: uuid.New(), Service: "Greeter1", Method: "SayHello1"}
	stub2 := &stuber.Stub{ID: uuid.New(), Service: "Greeter2", Method: "SayHello2"}

	s.PutMany(stub1, stub2)

	// Test relationships
	require.Len(t, s.All(), 2)
	require.NotNil(t, s.FindByID(stub1.ID))
	require.NotNil(t, s.FindByID(stub2.ID))
}

func TestBudgerigar_UnusedV2(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Create stubs
	stub1 := &stuber.Stub{ID: uuid.New(), Service: "Greeter1", Method: "SayHello1"}
	stub2 := &stuber.Stub{ID: uuid.New(), Service: "Greeter2", Method: "SayHello2"}

	s.PutMany(stub1, stub2)

	// Initially all stubs are unused
	unused := s.Unused()
	require.Len(t, unused, 2)

	// Use one stub by finding it with QueryV2
	// First, update stub1 to have matching input
	stub1.Input = stuber.InputData{
		Equals: map[string]any{"key": "value"},
	}
	s.UpdateMany(stub1)

	query := stuber.QueryV2{
		Service: "Greeter1",
		Method:  "SayHello1",
		Input:   []map[string]any{{"key": "value"}},
	}

	result, err := s.FindByQueryV2(query)
	require.NoError(t, err)
	require.NotNil(t, result.Found())

	// Now only one stub should be unused (the original stub2)
	unused = s.Unused()
	require.Len(t, unused, 1)
	require.Equal(t, stub2.ID, unused[0].ID)
}

func TestBudgerigar_SearchWithHeadersV2(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Create stub with headers
	stub := &stuber.Stub{
		Service: "Greeter1",
		Method:  "SayHello1",
		Headers: stuber.InputHeader{
			Equals: map[string]any{"authorization": "Bearer token123"},
		},
		Input: stuber.InputData{
			Equals: map[string]any{"name": "John"},
		},
	}

	s.PutMany(stub)

	// Test matching query
	query := stuber.QueryV2{
		Service: "Greeter1",
		Method:  "SayHello1",
		Headers: map[string]any{"authorization": "Bearer token123"},
		Input:   []map[string]any{{"name": "John"}},
	}

	result, err := s.FindByQueryV2(query)
	require.NoError(t, err)
	require.NotNil(t, result.Found())

	// Test non-matching headers
	queryNonMatching := stuber.QueryV2{
		Service: "Greeter1",
		Method:  "SayHello1",
		Headers: map[string]any{"authorization": "Bearer different"},
		Input:   []map[string]any{{"name": "John"}},
	}

	result, err = s.FindByQueryV2(queryNonMatching)
	require.NoError(t, err) // Should find similar match
	require.Nil(t, result.Found())
	require.NotNil(t, result.Similar()) // Should find similar match
}

func TestBudgerigar_SearchEmptyV2(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Test search with empty service/method
	_, err := s.FindBy("", "")
	require.Error(t, err)

	// Test search with non-existent service/method
	_, err = s.FindBy("NonExistent", "NonExistent")
	require.Error(t, err)
}

func TestBudgerigar_SearchWithHeaders_SimilarV2(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Create stub with headers
	stub := &stuber.Stub{
		Service: "Greeter1",
		Method:  "SayHello1",
		Headers: stuber.InputHeader{
			Equals: map[string]any{"authorization": "Bearer token123"},
		},
		Input: stuber.InputData{
			Equals: map[string]any{"name": "John"},
		},
	}

	s.PutMany(stub)

	// Test similar match (different headers but same service/method)
	query := stuber.QueryV2{
		Service: "Greeter1",
		Method:  "SayHello1",
		Headers: map[string]any{"authorization": "Bearer different"},
		Input:   []map[string]any{{"name": "John"}},
	}

	result, err := s.FindByQueryV2(query)
	require.NoError(t, err) // Should find similar match
	require.Nil(t, result.Found())
	require.NotNil(t, result.Similar()) // Should find similar match
}

func TestResult_SimilarV2(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Create stub
	stub := &stuber.Stub{
		Service: "Greeter1",
		Method:  "SayHello1",
		Input: stuber.InputData{
			Equals: map[string]any{"name": "John"},
		},
	}

	s.PutMany(stub)

	// Test query that doesn't match exactly but is similar
	query := stuber.QueryV2{
		Service: "Greeter1",
		Method:  "SayHello1",
		Input:   []map[string]any{{"name": "Jane"}}, // Different name
	}

	result, err := s.FindByQueryV2(query)
	require.NoError(t, err)
	require.Nil(t, result.Found())
	require.NotNil(t, result.Similar())
}

func TestStuber_MatchesEqualsFoundV2(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Create stub with equals
	stub := &stuber.Stub{
		Service: "Greeter1",
		Method:  "SayHello1",
		Input: stuber.InputData{
			Equals: map[string]any{"name": "John", "age": 30},
		},
	}

	s.PutMany(stub)

	// Test exact match
	query := stuber.QueryV2{
		Service: "Greeter1",
		Method:  "SayHello1",
		Input:   []map[string]any{{"name": "John", "age": 30}},
	}

	result, err := s.FindByQueryV2(query)
	require.NoError(t, err)
	require.NotNil(t, result.Found())

	// Test partial match (should not match)
	queryPartial := stuber.QueryV2{
		Service: "Greeter1",
		Method:  "SayHello1",
		Input:   []map[string]any{{"name": "John"}}, // Missing age
	}

	result, err = s.FindByQueryV2(queryPartial)
	require.NoError(t, err) // Should find similar match
	require.Nil(t, result.Found())
	require.NotNil(t, result.Similar()) // Should find similar match
}

func TestDeleteV2(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Create stub
	stub := &stuber.Stub{ID: uuid.New(), Service: "Greeter1", Method: "SayHello1"}

	s.PutMany(stub)

	// Verify stub exists
	require.Len(t, s.All(), 1)

	// Delete stub
	s.DeleteByID(stub.ID)

	// Verify stub is deleted
	require.Empty(t, s.All())
	require.Nil(t, s.FindByID(stub.ID))
}

func TestBudgerigar_ClearV2(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Create stubs
	stub1 := &stuber.Stub{ID: uuid.New(), Service: "Greeter1", Method: "SayHello1"}
	stub2 := &stuber.Stub{ID: uuid.New(), Service: "Greeter2", Method: "SayHello2"}

	s.PutMany(stub1, stub2)

	// Verify stubs exist
	require.Len(t, s.All(), 2)

	// Clear all stubs
	s.Clear()

	// Verify all stubs are cleared
	require.Empty(t, s.All())
	require.Nil(t, s.FindByID(stub1.ID))
	require.Nil(t, s.FindByID(stub2.ID))
}

func TestBudgerigar_FindByQuery_FoundWithPriorityV2(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Create stubs with different priorities
	stub1 := &stuber.Stub{
		ID:       uuid.New(),
		Service:  "Greeter1",
		Method:   "SayHello1",
		Priority: 1,
		Input: stuber.InputData{
			Equals: map[string]any{"name": "John"},
		},
		Output: stuber.Output{
			Data: map[string]any{"message": "Hello from stub1"},
		},
	}

	stub2 := &stuber.Stub{
		ID:       uuid.New(),
		Service:  "Greeter1",
		Method:   "SayHello1",
		Priority: 2, // Higher priority
		Input: stuber.InputData{
			Equals: map[string]any{"name": "John"},
		},
		Output: stuber.Output{
			Data: map[string]any{"message": "Hello from stub2"},
		},
	}

	s.PutMany(stub1, stub2)

	// Test query
	query := stuber.QueryV2{
		Service: "Greeter1",
		Method:  "SayHello1",
		Input:   []map[string]any{{"name": "John"}},
	}

	result, err := s.FindByQueryV2(query)
	require.NoError(t, err)
	require.NotNil(t, result.Found())

	// Should match the higher priority stub
	require.Equal(t, "Hello from stub2", result.Found().Output.Data["message"])
}

func TestBudgerigar_UsedV2(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Create stubs
	stub1 := &stuber.Stub{ID: uuid.New(), Service: "Greeter1", Method: "SayHello1"}
	stub2 := &stuber.Stub{ID: uuid.New(), Service: "Greeter2", Method: "SayHello2"}

	s.PutMany(stub1, stub2)

	// Initially no stubs are used
	used := s.Used()
	require.Empty(t, used)

	// Use one stub
	query := stuber.QueryV2{
		Service: "Greeter1",
		Method:  "SayHello1",
		Input:   []map[string]any{{"key": "value"}},
	}

	// Update stub1 to have matching input
	stub1.Input = stuber.InputData{
		Equals: map[string]any{"key": "value"},
	}
	s.UpdateMany(stub1)
	result, err := s.FindByQueryV2(query)
	require.NoError(t, err)
	require.NotNil(t, result.Found())

	// Now one stub should be used
	used = s.Used()
	require.Len(t, used, 1)
}

func TestBudgerigar_FindByQuery_WithIDV2(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Create stub
	stub := &stuber.Stub{
		ID:      uuid.New(),
		Service: "Greeter1",
		Method:  "SayHello1",
		Input: stuber.InputData{
			Equals: map[string]any{"name": "John"},
		},
	}

	s.PutMany(stub)

	// Test query with ID
	query := stuber.QueryV2{
		ID:      &stub.ID,
		Service: "Greeter1",
		Method:  "SayHello1",
	}

	result, err := s.FindByQueryV2(query)
	require.NoError(t, err)
	require.NotNil(t, result.Found())
	require.Equal(t, stub.ID, result.Found().ID)
}

// Additional V2-specific tests

func TestNewQueryV2(t *testing.T) {
	// Test creating QueryV2 from HTTP request
	jsonBody := `{"service":"test","method":"test","input":[{"key":"value"}]}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	query, err := stuber.NewQueryV2(req)
	require.NoError(t, err)
	require.Equal(t, "test", query.Service)
	require.Equal(t, "test", query.Method)
	require.Len(t, query.Input, 1)
	require.Equal(t, "value", query.Input[0]["key"])
}

// TestV2OptimizerIntegration and TestV2OptimizerHeapOperations removed
// as they tested internal StreamMatcher which is no longer exported

func TestV2StubMethods(t *testing.T) {
	// Test stub methods
	stub := &stuber.Stub{
		Service: "test",
		Method:  "test",
		Input: stuber.InputData{
			Equals:   map[string]any{"key1": "value1"},
			Contains: map[string]any{"key2": "value2"},
			Matches:  map[string]any{"key3": "value3"},
		},
		Headers: stuber.InputHeader{
			Equals:   map[string]any{"header1": "value1"},
			Contains: map[string]any{"header2": "value2"},
			Matches:  map[string]any{"header3": "value3"},
		},
	}

	// Test GetEquals, GetContains, GetMatches for Input
	require.Equal(t, map[string]any{"key1": "value1"}, stub.Input.GetEquals())
	require.Equal(t, map[string]any{"key2": "value2"}, stub.Input.GetContains())
	require.Equal(t, map[string]any{"key3": "value3"}, stub.Input.GetMatches())

	// Test GetEquals, GetContains, GetMatches for Headers
	require.Equal(t, map[string]any{"header1": "value1"}, stub.Headers.GetEquals())
	require.Equal(t, map[string]any{"header2": "value2"}, stub.Headers.GetContains())
	require.Equal(t, map[string]any{"header3": "value3"}, stub.Headers.GetMatches())

	// Test Len for Headers
	require.Equal(t, 3, stub.Headers.Len())
}

func TestV2QueryFunctions(t *testing.T) {
	// Test NewQuery and RequestInternal for V1 Query
	jsonBody := `{"service":"test","method":"test","data":{"key":"value"}}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	query, err := stuber.NewQuery(req)
	require.NoError(t, err)
	require.Equal(t, "test", query.Service)
	require.Equal(t, "test", query.Method)
	require.Equal(t, "value", query.Data["key"])
	require.False(t, query.RequestInternal())
}

func TestV2SearcherFunctions(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Test find function (V1)
	stub := &stuber.Stub{
		Service: "test",
		Method:  "test",
		Input: stuber.InputData{
			Equals: map[string]any{"key": "value"},
		},
	}

	s.PutMany(stub)

	query := stuber.Query{
		Service: "test",
		Method:  "test",
		Data:    map[string]any{"key": "value"},
	}

	result, err := s.FindByQuery(query)
	require.NoError(t, err)
	require.NotNil(t, result.Found())

	// Test searchByID function
	queryWithID := stuber.Query{
		ID:      &stub.ID,
		Service: "test",
		Method:  "test",
	}

	result, err = s.FindByQuery(queryWithID)
	require.NoError(t, err)
	require.NotNil(t, result.Found())
}

func TestV2StorageFunctions(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Test storage functions through searcher
	stub1 := &stuber.Stub{ID: uuid.New(), Service: "test1", Method: "test1"}
	stub2 := &stuber.Stub{ID: uuid.New(), Service: "test2", Method: "test2"}

	s.PutMany(stub1, stub2)

	// Test values function
	all := s.All()
	require.Len(t, all, 2)

	// Test findByID function
	found := s.FindByID(stub1.ID)
	require.NotNil(t, found)
	require.Equal(t, stub1.ID, found.ID)

	// Test delete function
	s.DeleteByID(stub1.ID)
	all = s.All()
	require.Len(t, all, 1)
	require.Equal(t, stub2.ID, all[0].ID)

	// Test clear function
	s.Clear()
	all = s.All()
	require.Empty(t, all)
}

func TestV2MatcherFunctions(t *testing.T) {
	// Test V2 matcher functions through the public API
	s := stuber.NewBudgerigar(features.New())

	stub := &stuber.Stub{
		ID:      uuid.New(),
		Service: "TestService",
		Method:  "TestMethod",
		Headers: stuber.InputHeader{
			Equals: map[string]any{"content-type": "application/json"},
		},
		Input: stuber.InputData{
			Equals: map[string]any{"key": "value"},
		},
	}

	s.PutMany(stub)

	query := stuber.QueryV2{
		Service: "TestService",
		Method:  "TestMethod",
		Headers: map[string]any{"content-type": "application/json"},
		Input:   []map[string]any{{"key": "value"}},
	}

	// Test matching through public API
	result, err := s.FindByQueryV2(query)
	require.NoError(t, err)
	require.NotNil(t, result.Found())
	require.Equal(t, stub.ID, result.Found().ID)
}

// TestBidiStreaming tests bidirectional streaming functionality.
//
//nolint:funlen
func TestBidiStreaming(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Create bidirectional stubs for bidirectional streaming
	// Each stub has Stream data for input matching
	bidiStub1 := &stuber.Stub{
		ID:      uuid.New(),
		Service: "ChatService",
		Method:  "Chat",
		Headers: stuber.InputHeader{
			Equals: map[string]any{"content-type": "application/json"},
		},
		Stream: []stuber.InputData{
			{Equals: map[string]any{"message": "hello"}},
		},
		Output: stuber.Output{
			Stream: []any{
				map[string]any{"message": "Hello! How can I help you?"},
				map[string]any{"message": "I'm doing well, thank you!"},
				map[string]any{"message": "Have a great day!"},
			},
		},
	}

	bidiStub2 := &stuber.Stub{
		ID:      uuid.New(),
		Service: "ChatService",
		Method:  "Chat",
		Headers: stuber.InputHeader{
			Equals: map[string]any{"content-type": "application/json"},
		},
		Stream: []stuber.InputData{
			{Equals: map[string]any{"message": "how are you"}},
		},
		Output: stuber.Output{
			Data: map[string]any{"response": "I'm doing great!"},
		},
	}

	bidiStub3 := &stuber.Stub{
		ID:      uuid.New(),
		Service: "ChatService",
		Method:  "Chat",
		Headers: stuber.InputHeader{
			Equals: map[string]any{"content-type": "application/json"},
		},
		Stream: []stuber.InputData{
			{Equals: map[string]any{"message": "goodbye"}},
		},
		Output: stuber.Output{
			Data: map[string]any{"response": "Goodbye! See you later!"},
		},
	}

	s.PutMany(bidiStub1, bidiStub2, bidiStub3)

	// Test bidirectional streaming
	t.Run("BidiStreamingWithUnaryStubs", func(t *testing.T) {
		query := stuber.QueryBidi{
			Service: "ChatService",
			Method:  "Chat",
			Headers: map[string]any{"content-type": "application/json"},
		}

		result, err := s.FindByQueryBidi(query)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Test first message - should match bidiStub1 (server streaming)
		stub, err := result.Next(map[string]any{"message": "hello"})
		require.NoError(t, err)
		require.Equal(t, bidiStub1.ID, stub.ID)
		require.Len(t, stub.Output.Stream, 3)
		require.Empty(t, stub.Output.Data)

		// Test second message - should match bidiStub2 (unary)
		stub, err = result.Next(map[string]any{"message": "how are you"})
		require.NoError(t, err)
		require.Equal(t, bidiStub2.ID, stub.ID)
		require.Equal(t, "I'm doing great!", stub.Output.Data["response"])
		require.Empty(t, stub.Output.Stream)

		// Test third message - should match bidiStub3 (unary)
		stub, err = result.Next(map[string]any{"message": "goodbye"})
		require.NoError(t, err)
		require.Equal(t, bidiStub3.ID, stub.ID)
		require.Equal(t, "Goodbye! See you later!", stub.Output.Data["response"])
		require.Empty(t, stub.Output.Stream)

		// Test message that doesn't match any stub - should return error
		_, err = result.Next(map[string]any{"message": "unknown"})
		require.Error(t, err)
		require.ErrorIs(t, err, stuber.ErrStubNotFound)
	})
}

// TestBidiStreamingFallback tests fallback behavior when no stubs are available.
func TestBidiStreamingFallback(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Add a stub for the same service but different method to ensure the service exists
	otherStub := &stuber.Stub{
		ID:      uuid.New(),
		Service: "ChatService",
		Method:  "Chat",
		Input: stuber.InputData{
			Equals: map[string]any{"message": "other"},
		},
		Output: stuber.Output{
			Data: map[string]any{"response": "other"},
		},
	}
	s.PutMany(otherStub)

	// Query for the same service and method
	query := stuber.QueryBidi{
		Service: "ChatService",
		Method:  "Chat",
		Headers: map[string]any{"content-type": "application/json"},
	}

	result, err := s.FindByQueryBidi(query)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Test message that doesn't match any stub
	_, err = result.Next(map[string]any{"message": "hello"})
	require.Error(t, err)
	require.ErrorIs(t, err, stuber.ErrStubNotFound)
}

// TestBidiStreamingWithID tests bidirectional streaming with ID-based queries.
func TestBidiStreamingWithID(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	unaryStub := &stuber.Stub{
		ID:      uuid.New(),
		Service: "ChatService",
		Method:  "Chat",
		Headers: stuber.InputHeader{
			Equals: map[string]any{"content-type": "application/json"},
		},
		Input: stuber.InputData{
			Equals: map[string]any{"message": "hello"},
		},
		Output: stuber.Output{
			Data: map[string]any{"response": "Hello!"},
		},
	}

	s.PutMany(unaryStub)

	query := stuber.QueryBidi{
		ID:      &unaryStub.ID,
		Service: "ChatService",
		Method:  "Chat",
		Headers: map[string]any{"content-type": "application/json"},
	}

	result, err := s.FindByQueryBidi(query)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Test that we can still use the result even with ID-based query
	stub, err := result.Next(map[string]any{"message": "hello"})
	require.NoError(t, err)
	require.Equal(t, unaryStub.ID, stub.ID)
}

// TestBidiStreamingEmptyService tests bidirectional streaming with empty service/method.
func TestBidiStreamingEmptyService(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	query := stuber.QueryBidi{
		Service: "NonExistentService",
		Method:  "NonExistentMethod",
		Headers: map[string]any{"content-type": "application/json"},
	}

	_, err := s.FindByQueryBidi(query)
	require.Error(t, err)
	require.ErrorIs(t, err, stuber.ErrServiceNotFound)
}

// TestBidiStreamingWithServerStream tests bidirectional streaming with server streaming responses.
func TestBidiStreamingWithServerStream(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Create a stub that can handle bidirectional streaming (unary input + server stream output)
	bidiStub := &stuber.Stub{
		ID:      uuid.New(),
		Service: "ChatService",
		Method:  "Chat",
		Headers: stuber.InputHeader{
			Equals: map[string]any{"content-type": "application/json"},
		},
		Input: stuber.InputData{
			Equals: map[string]any{"message": "hello"},
		},
		Output: stuber.Output{
			Stream: []any{
				map[string]any{"message": "Hello! How can I help you?"},
				map[string]any{"message": "I'm doing well, thank you!"},
				map[string]any{"message": "Have a great day!"},
			},
		},
	}

	s.PutMany(bidiStub)

	query := stuber.QueryBidi{
		Service: "ChatService",
		Method:  "Chat",
		Headers: map[string]any{"content-type": "application/json"},
	}

	result, err := s.FindByQueryBidi(query)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Test that the stub is correctly identified as server streaming but not bidirectional
	require.True(t, bidiStub.IsServerStream())
	require.False(t, bidiStub.IsBidirectional()) // This stub has Input (unary) + Output.Stream (server streaming), not bidirectional

	// Test message matching
	stub, err := result.Next(map[string]any{"message": "hello"})
	require.NoError(t, err)
	require.Equal(t, bidiStub.ID, stub.ID)
	require.Len(t, stub.Output.Stream, 3)
	require.Empty(t, stub.Output.Data)
}

// TestBidiStreamingStatefulLogic tests the stateful logic of bidirectional streaming
// where stubs are filtered based on incoming messages.
//
//nolint:funlen
func TestBidiStreamingStatefulLogic(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Create multiple stubs with different patterns
	// All start with "hello" but diverge after that
	stub1 := &stuber.Stub{
		ID:      uuid.New(),
		Service: "ChatService",
		Method:  "Chat",
		Headers: stuber.InputHeader{
			Equals: map[string]any{"content-type": "application/json"},
		},
		Stream: []stuber.InputData{
			{Equals: map[string]any{"message": "hello"}},
			{Equals: map[string]any{"message": "world"}},
			{Equals: map[string]any{"message": "goodbye"}},
		},
		Output: stuber.Output{
			Data: map[string]any{"response": "Pattern 1 completed"},
		},
	}

	stub2 := &stuber.Stub{
		ID:      uuid.New(),
		Service: "ChatService",
		Method:  "Chat",
		Headers: stuber.InputHeader{
			Equals: map[string]any{"content-type": "application/json"},
		},
		Stream: []stuber.InputData{
			{Equals: map[string]any{"message": "hello"}},
			{Equals: map[string]any{"message": "universe"}},
			{Equals: map[string]any{"message": "farewell"}},
		},
		Output: stuber.Output{
			Data: map[string]any{"response": "Pattern 2 completed"},
		},
	}

	stub3 := &stuber.Stub{
		ID:      uuid.New(),
		Service: "ChatService",
		Method:  "Chat",
		Headers: stuber.InputHeader{
			Equals: map[string]any{"content-type": "application/json"},
		},
		Stream: []stuber.InputData{
			{Equals: map[string]any{"message": "hello"}},
			{Equals: map[string]any{"message": "galaxy"}},
			{Equals: map[string]any{"message": "adios"}},
		},
		Output: stuber.Output{
			Data: map[string]any{"response": "Pattern 3 completed"},
		},
	}

	s.PutMany(stub1, stub2, stub3)

	query := stuber.QueryBidi{
		Service: "ChatService",
		Method:  "Chat",
		Headers: map[string]any{"content-type": "application/json"},
	}

	result, err := s.FindByQueryBidi(query)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Test first message - all stubs should match "hello"
	stub, err := result.Next(map[string]any{"message": "hello"})
	require.NoError(t, err)
	// Should return one of the matching stubs (could be any of the three)
	require.Contains(t, []uuid.UUID{stub1.ID, stub2.ID, stub3.ID}, stub.ID)

	// Test second message - should filter based on the pattern
	// If we send "world", only stub1 should match
	stub, err = result.Next(map[string]any{"message": "world"})
	require.NoError(t, err)
	require.Equal(t, stub1.ID, stub.ID)

	// Test third message - should continue with stub1 pattern
	stub, err = result.Next(map[string]any{"message": "goodbye"})
	require.NoError(t, err)
	require.Equal(t, stub1.ID, stub.ID)

	// Test that we get the expected response
	require.Equal(t, "Pattern 1 completed", stub.Output.Data["response"])
}

// TestBidiStreamingStatefulLogicDifferentPattern tests with a different pattern.
func TestBidiStreamingStatefulLogicDifferentPattern(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Create multiple stubs with different patterns
	stub1 := &stuber.Stub{
		ID:      uuid.New(),
		Service: "ChatService",
		Method:  "Chat",
		Headers: stuber.InputHeader{
			Equals: map[string]any{"content-type": "application/json"},
		},
		Stream: []stuber.InputData{
			{Equals: map[string]any{"message": "hello"}},
			{Equals: map[string]any{"message": "world"}},
		},
		Output: stuber.Output{
			Data: map[string]any{"response": "Pattern 1"},
		},
	}

	stub2 := &stuber.Stub{
		ID:      uuid.New(),
		Service: "ChatService",
		Method:  "Chat",
		Headers: stuber.InputHeader{
			Equals: map[string]any{"content-type": "application/json"},
		},
		Stream: []stuber.InputData{
			{Equals: map[string]any{"message": "hello"}},
			{Equals: map[string]any{"message": "universe"}},
		},
		Output: stuber.Output{
			Data: map[string]any{"response": "Pattern 2"},
		},
	}

	s.PutMany(stub1, stub2)

	query := stuber.QueryBidi{
		Service: "ChatService",
		Method:  "Chat",
		Headers: map[string]any{"content-type": "application/json"},
	}

	result, err := s.FindByQueryBidi(query)
	require.NoError(t, err)
	require.NotNil(t, result)

	// First message - both stubs match
	stub, err := result.Next(map[string]any{"message": "hello"})
	require.NoError(t, err)
	require.Contains(t, []uuid.UUID{stub1.ID, stub2.ID}, stub.ID)

	// Second message - if we send "universe", only stub2 should match
	stub, err = result.Next(map[string]any{"message": "universe"})
	require.NoError(t, err)
	require.Equal(t, stub2.ID, stub.ID)
	require.Equal(t, "Pattern 2", stub.Output.Data["response"])
}

// TestBidiStreamingStatefulLogicNoMatch tests when no stubs match the pattern.
func TestBidiStreamingStatefulLogicNoMatch(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Create a stub with a specific pattern
	stub := &stuber.Stub{
		ID:      uuid.New(),
		Service: "ChatService",
		Method:  "Chat",
		Headers: stuber.InputHeader{
			Equals: map[string]any{"content-type": "application/json"},
		},
		Stream: []stuber.InputData{
			{Equals: map[string]any{"message": "hello"}},
			{Equals: map[string]any{"message": "world"}},
		},
		Output: stuber.Output{
			Data: map[string]any{"response": "Pattern completed"},
		},
	}

	s.PutMany(stub)

	query := stuber.QueryBidi{
		Service: "ChatService",
		Method:  "Chat",
		Headers: map[string]any{"content-type": "application/json"},
	}

	result, err := s.FindByQueryBidi(query)
	require.NoError(t, err)
	require.NotNil(t, result)

	// First message - should match
	stub, err = result.Next(map[string]any{"message": "hello"})
	require.NoError(t, err)
	require.NotNil(t, stub)

	// Second message - should not match (sending "unknown" instead of "world")
	_, err = result.Next(map[string]any{"message": "unknown"})
	require.Error(t, err)
	require.ErrorIs(t, err, stuber.ErrStubNotFound)
}

// TestBidiStreamingEdgeCases tests edge cases for bidirectional streaming.
func TestBidiStreamingEdgeCases(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Add a stub for testing
	stub := &stuber.Stub{
		ID:      uuid.New(),
		Service: "ChatService",
		Method:  "Chat",
		Headers: stuber.InputHeader{
			Equals: map[string]any{"content-type": "application/json"},
		},
		Input: stuber.InputData{
			Equals: map[string]any{"message": "hello"},
		},
		Output: stuber.Output{
			Data: map[string]any{"response": "Hello!"},
		},
	}

	s.PutMany(stub)

	// Test with valid query first
	query := stuber.QueryBidi{
		Service: "ChatService",
		Method:  "Chat",
		Headers: map[string]any{"content-type": "application/json"},
	}

	result, err := s.FindByQueryBidi(query)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Test with nil messageData
	_, err = result.Next(nil)
	require.Error(t, err)
	require.ErrorIs(t, err, stuber.ErrStubNotFound)

	// Test with empty messageData
	_, err = result.Next(map[string]any{})
	require.Error(t, err)
	require.ErrorIs(t, err, stuber.ErrStubNotFound)

	// Test with valid messageData
	stubResult, err := result.Next(map[string]any{"message": "hello"})
	require.NoError(t, err)
	require.NotNil(t, stubResult)
	require.Equal(t, "Hello!", stubResult.Output.Data["response"])
}

// TestFieldNameVariations tests the field name variation matching.
func TestFieldNameVariations(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Test stub with snake_case field
	stub := &stuber.Stub{
		ID:      uuid.New(),
		Service: "TestService",
		Method:  "Test",
		Input: stuber.InputData{
			Equals: map[string]any{"user_name": "john"},
		},
		Output: stuber.Output{
			Data: map[string]any{"response": "Hello John!"},
		},
	}

	s.PutMany(stub)

	query := stuber.QueryBidi{
		Service: "TestService",
		Method:  "Test",
	}

	result, err := s.FindByQueryBidi(query)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Test with camelCase field (should match snake_case)
	stubResult, err := result.Next(map[string]any{"userName": "john"})
	require.NoError(t, err)
	require.NotNil(t, stubResult)
	require.Equal(t, "Hello John!", stubResult.Output.Data["response"])

	// Test with exact snake_case field
	stubResult2, err := result.Next(map[string]any{"user_name": "john"})
	require.NoError(t, err)
	require.NotNil(t, stubResult2)
	require.Equal(t, "Hello John!", stubResult2.Output.Data["response"])
}

// TestCamelCaseVariations tests camelCase field matching.
func TestCamelCaseVariations(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Test stub with camelCase field
	stub := &stuber.Stub{
		ID:      uuid.New(),
		Service: "TestService",
		Method:  "Test",
		Input: stuber.InputData{
			Equals: map[string]any{"userName": "john"},
		},
		Output: stuber.Output{
			Data: map[string]any{"response": "Hello John!"},
		},
	}

	s.PutMany(stub)

	query := stuber.QueryBidi{
		Service: "TestService",
		Method:  "Test",
	}

	result, err := s.FindByQueryBidi(query)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Test with snake_case field (should match camelCase)
	stubResult, err := result.Next(map[string]any{"user_name": "john"})
	require.NoError(t, err)
	require.NotNil(t, stubResult)
	require.Equal(t, "Hello John!", stubResult.Output.Data["response"])

	// Test with exact camelCase field
	stubResult2, err := result.Next(map[string]any{"userName": "john"})
	require.NoError(t, err)
	require.NotNil(t, stubResult2)
	require.Equal(t, "Hello John!", stubResult2.Output.Data["response"])
}

// TestComplexFieldVariations tests complex field name variations.
func TestComplexFieldVariations(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Test stub with complex field names
	stub := &stuber.Stub{
		ID:      uuid.New(),
		Service: "TestService",
		Method:  "Test",
		Input: stuber.InputData{
			Equals: map[string]any{
				"user_profile_data": "data",
				"apiKey":            "key123",
				"simple_field":      "value",
			},
		},
		Output: stuber.Output{
			Data: map[string]any{"response": "Success!"},
		},
	}

	s.PutMany(stub)

	query := stuber.QueryBidi{
		Service: "TestService",
		Method:  "Test",
	}

	result, err := s.FindByQueryBidi(query)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Test with camelCase variations
	messageData := map[string]any{
		"userProfileData": "data",   // should match user_profile_data
		"api_key":         "key123", // should match apiKey
		"simpleField":     "value",  // should match simple_field
	}

	stubResult, err := result.Next(messageData)
	require.NoError(t, err)
	require.NotNil(t, stubResult)
	require.Equal(t, "Success!", stubResult.Output.Data["response"])
}

// TestEmptyFieldVariations tests edge cases with empty fields.
func TestEmptyFieldVariations(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	stub := &stuber.Stub{
		ID:      uuid.New(),
		Service: "TestService",
		Method:  "Test",
		Input: stuber.InputData{
			Equals: map[string]any{"": "empty_key"},
		},
		Output: stuber.Output{
			Data: map[string]any{"response": "Empty key!"},
		},
	}

	s.PutMany(stub)

	query := stuber.QueryBidi{
		Service: "TestService",
		Method:  "Test",
	}

	result, err := s.FindByQueryBidi(query)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Test with empty key
	stubResult, err := result.Next(map[string]any{"": "empty_key"})
	require.NoError(t, err)
	require.NotNil(t, stubResult)
	require.Equal(t, "Empty key!", stubResult.Output.Data["response"])
}

// TestStableSortingOptimized tests that results are stable across multiple runs with optimized sorting.
//
//nolint:funlen
func TestStableSortingOptimized(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Create multiple stubs with same priority but different IDs
	stub1 := &stuber.Stub{
		ID:       uuid.New(),
		Service:  "TestService",
		Method:   "Test",
		Priority: 1,
		Input: stuber.InputData{
			Equals: map[string]any{"field": "value"},
		},
		Output: stuber.Output{
			Data: map[string]any{"response": "Stub1"},
		},
	}

	stub2 := &stuber.Stub{
		ID:       uuid.New(),
		Service:  "TestService",
		Method:   "Test",
		Priority: 1, // Same priority
		Input: stuber.InputData{
			Equals: map[string]any{"field": "value"},
		},
		Output: stuber.Output{
			Data: map[string]any{"response": "Stub2"},
		},
	}

	stub3 := &stuber.Stub{
		ID:       uuid.New(),
		Service:  "TestService",
		Method:   "Test",
		Priority: 1, // Same priority
		Input: stuber.InputData{
			Equals: map[string]any{"field": "value"},
		},
		Output: stuber.Output{
			Data: map[string]any{"response": "Stub3"},
		},
	}

	s.PutMany(stub1, stub2, stub3)

	query := stuber.QueryBidi{
		Service: "TestService",
		Method:  "Test",
	}

	// Run multiple times to ensure stable results
	var firstResult *stuber.Stub

	for range 10 {
		result, err := s.FindByQueryBidi(query)
		require.NoError(t, err)
		require.NotNil(t, result)

		stubResult, err := result.Next(map[string]any{"field": "value"})
		require.NoError(t, err)
		require.NotNil(t, stubResult)

		if firstResult == nil {
			firstResult = stubResult
		} else {
			// Should always return the same stub due to stable sorting
			require.Equal(t, firstResult.ID, stubResult.ID, "Stable sorting failed")
		}
	}
}
