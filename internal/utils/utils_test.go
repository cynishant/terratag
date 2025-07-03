package utils

import (
	"reflect"
	"testing"
)

func TestSortObjectKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		expected []string
	}{
		{
			name: "basic sorting",
			input: map[string]string{
				"zebra":       "value1",
				"apple":       "value2",
				"banana":      "value3",
				"cherry":      "value4",
			},
			expected: []string{"apple", "banana", "cherry", "zebra"},
		},
		{
			name: "mixed case sorting",
			input: map[string]string{
				"Zebra":       "value1",
				"apple":       "value2",
				"Banana":      "value3",
				"cherry":      "value4",
			},
			expected: []string{"Banana", "Zebra", "apple", "cherry"},
		},
		{
			name: "numbers and letters",
			input: map[string]string{
				"2_item":      "value1",
				"1_item":      "value2",
				"a_item":      "value3",
				"10_item":     "value4",
			},
			expected: []string{"10_item", "1_item", "2_item", "a_item"},
		},
		{
			name:     "empty map",
			input:    map[string]string{},
			expected: []string{},
		},
		{
			name: "single item",
			input: map[string]string{
				"only_item": "value",
			},
			expected: []string{"only_item"},
		},
		{
			name: "tag-like keys",
			input: map[string]string{
				"Environment": "prod",
				"Name":        "test-instance",
				"Team":        "platform",
				"CostCenter":  "CC1001",
				"Application": "web-server",
			},
			expected: []string{"Application", "CostCenter", "Environment", "Name", "Team"},
		},
		{
			name: "keys with special characters",
			input: map[string]string{
				"key-with-dashes": "value1",
				"key_with_under":  "value2",
				"key.with.dots":   "value3",
				"key:with:colons": "value4",
			},
			expected: []string{"key-with-dashes", "key.with.dots", "key:with:colons", "key_with_under"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SortObjectKeys(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("SortObjectKeys() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSortObjectKeysConsistency(t *testing.T) {
	// Test that sorting is consistent across multiple runs
	input := map[string]string{
		"delta":   "value1",
		"alpha":   "value2",
		"charlie": "value3",
		"bravo":   "value4",
	}

	result1 := SortObjectKeys(input)
	result2 := SortObjectKeys(input)
	result3 := SortObjectKeys(input)

	if !reflect.DeepEqual(result1, result2) {
		t.Error("SortObjectKeys() is not consistent between runs")
	}
	if !reflect.DeepEqual(result2, result3) {
		t.Error("SortObjectKeys() is not consistent between runs")
	}

	expected := []string{"alpha", "bravo", "charlie", "delta"}
	if !reflect.DeepEqual(result1, expected) {
		t.Errorf("SortObjectKeys() = %v, want %v", result1, expected)
	}
}

func TestSortObjectKeysPreservesOriginalMap(t *testing.T) {
	// Test that the original map is not modified
	original := map[string]string{
		"key3": "value3",
		"key1": "value1",
		"key2": "value2",
	}

	// Create a copy to compare later
	originalCopy := make(map[string]string)
	for k, v := range original {
		originalCopy[k] = v
	}

	_ = SortObjectKeys(original)

	// Original map should be unchanged
	if !reflect.DeepEqual(original, originalCopy) {
		t.Error("SortObjectKeys() modified the original map")
	}
}

func TestSortObjectKeysPerformance(t *testing.T) {
	// Test with a larger map to ensure reasonable performance
	largeMap := make(map[string]string)
	for i := 0; i < 1000; i++ {
		key := string(rune('a' + (i % 26))) + string(rune('a' + ((i/26) % 26))) + string(rune('0' + (i % 10)))
		largeMap[key] = "value" + string(rune('0' + (i % 10)))
	}

	result := SortObjectKeys(largeMap)
	
	if len(result) != len(largeMap) {
		t.Errorf("SortObjectKeys() result length %d doesn't match input length %d", len(result), len(largeMap))
	}

	// Verify all keys are present
	keySet := make(map[string]bool)
	for _, key := range result {
		keySet[key] = true
	}

	for key := range largeMap {
		if !keySet[key] {
			t.Errorf("SortObjectKeys() missing key %s in result", key)
		}
	}
}

func TestSortObjectKeysWithUnicodeKeys(t *testing.T) {
	// Test with Unicode characters
	input := map[string]string{
		"ñame":        "value1",
		"café":        "value2",
		"naïve":       "value3",
		"résumé":      "value4",
		"normal":      "value5",
	}

	result := SortObjectKeys(input)
	
	// Should not panic and should return all keys
	if len(result) != len(input) {
		t.Errorf("SortObjectKeys() returned %d keys, expected %d", len(result), len(input))
	}

	// Check all original keys are present
	keySet := make(map[string]bool)
	for _, key := range result {
		keySet[key] = true
	}

	for key := range input {
		if !keySet[key] {
			t.Errorf("SortObjectKeys() missing key %s in result", key)
		}
	}
}

func TestSortObjectKeysEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		validate func(t *testing.T, result []string)
	}{
		{
			name: "keys with only numbers",
			input: map[string]string{
				"123": "value1",
				"456": "value2",
				"789": "value3",
			},
			validate: func(t *testing.T, result []string) {
				expected := []string{"123", "456", "789"}
				if !reflect.DeepEqual(result, expected) {
					t.Errorf("Expected %v, got %v", expected, result)
				}
			},
		},
		{
			name: "keys with whitespace",
			input: map[string]string{
				" key with space": "value1",
				"key_no_space":    "value2",
				"  double_space":  "value3",
			},
			validate: func(t *testing.T, result []string) {
				if len(result) != 3 {
					t.Errorf("Expected 3 keys, got %d", len(result))
				}
			},
		},
		{
			name: "very long keys",
			input: map[string]string{
				"very_long_key_name_that_goes_on_and_on": "value1",
				"short":                                  "value2",
				"medium_length_key":                      "value3",
			},
			validate: func(t *testing.T, result []string) {
				if len(result) != 3 {
					t.Errorf("Expected 3 keys, got %d", len(result))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SortObjectKeys(tt.input)
			tt.validate(t, result)
		})
	}
}