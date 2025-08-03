# Stuber

A high-performance gRPC stub management library for Go with comprehensive support for all gRPC streaming patterns.

## Features

- **Complete gRPC Streaming Support**: Unary, Client Streaming, Server Streaming, and Bidirectional Streaming
- **Dual API Support**: V1 (legacy) and V2 (optimized) APIs
- **Priority-based stub sorting**
- **Efficient memory management** with object pooling
- **Thread-safe operations**
- **Comprehensive test coverage**
- **43x performance improvement** over V1 API

## gRPC Streaming Patterns

Stuber supports all four gRPC streaming patterns:

### 1. Unary RPC
- **Pattern**: One request, one response
- **Stub Configuration**: Use `Input` field for request matching, `Output.Data` for response
- **Use Case**: Simple request-response operations

### 2. Server Streaming RPC
- **Pattern**: One request, multiple responses
- **Stub Configuration**: Use `Input` field for request matching, `Output.Stream` for multiple responses
- **Use Case**: Real-time data streaming, notifications

### 3. Client Streaming RPC
- **Pattern**: Multiple requests, one response
- **Stub Configuration**: Use `Stream` field for request sequence matching, `Output.Data` for response
- **Use Case**: File uploads, batch processing

### 4. Bidirectional Streaming RPC
- **Pattern**: Multiple requests, multiple responses with stateful pattern matching
- **Stub Configuration**: Use `Stream` field for request sequence matching, `Output.Data` for response
- **Use Case**: Chat applications, real-time collaboration, complex conversation flows
- **Key Feature**: Stateful filtering - maintains a pool of candidate stubs and filters them as messages arrive

## Usage

### Basic Stub Definition

```go
import "github.com/gripmock/stuber"

stub := &stuber.Stub{
    ID:       uuid.New(),
    Service:  "GreeterService",
    Method:   "SayHello",
    Priority: 10,
    Input: stuber.InputData{
        Equals: map[string]any{
            "name": "World",
        },
    },
    Output: stuber.Output{
        Data: map[string]any{
            "message": "Hello, World!",
        },
    },
}
```

### Unary RPC Example

```go
// Simple request-response
unaryStub := &stuber.Stub{
    ID:      uuid.New(),
    Service: "UserService",
    Method:  "GetUser",
    Input: stuber.InputData{
        Equals: map[string]any{"user_id": "123"},
    },
    Output: stuber.Output{
        Data: map[string]any{
            "user": map[string]any{
                "id":   "123",
                "name": "John Doe",
                "email": "john@example.com",
            },
        },
    },
}
```

### Server Streaming RPC Example

```go
// One request, multiple responses
serverStreamStub := &stuber.Stub{
    ID:      uuid.New(),
    Service: "NotificationService",
    Method:  "Subscribe",
    Input: stuber.InputData{
        Equals: map[string]any{"user_id": "123"},
    },
    Output: stuber.Output{
        Stream: []any{
            map[string]any{"type": "notification", "message": "Welcome!"},
            map[string]any{"type": "notification", "message": "New message received"},
            map[string]any{"type": "notification", "message": "System update available"},
        },
    },
}
```

### Client Streaming RPC Example

```go
// Multiple requests, one response
clientStreamStub := &stuber.Stub{
    ID:      uuid.New(),
    Service: "FileService",
    Method:  "UploadFile",
    Stream: []stuber.InputData{
        {Equals: map[string]any{"chunk": 1, "data": "file_header"}},
        {Equals: map[string]any{"chunk": 2, "data": "file_content"}},
        {Equals: map[string]any{"chunk": 3, "data": "file_footer"}},
    },
    Output: stuber.Output{
        Data: map[string]any{
            "status": "uploaded",
            "file_id": "abc123",
        },
    },
}
```

### Bidirectional Streaming RPC Example

