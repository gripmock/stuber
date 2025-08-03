package stuber

import (
	"reflect"

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

// rankInput ranks query data against stub input.
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

	// Convert actual to map if it's not already
	actualMap, ok := actual.(map[string]any)
	if !ok {
		return false
	}

	// Check if all expected fields are present and equal
	for key, expectedValue := range expected {
		actualValue, exists := actualMap[key]
		if !exists {
			return false
		}

		// Use simple string comparison for strings
		expectedStr, expectedIsStr := expectedValue.(string)

		actualStr, actualIsStr := actualValue.(string)
		if expectedIsStr && actualIsStr {
			if expectedStr != actualStr {
				return false
			}

			continue
		}

		// Use simple number comparison for numbers (handle both int and float64)
		switch expectedNum := expectedValue.(type) {
		case float64:
			switch actualNum := actualValue.(type) {
			case float64:
				if expectedNum != actualNum {
					return false
				}

				continue
			case int:
				if expectedNum != float64(actualNum) {
					return false
				}

				continue
			case int64:
				if expectedNum != float64(actualNum) {
					return false
				}

				continue
			}
		case int:
			switch actualNum := actualValue.(type) {
			case float64:
				if float64(expectedNum) != actualNum {
					return false
				}

				continue
			case int:
				if expectedNum != actualNum {
					return false
				}

				continue
			case int64:
				if expectedNum != int(actualNum) {
					return false
				}

				continue
			}
		case int64:
			switch actualNum := actualValue.(type) {
			case float64:
				if float64(expectedNum) != actualNum {
					return false
				}

				continue
			case int:
				if int64(actualNum) != expectedNum {
					return false
				}

				continue
			case int64:
				if expectedNum != actualNum {
					return false
				}

				continue
			}
		}

		// Use simple bool comparison for booleans
		expectedBool, expectedIsBool := expectedValue.(bool)

		actualBool, actualIsBool := actualValue.(bool)
		if expectedIsBool && actualIsBool {
			if expectedBool != actualBool {
				return false
			}

			continue
		}

		// For other types, check if we need to ignore array order
		if orderIgnore {
			// Use deeply package for array order insensitive comparison
			if !deeply.EqualsIgnoreArrayOrder(expectedValue, actualValue) {
				return false
			}
		} else {
			// Use reflect.DeepEqual for exact match
			if !reflect.DeepEqual(expectedValue, actualValue) {
				return false
			}
		}
	}

	// For streaming cases, we don't need to check extra fields
	// as the client might send additional fields that we don't care about
	return true
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

// matchStreamElements checks if query input matches stub stream element by element.
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
			if len(stubItem.Equals) > 0 {
				if equals(stubItem.Equals, queryItem, stubItem.IgnoreArrayOrder) {
					return true
				}
			}

			// Check contains matcher
			if len(stubItem.Contains) > 0 {
				if contains(stubItem.Contains, queryItem, stubItem.IgnoreArrayOrder) {
					return true
				}
			}

			// Check matches matcher
			if len(stubItem.Matches) > 0 {
				if matches(stubItem.Matches, queryItem, stubItem.IgnoreArrayOrder) {
					return true
				}
			}
		}

		return false
	}

	// For client streaming, allow partial matching for ranking purposes
	if effectiveQueryLength != len(stubStream) {
		// Don't return false here - let ranking handle it
	}

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

// rankStreamElements ranks how well query input matches stub stream element by element.
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
	lengthBonus := float64(effectiveQueryLength) * 10.0   // Moderate bonus for length // nolint:mnd
	perfectMatchBonus := float64(perfectMatches) * 1000.0 // High bonus for perfect matches // nolint:mnd

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
