package stuber

import (
	"encoding/json"
	"net/http"

	"github.com/bavix/features"
	"github.com/google/uuid"
)

const (
	RequestInternalFlag features.Flag = iota
)

type Query struct {
	ID      *uuid.UUID     `json:"id,omitempty"`
	Service string         `json:"service"`
	Method  string         `json:"method"`
	Headers map[string]any `json:"headers"`
	Data    map[string]any `json:"data"`

	toggles features.Toggles
}

func toggles(r *http.Request) features.Toggles {
	var flags []features.Flag

	if len(r.Header.Values("X-Gripmock-Requestinternal")) > 0 {
		flags = append(flags, RequestInternalFlag)
	}

	return features.New(flags...)
}

func NewQuery(r *http.Request) (Query, error) {
	q := Query{
		toggles: toggles(r),
	}

	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()

	if err := decoder.Decode(&q); err != nil {
		return q, err
	}

	return q, nil
}

func (q Query) RequestInternal() bool {
	return q.toggles.Has(RequestInternalFlag)
}