```go
// Multiple requests, multiple responses with stateful pattern matching
bidiStub := &stuber.Stub{
    ID:      uuid.New(),
    Service: "ChatService",
    Method:  "Chat",
    Stream: []stuber.InputData{
        {Equals: map[string]any{"message": "hello"}},
        {Equals: map[string]any{"message": "world"}},
        {Equals: map[string]any{"message": "goodbye"}},
    },
    Output: stuber.Output{
        Data: map[string]any{"response": "Conversation completed successfully!"},
    },
}
```

## Performance Comparison: V1 vs V2

Stuber provides two API versions with significant performance differences:

### V1 API (Legacy)
- **Performance**: ~73,872 ns/op
- **Memory**: 32,000 B/op
- **Allocations**: 4,000 allocs/op
- **Features**: Unary requests only
- **Status**: Deprecated, maintained for backward compatibility

### V2 API (Optimized)
- **Unary Performance**: ~1,525 ns/op (**48x faster**)
- **Stream Performance**: ~2,855 ns/op (**26x faster**)
- **Memory**: 800-1,336 B/op (**24x less memory**)
- **Allocations**: 34-58 allocs/op (**118x fewer allocations**)
- **Features**: All streaming patterns supported
- **Status**: Recommended for new projects

### Performance Benchmarks

| Metric | V1 (FindByQuery) | V2 Unary | V2 Stream | Improvement |
|--------|------------------|----------|-----------|-------------|
| **Execution Time** | 73,872 ns/op | 1,525 ns/op | 2,855 ns/op | **48x faster** |
| **Memory Usage** | 32,000 B/op | 800 B/op | 1,336 B/op | **24x less** |
| **Allocations** | 4,000 allocs/op | 34 allocs/op | 58 allocs/op | **118x fewer** |

### Migration Guide

**For new projects**: Use V2 API directly
```go
// V2 API - Recommended
result, err := budgerigar.FindByQueryV2(queryV2)
```

**For existing projects**: Gradual migration supported
```go
// V1 API - Still works, but deprecated
result, err := budgerigar.FindByQuery(queryV1)
```

## API Reference

### Core Structures

#### Stub

The main structure representing a gRPC service method stub.

```go
type Stub struct {
    ID       uuid.UUID   // Unique identifier
    Service  string      // Service name
    Method   string      // Method name
    Priority int         // Priority score for sorting
    Headers  InputHeader // Request headers
    Input    InputData   // Request input data (for unary/server streaming)
    Stream   []InputData // Request stream data (for client streaming)
    Output   Output      // Response output data
}
```

#### Stub Type Detection

```go
// Check stub type
stub.IsUnary()        // Returns true for unary/server streaming stubs
stub.IsClientStream() // Returns true for client streaming stubs
stub.IsServerStream() // Returns true for server streaming stubs
stub.IsBidirectional() // Returns true for bidirectional streaming stubs
```

#### Query (V1 - Deprecated)

```go
type Query struct {
    ID      *uuid.UUID       // Optional unique identifier
    Service string           // Service name
    Method  string           // Method name
    Headers map[string]any   // Request headers
    Data    map[string]any   // Request data (unary only)
}
```

#### QueryV2 (Recommended)

```go
type QueryV2 struct {
    ID      *uuid.UUID       // Optional unique identifier
    Service string           // Service name
    Method  string           // Method name
    Headers map[string]any   // Request headers
    Input   []map[string]any // Request input data (supports all patterns)
}
```

#### QueryBidi (Bidirectional Streaming)

```go
type QueryBidi struct {
    ID      *uuid.UUID     // Optional unique identifier
    Service string         // Service name
    Method  string         // Method name
    Headers map[string]any // Request headers
}
```

### Output

Represents the output data of a gRPC response.

```go
type Output struct {
    Headers map[string]string // Response headers
    Data    map[string]any    // Response data (omitempty)
    Stream  []any             // Stream data for server streaming (omitempty)
    Error   string            // Error message
    Code    *codes.Code       // Status code (omitempty)
    Delay   time.Duration     // Response delay (omitempty)
}
```

## Usage Examples

### V1 API (Legacy)

