package stuber

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestEmpty(t *testing.T) {
	// Empty test file to avoid package import issues
	//nolint:testifylint
	require.True(t, true)
}

func TestSearch_IgnoreArrayOrderAndFields(t *testing.T) {
	s := newSearcher()

	stub1 := &Stub{
		ID:      uuid.New(),
		Service: "IdentifierService",
		Method:  "ProcessUUIDs",
		Input: InputData{
			Equals: map[string]any{
				"string_uuids": []any{
					"f1e9ed24-93ba-4e4f-ab9f-3942196d5c03",
					"e3484119-24e1-42d9-b4c2-7d6004ee86d9",
					"cc991218-a920-40c8-9f42-3b329c8723f2",
					"c30f45d2-f8a4-4a94-a994-4cc349bca457",
				},
			},
			IgnoreArrayOrder: true,
		},
		Output: Output{
			Data: map[string]any{"process_id": 1, "status_code": 200},
		},
	}
	stub2 := &Stub{
		ID:      uuid.New(),
		Service: "IdentifierService",
		Method:  "ProcessUUIDs",
		Input: InputData{
			Equals: map[string]any{
				"string_uuids": []any{
					"f1e9ed24-93ba-4e4f-ab9f-3942196d5c03",
					"e3484119-24e1-42d9-b4c2-7d6004ee86d9",
					"cc991218-a920-40c8-9f42-3b329c8723f2",
					"c30f45d2-f8a4-4a94-a994-4cc349bca457",
				},
				"request_timestamp": 1745081266,
			},
			IgnoreArrayOrder: true,
		},
		Output: Output{
			Data: map[string]any{"process_id": 2, "status_code": 200},
		},
	}
	s.upsert(stub1, stub2)

	// Request with array in any order and request_timestamp
	query := QueryV2{
		Service: "IdentifierService",
		Method:  "ProcessUUIDs",
		Input: []map[string]any{{
			"string_uuids": []any{
				"e3484119-24e1-42d9-b4c2-7d6004ee86d9",
				"c30f45d2-f8a4-4a94-a994-4cc349bca457",
				"f1e9ed24-93ba-4e4f-ab9f-3942196d5c03",
				"cc991218-a920-40c8-9f42-3b329c8723f2",
			},
			"request_timestamp": 1745081266,
		}},
	}
	res, err := s.findV2(query)
	require.NoError(t, err)
	require.NotNil(t, res.Found())
	require.Equal(t, 2, int(res.Found().Output.Data["process_id"].(int)))

	// Request with the same array, but without request_timestamp
	query2 := QueryV2{
		Service: "IdentifierService",
		Method:  "ProcessUUIDs",
		Input: []map[string]any{{
			"string_uuids": []any{
				"cc991218-a920-40c8-9f42-3b329c8723f2",
				"f1e9ed24-93ba-4e4f-ab9f-3942196d5c03",
				"c30f45d2-f8a4-4a94-a994-4cc349bca457",
				"e3484119-24e1-42d9-b4c2-7d6004ee86d9",
			},
		}},
	}
	res2, err2 := s.findV2(query2)
	require.NoError(t, err2)
	require.NotNil(t, res2.Found())
	require.Equal(t, 1, int(res2.Found().Output.Data["process_id"].(int)))
}

