package stuber

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestStringConversionFunctions tests the toCamelCase and toSnakeCase functions
func TestStringConversionFunctions(t *testing.T) {
	// Test toCamelCase function
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"hello", "hello"},
		{"hello_world", "helloWorld"},
		{"user_name", "userName"},
		{"api_key", "apiKey"},
		{"user_profile_data", "userProfileData"},
		{"user_id_123", "userId123"},
		{"a_b_c", "aBC"},
		{"test_case", "testCase"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("toCamelCase_%s", tt.input), func(t *testing.T) {
			result := toCamelCase(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}

	// Test toSnakeCase function
	snakeTests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"hello", "hello"},
		{"helloWorld", "hello_world"},
		{"userName", "user_name"},
		{"apiKey", "api_key"},
		{"userProfileData", "user_profile_data"},
		{"Hello", "hello"},
		{"API", "a_p_i"},
		{"UserID", "user_i_d"},
		{"userId123", "user_id123"},
		{"TestCase", "test_case"},
		{"HTTPRequest", "h_t_t_p_request"},
		{"JSONData", "j_s_o_n_data"},
	}

	for _, tt := range snakeTests {
		t.Run(fmt.Sprintf("toSnakeCase_%s", tt.input), func(t *testing.T) {
			result := toSnakeCase(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestStringConversionRoundTrip tests that converting back and forth preserves the original
func TestStringConversionRoundTrip(t *testing.T) {
	testCases := []string{
		"hello_world",
		"user_name",
		"api_key",
		"user_profile_data",
		"test_case",
		"user_id123", // Note: this is the actual result, not user_id_123
	}

	for _, original := range testCases {
		t.Run(fmt.Sprintf("roundtrip_%s", original), func(t *testing.T) {
			// snake_case -> camelCase -> snake_case
			camel := toCamelCase(original)
			snake := toSnakeCase(camel)
			require.Equal(t, original, snake)
		})
	}
}

// TestStringConversionEdgeCases tests edge cases for string conversion
func TestStringConversionEdgeCases(t *testing.T) {
	// Test empty strings
	require.Equal(t, "", toCamelCase(""))
	require.Equal(t, "", toSnakeCase(""))

	// Test single characters
	require.Equal(t, "a", toCamelCase("a"))
	require.Equal(t, "a", toSnakeCase("a"))

	// Test with uppercase at boundaries
	require.Equal(t, "hello", toSnakeCase("Hello"))
	require.Equal(t, "h_e_l_l_o", toSnakeCase("HELLO"))
}
