package utils

import (
	"regexp"
	"testing"
)

func TestRandString(t *testing.T) {
	// Test with length 0
	result := RandString(0)
	if len(result) != 0 {
		t.Errorf("Expected empty string for length 0, got %q", result)
	}

	// Test with positive length
	length := 10
	result = RandString(length)
	if len(result) != length {
		t.Errorf("Expected string of length %d, got %d", length, len(result))
	}

	// Test that result contains only valid characters
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9]*$`)
	if !validPattern.MatchString(result) {
		t.Errorf("String contains invalid characters: %q", result)
	}

	// Test that multiple calls return different strings (with high probability)
	result1 := RandString(20)
	result2 := RandString(20)
	if result1 == result2 {
		t.Log("Warning: Two random strings are identical (very low probability but possible)")
	}

	// Test with various lengths
	for _, length := range []int{1, 5, 50, 100} {
		result := RandString(length)
		if len(result) != length {
			t.Errorf("Expected string of length %d, got %d", length, len(result))
		}
		if !validPattern.MatchString(result) {
			t.Errorf("String contains invalid characters for length %d: %q", length, result)
		}
	}
}

func TestRandHexString(t *testing.T) {
	// Test with length 0
	result := RandHexString(0)
	if len(result) != 0 {
		t.Errorf("Expected empty string for length 0, got %q", result)
	}

	// Test with positive length
	length := 10
	result = RandHexString(length)
	if len(result) != length {
		t.Errorf("Expected string of length %d, got %d", length, len(result))
	}

	// Test that result contains only valid hex characters
	validHexPattern := regexp.MustCompile(`^[0-9a-f]*$`)
	if !validHexPattern.MatchString(result) {
		t.Errorf("String contains invalid hex characters: %q", result)
	}

	// Test that multiple calls return different strings (with high probability)
	result1 := RandHexString(20)
	result2 := RandHexString(20)
	if result1 == result2 {
		t.Log("Warning: Two random hex strings are identical (very low probability but possible)")
	}

	// Test with various lengths
	for _, length := range []int{1, 8, 16, 32} {
		result := RandHexString(length)
		if len(result) != length {
			t.Errorf("Expected hex string of length %d, got %d", length, len(result))
		}
		if !validHexPattern.MatchString(result) {
			t.Errorf("String contains invalid hex characters for length %d: %q", length, result)
		}
	}
}

func TestRandIndex(t *testing.T) {
	// Test with size 0
	result := RandIndex(0)
	if len(result) != 0 {
		t.Errorf("Expected empty slice for size 0, got %v", result)
	}

	// Test with size 1
	result = RandIndex(1)
	if len(result) != 1 || result[0] != 0 {
		t.Errorf("Expected [0] for size 1, got %v", result)
	}

	// Test with various sizes
	for _, size := range []int{2, 5, 10, 50} {
		result := RandIndex(size)
		
		// Check length
		if len(result) != size {
			t.Errorf("Expected slice of length %d, got %d", size, len(result))
			continue
		}

		// Check that all indices from 0 to size-1 are present
		found := make(map[int]bool)
		for _, index := range result {
			if index < 0 || index >= size {
				t.Errorf("Index %d is out of bounds for size %d", index, size)
				break
			}
			if found[index] {
				t.Errorf("Index %d appears multiple times", index)
				break
			}
			found[index] = true
		}

		// Check that all indices are present
		if len(found) != size {
			t.Errorf("Not all indices are present, expected %d, got %d", size, len(found))
		}
	}

	// Test that multiple calls return different permutations (with high probability)
	size := 10
	result1 := RandIndex(size)
	result2 := RandIndex(size)
	
	identical := true
	for i := 0; i < size; i++ {
		if result1[i] != result2[i] {
			identical = false
			break
		}
	}
	
	if identical {
		t.Log("Warning: Two random permutations are identical (very low probability but possible)")
	}
}

func TestRandIndexDistribution(t *testing.T) {
	// Test that RandIndex produces different permutations
	// This is a statistical test, so we use a smaller size and multiple runs
	size := 5
	runs := 100
	
	// Count how many times each index appears in the first position
	firstPositionCounts := make(map[int]int)
	
	for i := 0; i < runs; i++ {
		result := RandIndex(size)
		firstPositionCounts[result[0]]++
	}
	
	// Each index should appear in the first position at least once in 100 runs
	// (this is probabilistic, so it might fail very rarely)
	if len(firstPositionCounts) < size {
		t.Logf("Warning: Not all indices appeared in first position over %d runs", runs)
		t.Logf("First position counts: %v", firstPositionCounts)
	}
}

func TestRandStringCharacterSet(t *testing.T) {
	// Test that RandString uses the expected character set
	// We'll generate a large string and check that it contains characters from all expected categories
	length := 1000
	result := RandString(length)
	
	hasLowercase := false
	hasUppercase := false
	hasDigit := false
	
	for _, char := range result {
		if char >= 'a' && char <= 'z' {
			hasLowercase = true
		} else if char >= 'A' && char <= 'Z' {
			hasUppercase = true
		} else if char >= '0' && char <= '9' {
			hasDigit = true
		}
	}
	
	// With a string of length 1000, we should see all character types
	if !hasLowercase {
		t.Error("Generated string does not contain lowercase letters")
	}
	if !hasUppercase {
		t.Error("Generated string does not contain uppercase letters")
	}
	if !hasDigit {
		t.Error("Generated string does not contain digits")
	}
}

func TestRandHexStringCharacterSet(t *testing.T) {
	// Test that RandHexString uses only hex characters
	length := 1000
	result := RandHexString(length)
	
	hasDigit := false
	hasLowerHex := false
	
	for _, char := range result {
		if char >= '0' && char <= '9' {
			hasDigit = true
		} else if char >= 'a' && char <= 'f' {
			hasLowerHex = true
		} else {
			t.Errorf("Invalid hex character found: %c", char)
		}
	}
	
	// With a string of length 1000, we should see both digits and hex letters
	if !hasDigit {
		t.Error("Generated hex string does not contain digits")
	}
	if !hasLowerHex {
		t.Error("Generated hex string does not contain hex letters")
	}
}