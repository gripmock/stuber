# Stuber

A high-performance gRPC stub management library for Go.

## Features

- Priority-based stub sorting
- Efficient memory management
- Thread-safe operations
- Comprehensive test coverage

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

### Server-Side Streaming Support

For server-side streaming, you can use the `Stream` field in the `Output` structure:

```go
stub := &stuber.Stub{
    ID:       uuid.New(),
    Service:  "TrackService",
    Method:   "StreamTrack",
    Priority: 10,
    Input: stuber.InputData{
        Equals: map[string]any{
            "stn": "MS#00005",
        },
    },
    Output: stuber.Output{
        Stream: []any{
            map[string]any{
                "stn":       "MS#00005",
                "identity":  "00",
                "latitude":  0.11000,
                "longitude": 0.00600,
                "speed":     50.0,
                "updatedAt": "2024-01-01T13:00:00.000Z",
            },
            map[string]any{
                "stn":       "MS#00005",
                "identity":  "01",
                "latitude":  0.11001,
                "longitude": 0.00601,
                "speed":     51.0,
                "updatedAt": "2024-01-01T13:00:01.000Z",
            },
            map[string]any{
                "stn":       "MS#00005",
                "identity":  "02",
                "latitude":  0.11002,
                "longitude": 0.00602,
                "speed":     52.0,
                "updatedAt": "2024-01-01T13:00:02.000Z",
            },
        },
    },
}
```

Each element in the `Stream` array represents a message to be sent to the client during server-side streaming.

## API Reference

### Stub

The main structure representing a gRPC service method stub.

```go
type Stub struct {
    ID       uuid.UUID   // Unique identifier
    Service  string      // Service name
    Method   string      // Method name
    Priority int         // Priority score for sorting
    Headers  InputHeader // Request headers
    Input    InputData   // Request input data
    Output   Output      // Response output data
}
```

### Output

Represents the output data of a gRPC response.

```go
type Output struct {
    Headers map[string]string // Response headers
    Data    map[string]any    // Response data (omitempty)
    Stream  []any             // Stream data for server-side streaming (omitempty)
    Error   string            // Error message
    Code    *codes.Code       // Status code (omitempty)
    Delay   time.Duration     // Response delay (omitempty)
}
```



## Testing

Run the test suite:

```bash
go test -v ./...
```

Run with coverage:

```bash
go test -cover ./...
```

## Linting

Run the linter:

```bash
make lint
```

## License

MIT License
