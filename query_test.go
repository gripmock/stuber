package stuber_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gripmock/stuber"
)

func TestQuery_NewInternal(t *testing.T) {
	payload := `{"service":"Testing","method":"TestMethod","data":{"Hola":"Mundo"}}`

	req := httptest.NewRequest(http.MethodPost, "/api/stubs/search", bytes.NewReader([]byte(payload)))
	req.Header.Add(strings.ToUpper("X-GripMock-RequestInternal"), "ok") //enable

	q, err := stuber.NewQuery(req)
	require.NoError(t, err)

	require.Equal(t, "Testing", q.Service)
	require.Equal(t, "TestMethod", q.Method)
	require.Equal(t, "Mundo", q.Data["Hola"])
	require.True(t, q.RequestInternal())
}

func TestQuery_NewExternal(t *testing.T) {
	payload := `{"service":"Testing","method":"TestMethod","data":{"Hola":"Mundo"}}`

	req := httptest.NewRequest(http.MethodPost, "/api/stubs/search", bytes.NewReader([]byte(payload)))

	q, err := stuber.NewQuery(req)
	require.NoError(t, err)

	require.Equal(t, "Testing", q.Service)
	require.Equal(t, "TestMethod", q.Method)
	require.Equal(t, "Mundo", q.Data["Hola"])
	require.False(t, q.RequestInternal())
}
