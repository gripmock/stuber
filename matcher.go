package stuber

import (
	"github.com/gripmock/deeply"
)

// match checks if a given query matches a given stub.
//
// It checks if the query matches the stub's input data and headers using
// the equals, contains, and matches methods.
func match(query Query, stub *Stub) bool {
	// Check if the query's input data matches the stub's input data.
	dataMatch := equals(stub.Input.Equals, query.Data, stub.Input.IgnoreArrayOrder) &&
		contains(stub.Input.Contains, query.Data, stub.Input.IgnoreArrayOrder) &&
		matches(stub.Input.Matches, query.Data, stub.Input.IgnoreArrayOrder)

	// Check if the query's headers match the stub's headers.
	headersMatch := equals(stub.Headers.Equals, query.Headers, false) &&
		contains(stub.Headers.Contains, query.Headers, false) &&
		matches(stub.Headers.Matches, query.Headers, false)

	// Return true if both the data and headers match, otherwise false.
	return dataMatch && headersMatch
}

// rankMatch ranks how well a given query matches a given stub.
//
// It ranks the query's input data and headers against the stub's input data
// and headers using the RankMatch method from the deeply package.
func rankMatch(query Query, stub *Stub) float64 {
	// Rank the query's input data against the stub's input data.
	dataRank := deeply.RankMatch(stub.Input.Equals, query.Data) +
		deeply.RankMatch(stub.Input.Contains, query.Data) +
		deeply.RankMatch(stub.Input.Matches, query.Data)

	// If the stub has headers, rank the query's headers against the stub's headers.
	var headersRank float64
	if stub.Headers.Len() > 0 {
		headersRank = deeply.RankMatch(stub.Headers.Equals, query.Headers) +
			deeply.RankMatch(stub.Headers.Contains, query.Headers) +
			deeply.RankMatch(stub.Headers.Matches, query.Headers)
	}

	// Return the sum of the data and headers ranks.
	return dataRank + headersRank
}

// equals checks if the expected map matches the actual value.
//
// It returns true if the expected map matches the actual value,
// otherwise false.
func equals(expected map[string]any, actual any, orderIgnore bool) bool {
	// If the expected map is empty or nil, return true.
	if len(expected) == 0 {
		return true
	}

	// If orderIgnore is true, use the EqualsIgnoreArrayOrder method from the deeply package.
	if orderIgnore {
		return deeply.EqualsIgnoreArrayOrder(expected, actual)
	}

	// Otherwise, use the Equals method from the deeply package.
	return deeply.Equals(expected, actual)
}

// contains checks if the expected map is a subset of the actual value.
//
// It returns true if the expected map is a subset of the actual value,
// otherwise false.
func contains(expected map[string]any, actual any, _ bool) bool {
	// If the expected map is empty or nil, return true.
	if len(expected) == 0 {
		return true
	}

	// Use the ContainsIgnoreArrayOrder method from the deeply package.
	return deeply.ContainsIgnoreArrayOrder(expected, actual)
}

// matches checks if the expected map matches the actual value using regular expressions.
//
// It returns true if the expected map matches the actual value using regular expressions,
// otherwise false.
func matches(expected map[string]any, actual any, _ bool) bool {
	// If the expected map is empty or nil, return true.
	if len(expected) == 0 {
		return true
	}

	// Use the MatchesIgnoreArrayOrder method from the deeply package.
	return deeply.MatchesIgnoreArrayOrder(expected, actual)
}
