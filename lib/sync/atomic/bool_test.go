package atomic

import (
	"sync"
	"testing"
)

func TestBooleanSetGet(t *testing.T) {
	var b Boolean

	// Test initial value (should be false)
	if b.Get() != false {
		t.Error("Initial value should be false")
	}

	// Test setting to true
	b.Set(true)
	if b.Get() != true {
		t.Error("Value should be true after Set(true)")
	}

	// Test setting to false
	b.Set(false)
	if b.Get() != false {
		t.Error("Value should be false after Set(false)")
	}

	// Test setting to true again
	b.Set(true)
	if b.Get() != true {
		t.Error("Value should be true after Set(true) again")
	}
}

func TestBooleanConcurrency(t *testing.T) {
	var b Boolean
	const numGoroutines = 100
	const numOperations = 1000

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // numGoroutines for setters, numGoroutines for getters

	// Start goroutines that set the value
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				// Alternate between true and false based on operation number
				b.Set(j%2 == 0)
			}
		}(i)
	}

	// Start goroutines that read the value
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				// Just read the value, don't care about the result
				// This tests that Get() is atomic and doesn't race with Set()
				_ = b.Get()
			}
		}(i)
	}

	wg.Wait()

	// The test passes if no race conditions are detected
	// The final value can be either true or false depending on timing
	finalValue := b.Get()
	t.Logf("Final value after concurrent operations: %v", finalValue)
}

func TestBooleanMultipleInstances(t *testing.T) {
	var b1, b2, b3 Boolean

	// Test that different instances are independent
	b1.Set(true)
	b2.Set(false)
	b3.Set(true)

	if b1.Get() != true {
		t.Error("b1 should be true")
	}
	if b2.Get() != false {
		t.Error("b2 should be false")
	}
	if b3.Get() != true {
		t.Error("b3 should be true")
	}

	// Change one and verify others are unchanged
	b2.Set(true)
	if b1.Get() != true {
		t.Error("b1 should still be true")
	}
	if b2.Get() != true {
		t.Error("b2 should now be true")
	}
	if b3.Get() != true {
		t.Error("b3 should still be true")
	}
}

func TestBooleanZeroValue(t *testing.T) {
	// Test that zero value of Boolean is false
	var b Boolean
	if b.Get() != false {
		t.Error("Zero value of Boolean should be false")
	}

	// Test with array of Booleans
	var bools [10]Boolean
	for i := range bools {
		if bools[i].Get() != false {
			t.Errorf("Boolean at index %d should be false", i)
		}
	}
}

func TestBooleanPointer(t *testing.T) {
	// Test Boolean used as pointer
	var b Boolean
	ptr := &b
	if ptr.Get() != false {
		t.Error("Pointer to Boolean should initially be false")
	}

	ptr.Set(true)
	if ptr.Get() != true {
		t.Error("Pointer to Boolean should be true after Set(true)")
	}
}

func TestBooleanStruct(t *testing.T) {
	// Test Boolean as struct field
	type TestStruct struct {
		Flag Boolean
		Name string
	}

	s := TestStruct{Name: "test"}
	if s.Flag.Get() != false {
		t.Error("Boolean field should initially be false")
	}

	s.Flag.Set(true)
	if s.Flag.Get() != true {
		t.Error("Boolean field should be true after Set(true)")
	}
	if s.Name != "test" {
		t.Error("Other fields should be unchanged")
	}
}

func TestBooleanRapidToggle(t *testing.T) {
	// Test rapid toggling between true and false
	var b Boolean
	const iterations = 10000

	for i := 0; i < iterations; i++ {
		expected := i%2 == 0
		b.Set(expected)
		actual := b.Get()
		if actual != expected {
			t.Errorf("Iteration %d: expected %v, got %v", i, expected, actual)
		}
	}
}

func TestBooleanConcurrentToggle(t *testing.T) {
	// Test concurrent toggling to ensure no race conditions
	var b Boolean
	const numGoroutines = 50
	const numToggles = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numToggles; j++ {
				// Read current value and toggle it
				current := b.Get()
				b.Set(!current)
			}
		}(i)
	}

	wg.Wait()

	// Test passes if no race conditions detected
	// Final value can be either true or false
	finalValue := b.Get()
	t.Logf("Final value after concurrent toggles: %v", finalValue)
}