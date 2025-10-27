package wait

import (
	"testing"
	"time"
)

func TestWaitBasic(t *testing.T) {
	var w Wait
	
	// Test that Wait() returns immediately when counter is 0
	done := make(chan bool, 1)
	go func() {
		w.Wait()
		done <- true
	}()
	
	select {
	case <-done:
		// Expected - Wait() should return immediately
	case <-time.After(100 * time.Millisecond):
		t.Error("Wait() should return immediately when counter is 0")
	}
}

func TestWaitAddDone(t *testing.T) {
	var w Wait
	
	// Add 1 to counter
	w.Add(1)
	
	// Start a goroutine that waits
	done := make(chan bool, 1)
	go func() {
		w.Wait()
		done <- true
	}()
	
	// Wait() should not return immediately
	select {
	case <-done:
		t.Error("Wait() should not return when counter > 0")
	case <-time.After(50 * time.Millisecond):
		// Expected - Wait() should block
	}
	
	// Call Done() to decrement counter
	w.Done()
	
	// Now Wait() should return
	select {
	case <-done:
		// Expected - Wait() should return after Done()
	case <-time.After(100 * time.Millisecond):
		t.Error("Wait() should return after Done() is called")
	}
}

func TestWaitMultipleAddDone(t *testing.T) {
	var w Wait
	const numWorkers = 5
	
	// Add multiple to counter
	w.Add(numWorkers)
	
	// Start a goroutine that waits
	done := make(chan bool, 1)
	go func() {
		w.Wait()
		done <- true
	}()
	
	// Wait() should not return yet
	select {
	case <-done:
		t.Error("Wait() should not return when counter > 0")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}
	
	// Call Done() multiple times, but not enough
	for i := 0; i < numWorkers-1; i++ {
		w.Done()
		
		// Should still be waiting
		select {
		case <-done:
			t.Error("Wait() should not return until all Done() calls are made")
		case <-time.After(10 * time.Millisecond):
			// Expected
		}
	}
	
	// Final Done() call
	w.Done()
	
	// Now Wait() should return
	select {
	case <-done:
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Wait() should return after all Done() calls are made")
	}
}

func TestWaitWithTimeoutNoTimeout(t *testing.T) {
	var w Wait
	
	// Test WaitWithTimeout when no timeout occurs
	timedOut := w.WaitWithTimeout(100 * time.Millisecond)
	if timedOut {
		t.Error("WaitWithTimeout should not timeout when counter is 0")
	}
}

func TestWaitWithTimeoutTimeout(t *testing.T) {
	var w Wait
	
	// Add to counter but don't call Done()
	w.Add(1)
	
	start := time.Now()
	timedOut := w.WaitWithTimeout(100 * time.Millisecond)
	elapsed := time.Since(start)
	
	if !timedOut {
		t.Error("WaitWithTimeout should timeout when Done() is not called")
	}
	
	// Check that it actually waited approximately the timeout duration
	if elapsed < 90*time.Millisecond || elapsed > 200*time.Millisecond {
		t.Errorf("Expected timeout around 100ms, got %v", elapsed)
	}
}

func TestWaitWithTimeoutCompleteBeforeTimeout(t *testing.T) {
	var w Wait
	
	w.Add(1)
	
	// Start a goroutine that will call Done() after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		w.Done()
	}()
	
	start := time.Now()
	timedOut := w.WaitWithTimeout(200 * time.Millisecond)
	elapsed := time.Since(start)
	
	if timedOut {
		t.Error("WaitWithTimeout should not timeout when Done() is called before timeout")
	}
	
	// Should complete in about 50ms, not 200ms
	if elapsed < 40*time.Millisecond || elapsed > 100*time.Millisecond {
		t.Errorf("Expected completion around 50ms, got %v", elapsed)
	}
}

func TestWaitConcurrentWorkers(t *testing.T) {
	var w Wait
	const numWorkers = 10
	const workDuration = 50 * time.Millisecond
	
	w.Add(numWorkers)
	
	// Start workers
	for i := 0; i < numWorkers; i++ {
		go func(id int) {
			time.Sleep(workDuration)
			w.Done()
		}(i)
	}
	
	start := time.Now()
	w.Wait()
	elapsed := time.Since(start)
	
	// All workers run concurrently, so total time should be close to workDuration
	if elapsed < 40*time.Millisecond || elapsed > 100*time.Millisecond {
		t.Errorf("Expected concurrent completion around %v, got %v", workDuration, elapsed)
	}
}

func TestWaitConcurrentWorkersWithTimeout(t *testing.T) {
	var w Wait
	const numWorkers = 5
	const workDuration = 30 * time.Millisecond
	
	w.Add(numWorkers)
	
	// Start workers
	for i := 0; i < numWorkers; i++ {
		go func(id int) {
			time.Sleep(workDuration)
			w.Done()
		}(i)
	}
	
	// Timeout is longer than work duration
	timedOut := w.WaitWithTimeout(100 * time.Millisecond)
	if timedOut {
		t.Error("WaitWithTimeout should not timeout when work completes before timeout")
	}
}

func TestWaitNegativeAdd(t *testing.T) {
	var w Wait
	
	// Add positive count
	w.Add(3)
	
	// Add negative count to partially reduce
	w.Add(-2)
	
	// Should still be waiting for 1 more Done()
	done := make(chan bool, 1)
	go func() {
		w.Wait()
		done <- true
	}()
	
	select {
	case <-done:
		t.Error("Wait() should not return when counter is still > 0")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}
	
	// One more Done() should complete it
	w.Done()
	
	select {
	case <-done:
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Wait() should return after final Done()")
	}
}

func TestWaitMultipleWaiters(t *testing.T) {
	var w Wait
	const numWaiters = 3
	
	w.Add(1)
	
	// Start multiple goroutines waiting
	waitersComplete := make(chan bool, numWaiters)
	for i := 0; i < numWaiters; i++ {
		go func(id int) {
			w.Wait()
			waitersComplete <- true
		}(i)
	}
	
	// No waiters should complete yet
	for i := 0; i < numWaiters; i++ {
		select {
		case <-waitersComplete:
			t.Error("Wait() should not return before Done() is called")
		case <-time.After(50 * time.Millisecond):
			// Expected
		}
	}
	
	// Call Done() once
	w.Done()
	
	// All waiters should complete
	for i := 0; i < numWaiters; i++ {
		select {
		case <-waitersComplete:
			// Expected
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Waiter %d should complete after Done()", i)
		}
	}
}

func TestWaitZeroTimeout(t *testing.T) {
	var w Wait
	
	w.Add(1)
	
	// Zero timeout should timeout immediately
	timedOut := w.WaitWithTimeout(0)
	if !timedOut {
		t.Error("WaitWithTimeout(0) should timeout immediately")
	}
}

func TestWaitReuse(t *testing.T) {
	var w Wait
	
	// First use
	w.Add(2)
	go func() {
		time.Sleep(20 * time.Millisecond)
		w.Done()
		w.Done()
	}()
	
	w.Wait()
	
	// Reuse the same Wait instance
	w.Add(1)
	go func() {
		time.Sleep(20 * time.Millisecond)
		w.Done()
	}()
	
	timedOut := w.WaitWithTimeout(100 * time.Millisecond)
	if timedOut {
		t.Error("Reused Wait should work correctly")
	}
}