```go
// Create a stub
stub := &stuber.Stub{
    Service: "GreeterService",
    Method:  "SayHello",
    Input: stuber.InputData{
        Equals: map[string]any{"name": "World"},
    },
    Output: stuber.Output{
        Data: map[string]any{"message": "Hello, World!"},
    },
}

// Add to searcher
budgerigar.PutMany(stub)

// Search using V1 API
query := stuber.Query{
    Service: "GreeterService",
    Method:  "SayHello",
    Data:    map[string]any{"name": "World"},
}

result, err := budgerigar.FindByQuery(query)
if err != nil {
    log.Fatal(err)
}

fmt.Println(result.Found().Output.Data["message"])
```

### V2 API (Recommended)

```go
// Create a stub
stub := &stuber.Stub{
    Service: "GreeterService",
    Method:  "SayHello",
    Input: stuber.InputData{
        Equals: map[string]any{"name": "World"},
    },
    Output: stuber.Output{
        Data: map[string]any{"message": "Hello, World!"},
    },
}

// Add to searcher
budgerigar.PutMany(stub)

// Search using V2 API
query := stuber.QueryV2{
    Service: "GreeterService",
    Method:  "SayHello",
    Input:   []map[string]any{{"name": "World"}},
}

result, err := budgerigar.FindByQueryV2(query)
if err != nil {
    log.Fatal(err)
}

fmt.Println(result.Found().Output.Data["message"])
```

### Bidirectional Streaming

```go
// Create bidirectional streaming stubs with different patterns
pattern1Stub := &stuber.Stub{
	ID:      uuid.New(),
	Service: "ChatService",
	Method:  "Chat",
	Stream: []stuber.InputData{
		{Equals: map[string]any{"message": "hello"}},
		{Equals: map[string]any{"message": "world"}},
		{Equals: map[string]any{"message": "goodbye"}},
	},
	Output: stuber.Output{
		Data: map[string]any{"response": "Pattern 1 completed!"},
	},
}

pattern2Stub := &stuber.Stub{
	ID:      uuid.New(),
	Service: "ChatService",
	Method:  "Chat",
	Stream: []stuber.InputData{
		{Equals: map[string]any{"message": "hello"}},
		{Equals: map[string]any{"message": "universe"}},
		{Equals: map[string]any{"message": "farewell"}},
	},
	Output: stuber.Output{
		Data: map[string]any{"response": "Pattern 2 completed!"},
	},
}

budgerigar.PutMany(pattern1Stub, pattern2Stub)

// Use bidirectional streaming
query := stuber.QueryBidi{
	Service: "ChatService",
	Method:  "Chat",
	Headers: map[string]any{"content-type": "application/json"},
}

result, err := budgerigar.FindByQueryBidi(query)
if err != nil {
	log.Fatal(err)
}

// First message - both patterns match "hello"
stub, err := result.Next(map[string]any{"message": "hello"})
if err != nil {
	log.Fatal(err)
}

// Second message - filters to pattern1 if "world", pattern2 if "universe"
stub, err = result.Next(map[string]any{"message": "world"})
if err != nil {
	log.Fatal(err)
}

// Third message - continues with the selected pattern
stub, err = result.Next(map[string]any{"message": "goodbye"})
if err != nil {
	log.Fatal(err)
}

fmt.Println(stub.Output.Data["response"])
// Output: "Pattern 1 completed!"
```

## Stateful Bidirectional Streaming Logic

Bidirectional streaming in stuber uses **stateful pattern matching** to handle complex conversation flows:

### How It Works

1. **Initial Pool**: When you start a bidirectional session, stuber creates a pool of all available stubs for the service/method
2. **First Message**: The first message filters the pool to stubs that could potentially match the pattern
3. **Subsequent Messages**: Each new message further filters the candidate stubs based on the pattern
4. **Pattern Matching**: Only stubs that match the entire sequence up to the current message remain candidates
5. **Best Match**: Among remaining candidates, stuber selects the best match based on priority and ranking

### Example Scenario

