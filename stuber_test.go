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

func TestServiceNotFound(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	_, err := s.FindBy("hello", "world")

	require.ErrorIs(t, err, stuber.ErrServiceNotFound)
}

func TestMethodNotFound(t *testing.T) {
	s := stuber.NewBudgerigar(features.New())

	s.PutMany(
		&stuber.Stub{ID: uuid.New(), Service: "Greeter1", Method: "SayHello1"},
	)

	_, err := s.FindBy("Greeter1", "world")

	require.ErrorIs(t, err, stuber.ErrMethodNotFound)
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
			Input: stuber.InputData{Contains: map[string]interface{}{
				"field1": "hello field1",
			}},
			Output: stuber.Output{Data: map[string]interface{}{"message": "hello world"}},
		},
		&stuber.Stub{
			ID:      uuid.New(),
			Service: "Greeter2",
			Method:  "SayHello1",
			Input: stuber.InputData{Contains: map[string]interface{}{
				"field1": "hello field1",
			}},
			Output: stuber.Output{Data: map[string]interface{}{"message": "greeter2"}},
		},
		&stuber.Stub{
			ID:      uuid.New(),
			Service: "Greeter1",
			Method:  "SayHello1",
			Input: stuber.InputData{Contains: map[string]interface{}{
				"field1": "hello field2",
			}},
			Output: stuber.Output{Data: map[string]interface{}{"message": "say hello world"}},
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

	require.Equal(t, map[string]interface{}{"message": "hello world"}, r.Found().Output.Data)
}

func TestBudgerigar_SearchWithHeaders(t *testing.T) {
	s := stuber.NewBudgerigar(features.New(stuber.MethodTitle))

	require.Empty(t, s.Unused())

	s.PutMany(
		&stuber.Stub{
			ID:      uuid.New(),
			Service: "Gripmock",
			Method:  "SayHello",
			Input: stuber.InputData{Equals: map[string]interface{}{
				"name": "simple3",
			}},
			Output: stuber.Output{Data: map[string]interface{}{
				"message": "Hello Simple3",
			}},
		},
		&stuber.Stub{
			ID:      uuid.New(),
			Service: "Gripmock",
			Method:  "SayHello",
			Headers: stuber.InputHeader{Equals: map[string]interface{}{
				"authorization": "Basic dXNlcjp1c2Vy",
			}},
			Input: stuber.InputData{Equals: map[string]interface{}{
				"name": "simple3",
			}},
			Output: stuber.Output{Data: map[string]interface{}{
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

	require.Equal(t, map[string]interface{}{
		"message":     "Hello Simple3",
		"return_code": 3,
	}, r.Found().Output.Data)
}

func TestBudgerigar_SearchEmpty(t *testing.T) {
	s := stuber.NewBudgerigar(features.New(stuber.MethodTitle))

	require.Empty(t, s.Unused())

	s.PutMany(
		&stuber.Stub{
			ID:      uuid.New(),
			Service: "Gripmock",
			Method:  "ApiInfo",
			Input:   stuber.InputData{Equals: map[string]interface{}{}},
			Output: stuber.Output{Data: map[string]interface{}{
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

	require.Equal(t, map[string]interface{}{
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
			Input: stuber.InputData{Equals: map[string]interface{}{
				"name": "simple3",
			}},
			Output: stuber.Output{Data: map[string]interface{}{
				"message":     "Hello Simple3",
				"return_code": 3,
			}},
		},
		&stuber.Stub{
			ID:      uuid.New(),
			Service: "Gripmock",
			Method:  "SayHello",
			Headers: stuber.InputHeader{Equals: map[string]interface{}{
				"authorization": "Basic dXNlcjp1c2Vy",
			}},
			Input: stuber.InputData{Equals: map[string]interface{}{
				"name": "simple3",
			}},
			Output: stuber.Output{Data: map[string]interface{}{
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

	require.Equal(t, map[string]interface{}{
		"message":     "Hello Simple3",
		"return_code": 3,
	}, r.Similar().Output.Data)
}

func TestResult_Similar(t *testing.T) {
	s := stuber.NewBudgerigar(features.New(stuber.MethodTitle))

	s.PutMany(
		&stuber.Stub{
			ID:      uuid.New(),
			Service: "Greeter1",
			Method:  "SayHello1",
			Input: stuber.InputData{Contains: map[string]interface{}{
				"field1": "hello field1",
				"field3": "hello field3",
			}},
			Output: stuber.Output{Data: map[string]interface{}{"message": "hello world"}},
		},
	)

	r, err := s.FindByQuery(stuber.Query{
		ID:      nil,
		Service: "Greeter1",
		Method:  "SayHello1",
		Headers: nil,
		Data: map[string]interface{}{
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
	require.NoError(t, err)
	require.Empty(t, all)

	all, err = s.FindBy("Greeter2", "SayHello2")
	require.NoError(t, err)
	require.Empty(t, all)

	all, err = s.FindBy("Greeter3", "SayHello3")
	require.NoError(t, err)
	require.Empty(t, all)
}

func TestBudgerigar_Clear(t *testing.T) {
	s := stuber.NewBudgerigar(features.New(stuber.MethodTitle))

	s.PutMany(
		&stuber.Stub{
			ID:      uuid.New(),
			Service: "Service1",
			Method:  "Method1",
			Input:   stuber.InputData{Equals: map[string]interface{}{}},
			Output:  stuber.Output{Data: map[string]interface{}{}},
		},
		&stuber.Stub{
			ID:      uuid.New(),
			Service: "Service2",
			Method:  "Method2",
			Input:   stuber.InputData{Equals: map[string]interface{}{}},
			Output:  stuber.Output{Data: map[string]interface{}{}},
		},
	)

	require.Len(t, s.All(), 2)

	s.Clear()

	require.Empty(t, s.All())
}