func TestSearch_IgnoreArrayOrder_UserScenario(t *testing.T) {
	s := newSearcher()

	// Stub 1: without request_timestamp
	stub1 := &Stub{
		ID:      uuid.New(),
		Service: "IdentifierService",
		Method:  "ProcessUUIDs",
		Input: InputData{
			Equals: map[string]any{
				"string_uuids": []any{
					"f1e9ed24-93ba-4e4f-ab9f-3942196d5c03", // 0
					"e3484119-24e1-42d9-b4c2-7d6004ee86d9", // 1
					"cc991218-a920-40c8-9f42-3b329c8723f2", // 2
					"c30f45d2-f8a4-4a94-a994-4cc349bca457", // 3
				},
			},
			IgnoreArrayOrder: true,
		},
		Output: Output{
			Data: map[string]any{"processId": "1", "statusCode": "200"},
		},
	}

	// Stub 2: with request_timestamp
	stub2 := &Stub{
		ID:      uuid.New(),
		Service: "IdentifierService",
		Method:  "ProcessUUIDs",
		Input: InputData{
			Equals: map[string]any{
				"string_uuids": []any{
					"f1e9ed24-93ba-4e4f-ab9f-3942196d5c03", // 0
					"e3484119-24e1-42d9-b4c2-7d6004ee86d9", // 1
					"cc991218-a920-40c8-9f42-3b329c8723f2", // 2
					"c30f45d2-f8a4-4a94-a994-4cc349bca457", // 3
				},
				"request_timestamp": 1745081266,
			},
			IgnoreArrayOrder: true,
		},
		Output: Output{
			Data: map[string]any{"processId": "2", "statusCode": "200"},
		},
	}

	s.upsert(stub1, stub2)

	// Test case 1: Request with different order and request_timestamp
	query1 := QueryV2{
		Service: "IdentifierService",
		Method:  "ProcessUUIDs",
		Input: []map[string]any{{
			"string_uuids": []any{
				"e3484119-24e1-42d9-b4c2-7d6004ee86d9", // 1
				"c30f45d2-f8a4-4a94-a994-4cc349bca457", // 3
				"f1e9ed24-93ba-4e4f-ab9f-3942196d5c03", // 0
				"cc991218-a920-40c8-9f42-3b329c8723f2", // 2
			},
			"request_timestamp": 1745081266,
		}},
	}
	res1, err1 := s.findV2(query1)
	require.NoError(t, err1)
	require.NotNil(t, res1.Found())
	require.Equal(t, "2", res1.Found().Output.Data["processId"])

	// Test case 2: Request with different order and NO request_timestamp
	query2 := QueryV2{
		Service: "IdentifierService",
		Method:  "ProcessUUIDs",
		Input: []map[string]any{{
			"string_uuids": []any{
				"e3484119-24e1-42d9-b4c2-7d6004ee86d9", // 1
				"c30f45d2-f8a4-4a94-a994-4cc349bca457", // 3
				"f1e9ed24-93ba-4e4f-ab9f-3942196d5c03", // 0
				"cc991218-a920-40c8-9f42-3b329c8723f2", // 2
			},
		}},
	}
	res2, err2 := s.findV2(query2)
	require.NoError(t, err2)
	require.NotNil(t, res2.Found())
	require.Equal(t, "1", res2.Found().Output.Data["processId"])

	// Test case 3: Request with different order and request_timestamp (same as case 1)
	query3 := QueryV2{
		Service: "IdentifierService",
		Method:  "ProcessUUIDs",
		Input: []map[string]any{{
			"string_uuids": []any{
				"e3484119-24e1-42d9-b4c2-7d6004ee86d9", // 1
				"c30f45d2-f8a4-4a94-a994-4cc349bca457", // 3
				"f1e9ed24-93ba-4e4f-ab9f-3942196d5c03", // 0
				"cc991218-a920-40c8-9f42-3b329c8723f2", // 2
			},
		}},
	}
	res3, err3 := s.findV2(query3)
	require.NoError(t, err3)
	require.NotNil(t, res3.Found())
	require.Equal(t, "1", res3.Found().Output.Data["processId"])
}

func TestSearch_IgnoreArrayOrder_V1API(t *testing.T) {
	s := newSearcher()

	stub1 := &Stub{
		ID:      uuid.New(),
		Service: "IdentifierService",
		Method:  "ProcessUUIDs",
		Input: InputData{
			Equals: map[string]any{
				"string_uuids": []any{
					"f1e9ed24-93ba-4e4f-ab9f-3942196d5c03",
					"e3484119-24e1-42d9-b4c2-7d6004ee86d9",
					"cc991218-a920-40c8-9f42-3b329c8723f2",
					"c30f45d2-f8a4-4a94-a994-4cc349bca457",
				},
			},
			IgnoreArrayOrder: true,
		},
		Output: Output{
			Data: map[string]any{"processId": "1", "statusCode": "200"},
		},
	}

	stub2 := &Stub{
		ID:      uuid.New(),
		Service: "IdentifierService",
		Method:  "ProcessUUIDs",
		Input: InputData{
			Equals: map[string]any{
				"string_uuids": []any{
					"f1e9ed24-93ba-4e4f-ab9f-3942196d5c03",
					"e3484119-24e1-42d9-b4c2-7d6004ee86d9",
					"cc991218-a920-40c8-9f42-3b329c8723f2",
					"c30f45d2-f8a4-4a94-a994-4cc349bca457",
				},
				"request_timestamp": 1745081266,
			},
			IgnoreArrayOrder: true,
		},
		Output: Output{
			Data: map[string]any{"processId": "2", "statusCode": "200"},
		},
	}

	s.upsert(stub1, stub2)

	// Test with V1 API
	query := Query{
		Service: "IdentifierService",
		Method:  "ProcessUUIDs",
		Data: map[string]any{
			"string_uuids": []any{
				"e3484119-24e1-42d9-b4c2-7d6004ee86d9",
				"c30f45d2-f8a4-4a94-a994-4cc349bca457",
				"f1e9ed24-93ba-4e4f-ab9f-3942196d5c03",
				"cc991218-a920-40c8-9f42-3b329c8723f2",
			},
			"request_timestamp": 1745081266,
		},
	}
	res, err := s.find(query)
	require.NoError(t, err)
	require.NotNil(t, res.Found())
	require.Equal(t, "2", res.Found().Output.Data["processId"])

	// Test without request_timestamp
	query2 := Query{
		Service: "IdentifierService",
		Method:  "ProcessUUIDs",
		Data: map[string]any{
			"string_uuids": []any{
				"cc991218-a920-40c8-9f42-3b329c8723f2",
				"f1e9ed24-93ba-4e4f-ab9f-3942196d5c03",
				"c30f45d2-f8a4-4a94-a994-4cc349bca457",
				"e3484119-24e1-42d9-b4c2-7d6004ee86d9",
			},
		},
	}
	res2, err2 := s.find(query2)
	require.NoError(t, err2)
	require.NotNil(t, res2.Found())
	require.Equal(t, "1", res2.Found().Output.Data["processId"])
}

