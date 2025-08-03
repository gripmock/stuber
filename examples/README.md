# Stuber Examples

This directory contains practical examples demonstrating how to use stuber for different streaming scenarios.

## Bidirectional Streaming Example

The `bidirectional_streaming_example.go` file demonstrates how to use bidirectional streaming with stuber.

### Running the Example

```bash
go run examples/bidirectional_streaming_example.go
```

### What it demonstrates

1. **Creating Stream Stubs**: How to define stubs with stream patterns for bidirectional communication
2. **Iterator Pattern**: How to use the iterator to process messages one by one
3. **Pattern Matching**: How the first message determines the stream pattern to follow
4. **Response Handling**: How to handle both data and stream responses from stubs

### Expected Output

```
=== Bidirectional Streaming Example ===
Simulating chat conversation...

[Message 1] Client sends: hello
[Response 1] Server responds: Hello! How can I help you?
[Stream 1-1] Server sends: I'm doing well, thank you!
[Stream 1-2] Server sends: Have a great day!

[Message 2] Client sends: how are you
[Response 2] Server responds: Hello! How can I help you?
[Stream 2-1] Server sends: I'm doing well, thank you!
[Stream 2-2] Server sends: Have a great day!

[Message 3] Client sends: goodbye
[Response 3] Server responds: Hello! How can I help you?
[Stream 3-1] Server sends: I'm doing well, thank you!
[Stream 3-2] Server sends: Have a great day!

=== Example completed ===
```

### Key Concepts Demonstrated

- **Stateful Iterator**: The iterator maintains state between calls and advances through the stream pattern
- **Pattern Matching**: Each message must match the expected pattern in the stream
- **Multiple Responses**: The server can send both a main response and additional stream messages
- **Error Handling**: The example shows how to handle cases where no matching stub is found

### Use Cases

This pattern is useful for:
- Chat applications with conversation flows
- Game state management with action sequences
- IoT device communication protocols
- Any scenario requiring stateful bidirectional communication 