```go
// Three different conversation patterns
pattern1: hello -> world -> goodbye
pattern2: hello -> universe -> farewell  
pattern3: hello -> galaxy -> adios

// Client sends: "hello"
// Result: All 3 patterns are candidates

// Client sends: "world"  
// Result: Only pattern1 remains (pattern2,3 are filtered out)

// Client sends: "goodbye"
// Result: pattern1 matches and returns response
```

### Key Benefits

- **Flexible Matching**: Supports multiple conversation patterns for the same service
- **Progressive Filtering**: Reduces candidate pool as conversation progresses
- **Error Handling**: Returns error if no patterns match at any point
- **Priority Support**: Uses stub priorities when multiple candidates remain

## Integration with Gripmock

When using stuber with gripmock, all streaming patterns work seamlessly with the existing HTTP API:

### Adding Stubs via HTTP API

```bash
# Add a unary stub
curl -X POST http://localhost:4770/add \
  -H "Content-Type: application/json" \
  -d '{
    "service": "UserService",
    "method": "GetUser",
    "input": {
      "equals": {"user_id": "123"}
    },
    "output": {
      "data": {
        "user": {
          "id": "123",
          "name": "John Doe"
        }
      }
    }
  }'

# Add a server streaming stub
curl -X POST http://localhost:4770/add \
  -H "Content-Type: application/json" \
  -d '{
    "service": "NotificationService",
    "method": "Subscribe",
    "input": {
      "equals": {"user_id": "123"}
    },
    "output": {
      "stream": [
        {"type": "notification", "message": "Welcome!"},
        {"type": "notification", "message": "New message received"}
      ]
    }
  }'

# Add a client streaming stub
curl -X POST http://localhost:4770/add \
  -H "Content-Type: application/json" \
  -d '{
    "service": "FileService",
    "method": "UploadFile",
    "stream": [
      {"equals": {"chunk": 1, "data": "file_header"}},
      {"equals": {"chunk": 2, "data": "file_content"}},
      {"equals": {"chunk": 3, "data": "file_footer"}}
    ],
    "output": {
      "data": {"status": "uploaded", "file_id": "abc123"}
    }
  }'

# Add a bidirectional streaming stub
curl -X POST http://localhost:4770/add \
  -H "Content-Type: application/json" \
  -d '{
    "service": "ChatService",
    "method": "Chat",
    "stream": [
      {"equals": {"message": "hello"}},
      {"equals": {"message": "world"}},
      {"equals": {"message": "goodbye"}}
    ],
    "output": {
      "data": {"response": "Conversation completed successfully!"}
    }
  }'
```

### Using Bidirectional Streaming

```bash
# Start bidirectional streaming session
curl -X POST http://localhost:4770/find-bidi \
  -H "Content-Type: application/json" \
  -d '{
    "service": "ChatService",
    "method": "Chat",
    "headers": {"content-type": "application/json"}
  }'

# Send messages one by one
curl -X POST http://localhost:4770/next \
  -H "Content-Type: application/json" \
  -d '{"message": "hello"}'

curl -X POST http://localhost:4770/next \
  -H "Content-Type: application/json" \
  -d '{"message": "goodbye"}'
```

## Best Practices

### 1. Stub Design

- **Use appropriate fields**: `Input` for unary/server streaming, `Stream` for client streaming
- **Keep stubs focused**: One stub per specific request pattern
- **Use priorities**: Higher priority stubs are matched first
- **Validate data**: Use `Equals`, `Contains`, and `Matches` appropriately

### 2. Performance

- **Use V2 API**: Significantly better performance than V1
- **Optimize headers**: Include only necessary headers for matching
- **Use priorities**: Order stubs by frequency of use

### 3. Testing

- **Test all patterns**: Ensure your stubs work for all streaming types
- **Mock edge cases**: Create stubs for error conditions
- **Validate responses**: Check both `Data` and `Stream` outputs

### 4. Integration

- **Start with unary**: Begin with simple request-response patterns
- **Add streaming gradually**: Introduce streaming patterns as needed
- **Monitor performance**: Use the performance benchmarks as guidelines

## Contributing

Contributions are welcome! Please ensure all tests pass and new features include appropriate test coverage.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
