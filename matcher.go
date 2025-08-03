package stuber

import (
	"github.com/gripmock/deeply"
)

// match checks if a given query matches a given stub.
//
// It checks if the query matches the stub's input data and headers using
// the equals, contains, and matches methods.
func match(query Query, stub *Stub) bool {
	// Check headers first
	if !matchHeaders(query.Headers, stub.Headers) {
		return false
	}

	// Check if the query's input data matches the stub's input data
	return matchInput(query.Data, stub.Input)
}

// matchHeaders checks if query headers match stub headers.
func matchHeaders(queryHeaders map[string]any, stubHeaders InputHeader) bool {
	return equals(stubHeaders.Equals, queryHeaders, false) &&
		contains(stubHeaders.Contains, queryHeaders, false) &&
		matches(stubHeaders.Matches, queryHeaders, false)
}

// matchInput checks if query data matches stub input.
func matchInput(queryData map[string]any, stubInput InputData) bool {
	return equals(stubInput.Equals, queryData, stubInput.IgnoreArrayOrder) &&
		contains(stubInput.Contains, queryData, stubInput.IgnoreArrayOrder) &&
		matches(stubInput.Matches, queryData, stubInput.IgnoreArrayOrder)
}

// rankMatch ranks how well a given query matches a given stub.
//
// It ranks the query's input data and headers against the stub's input data
// and headers using the RankMatch method from the deeply package.
func rankMatch(query Query, stub *Stub) float64 {
	// Rank headers first
	headersRank := rankHeaders(query.Headers, stub.Headers)

	// Rank the query's input data against the stub's input data
	return headersRank + rankInput(query.Data, stub.Input)
}

// rankHeaders ranks query headers against stub headers.
func rankHeaders(queryHeaders map[string]any, stubHeaders InputHeader) float64 {
	if stubHeaders.Len() == 0 {
		return 0
	}

	return deeply.RankMatch(stubHeaders.Equals, queryHeaders) +
		deeply.RankMatch(stubHeaders.Contains, queryHeaders) +
		deeply.RankMatch(stubHeaders.Matches, queryHeaders)
}

// rankInput ranks query data against stub input
func rankInput(queryData map[string]any, stubInput InputData) float64 {
	return deeply.RankMatch(stubInput.Equals, queryData) +
		deeply.RankMatch(stubInput.Contains, queryData) +
		deeply.RankMatch(stubInput.Matches, queryData)
}

// equals checks if the expected map matches the actual value.
//
// It returns true if the expected map matches the actual value,
// otherwise false.
func equals(expected map[string]any, actual any, orderIgnore bool) bool {
	if len(expected) == 0 {
		return true
	}

	if orderIgnore {
		return deeply.EqualsIgnoreArrayOrder(expected, actual)
	}

	return deeply.Equals(expected, actual)
}

// contains checks if the expected map is a subset of the actual value.
//
// It returns true if the expected map is a subset of the actual value,
// otherwise false.
func contains(expected map[string]any, actual any, _ bool) bool {
	if len(expected) == 0 {
		return true
	}

	return deeply.ContainsIgnoreArrayOrder(expected, actual)
}

// matches checks if the expected map matches the actual value using regular expressions.
//
// It returns true if the expected map matches the actual value using regular expressions,
// otherwise false.
func matches(expected map[string]any, actual any, _ bool) bool {
	if len(expected) == 0 {
		return true
	}

	return deeply.MatchesIgnoreArrayOrder(expected, actual)
}

// matchV2 checks if a given QueryV2 matches a given stub.
// It first tries to match against stream elements, then against input.
// Maintains 100% compatibility with V1 behavior.
func matchV2(query QueryV2, stub *Stub) bool {
	// Check headers first
	if !matchHeaders(query.Headers, stub.Headers) {
		return false
	}

	// Check stream if stub has stream data
	if len(stub.Stream) > 0 {
		return matchStreamElements(query.Input, stub.Stream)
	}

	// If no stream in stub, check unary case (one element in input)
	if len(query.Input) == 1 {
		queryItem := query.Input[0]

		return matchInput(queryItem, stub.Input)
	}

	// Multiple stream items but no stream in stub - no match
	return false
}

// rankMatchV2 ranks how well a given QueryV2 matches a given stub.
// It first tries to rank against stream elements, then against input.
// Maintains 100% compatibility with V1 behavior.
func rankMatchV2(query QueryV2, stub *Stub) float64 {
	// Rank headers first
	headersRank := rankHeaders(query.Headers, stub.Headers)

	// Rank stream if stub has stream data
	if len(stub.Stream) > 0 {
		return headersRank + rankStreamElements(query.Input, stub.Stream)
	}

	// If no stream in stub, rank unary case (one element in input)
	if len(query.Input) == 1 {
		queryItem := query.Input[0]

		return headersRank + rankInput(queryItem, stub.Input)
	}

	// Multiple stream items but no stream in stub - no rank
	return headersRank
}

// matchStreamElements checks if query input matches stub stream element by element
func matchStreamElements(queryStream []map[string]any, stubStream []InputData) bool {
	// If lengths don't match, return false
	if len(queryStream) != len(stubStream) {
		return false
	}

	// Check each element
	for i, queryItem := range queryStream {
		stubItem := stubStream[i]
		if !equals(stubItem.Equals, queryItem, stubItem.IgnoreArrayOrder) ||
			!contains(stubItem.Contains, queryItem, stubItem.IgnoreArrayOrder) ||
			!matches(stubItem.Matches, queryItem, stubItem.IgnoreArrayOrder) {
			return false
		}
	}

	return true
}

// rankStreamElements ranks how well query input matches stub stream element by element
func rankStreamElements(queryStream []map[string]any, stubStream []InputData) float64 {
	// If lengths don't match, return 0
	if len(queryStream) != len(stubStream) {
		return 0
	}

	var totalRank float64
	// Rank each element
	for i, queryItem := range queryStream {
		stubItem := stubStream[i]
		totalRank += deeply.RankMatch(stubItem.Equals, queryItem) +
			deeply.RankMatch(stubItem.Contains, queryItem) +
			deeply.RankMatch(stubItem.Matches, queryItem)
	}

	return totalRank
}