func TestSearch_Specificity_AllCases(t *testing.T) {
	s := newSearcher()

	// Test case 1: Unary with equals fields
	stub1 := &Stub{
		ID:      uuid.New(),
		Service: "TestService",
		Method:  "UnaryMethod",
		Input: InputData{
			Equals: map[string]any{
				"field1": "value1",
				"field2": "value2",
			},
		},
		Output: Output{
			Data: map[string]any{"result": "stub1"},
		},
	}

	stub2 := &Stub{
		ID:      uuid.New(),
		Service: "TestService",
		Method:  "UnaryMethod",
		Input: InputData{
			Equals: map[string]any{
				"field1": "value1",
				"field2": "value2",
				"field3": "value3",
			},
		},
		Output: Output{
			Data: map[string]any{"result": "stub2"},
		},
	}

	s.upsert(stub1, stub2)

	// Query with field1 and field2 only - should match stub1
	query1 := QueryV2{
		Service: "TestService",
		Method:  "UnaryMethod",
		Input: []map[string]any{{
			"field1": "value1",
			"field2": "value2",
		}},
	}
	res1, err1 := s.findV2(query1)
	require.NoError(t, err1)
	require.NotNil(t, res1.Found())
	require.Equal(t, "stub1", res1.Found().Output.Data["result"])

	// Query with field1, field2, and field3 - should match stub2 (higher specificity)
	query2 := QueryV2{
		Service: "TestService",
		Method:  "UnaryMethod",
		Input: []map[string]any{{
			"field1": "value1",
			"field2": "value2",
			"field3": "value3",
		}},
	}
	res2, err2 := s.findV2(query2)
	require.NoError(t, err2)
	require.NotNil(t, res2.Found())
	require.Equal(t, "stub2", res2.Found().Output.Data["result"])
}

func TestSearch_Specificity_StreamCase(t *testing.T) {
	s := newSearcher()

	// Test case 2: Stream with equals fields
	stub1 := &Stub{
		ID:      uuid.New(),
		Service: "TestService",
		Method:  "StreamMethod",
		Stream: []InputData{
			{
				Equals: map[string]any{
					"field1": "value1",
				},
			},
			{
				Equals: map[string]any{
					"field2": "value2",
				},
			},
		},
		Output: Output{
			Data: map[string]any{"result": "stub1"},
		},
	}

	stub2 := &Stub{
		ID:      uuid.New(),
		Service: "TestService",
		Method:  "StreamMethod",
		Stream: []InputData{
			{
				Equals: map[string]any{
					"field1": "value1",
					"field3": "value3",
				},
			},
			{
				Equals: map[string]any{
					"field2": "value2",
					"field4": "value4",
				},
			},
		},
		Output: Output{
			Data: map[string]any{"result": "stub2"},
		},
	}

	s.upsert(stub1, stub2)

	// Stream query with basic fields - should match stub1
	query1 := QueryV2{
		Service: "TestService",
		Method:  "StreamMethod",
		Input: []map[string]any{
			{"field1": "value1"},
			{"field2": "value2"},
		},
	}
	res1, err1 := s.findV2(query1)
	require.NoError(t, err1)
	require.NotNil(t, res1.Found())
	require.Equal(t, "stub1", res1.Found().Output.Data["result"])

	// Stream query with additional fields - should match stub2 (higher specificity)
	query2 := QueryV2{
		Service: "TestService",
		Method:  "StreamMethod",
		Input: []map[string]any{
			{"field1": "value1", "field3": "value3"},
			{"field2": "value2", "field4": "value4"},
		},
	}
	res2, err2 := s.findV2(query2)
	require.NoError(t, err2)
	require.NotNil(t, res2.Found())
	require.Equal(t, "stub2", res2.Found().Output.Data["result"])
}

