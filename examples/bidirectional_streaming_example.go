package main

import (
	"fmt"
	"log"

	"github.com/bavix/features"
	"github.com/google/uuid"
	"github.com/gripmock/stuber"
)

// Example demonstrating bidirectional streaming usage with stateful logic
func main() {
	// Initialize stuber
	budgerigar := stuber.NewBudgerigar(features.New())

	// Create multiple stubs with different patterns
	// All start with "hello" but diverge after that
	pattern1Stub := &stuber.Stub{
		ID:       uuid.New(),
		Service:  "ChatService",
		Method:   "Chat",
		Priority: 1, // Lower priority
		Headers: stuber.InputHeader{
			Equals: map[string]any{"content-type": "application/json"},
		},
		Stream: []stuber.InputData{
			{Equals: map[string]any{"message": "hello"}},
			{Equals: map[string]any{"message": "world"}},
			{Equals: map[string]any{"message": "goodbye"}},
		},
		Output: stuber.Output{
			Data: map[string]any{"response": "Pattern 1 completed successfully!"},
		},
	}

	pattern2Stub := &stuber.Stub{
		ID:       uuid.New(),
		Service:  "ChatService",
		Method:   "Chat",
		Priority: 2, // Higher priority
		Headers: stuber.InputHeader{
			Equals: map[string]any{"content-type": "application/json"},
		},
		Stream: []stuber.InputData{
			{Equals: map[string]any{"message": "hello"}},
			{Equals: map[string]any{"message": "universe"}},
			{Equals: map[string]any{"message": "farewell"}},
		},
		Output: stuber.Output{
			Data: map[string]any{"response": "Pattern 2 completed successfully!"},
		},
	}

	pattern3Stub := &stuber.Stub{
		ID:       uuid.New(),
		Service:  "ChatService",
		Method:   "Chat",
		Priority: 3, // Highest priority
		Headers: stuber.InputHeader{
			Equals: map[string]any{"content-type": "application/json"},
		},
		Stream: []stuber.InputData{
			{Equals: map[string]any{"message": "hello"}},
			{Equals: map[string]any{"message": "galaxy"}},
			{Equals: map[string]any{"message": "adios"}},
		},
		Output: stuber.Output{
			Data: map[string]any{"response": "Pattern 3 completed successfully!"},
		},
	}

	// Add the stubs to the searcher
	budgerigar.PutMany(pattern1Stub, pattern2Stub, pattern3Stub)

	// Create a bidirectional streaming query
	query := stuber.QueryBidi{
		Service: "ChatService",
		Method:  "Chat",
		Headers: map[string]any{"content-type": "application/json"},
	}

	// Get the bidirectional result
	result, err := budgerigar.FindByQueryBidi(query)
	if err != nil {
		log.Fatalf("Failed to find stubs: %v", err)
	}

	fmt.Println("=== Bidirectional Streaming with Stateful Logic ===")
	fmt.Println("Available patterns:")
	fmt.Println("  Pattern 1: hello -> world -> goodbye")
	fmt.Println("  Pattern 2: hello -> universe -> farewell")
	fmt.Println("  Pattern 3: hello -> galaxy -> adios")
	fmt.Println()

	// Simulate Pattern 1
	fmt.Println("--- Testing Pattern 1 ---")
	simulatePattern(result, []string{"hello", "world", "goodbye"})

	// Create a new result for Pattern 2 with higher priority
	result2, err := budgerigar.FindByQueryBidi(query)
	if err != nil {
		log.Fatalf("Failed to find stubs: %v", err)
	}

	fmt.Println("\n--- Testing Pattern 2 ---")
	simulatePattern(result2, []string{"hello", "universe", "farewell"})

	// Create a new result for Pattern 3 with highest priority
	result3, err := budgerigar.FindByQueryBidi(query)
	if err != nil {
		log.Fatalf("Failed to find stubs: %v", err)
	}

	fmt.Println("\n--- Testing Pattern 3 ---")
	simulatePattern(result3, []string{"hello", "galaxy", "adios"})

	// Test error case - invalid pattern
	result4, err := budgerigar.FindByQueryBidi(query)
	if err != nil {
		log.Fatalf("Failed to find stubs: %v", err)
	}

	fmt.Println("\n--- Testing Invalid Pattern ---")
	simulatePattern(result4, []string{"hello", "unknown"})

	// Test priority-based selection
	fmt.Println("\n--- Testing Priority-Based Selection ---")
	testPrioritySelection(budgerigar, query)

	fmt.Println("\n=== Example completed ===")
}

// testPrioritySelection tests priority-based stub selection
func testPrioritySelection(budgerigar *stuber.Budgerigar, query stuber.QueryBidi) {
	result, err := budgerigar.FindByQueryBidi(query)
	if err != nil {
		log.Fatalf("Failed to find stubs: %v", err)
	}

	// Send "hello" - should select highest priority stub (Pattern 3)
	stub, err := result.Next(map[string]any{"message": "hello"})
	if err != nil {
		log.Fatalf("Failed to get stub: %v", err)
	}

	fmt.Printf("Selected stub priority: %d, response: %v\n", stub.Priority, stub.Output.Data["response"])
}

// simulatePattern simulates a conversation pattern
func simulatePattern(result *stuber.BidiResult, messages []string) {
	for i, message := range messages {
		fmt.Printf("[Message %d] Client sends: %s\n", i+1, message)

		stub, err := result.Next(map[string]any{"message": message})
		if err != nil {
			fmt.Printf("[Error] No matching stub found: %v\n", err)
			return
		}

		if len(stub.Output.Data) > 0 {
			fmt.Printf("[Response %d] Server responds: %v\n", i+1, stub.Output.Data["response"])
		} else if len(stub.Output.Stream) > 0 {
			fmt.Printf("[Response %d] Server sends stream:\n", i+1)
			for j, streamMsg := range stub.Output.Stream {
				if msg, ok := streamMsg.(map[string]any); ok {
					fmt.Printf("  [Stream %d-%d] %v\n", i+1, j+1, msg["message"])
				}
			}
		}
	}
}

// Example with different stub types
func differentStubTypesExample() {
	budgerigar := stuber.NewBudgerigar(features.New())

	// Unary stub - Input + Output.Data
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
				},
			},
		},
	}

	// Server streaming stub - Input + Output.Stream
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
			},
		},
	}

	// Client streaming stub - Stream + Output.Data
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
				"status":  "uploaded",
				"file_id": "abc123",
			},
		},
	}

	budgerigar.PutMany(unaryStub, serverStreamStub, clientStreamStub)

	fmt.Println("\n=== Different Stub Types Example ===")

	// Check stub types
	fmt.Printf("Unary stub: IsUnary=%v, IsClientStream=%v, IsServerStream=%v, IsBidirectional=%v\n",
		unaryStub.IsUnary(), unaryStub.IsClientStream(), unaryStub.IsServerStream(), unaryStub.IsBidirectional())

	fmt.Printf("Server streaming stub: IsUnary=%v, IsClientStream=%v, IsServerStream=%v, IsBidirectional=%v\n",
		serverStreamStub.IsUnary(), serverStreamStub.IsClientStream(), serverStreamStub.IsServerStream(), serverStreamStub.IsBidirectional())

	fmt.Printf("Client streaming stub: IsUnary=%v, IsClientStream=%v, IsServerStream=%v, IsBidirectional=%v\n",
		clientStreamStub.IsUnary(), clientStreamStub.IsClientStream(), clientStreamStub.IsServerStream(), clientStreamStub.IsBidirectional())

	fmt.Println("=== Different Stub Types Example completed ===")
}
