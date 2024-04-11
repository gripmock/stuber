package stuber

import (
	"github.com/gripmock/deeply"
)

func match(query Query, stub *Stub) bool {
	return equals(stub.Input.Equals, query.Data, stub.Input.IgnoreArrayOrder) &&
		contains(stub.Input.Contains, query.Data, stub.Input.IgnoreArrayOrder) &&
		matches(stub.Input.Matches, query.Data, stub.Input.IgnoreArrayOrder) &&
		equals(stub.Headers.Equals, query.Data, false) &&
		contains(stub.Headers.Contains, query.Data, false) &&
		matches(stub.Headers.Matches, query.Data, false)
}

func rankMatch(query Query, stub *Stub) float64 {
	return deeply.RankMatch(stub.Input.Equals, query.Data) +
		deeply.RankMatch(stub.Input.Contains, query.Data) +
		deeply.RankMatch(stub.Input.Matches, query.Data)
}

func equals(expected map[string]any, actual any, orderIgnore bool) bool {
	if expected == nil || len(expected) == 0 {
		return true
	}

	if orderIgnore {
		return deeply.EqualsIgnoreArrayOrder(expected, actual)
	}

	return deeply.Equals(expected, actual)
}

func contains(expected map[string]any, actual any, _ bool) bool {
	if expected == nil || len(expected) == 0 {
		return true
	}

	return deeply.ContainsIgnoreArrayOrder(expected, actual)
}

func matches(expected map[string]any, actual any, _ bool) bool {
	if expected == nil || len(expected) == 0 {
		return true
	}

	return deeply.MatchesIgnoreArrayOrder(expected, actual)
}