func TestSearch_Specificity_WithContainsAndMatches(t *testing.T) {
	s := newSearcher()

	// Test case 4: Mixed field types (equals, contains, matches)
	stub1 := &Stub{
		ID:      uuid.New(),
		Service: "TestService",
		Method:  "MixedMethod",
		Input: InputData{
			Equals: map[string]any{
				"field1": "value1",
			},
			Contains: map[string]any{
				"field2": "value2",
			},
		},
		Output: Output{
			Data: map[string]any{"result": "stub1"},
		},
	}

	stub2 := &Stub{
		ID:      uuid.New(),
		Service: "TestService",
		Method:  "MixedMethod",
		Input: InputData{
			Equals: map[string]any{
				"field1": "value1",
			},
			Contains: map[string]any{
				"field2": "value2",
			},
			Matches: map[string]any{
				"field3": "value3",
			},
		},
		Output: Output{
			Data: map[string]any{"result": "stub2"},
		},
	}

	s.upsert(stub1, stub2)

	// Query with equals and contains - should match stub1
	query1 := QueryV2{
		Service: "TestService",
		Method:  "MixedMethod",
		Input: []map[string]any{{
			"field1": "value1",
			"field2": "value2",
		}},
	}
	res1, err1 := s.findV2(query1)
	require.NoError(t, err1)
	require.NotNil(t, res1.Found())
	require.Equal(t, "stub1", res1.Found().Output.Data["result"])

	// Query with equals, contains, and matches - should match stub2 (higher specificity)
	query2 := QueryV2{
		Service: "TestService",
		Method:  "MixedMethod",
		Input: []map[string]any{{
			"field1": "value1",
			"field2": "value2",
			"field3": "value3",
		}},
	}
	res2, err2 := s.findV2(query2)
	require.NoError(t, err2)
	require.NotNil(t, res2.Found())
	require.Equal(t, "stub2", res2.Found().Output.Data["result"])
}

func TestSearch_Specificity_WithIgnoreArrayOrder(t *testing.T) {
	s := newSearcher()

	// Test case 5: With ignoreArrayOrder
	stub1 := &Stub{
		ID:      uuid.New(),
		Service: "TestService",
		Method:  "ArrayMethod",
		Input: InputData{
			Equals: map[string]any{
				"array1": []any{"a", "b", "c"},
			},
			IgnoreArrayOrder: true,
		},
		Output: Output{
			Data: map[string]any{"result": "stub1"},
		},
	}

	stub2 := &Stub{
		ID:      uuid.New(),
		Service: "TestService",
		Method:  "ArrayMethod",
		Input: InputData{
			Equals: map[string]any{
				"array1": []any{"a", "b", "c"},
				"field1": "value1",
			},
			IgnoreArrayOrder: true,
		},
		Output: Output{
			Data: map[string]any{"result": "stub2"},
		},
	}

	s.upsert(stub1, stub2)

	// Query with array only - should match stub1
	query1 := QueryV2{
		Service: "TestService",
		Method:  "ArrayMethod",
		Input: []map[string]any{{
			"array1": []any{"c", "a", "b"}, // Different order
		}},
	}
	res1, err1 := s.findV2(query1)
	require.NoError(t, err1)
	require.NotNil(t, res1.Found())
	require.Equal(t, "stub1", res1.Found().Output.Data["result"])

	// Query with array and additional field - should match stub2 (higher specificity)
	query2 := QueryV2{
		Service: "TestService",
		Method:  "ArrayMethod",
		Input: []map[string]any{{
			"array1": []any{"b", "c", "a"}, // Different order
			"field1": "value1",
		}},
	}
	res2, err2 := s.findV2(query2)
	require.NoError(t, err2)
	require.NotNil(t, res2.Found())
	require.Equal(t, "stub2", res2.Found().Output.Data["result"])
}
