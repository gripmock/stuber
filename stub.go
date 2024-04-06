package stuber

import (
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
)

type Stub struct {
	ID      uuid.UUID   `json:"id"`
	Service string      `json:"service"`
	Method  string      `json:"method"`
	Headers InputHeader `json:"headers"`
	Input   InputData   `json:"input"`
	Output  Output      `json:"output"`
}

func (s Stub) Key() uuid.UUID {
	return s.ID
}

func (s Stub) Left() string {
	return s.Service
}

func (s Stub) Right() string {
	return s.Method
}

type InputData struct {
	IgnoreArrayOrder bool                   `json:"ignoreArrayOrder,omitempty"`
	Equals           map[string]interface{} `json:"equals"`
	Contains         map[string]interface{} `json:"contains"`
	Matches          map[string]interface{} `json:"matches"`
}

func (i InputData) GetEquals() map[string]interface{} {
	return i.Equals
}

func (i InputData) GetContains() map[string]interface{} {
	return i.Contains
}

func (i InputData) GetMatches() map[string]interface{} {
	return i.Matches
}

type InputHeader struct {
	Equals   map[string]interface{} `json:"equals"`
	Contains map[string]interface{} `json:"contains"`
	Matches  map[string]interface{} `json:"matches"`
}

func (i InputHeader) GetEquals() map[string]interface{} {
	return i.Equals
}

func (i InputHeader) GetContains() map[string]interface{} {
	return i.Contains
}

func (i InputHeader) GetMatches() map[string]interface{} {
	return i.Matches
}

type Output struct {
	Headers map[string]string      `json:"headers"`
	Data    map[string]interface{} `json:"data"`
	Error   string                 `json:"error"`
	Code    *codes.Code            `json:"code,omitempty"`
}
