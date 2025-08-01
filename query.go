package stuber

import (
	"encoding/json"
	"net/http"

	"github.com/bavix/features"
	"github.com/google/uuid"
)

const (
	// RequestInternalFlag is a feature flag for internal requests.
	RequestInternalFlag features.Flag = iota
)

// Query represents a query for finding stubs.
type Query struct {
	ID      *uuid.UUID     `json:"id,omitempty"` // The unique identifier of the stub (optional).
	Service string         `json:"service"`      // The service name to search for.
	Method  string         `json:"method"`       // The method name to search for.
	Headers map[string]any `json:"headers"`      // The headers to match.
	Data    map[string]any `json:"data"`         // The data to match.

	toggles features.Toggles
}

func toggles(r *http.Request) features.Toggles {
	var flags []features.Flag

	if len(r.Header.Values("X-Gripmock-Requestinternal")) > 0 {
		flags = append(flags, RequestInternalFlag)
	}

	return features.New(flags...)
}

// NewQuery creates a new Query from an HTTP request.
//
// Parameters:
// - r: The HTTP request to parse.
//
// Returns:
// - Query: The parsed query.
// - error: An error if the request body cannot be parsed.
func NewQuery(r *http.Request) (Query, error) {
	q := Query{
		toggles: toggles(r),
	}

	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()

	err := decoder.Decode(&q)

	return q, err
}

// RequestInternal returns true if the query is marked as internal.
func (q Query) RequestInternal() bool {
	return q.toggles.Has(RequestInternalFlag)
}
