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

type searcher struct {
	stubUsed map[uuid.UUID]struct{}
}

type Query struct {
	ID      *uuid.UUID             `json:"id,omitempty"`
	Service string                 `json:"service"`
	Method  string                 `json:"method"`
	Headers map[string]interface{} `json:"headers"`
	Data    map[string]interface{} `json:"data"`

	toggles features.Toggles
}

func toggles(r *http.Request) features.Toggles {
	var flags []features.Flag

	if len(r.Header.Values("X-GripMock-RequestInternal")) > 0 {
		flags = append(flags, RequestInternalFlag)
	}

	return features.New(flags...)
}

func NewQuery(r *http.Request) (*Query, error) {
	q := &Query{
		toggles: toggles(r),
	}

	decoder := json.NewDecoder(r.Body)
	decoder.UseNumber()

	if err := decoder.Decode(q); err != nil {
		return nil, err
	}

	return q, nil
}

type Result struct {
	found   *Stub
	similar *Stub
	err     error
}

func (searcher *searcher) Find(query Query) Result {
	return Result{similar: nil, found: nil, err: nil}
}
