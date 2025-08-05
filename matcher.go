package stuber

import (
	"reflect"
	"regexp"

	"github.com/gripmock/deeply"
	lru "github.com/hashicorp/golang-lru/v2"
)

// Global LRU cache for regex patterns with size limit
var regexCache *lru.Cache[string, *regexp.Regexp]

func init() {
	var err error
	// Create LRU cache with size limit of 1000 regex patterns
	regexCache, err = lru.New[string, *regexp.Regexp](1000)
	if err != nil {
		panic("failed to create regex cache: " + err.Error())
	}
}

// Get retrieves a compiled regex from cache or compiles it if not found
func getRegex(pattern string) (*regexp.Regexp, error) {
	// Try to get from cache first
	if re, exists := regexCache.Get(pattern); exists {
		return re, nil
	}

	// Compile and cache
	re, err := regexp.Compile(pattern)
	if err == nil {
		regexCache.Add(pattern, re)
	}
	return re, err
}

// getRegexCacheStats returns regex cache statistics
func getRegexCacheStats() (int, int) {
	return regexCache.Len(), 1000 // Fixed capacity
}

// clearRegexCache clears the regex cache
func clearRegexCache() {
	regexCache.Purge()
}

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

// rankInput ranks query data against stub input.
func rankInput(queryData map[string]any, stubInput InputData) float64 {
	return deeply.RankMatch(stubInput.Equals, queryData) +
		deeply.RankMatch(stubInput.Contains, queryData) +
		deeply.RankMatch(stubInput.Matches, queryData)
}

// equals compares two values for deep equality.
//
//nolint:gocognit,cyclop,gocyclo,funlen
func equals(expected map[string]any, actual any, orderIgnore bool) bool {
	if len(expected) == 0 {
		return true
	}

	// Convert actual to map if it's not already
	actualMap, ok := actual.(map[string]any)
	if !ok {
		return false
	}

	// Ultra-fast path: single field comparison (most common case)
	if len(expected) == 1 {
		for key, expectedValue := range expected {
			actualValue, exists := actualMap[key]
			if !exists {
				return false
			}
			return ultraFastSpecializedEquals(expectedValue, actualValue)
		}
	}

	// Check if all expected fields are present and equal
	for key, expectedValue := range expected {
		actualValue, exists := actualMap[key]
		if !exists {
			return false
		}

		if !ultraFastSpecializedEquals(expectedValue, actualValue) {
			return false
		}
	}

	// For streaming cases, we don't need to check extra fields
	// as the client might send additional fields that we don't care about
	return true
}

// fastStringEquals provides ultra-fast string comparison
func fastStringEquals(expected, actual any) bool {
	if e, ok := expected.(string); ok {
		if a, ok := actual.(string); ok {
			return e == a
		}
	}
	return false
}

// fastNumberEquals provides ultra-fast number comparison
func fastNumberEquals(expected, actual any) bool {
	switch e := expected.(type) {
	case float64:
		switch a := actual.(type) {
		case float64:
			return e == a
		case int:
			return e == float64(a)
		case int64:
			return e == float64(a)
		}
	case int:
		switch a := actual.(type) {
		case float64:
			return float64(e) == a
		case int:
			return e == a
		case int64:
			return e == int(actual.(int64))
		}
	case int64:
		switch a := actual.(type) {
		case float64:
			return float64(e) == a
		case int:
			return int64(a) == e
		case int64:
			return e == a
		}
	}
	return false
}

// fastBoolEquals provides ultra-fast boolean comparison
func fastBoolEquals(expected, actual any) bool {
	if e, ok := expected.(bool); ok {
		if a, ok := actual.(bool); ok {
			return e == a
		}
	}
	return false
}

