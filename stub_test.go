package stuber //nolint:testpackage

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
)

func TestStub_Methods(t *testing.T) {
	id := uuid.New()
	stub := &Stub{
		ID:       id,
		Service:  "TestService",
		Method:   "TestMethod",
		Priority: 10,
	}

	require.Equal(t, id, stub.Key())
	require.Equal(t, "TestService", stub.Left())
	require.Equal(t, "TestMethod", stub.Right())
	require.Equal(t, 10, stub.Score())
}

func TestInputData_Methods(t *testing.T) {
	inputData := InputData{
		IgnoreArrayOrder: true,
		Equals:           map[string]any{"key1": "value1"},
		Contains:         map[string]any{"key2": "value2"},
		Matches:          map[string]any{"key3": "value3"},
	}

	require.Equal(t, map[string]any{"key1": "value1"}, inputData.GetEquals())
	require.Equal(t, map[string]any{"key2": "value2"}, inputData.GetContains())
	require.Equal(t, map[string]any{"key3": "value3"}, inputData.GetMatches())
}

func TestInputHeader_Methods(t *testing.T) {
	inputHeader := InputHeader{
		Equals:   map[string]any{"header1": "value1"},
		Contains: map[string]any{"header2": "value2"},
		Matches:  map[string]any{"header3": "value3"},
	}

	require.Equal(t, map[string]any{"header1": "value1"}, inputHeader.GetEquals())
	require.Equal(t, map[string]any{"header2": "value2"}, inputHeader.GetContains())
	require.Equal(t, map[string]any{"header3": "value3"}, inputHeader.GetMatches())
	require.Equal(t, 3, inputHeader.Len())
}

func TestInputHeader_Len_Empty(t *testing.T) {
	inputHeader := InputHeader{}

	require.Equal(t, 0, inputHeader.Len())
}

func TestOutput_Fields(t *testing.T) {
	code := codes.OK
	output := Output{
		Headers: map[string]string{"header1": "value1"},
		Data:    map[string]any{"data1": "value1"},
		Error:   "test error",
		Code:    &code,
		Delay:   100,
	}

	require.Equal(t, map[string]string{"header1": "value1"}, output.Headers)
	require.Equal(t, map[string]any{"data1": "value1"}, output.Data)
	require.Equal(t, "test error", output.Error)
	require.Equal(t, &code, output.Code)
	require.Equal(t, 100, int(output.Delay))
}
