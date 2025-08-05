# Stuber

High-performance gRPC stub management for Go.

## Quick Start

```go
package main

import "github.com/gripmock/stuber"

func main() {
	// Create stub
	stub := &stuber.Stub{
		Service: "Greeter",
		Method:  "SayHello",
		Input: stuber.InputData{
			Equals: map[string]any{"name": "World"},
		},
		Output: stuber.Output{
			Data: map[string]any{"message": "Hello, World!"},
		},
	}

	// Use
	s := stuber.NewBudgerigar(nil)
	s.PutMany(stub)

	query := stuber.QueryV2{
		Service: "Greeter",
		Method:  "SayHello",
		Input:   []map[string]any{{"name": "World"}},
	}
	result, _ := s.FindByQueryV2(query)
}
```

## Examples

### Unary RPC
```go
stub := &stuber.Stub{
	Service: "UserService",
	Method:  "GetUser",
	Input: stuber.InputData{
		Equals: map[string]any{"user_id": "123"},
	},
	Output: stuber.Output{
		Data: map[string]any{
			"user": map[string]any{
				"id":   "123",
				"name": "John",
			},
		},
	},
}
```

### Server Streaming
```go
stub := &stuber.Stub{
	Service: "NotificationService",
	Method:  "Subscribe",
	Input: stuber.InputData{
		Equals: map[string]any{"user_id": "123"},
	},
	Output: stuber.Output{
		Stream: []any{
			map[string]any{"type": "notification", "message": "Welcome!"},
			map[string]any{"type": "notification", "message": "New message"},
		},
	},
}
```

### Client Streaming
```go
stub := &stuber.Stub{
	Service: "FileService",
	Method:  "UploadFile",
	Stream: []stuber.InputData{
		{Equals: map[string]any{"chunk": 1, "data": "header"}},
		{Equals: map[string]any{"chunk": 2, "data": "content"}},
	},
	Output: stuber.Output{
		Data: map[string]any{
			"status":  "uploaded",
			"file_id": "abc123",
		},
	},
}
```

### Bidirectional Streaming
```go
stub := &stuber.Stub{
	Service: "ChatService",
	Method:  "Chat",
	Stream: []stuber.InputData{
		{Equals: map[string]any{"message": "hello"}},
		{Equals: map[string]any{"message": "world"}},
	},
	Output: stuber.Output{
		Data: map[string]any{"response": "Conversation completed!"},
	},
}

// Usage
query := stuber.QueryBidi{
	Service: "ChatService",
	Method:  "Chat",
}
result, _ := s.FindByQueryBidi(query)

stub, _ := result.Next(map[string]any{"message": "hello"})
stub, _ := result.Next(map[string]any{"message": "world"})
```

## Performance

| Scenario | V1 | V2 | Improvement |
|----------|----|----|-------------|
| Found | 1,421 ns/op | 322 ns/op | **4.4x faster** |
| Not Found | 71 ns/op | 59 ns/op | **1.2x faster** |
| Multiple Stubs | 15,327 ns/op | 1,445 ns/op | **10.6x faster** |

## API

### Stub
```go
type Stub struct {
	ID       uuid.UUID
	Service  string
	Method   string
	Priority int
	Input    InputData   // For unary/server streaming
	Stream   []InputData // For client/bidirectional streaming
	Output   Output
}
```

### QueryV2
```go
type QueryV2 struct {
	Service string
	Method  string
	Headers map[string]any
	Input   []map[string]any
}
```

### Output
```go
type Output struct {
	Data    map[string]any // For unary/client streaming
	Stream  []any          // For server streaming
	Error   string
	Code    *codes.Code
	Delay   time.Duration
}
```

## Tips

- Use **V2 API** for new projects
- Use **Input** for unary/server streaming, **Stream** for client/bidirectional
- Set **priorities** for stub ordering
- Use **Equals**, **Contains**, **Matches** for flexible matching

## License

MIT
