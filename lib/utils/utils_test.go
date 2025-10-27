package utils

import (
	"bytes"
	"reflect"
	"testing"
)

func TestToCmdLine(t *testing.T) {
	// Test empty input
	result := ToCmdLine()
	if len(result) != 0 {
		t.Errorf("Expected empty result for no input, got %v", result)
	}

	// Test single command
	result = ToCmdLine("GET")
	expected := [][]byte{[]byte("GET")}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	// Test multiple commands
	result = ToCmdLine("SET", "key", "value")
	expected = [][]byte{[]byte("SET"), []byte("key"), []byte("value")}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	// Test with empty strings
	result = ToCmdLine("", "test", "")
	expected = [][]byte{[]byte(""), []byte("test"), []byte("")}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestToCmdLine2(t *testing.T) {
	// Test with command name only
	result := ToCmdLine2("GET")
	expected := [][]byte{[]byte("GET")}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	// Test with command name and arguments
	result = ToCmdLine2("SET", "key", "value")
	expected = [][]byte{[]byte("SET"), []byte("key"), []byte("value")}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	// Test with empty arguments
	result = ToCmdLine2("DEL", "", "key2", "")
	expected = [][]byte{[]byte("DEL"), []byte(""), []byte("key2"), []byte("")}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestToCmdLine3(t *testing.T) {
	// Test with command name only
	result := ToCmdLine3("GET")
	expected := [][]byte{[]byte("GET")}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	// Test with command name and byte arguments
	result = ToCmdLine3("SET", []byte("key"), []byte("value"))
	expected = [][]byte{[]byte("SET"), []byte("key"), []byte("value")}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	// Test with empty byte slices
	result = ToCmdLine3("DEL", []byte(""), []byte("key2"))
	expected = [][]byte{[]byte("DEL"), []byte(""), []byte("key2")}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}

	// Test with nil byte slices
	result = ToCmdLine3("TEST", nil, []byte("value"))
	expected = [][]byte{[]byte("TEST"), nil, []byte("value")}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestEquals(t *testing.T) {
	// Test with byte slices
	a := []byte("hello")
	b := []byte("hello")
	if !Equals(a, b) {
		t.Error("Equal byte slices should return true")
	}

	// Test with different byte slices
	c := []byte("world")
	if Equals(a, c) {
		t.Error("Different byte slices should return false")
	}

	// Test with non-byte slice types
	if !Equals("hello", "hello") {
		t.Error("Equal strings should return true")
	}

	if Equals("hello", "world") {
		t.Error("Different strings should return false")
	}

	if !Equals(42, 42) {
		t.Error("Equal integers should return true")
	}

	if Equals(42, 43) {
		t.Error("Different integers should return false")
	}

	// Test mixed types
	if Equals([]byte("hello"), "hello") {
		t.Error("Different types should return false")
	}
}

func TestBytesEquals(t *testing.T) {
	// Test equal byte slices
	a := []byte("hello")
	b := []byte("hello")
	if !BytesEquals(a, b) {
		t.Error("Equal byte slices should return true")
	}

	// Test different byte slices
	c := []byte("world")
	if BytesEquals(a, c) {
		t.Error("Different byte slices should return false")
	}

	// Test different lengths
	d := []byte("hello world")
	if BytesEquals(a, d) {
		t.Error("Byte slices with different lengths should return false")
	}

	// Test empty byte slices
	empty1 := []byte{}
	empty2 := []byte{}
	if !BytesEquals(empty1, empty2) {
		t.Error("Empty byte slices should be equal")
	}

	// Test nil byte slices
	if !BytesEquals(nil, nil) {
		t.Error("Both nil byte slices should be equal")
	}

	// Test one nil, one non-nil
	if BytesEquals(nil, []byte("hello")) {
		t.Error("nil and non-nil byte slices should not be equal")
	}

	if BytesEquals([]byte("hello"), nil) {
		t.Error("non-nil and nil byte slices should not be equal")
	}

	// Test byte-by-byte comparison
	different := []byte("hellp") // Last character different
	if BytesEquals(a, different) {
		t.Error("Byte slices with different content should return false")
	}
}

func TestConvertRange(t *testing.T) {
	size := int64(10)

	// Test normal positive range
	start, end := ConvertRange(0, 5, size)
	if start != 0 || end != 6 {
		t.Errorf("Expected (0, 6), got (%d, %d)", start, end)
	}

	// Test negative indices
	start, end = ConvertRange(-1, -1, size)
	if start != 9 || end != 10 {
		t.Errorf("Expected (9, 10), got (%d, %d)", start, end)
	}

	// Test negative start, positive end
	start, end = ConvertRange(-5, 5, size)
	if start != 5 || end != 6 {
		t.Errorf("Expected (5, 6), got (%d, %d)", start, end)
	}

	// Test out of bounds start (too negative)
	start, end = ConvertRange(-15, 5, size)
	if start != -1 || end != -1 {
		t.Errorf("Expected (-1, -1), got (%d, %d)", start, end)
	}

	// Test out of bounds start (too positive)
	start, end = ConvertRange(15, 20, size)
	if start != -1 || end != -1 {
		t.Errorf("Expected (-1, -1), got (%d, %d)", start, end)
	}

	// Test out of bounds end (too negative)
	start, end = ConvertRange(0, -15, size)
	if start != -1 || end != -1 {
		t.Errorf("Expected (-1, -1), got (%d, %d)", start, end)
	}

	// Test end exceeding size
	start, end = ConvertRange(0, 15, size)
	if start != 0 || end != 10 {
		t.Errorf("Expected (0, 10), got (%d, %d)", start, end)
	}

	// Test start > end after conversion
	start, end = ConvertRange(5, 2, size)
	if start != -1 || end != -1 {
		t.Errorf("Expected (-1, -1), got (%d, %d)", start, end)
	}

	// Test full range
	start, end = ConvertRange(0, -1, size)
	if start != 0 || end != 10 {
		t.Errorf("Expected (0, 10), got (%d, %d)", start, end)
	}

	// Test single element
	start, end = ConvertRange(5, 5, size)
	if start != 5 || end != 6 {
		t.Errorf("Expected (5, 6), got (%d, %d)", start, end)
	}
}

func TestRemoveDuplicates(t *testing.T) {
	// Test with duplicates
	input := [][]byte{
		[]byte("hello"),
		[]byte("world"),
		[]byte("hello"),
		[]byte("test"),
		[]byte("world"),
	}
	result := RemoveDuplicates(input)
	
	// Check that duplicates are removed
	if len(result) != 3 {
		t.Errorf("Expected 3 unique elements, got %d", len(result))
	}

	// Check that all expected elements are present
	expected := map[string]bool{
		"hello": false,
		"world": false,
		"test":  false,
	}
	
	for _, item := range result {
		key := string(item)
		if _, exists := expected[key]; !exists {
			t.Errorf("Unexpected element: %s", key)
		}
		expected[key] = true
	}
	
	for key, found := range expected {
		if !found {
			t.Errorf("Missing expected element: %s", key)
		}
	}

	// Test with no duplicates
	noDuplicates := [][]byte{
		[]byte("a"),
		[]byte("b"),
		[]byte("c"),
	}
	result = RemoveDuplicates(noDuplicates)
	if len(result) != 3 {
		t.Errorf("Expected 3 elements when no duplicates, got %d", len(result))
	}

	// Test with empty input
	empty := [][]byte{}
	result = RemoveDuplicates(empty)
	if len(result) != 0 {
		t.Errorf("Expected 0 elements for empty input, got %d", len(result))
	}

	// Test with all same elements
	allSame := [][]byte{
		[]byte("same"),
		[]byte("same"),
		[]byte("same"),
	}
	result = RemoveDuplicates(allSame)
	if len(result) != 1 {
		t.Errorf("Expected 1 element when all same, got %d", len(result))
	}
	if !bytes.Equal(result[0], []byte("same")) {
		t.Errorf("Expected 'same', got %s", string(result[0]))
	}

	// Test with empty byte slices
	withEmpty := [][]byte{
		[]byte(""),
		[]byte("test"),
		[]byte(""),
		[]byte("test"),
	}
	result = RemoveDuplicates(withEmpty)
	if len(result) != 2 {
		t.Errorf("Expected 2 elements with empty bytes, got %d", len(result))
	}
}