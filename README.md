# Stuber

High-performance gRPC stub server with advanced matching capabilities.

## Performance Results

### V1 (Old) vs V2 (New) Comparison

| Benchmark | V1 (Old) | V2 (New) | Improvement |
|-----------|----------|----------|-------------|
| **Found** | 1,568 ns/op | 294.1 ns/op | **5.3x faster** |
| **NotFound** | 70.0 ns/op | 61.0 ns/op | **15% faster** |
| **Multiple Stubs** | 15,635 ns/op | 2,022 ns/op | **7.7x faster** |

### Memory Usage

| Benchmark | V1 (Old) Memory | V2 (New) Memory | Improvement |
|-----------|-----------------|-----------------|-------------|
| **Found** | 608 B/op | 136 B/op | **78% less** |
| **Multiple Stubs** | 7,959 B/op | 616 B/op | **92% less** |
| **Stream** | 160 B/op | 136 B/op | **15% less** |

### Allocations

| Benchmark | V1 (Old) Allocs | V2 (New) Allocs | Improvement |
|-----------|-----------------|-----------------|-------------|
| **Found** | 27 allocs/op | 7 allocs/op | **74% less** |
| **Multiple Stubs** | 262 allocs/op | 12 allocs/op | **95% less** |
| **Stream** | 8 allocs/op | 7 allocs/op | **13% less** |

## Parallel Processing Results

| Stub Count | Sequential | Parallel | Improvement |
|------------|------------|----------|-------------|
| 50 | 7,023 ns/op | 7,060 ns/op | **1% slower** |
| 100 | 13,967 ns/op | 15,604 ns/op | **12% slower** |
| 200 | 30,658 ns/op | 26,657 ns/op | **13% faster** |
| 500 | 72,135 ns/op | 53,036 ns/op | **26% faster** |

**Note**: Parallel processing activates at 100+ stubs. Below 200 stubs, sequential is faster due to overhead.

## Key Features

- **Exact matching** with field-level precision
- **Array order ignoring** with `ignoreArrayOrder: true`
- **Regex pattern matching** with caching
- **Priority-based selection** with specificity scoring
- **Stream support**: Unary, Server, Client, Bidirectional
- **LRU caching** for string hashes and regex patterns
- **Memory pools** for zero-allocation operations

## Usage

```go
package main

import (
    "github.com/gripmock/stuber"
    "github.com/bavix/features"
)

func main() {
    toggles := features.New()
    stuber := stuber.NewBudgerigar(toggles)
    
    stub := &stuber.Stub{
        Service: "example.service",
        Method:  "ExampleMethod",
        Input: stuber.InputData{
            Equals: map[string]any{"id": "123"},
        },
        Output: stuber.Output{
            Data: map[string]any{"result": "success"},
        },
    }
    
    stuber.PutMany(stub)
}
```

## API Reference

```go
type Stub struct {
    ID       uuid.UUID
    Service  string
    Method   string
    Priority int
    Input    InputData
    Output   Output
    Stream   []InputData
}

type InputData struct {
    Equals          map[string]any
    Contains        map[string]any
    Matches         map[string]any
    IgnoreArrayOrder bool
}
```

## Testing

```bash
go test -v
go test -bench=. -benchmem
```

## Performance Notes

- **V2 (New)**: Use for best performance in most cases
- **Parallel processing**: Only beneficial for 200+ stubs
- **Memory**: V2 (New) significantly more memory efficient
- **Allocations**: V2 (New) reduces allocations by 74-95%

## License

MIT License
