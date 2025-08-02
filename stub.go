package stuber

import (
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
)

// Stub represents a gRPC service method and its associated data.
type Stub struct {
	ID       uuid.UUID   `json:"id"`       // The unique identifier of the stub.
	Service  string      `json:"service"`  // The name of the service.
	Method   string      `json:"method"`   // The name of the method.
	Priority int         `json:"priority"` // The priority score of the stub.
	Headers  InputHeader `json:"headers"`  // The headers of the request.
	Input    InputData   `json:"input"`    // The input data of the request.
	Output   Output      `json:"output"`   // The output data of the response.
}

// Key returns the unique identifier of the stub.
func (s Stub) Key() uuid.UUID {
	return s.ID
}

// Left returns the service name of the stub.
func (s Stub) Left() string {
	return s.Service
}

// Right returns the method name of the stub.
func (s Stub) Right() string {
	return s.Method
}

// Score returns the priority score of the stub.
func (s Stub) Score() int {
	return s.Priority
}

// InputData represents the input data of a gRPC request.
type InputData struct {
	IgnoreArrayOrder bool           `json:"ignoreArrayOrder,omitempty"` // Whether to ignore the order of arrays in the input data.
	Equals           map[string]any `json:"equals"`                     // The data to match exactly.
	Contains         map[string]any `json:"contains"`                   // The data to match partially.
	Matches          map[string]any `json:"matches"`                    // The data to match using regular expressions.
}

// GetEquals returns the data to match exactly.
func (i InputData) GetEquals() map[string]any {
	return i.Equals
}

// GetContains returns the data to match partially.
func (i InputData) GetContains() map[string]any {
	return i.Contains
}

// GetMatches returns the data to match using regular expressions.
func (i InputData) GetMatches() map[string]any {
	return i.Matches
}

// InputHeader represents the headers of a gRPC request.
type InputHeader struct {
	Equals   map[string]any `json:"equals"`   // The headers to match exactly.
	Contains map[string]any `json:"contains"` // The headers to match partially.
	Matches  map[string]any `json:"matches"`  // The headers to match using regular expressions.
}

// GetEquals returns the headers to match exactly.
func (i InputHeader) GetEquals() map[string]any {
	return i.Equals
}

// GetContains returns the headers to match partially.
func (i InputHeader) GetContains() map[string]any {
	return i.Contains
}

// GetMatches returns the headers to match using regular expressions.
func (i InputHeader) GetMatches() map[string]any {
	return i.Matches
}

// Len returns the total number of headers to match.
func (i InputHeader) Len() int {
	return len(i.Equals) + len(i.Matches) + len(i.Contains)
}

// Output represents the output data of a gRPC response.
type Output struct {
	Headers map[string]string `json:"headers"`          // The headers of the response.
	Data    map[string]any    `json:"data,omitempty"`   // The data of the response.
	Stream  []any             `json:"stream,omitempty"` // The stream data for server-side streaming.
	// Each element represents a message to be sent.
	Error string        `json:"error"`           // The error message of the response.
	Code  *codes.Code   `json:"code,omitempty"`  // The status code of the response.
	Delay time.Duration `json:"delay,omitempty"` // The delay of the response or error.
}