// ultraFastEquals is an ultra-optimized version of equals for single value comparison
func ultraFastEquals(expected, actual any, orderIgnore bool) bool {
	// Fast path: same type comparison
	if reflect.TypeOf(expected) == reflect.TypeOf(actual) {
		switch expected.(type) {
		case string:
			return fastStringEquals(expected, actual)
		case float64, int, int64:
			return fastNumberEquals(expected, actual)
		case bool:
			return fastBoolEquals(expected, actual)
		}
	}

	// Ultra-fast path: string comparison (most common)
	if expectedStr, expectedIsStr := expected.(string); expectedIsStr {
		if actualStr, actualIsStr := actual.(string); actualIsStr {
			return expectedStr == actualStr
		}
		return false
	}

	// Ultra-fast path: number comparison
	if expectedNum, expectedIsNum := expected.(float64); expectedIsNum {
		switch actualNum := actual.(type) {
		case float64:
			return expectedNum == actualNum
		case int:
			return expectedNum == float64(actualNum)
		case int64:
			return expectedNum == float64(actualNum)
		}
		return false
	}

	if expectedNum, expectedIsNum := expected.(int); expectedIsNum {
		switch actualNum := actual.(type) {
		case float64:
			return float64(expectedNum) == actualNum
		case int:
			return expectedNum == actualNum
		case int64:
			return expectedNum == int(actualNum)
		}
		return false
	}

	if expectedNum, expectedIsNum := expected.(int64); expectedIsNum {
		switch actualNum := actual.(type) {
		case float64:
			return float64(expectedNum) == actualNum
		case int:
			return int64(actualNum) == expectedNum
		case int64:
			return expectedNum == actualNum
		}
		return false
	}

	// Ultra-fast path: bool comparison
	if expectedBool, expectedIsBool := expected.(bool); expectedIsBool {
		if actualBool, actualIsBool := actual.(bool); actualIsBool {
			return expectedBool == actualBool
		}
		return false
	}

	// For other types, use optimized comparison
	if orderIgnore {
		return deeply.EqualsIgnoreArrayOrder(expected, actual)
	}
	return reflect.DeepEqual(expected, actual)
}

