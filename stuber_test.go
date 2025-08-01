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

func TestFindByNotFound(t *testing.T) {
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

func TestStubNil(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	require.Nil(t, s.FindByID(uuid.New()))
}

func TestFindBy(t *testing.T) {
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

func TestFindBySorted(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Create stubs with different priorities
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

func TestPutMany_FixID(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	require.Empty(t, s.All())

	stubs := []*stuber.Stub{
		{Service: "Greeter1", Method: "SayHello1"},
		{Service: "Greeter1", Method: "SayHello1"},
	}

	s.PutMany(stubs...)

	require.Len(t, s.All(), 2)
	require.NotEqual(t, uuid.Nil, stubs[0].Key())
	require.NotEqual(t, uuid.Nil, stubs[1].Key())
}

func TestUpdateMany(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	require.Empty(t, s.All())

	stubs := []*stuber.Stub{
		{Service: "Greeter1", Method: "SayHello1", ID: uuid.New()},
		{Service: "Greeter1", Method: "SayHello1"},
		{Service: "Greeter1", Method: "SayHello1"},
	}

	s.UpdateMany(stubs...)

	require.Len(t, s.All(), 1)
}

func TestRelationship(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	s.PutMany(
		&stuber.Stub{ID: uuid.New(), Service: "Greeter1", Method: "SayHello1"},
		&stuber.Stub{ID: uuid.New(), Service: "Greeter2", Method: "SayHello2"},
	)

	_, err := s.FindBy("Greeter1", "SayHello2")
	require.ErrorIs(t, err, stuber.ErrMethodNotFound)
}

func TestBudgerigar_Unused(t *testing.T) {
	s := stuber.NewBudgerigar(features.New(stuber.MethodTitle))

	require.Empty(t, s.Unused())

	s.PutMany(
		&stuber.Stub{
			ID:      uuid.New(),
			Service: "Greeter1",
			Method:  "SayHello1",
			Input: stuber.InputData{Contains: map[string]any{
				"field1": "hello field1",
			}},
			Output: stuber.Output{Data: map[string]any{"message": "hello world"}},
		},
		&stuber.Stub{
			ID:      uuid.New(),
			Service: "Greeter2",
			Method:  "SayHello1",
			Input: stuber.InputData{Contains: map[string]any{
				"field1": "hello field1",
			}},
			Output: stuber.Output{Data: map[string]any{"message": "greeter2"}},
		},
		&stuber.Stub{
			ID:      uuid.New(),
			Service: "Greeter1",
			Method:  "SayHello1",
			Input: stuber.InputData{Contains: map[string]any{
				"field1": "hello field2",
			}},
			Output: stuber.Output{Data: map[string]any{"message": "say hello world"}},
		},
	)

	require.Len(t, s.Unused(), 3)

	payload := `{"service":"Greeter1","method":"SayHello1","data":{"field1":"hello field1", "field2":"hello world"}}`

	req := httptest.NewRequest(http.MethodPost, "/api/stubs/search", bytes.NewReader([]byte(payload)))
	q, err := stuber.NewQuery(req)
	require.NoError(t, err)

	r, err := s.FindByQuery(q)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Nil(t, r.Similar())
	require.NotNil(t, r.Found())

	require.Equal(t, map[string]any{"message": "hello world"}, r.Found().Output.Data)
}

func TestBudgerigar_SearchWithHeaders(t *testing.T) {
	s := stuber.NewBudgerigar(features.New(stuber.MethodTitle))

	require.Empty(t, s.Unused())

	s.PutMany(
		&stuber.Stub{
			ID:      uuid.New(),
			Service: "Gripmock",
			Method:  "SayHello",
			Input: stuber.InputData{Equals: map[string]any{
				"name": "simple3",
			}},
			Output: stuber.Output{Data: map[string]any{
				"message": "Hello Simple3",
			}},
		},
		&stuber.Stub{
			ID:      uuid.New(),
			Service: "Gripmock",
			Method:  "SayHello",
			Headers: stuber.InputHeader{Equals: map[string]any{
				"authorization": "Basic dXNlcjp1c2Vy",
			}},
			Input: stuber.InputData{Equals: map[string]any{
				"name": "simple3",
			}},
			Output: stuber.Output{Data: map[string]any{
				"message":     "Hello Simple3",
				"return_code": 3,
			}},
		},
	)

	require.Len(t, s.Unused(), 2)

	payload := `{"service":"Gripmock","method":"SayHello",
		"headers": {"authorization": "Basic dXNlcjp1c2Vy"}, 
		"data":{"name":"simple3"}}`

	req := httptest.NewRequest(http.MethodPost, "/api/stubs/search", bytes.NewReader([]byte(payload)))
	q, err := stuber.NewQuery(req)
	require.NoError(t, err)

	r, err := s.FindByQuery(q)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.NotNil(t, r.Found())
	require.Nil(t, r.Similar())

	require.Equal(t, map[string]any{
		"message":     "Hello Simple3",
		"return_code": 3,
	}, r.Found().Output.Data)
}

//nolint:funlen
func TestBudgerigar_SearchWithPackageAndWithoutPackage(t *testing.T) {
	s := stuber.NewBudgerigar(features.New(stuber.MethodTitle))

	require.Empty(t, s.Unused())

	stubs := []*stuber.Stub{
		{
			ID:      uuid.New(),
			Service: "helloworld.v1.Gripmock",
			Method:  "SayHello",
			Input: stuber.InputData{Equals: map[string]any{
				"name": "simple3",
			}},
			Output: stuber.Output{Data: map[string]any{
				"message": "Hello Simple3. Package helloworld.v1",
			}},
		},
		{
			ID:      uuid.New(),
			Service: "Gripmock",
			Method:  "SayHello",
			Input: stuber.InputData{Equals: map[string]any{
				"name": "simple4",
			}},
			Output: stuber.Output{Data: map[string]any{
				"message": "Hello Simple4",
			}},
		},
		{
			ID:      uuid.New(),
			Service: "Gripmock",
			Method:  "SayHello",
			Input: stuber.InputData{Equals: map[string]any{
				"name": "simple3",
			}},
			Output: stuber.Output{Data: map[string]any{
				"message": "Hello Simple3",
			}},
		},
	}

	s.PutMany(stubs...)

	require.Len(t, s.Unused(), len(stubs))

	cases := []struct {
		payload string
		id      uuid.UUID
	}{
		{
			payload: `{"data":{"name":"simple3"},"method":"SayHello","service":"helloworld.v1.Gripmock"}`,
			id:      stubs[0].ID,
		},
		{
			payload: `{"data":{"name":"simple3"},"method":"SayHello","service":"Gripmock"}`,
			id:      stubs[2].ID,
		},
		{
			payload: `{"data":{"name":"simple4"},"method":"SayHello","service":"helloworld.v1.Gripmock"}`,
			id:      stubs[1].ID,
		},
	}

	for _, c := range cases {
		req := httptest.NewRequest(http.MethodPost, "/api/stubs/search", bytes.NewReader([]byte(c.payload)))
		q, err := stuber.NewQuery(req)
		require.NoError(t, err)

		r, err := s.FindByQuery(q)
		require.NoError(t, err)
		require.NotNil(t, r)
		require.NotNil(t, r.Found())
		require.Equal(t, c.id, r.Found().ID)
		require.Nil(t, r.Similar())
	}

	checkItems := func(service string, expectedCount int) {
		items, err := s.FindBy(service, "SayHello")
		require.NoError(t, err)
		require.Len(t, items, expectedCount)
	}

	checkItems("helloworld.v1.Gripmock", 3)
	checkItems("Gripmock", 2)
}

func TestBudgerigar_SearchEmpty(t *testing.T) {
	s := stuber.NewBudgerigar(features.New(stuber.MethodTitle))

	require.Empty(t, s.Unused())

	s.PutMany(
		&stuber.Stub{
			ID:      uuid.New(),
			Service: "Gripmock",
			Method:  "ApiInfo",
			Input:   stuber.InputData{Equals: map[string]any{}},
			Output: stuber.Output{Data: map[string]any{
				"name":    "Gripmock",
				"version": "1.0",
			}},
		},
	)

	require.Len(t, s.Unused(), 1)

	payload := `{"data":{},"method":"ApiInfo","service":"Gripmock"}`

	req := httptest.NewRequest(http.MethodPost, "/api/stubs/search", bytes.NewReader([]byte(payload)))
	q, err := stuber.NewQuery(req)
	require.NoError(t, err)

	r, err := s.FindByQuery(q)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.NotNil(t, r.Found())
	require.Nil(t, r.Similar())

	require.Equal(t, map[string]any{
		"name":    "Gripmock",
		"version": "1.0",
	}, r.Found().Output.Data)
}

func TestBudgerigar_SearchWithHeaders_Similar(t *testing.T) {
	s := stuber.NewBudgerigar(features.New(stuber.MethodTitle))

	require.Empty(t, s.Unused())

	s.PutMany(
		&stuber.Stub{
			ID:      uuid.New(),
			Service: "Gripmock",
			Method:  "SayHello",
			Input: stuber.InputData{Equals: map[string]any{
				"name": "simple3",
			}},
			Output: stuber.Output{Data: map[string]any{
				"message":     "Hello Simple3",
				"return_code": 3,
			}},
		},
		&stuber.Stub{
			ID:      uuid.New(),
			Service: "Gripmock",
			Method:  "SayHello",
			Headers: stuber.InputHeader{Equals: map[string]any{
				"authorization": "Basic dXNlcjp1c2Vy",
			}},
			Input: stuber.InputData{Equals: map[string]any{
				"name": "simple3",
			}},
			Output: stuber.Output{Data: map[string]any{
				"message":     "Hello Simple3",
				"return_code": 3,
			}},
		},
	)

	require.Len(t, s.Unused(), 2)

	payload := `{"service":"Gripmock","method":"SayHello",
		"headers": {"authorization": "Basic dXNlcjp1c2Vy"}, 
		"data":{"name":"simple2"}}`

	req := httptest.NewRequest(http.MethodPost, "/api/stubs/search", bytes.NewReader([]byte(payload)))
	q, err := stuber.NewQuery(req)
	require.NoError(t, err)

	r, err := s.FindByQuery(q)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.NotNil(t, r.Similar())
	require.Nil(t, r.Found())

	require.Equal(t, map[string]any{
		"message":     "Hello Simple3",
		"return_code": 3,
	}, r.Similar().Output.Data)
}

func TestResult_MatchesRegexInt(t *testing.T) {
	s := stuber.NewBudgerigar(features.New(stuber.MethodTitle))

	require.Empty(t, s.Unused())

	s.PutMany(
		&stuber.Stub{
			ID:      uuid.New(),
			Service: "Gripmock",
			Method:  "ApiInfo",
			Input: stuber.InputData{Matches: map[string]interface{}{
				"vint64": "^100[1-2]{2}\\d{0,3}$",
			}},
			Output: stuber.Output{Data: map[string]interface{}{
				"name":    "Gripmock",
				"version": "1.0",
			}},
		},
	)

	require.Len(t, s.Unused(), 1)

	payload := `{"data":{"vint64":"10012000"},"method":"ApiInfo","service":"Gripmock"}`

	req := httptest.NewRequest(http.MethodPost, "/api/stubs/search", bytes.NewReader([]byte(payload)))
	q, err := stuber.NewQuery(req)
	require.NoError(t, err)

	r, err := s.FindByQuery(q)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.NotNil(t, r.Found())
	require.Nil(t, r.Similar())

	require.Equal(t, map[string]interface{}{
		"name":    "Gripmock",
		"version": "1.0",
	}, r.Found().Output.Data)
}

func TestResult_Similar(t *testing.T) {
	s := stuber.NewBudgerigar(features.New(stuber.MethodTitle))

	s.PutMany(
		&stuber.Stub{
			ID:      uuid.New(),
			Service: "Greeter1",
			Method:  "SayHello1",
			Input: stuber.InputData{Contains: map[string]any{
				"field1": "hello field1",
				"field3": "hello field3",
			}},
			Output: stuber.Output{Data: map[string]any{"message": "hello world"}},
		},
	)

	r, err := s.FindByQuery(stuber.Query{
		ID:      nil,
		Service: "Greeter1",
		Method:  "SayHello1",
		Headers: nil,
		Data: map[string]any{
			"field1": "hello field1",
		},
	})
	require.NoError(t, err)
	require.Nil(t, r.Found())
	require.NotNil(t, r.Similar())
}

func TestStuber_MatchesEqualsFound(t *testing.T) {
	s := stuber.NewBudgerigar(features.New(stuber.MethodTitle))
	// "ignoreArrayOrder": true,
	// "equals": { "id": "123", "tags": ["grpc", "mock"] },
	// "matches": { "name": "^user_\\d+$" },
	s.PutMany(
		&stuber.Stub{
			ID:      uuid.New(),
			Service: "Greeter1",
			Method:  "SayHello1",
			Input: stuber.InputData{
				IgnoreArrayOrder: true,
				Contains: map[string]any{
					"id":   "123",
					"tags": []any{"grpc", "mock"},
				},
				Matches: map[string]any{
					"name": "^user_\\d+$",
				},
			},
			Output: stuber.Output{Data: map[string]any{"message": "hello world"}},
		},
	)

	// {
	// 	"id": 123,
	// 	"name": "user_456",
	// 	"tags": ["grpc", "mock"]
	// }
	r, err := s.FindByQuery(stuber.Query{
		ID:      nil,
		Service: "Greeter1",
		Method:  "SayHello1",
		Headers: nil,
		Data: map[string]any{
			"id":   "123",
			"name": "user_456",
			"tags": []any{"mock", "grpc"},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, r.Found())
	require.Nil(t, r.Similar())
}

func TestDelete(t *testing.T) {
	id1, id2, id3 := uuid.New(), uuid.New(), uuid.New()

	s := stuber.NewBudgerigar(features.New())

	s.PutMany(
		&stuber.Stub{ID: id1, Service: "Greeter1", Method: "SayHello1"},
		&stuber.Stub{ID: id2, Service: "Greeter2", Method: "SayHello2"},
		&stuber.Stub{ID: id3, Service: "Greeter3", Method: "SayHello3"},
	)

	require.NotNil(t, s.FindByID(id1))

	all, err := s.FindBy("Greeter1", "SayHello1")
	require.NoError(t, err)
	require.Len(t, all, 1)

	all, err = s.FindBy("Greeter2", "SayHello2")
	require.NoError(t, err)
	require.Len(t, all, 1)

	all, err = s.FindBy("Greeter3", "SayHello3")
	require.NoError(t, err)
	require.Len(t, all, 1)

	require.Equal(t, 0, s.DeleteByID(uuid.New())) // undefined
	require.Len(t, s.All(), 3)

	require.Equal(t, 1, s.DeleteByID(id1))
	require.Len(t, s.All(), 2)
	require.Nil(t, s.FindByID(id1))

	require.Equal(t, 2, s.DeleteByID(id2, id3))
	require.Empty(t, s.All())
	require.Nil(t, s.FindByID(id2))
	require.Nil(t, s.FindByID(id3))

	all, err = s.FindBy("Greeter1", "SayHello1")
	require.ErrorIs(t, err, stuber.ErrMethodNotFound)
	require.Empty(t, all)

	all, err = s.FindBy("Greeter2", "SayHello2")
	require.ErrorIs(t, err, stuber.ErrMethodNotFound)
	require.Empty(t, all)

	all, err = s.FindBy("Greeter3", "SayHello3")
	require.ErrorIs(t, err, stuber.ErrMethodNotFound)
	require.Empty(t, all)
}

func TestBudgerigar_Clear(t *testing.T) {
	s := stuber.NewBudgerigar(features.New(stuber.MethodTitle))

	s.PutMany(
		&stuber.Stub{
			ID:      uuid.New(),
			Service: "Service1",
			Method:  "Method1",
			Input:   stuber.InputData{Equals: map[string]any{}},
			Output:  stuber.Output{Data: map[string]any{}},
		},
		&stuber.Stub{
			ID:      uuid.New(),
			Service: "Service2",
			Method:  "Method2",
			Input:   stuber.InputData{Equals: map[string]any{}},
			Output:  stuber.Output{Data: map[string]any{}},
		},
	)

	require.Len(t, s.All(), 2)

	s.Clear()

	require.Empty(t, s.All())
}

func TestBudgerigar_FindByQuery_FoundWithPriority(t *testing.T) {
	t.Parallel()

	s := stuber.NewBudgerigar(features.New(stuber.MethodTitle))

	s.PutMany(
		&stuber.Stub{
			ID:       uuid.New(),
			Service:  "Service",
			Method:   "Method",
			Input:    stuber.InputData{Contains: map[string]any{"id": "1"}},
			Output:   stuber.Output{Data: map[string]any{"result": "fail"}},
			Priority: -1,
		},
		&stuber.Stub{
			ID:       uuid.New(),
			Service:  "Service",
			Method:   "Method",
			Input:    stuber.InputData{Matches: map[string]any{"id": "\\d+"}},
			Output:   stuber.Output{Data: map[string]any{"result": "fail"}},
			Priority: 0,
		},
		&stuber.Stub{
			ID:       uuid.New(),
			Service:  "Service",
			Method:   "Method",
			Input:    stuber.InputData{Equals: map[string]any{"id": "1"}},
			Output:   stuber.Output{Data: map[string]any{"result": "success"}},
			Priority: 10,
		},
		&stuber.Stub{
			ID:       uuid.New(),
			Service:  "Service",
			Method:   "Method",
			Input:    stuber.InputData{Equals: map[string]any{"id": "1"}},
			Output:   stuber.Output{Data: map[string]any{"result": "fail"}},
			Priority: 1,
		},
	)

	r, err := s.FindByQuery(stuber.Query{
		Service: "Service",
		Method:  "Method",
		Data:    map[string]any{"id": "1"},
	})

	require.NoError(t, err)
	require.NotNil(t, r.Found())
	require.Nil(t, r.Similar())

	require.Equal(t, "success", r.Found().Output.Data["result"])
}

func TestBudgerigar_Used(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	// Initially no used stubs
	require.Empty(t, s.Used())

	// Add some stubs
	stub1 := &stuber.Stub{ID: uuid.New(), Service: "Service1", Method: "Method1"}
	stub2 := &stuber.Stub{ID: uuid.New(), Service: "Service2", Method: "Method2"}
	s.PutMany(stub1, stub2)

	// Still no used stubs
	require.Empty(t, s.Used())

	// Use a stub by finding it
	_, err := s.FindByQuery(stuber.Query{
		Service: "Service1",
		Method:  "Method1",
	})
	require.NoError(t, err)

	// Now we have one used stub
	used := s.Used()
	require.Len(t, used, 1)
	require.Equal(t, stub1.ID, used[0].ID)
}

func TestBudgerigar_FindByQuery_WithID(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	stubID := uuid.New()
	stub := &stuber.Stub{
		ID:      stubID,
		Service: "Service",
		Method:  "Method",
		Output:  stuber.Output{Data: map[string]any{"result": "success"}},
	}
	s.PutMany(stub)

	// Test finding by ID
	result, err := s.FindByQuery(stuber.Query{
		ID:      &stubID,
		Service: "Service",
		Method:  "Method",
	})

	require.NoError(t, err)
	require.NotNil(t, result.Found())
	require.Equal(t, stubID, result.Found().ID)
	require.Equal(t, "success", result.Found().Output.Data["result"])

	// Test finding by non-existent ID
	nonExistentID := uuid.New()
	_, err = s.FindByQuery(stuber.Query{
		ID:      &nonExistentID,
		Service: "Service",
		Method:  "Method",
	})
	require.Error(t, err)
}

func TestBudgerigar_FindByQuery_InternalRequest(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	stub := &stuber.Stub{
		ID:      uuid.New(),
		Service: "Service",
		Method:  "Method",
		Output:  stuber.Output{Data: map[string]any{"result": "success"}},
	}
	s.PutMany(stub)

	// Test internal request (should not mark as used)
	// We can't directly test internal requests through the public API
	// but we can test that normal requests mark stubs as used
	result, err := s.FindByQuery(stuber.Query{
		Service: "Service",
		Method:  "Method",
	})
	require.NoError(t, err)
	require.NotNil(t, result.Found())

	// Should be marked as used for normal requests
	require.Len(t, s.Used(), 1)
}
