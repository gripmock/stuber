package stuber //nolint:testpackage

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/bavix/features"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestQuery_RequestInternal(t *testing.T) {
	// Test without internal flag
	q := Query{
		toggles: features.New(),
	}
	require.False(t, q.RequestInternal())

	// Test with internal flag
	q = Query{
		toggles: features.New(RequestInternalFlag),
	}
	require.True(t, q.RequestInternal())
}

func TestToggles(t *testing.T) {
	// Test without header
	req, _ := http.NewRequest("POST", "/", nil)
	togglesResult := toggles(req)
	require.False(t, togglesResult.Has(RequestInternalFlag))

	// Test with header
	req, _ = http.NewRequest("POST", "/", nil)
	req.Header.Set("X-Gripmock-Requestinternal", "true")
	togglesResult = toggles(req)
	require.True(t, togglesResult.Has(RequestInternalFlag))
}

func TestNewQuery_WithBody(t *testing.T) {
	data := map[string]any{
		"service": "TestService",
		"method":  "TestMethod",
		"data":    map[string]any{"key": "value"},
		"headers": map[string]any{"header": "value"},
	}

	body, err := json.Marshal(data)
	require.NoError(t, err)
	req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(body))

	q, err := NewQuery(req)
	require.NoError(t, err)
	require.Equal(t, "TestService", q.Service)
	require.Equal(t, "TestMethod", q.Method)
	require.Equal(t, map[string]any{"key": "value"}, q.Data)
	require.Equal(t, map[string]any{"header": "value"}, q.Headers)
}

func TestNewQuery_WithID(t *testing.T) {
	id := uuid.New()
	data := map[string]any{
		"id":      id.String(),
		"service": "TestService",
		"method":  "TestMethod",
	}

	body, err := json.Marshal(data)
	require.NoError(t, err)
	req, _ := http.NewRequest("POST", "/", bytes.NewBuffer(body))

	q, err := NewQuery(req)
	require.NoError(t, err)
	require.Equal(t, id, *q.ID)
	require.Equal(t, "TestService", q.Service)
	require.Equal(t, "TestMethod", q.Method)
}

func TestNewQuery_InvalidJSON(t *testing.T) {
	req, _ := http.NewRequest("POST", "/", bytes.NewBufferString("invalid json"))

	_, err := NewQuery(req)
	require.Error(t, err)
}