// ultraFastSpecializedEquals provides ultra-fast comparison for common types without reflect
func ultraFastSpecializedEquals(expected, actual any) bool {
	// Ultra-fast path: same type comparison (most common case)
	if reflect.TypeOf(expected) == reflect.TypeOf(actual) {
		switch e := expected.(type) {
		case string:
			return e == actual.(string)
		case int:
			return e == actual.(int)
		case float64:
			return e == actual.(float64)
		case bool:
			return e == actual.(bool)
		case int64:
			return e == actual.(int64)
		}
	}

	// Fast path: number type conversions (common case)
	switch e := expected.(type) {
	case int:
		switch a := actual.(type) {
		case int:
			return e == a
		case float64:
			return float64(e) == a
		case int64:
			return int64(e) == a
		}
	case float64:
		switch a := actual.(type) {
		case float64:
			return e == a
		case int:
			return e == float64(a)
		case int64:
			return e == float64(a)
		}
	case int64:
		switch a := actual.(type) {
		case int64:
			return e == a
		case float64:
			return float64(e) == a
		case int:
			return e == int64(a)
		}
	case string:
		if a, ok := actual.(string); ok {
			return e == a
		}
	case bool:
		if a, ok := actual.(bool); ok {
			return e == a
		}
	}

	// Fallback to reflect for complex types (rare case)
	return reflect.DeepEqual(expected, actual)
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
// Optimized version with minimal allocations and checks.
func matchV2(query QueryV2, stub *Stub) bool {
	// Fast path: check headers first (most common failure case)
	if len(query.Headers) > 0 && !matchHeaders(query.Headers, stub.Headers) {
		return false
	}

	// Fast path: unary case (most common case)
	if len(stub.Stream) == 0 && len(query.Input) == 1 {
		return matchInput(query.Input[0], stub.Input)
	}

	// Stream case
	if len(stub.Stream) > 0 {
		return matchStreamElements(query.Input, stub.Stream)
	}

	// Multiple stream items but no stream in stub - no match
	return false
}

// rankMatchV2 ranks how well a given QueryV2 matches a given stub.
// Optimized version with minimal allocations and checks.
func rankMatchV2(query QueryV2, stub *Stub) float64 {
	// Fast path: unary case (most common case)
	if len(stub.Stream) == 0 && len(query.Input) == 1 {
		return rankHeaders(query.Headers, stub.Headers) + rankInput(query.Input[0], stub.Input)
	}

	// Stream case
	if len(stub.Stream) > 0 {
		return rankHeaders(query.Headers, stub.Headers) + rankStreamElements(query.Input, stub.Stream)
	}

	// Multiple stream items but no stream in stub - no rank
	return rankHeaders(query.Headers, stub.Headers)
}

// matchStreamElements checks if the query stream matches the stub stream.
//
//nolint:gocognit,cyclop,funlen
func matchStreamElements(queryStream []map[string]any, stubStream []InputData) bool {
	// For client streaming, grpctestify sends an extra empty message at the end
	// We need to handle this case by checking if the last message is empty
	effectiveQueryLength := len(queryStream)
	if effectiveQueryLength > 0 {
		lastMessage := queryStream[effectiveQueryLength-1]
		if len(lastMessage) == 0 {
			effectiveQueryLength--
		}
	}

	// For bidirectional streaming, we need to handle single messages
	// If query has only one message, try to match it against any stub item
	if effectiveQueryLength == 1 && len(stubStream) > 1 {
		queryItem := queryStream[0]

		// Try to match against any stub item
		for _, stubItem := range stubStream {
			// Check if this stub item has any matchers defined
			hasMatchers := len(stubItem.Equals) > 0 || len(stubItem.Contains) > 0 || len(stubItem.Matches) > 0
			if !hasMatchers {
				continue
			}

			// Check equals matcher
			if len(stubItem.Equals) > 0 && equals(stubItem.Equals, queryItem, stubItem.IgnoreArrayOrder) {
				return true
			}

			// Check contains matcher
			if len(stubItem.Contains) > 0 && contains(stubItem.Contains, queryItem, stubItem.IgnoreArrayOrder) {
				return true
			}

			// Check matches matcher
			if len(stubItem.Matches) > 0 && matches(stubItem.Matches, queryItem, stubItem.IgnoreArrayOrder) {
				return true
			}
		}

		return false
	}

	// For client streaming, allow partial matching for ranking purposes
	// Length mismatch is handled in ranking function

	// STRICT: If query stream is empty but stub expects data, no match
	if effectiveQueryLength == 0 && len(stubStream) > 0 {
		return false
	}

	for i := range effectiveQueryLength {
		queryItem := queryStream[i]
		stubItem := stubStream[i]

		// Check if this stub item has any matchers defined
		hasMatchers := len(stubItem.Equals) > 0 || len(stubItem.Contains) > 0 || len(stubItem.Matches) > 0
		if !hasMatchers {
			return false
		}

		// Check equals matcher
		if len(stubItem.Equals) > 0 {
			if !equals(stubItem.Equals, queryItem, stubItem.IgnoreArrayOrder) {
				return false
			}
		}

		// Check contains matcher
		if len(stubItem.Contains) > 0 {
			if !contains(stubItem.Contains, queryItem, stubItem.IgnoreArrayOrder) {
				return false
			}
		}

		// Check matches matcher
		if len(stubItem.Matches) > 0 {
			if !matches(stubItem.Matches, queryItem, stubItem.IgnoreArrayOrder) {
				return false
			}
		}
	}

	return true
}

// rankStreamElements ranks the match between query stream and stub stream.
//
//nolint:gocognit,cyclop,funlen
func rankStreamElements(queryStream []map[string]any, stubStream []InputData) float64 {
	// For client streaming, grpctestify sends an extra empty message at the end
	// We need to handle this case by checking if the last message is empty
	effectiveQueryLength := len(queryStream)
	if effectiveQueryLength > 0 {
		lastMessage := queryStream[effectiveQueryLength-1]
		if len(lastMessage) == 0 {
			effectiveQueryLength--
		}
	}

	// For bidirectional streaming, we need to handle single messages
	// If query has only one message, try to rank it against any stub item
	if effectiveQueryLength == 1 && len(stubStream) > 1 {
		queryItem := queryStream[0]

		var bestRank float64
		// Try to rank against any stub item
		for _, stubItem := range stubStream {
			// Check if this stub item has any matchers defined
			hasMatchers := len(stubItem.Equals) > 0 || len(stubItem.Contains) > 0 || len(stubItem.Matches) > 0
			if !hasMatchers {
				continue
			}

			// Use the same logic as before for element rank
			equalsRank := 0.0

			if len(stubItem.Equals) > 0 && equals(stubItem.Equals, queryItem, stubItem.IgnoreArrayOrder) {
				equalsRank = 1.0
			}

			containsRank := deeply.RankMatch(stubItem.Contains, queryItem)
			matchesRank := deeply.RankMatch(stubItem.Matches, queryItem)
			elementRank := equalsRank*100.0 + containsRank*0.1 + matchesRank*0.1 //nolint:mnd

			if elementRank > bestRank {
				bestRank = elementRank
			}
		}

		// Give bonus for bidirectional streaming match
		bidirectionalBonus := 500.0
		finalRank := bestRank + bidirectionalBonus

		return finalRank
	}

	// For client streaming, if lengths don't match, return very low rank
	if effectiveQueryLength != len(stubStream) {
		// For client streaming, length must match exactly
		return 0.1 //nolint:mnd
	}

	// STRICT: If query stream is empty but stub expects data, no rank
	if effectiveQueryLength == 0 && len(stubStream) > 0 {
		return 0
	}

	var (
		totalRank      float64
		perfectMatches int
	)

	for i := range effectiveQueryLength {
		queryItem := queryStream[i]
		stubItem := stubStream[i]
		// Use the same logic as before for element rank
		equalsRank := 0.0

		if len(stubItem.Equals) > 0 {
			if equals(stubItem.Equals, queryItem, stubItem.IgnoreArrayOrder) {
				equalsRank = 1.0
			} else {
				equalsRank = 0.0
			}
		}

		containsRank := deeply.RankMatch(stubItem.Contains, queryItem)
		matchesRank := deeply.RankMatch(stubItem.Matches, queryItem)
		elementRank := equalsRank*100.0 + containsRank*0.1 + matchesRank*0.1 //nolint:mnd
		totalRank += elementRank

		if equalsRank > 0.99 { //nolint:mnd
			perfectMatches++
		}
	}
	// For client streaming, accumulate rank based on received messages
	// Each message contributes to the total rank
	//nolint:mnd
	lengthBonus := float64(effectiveQueryLength) * 10.0 // Moderate bonus for length
	//nolint:mnd
	perfectMatchBonus := float64(perfectMatches) * 1000.0 // High bonus for perfect matches

	// Give bonus for complete match (all received messages match perfectly)
	completeMatchBonus := 0.0
	if perfectMatches == effectiveQueryLength && effectiveQueryLength > 0 {
		completeMatchBonus = 10000.0 // Very high bonus for complete match
	}

	// Add specificity bonus - more specific matchers = higher specificity
	specificityBonus := 0.0

	for _, stubItem := range stubStream {
		// Count actual matchers, not just field count
		equalsCount := 0

		for _, v := range stubItem.Equals {
			if v != nil {
				equalsCount++
			}
		}

		containsCount := 0

		for _, v := range stubItem.Contains {
			if v != nil {
				containsCount++
			}
		}

		matchesCount := 0

		for _, v := range stubItem.Matches {
			if v != nil {
				matchesCount++
			}
		}

		specificityBonus += float64(equalsCount + containsCount + matchesCount)
	}

	specificityBonus *= 50.0 // Medium weight for specificity

	finalRank := totalRank + lengthBonus + perfectMatchBonus + completeMatchBonus + specificityBonus

	return finalRank
}
