package database

import (
	"testing"
	"time"

	"github.com/hdt3213/godis/interface/database"
	"github.com/hdt3213/godis/redis/connection"
)

// TestMakeDB tests DB creation
func TestMakeDB(t *testing.T) {
	db := makeDB()
	if db == nil {
		t.Error("makeDB returned nil")
	}
	if db.data == nil {
		t.Error("data dict is nil")
	}
	if db.ttlMap == nil {
		t.Error("ttlMap is nil")
	}
	if db.versionMap == nil {
		t.Error("versionMap is nil")
	}
}

// TestPutEntity tests putting entity into DB
func TestPutEntity(t *testing.T) {
	db := makeDB()
	key := "test_key"
	value := "test_value"
	entity := &database.DataEntity{
		Data: value,
	}

	result := db.PutEntity(key, entity)
	if result != 1 {
		t.Errorf("Expected 1, got %d", result)
	}

	// Get the entity back
	retrieved, exists := db.GetEntity(key)
	if !exists {
		t.Error("Entity not found")
	}
	if retrieved.Data != value {
		t.Errorf("Expected %s, got %v", value, retrieved.Data)
	}
}

// TestPutIfExists tests PutIfExists functionality
func TestPutIfExists(t *testing.T) {
	db := makeDB()
	key := "test_key"

	// Try to update non-existent key
	entity1 := &database.DataEntity{Data: "value1"}
	result := db.PutIfExists(key, entity1)
	if result != 0 {
		t.Errorf("Expected 0 for non-existent key, got %d", result)
	}

	// Put initial value
	db.PutEntity(key, entity1)

	// Update existing key
	entity2 := &database.DataEntity{Data: "value2"}
	result = db.PutIfExists(key, entity2)
	if result != 1 {
		t.Errorf("Expected 1 for existing key update, got %d", result)
	}

	// Verify the update
	retrieved, exists := db.GetEntity(key)
	if !exists {
		t.Error("Entity not found")
	}
	if retrieved.Data != "value2" {
		t.Errorf("Expected value2, got %v", retrieved.Data)
	}
}

// TestPutIfAbsent tests PutIfAbsent functionality
func TestPutIfAbsent(t *testing.T) {
	db := makeDB()
	key := "test_key"

	// Put new key
	entity1 := &database.DataEntity{Data: "value1"}
	result := db.PutIfAbsent(key, entity1)
	if result != 1 {
		t.Errorf("Expected 1 for new key, got %d", result)
	}

	// Try to put again with same key
	entity2 := &database.DataEntity{Data: "value2"}
	result = db.PutIfAbsent(key, entity2)
	if result != 0 {
		t.Errorf("Expected 0 for existing key, got %d", result)
	}

	// Verify original value remains
	retrieved, exists := db.GetEntity(key)
	if !exists {
		t.Error("Entity not found")
	}
	if retrieved.Data != "value1" {
		t.Errorf("Expected value1, got %v", retrieved.Data)
	}
}

// TestRemove tests removing entity from DB
func TestRemove(t *testing.T) {
	db := makeDB()
	key := "test_key"
	entity := &database.DataEntity{Data: "test_value"}

	db.PutEntity(key, entity)
	db.Remove(key)

	_, exists := db.GetEntity(key)
	if exists {
		t.Error("Entity should be removed")
	}
}

// TestRemoves tests removing multiple entities
func TestRemoves(t *testing.T) {
	db := makeDB()

	// Add multiple keys
	keys := []string{"key1", "key2", "key3"}
	for _, key := range keys {
		entity := &database.DataEntity{Data: key}
		db.PutEntity(key, entity)
	}

	// Remove all keys
	deleted := db.Removes(keys...)
	if deleted != 3 {
		t.Errorf("Expected 3 deleted, got %d", deleted)
	}

	// Verify all keys are removed
	for _, key := range keys {
		_, exists := db.GetEntity(key)
		if exists {
			t.Errorf("Key %s should be removed", key)
		}
	}
}

// TestRemovesNonExistent tests removing non-existent keys
func TestRemovesNonExistent(t *testing.T) {
	db := makeDB()

	deleted := db.Removes("key1", "key2", "key3")
	if deleted != 0 {
		t.Errorf("Expected 0 deleted for non-existent keys, got %d", deleted)
	}
}

// TestFlush tests flushing the database
func TestFlush(t *testing.T) {
	db := makeDB()

	// Add some data
	for i := 0; i < 10; i++ {
		key := "key" + string(rune(i))
		entity := &database.DataEntity{Data: i}
		db.PutEntity(key, entity)
	}

	// Flush
	db.Flush()

	// Verify all data is removed
	count := 0
	db.ForEach(func(key string, data *database.DataEntity, expiration *time.Time) bool {
		count++
		return true
	})

	if count != 0 {
		t.Errorf("Expected 0 keys after flush, got %d", count)
	}
}

// TestExpireInternal tests TTL functionality at DB level
func TestExpireInternal(t *testing.T) {
	db := makeDB()
	key := "test_key"
	entity := &database.DataEntity{Data: "test_value"}

	db.PutEntity(key, entity)

	// Set expiration to 100ms from now
	expireTime := time.Now().Add(100 * time.Millisecond)
	db.Expire(key, expireTime)

	// Key should still exist
	_, exists := db.GetEntity(key)
	if !exists {
		t.Error("Key should exist before expiration")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Key should be expired
	_, exists = db.GetEntity(key)
	if exists {
		t.Error("Key should be expired")
	}
}

// TestIsExpired tests IsExpired functionality
func TestIsExpired(t *testing.T) {
	db := makeDB()
	key := "test_key"
	entity := &database.DataEntity{Data: "test_value"}

	db.PutEntity(key, entity)

	// No expiration set
	if db.IsExpired(key) {
		t.Error("Key should not be expired")
	}

	// Set expiration in the past
	expireTime := time.Now().Add(-1 * time.Second)
	db.Expire(key, expireTime)

	// Key should be expired
	if !db.IsExpired(key) {
		t.Error("Key should be expired")
	}

	// Key should be removed after IsExpired check
	_, exists := db.GetEntity(key)
	if exists {
		t.Error("Expired key should be removed")
	}
}

// TestPersist tests canceling TTL
func TestPersist(t *testing.T) {
	db := makeDB()
	key := "test_key"
	entity := &database.DataEntity{Data: "test_value"}

	db.PutEntity(key, entity)

	// Set expiration
	expireTime := time.Now().Add(100 * time.Millisecond)
	db.Expire(key, expireTime)

	// Cancel expiration
	db.Persist(key)

	// Wait past original expiration time
	time.Sleep(150 * time.Millisecond)

	// Key should still exist
	_, exists := db.GetEntity(key)
	if !exists {
		t.Error("Key should still exist after persist")
	}
}

// TestGetVersion tests version management
func TestGetVersion(t *testing.T) {
	db := makeDB()
	key := "test_key"

	// Initial version should be 0
	version := db.GetVersion(key)
	if version != 0 {
		t.Errorf("Expected version 0, got %d", version)
	}

	// Add version
	db.addVersion(key)
	version = db.GetVersion(key)
	if version != 1 {
		t.Errorf("Expected version 1, got %d", version)
	}

	// Add version again
	db.addVersion(key)
	version = db.GetVersion(key)
	if version != 2 {
		t.Errorf("Expected version 2, got %d", version)
	}
}

// TestAddVersionMultipleKeys tests version management for multiple keys
func TestAddVersionMultipleKeys(t *testing.T) {
	db := makeDB()
	keys := []string{"key1", "key2", "key3"}

	db.addVersion(keys...)

	for _, key := range keys {
		version := db.GetVersion(key)
		if version != 1 {
			t.Errorf("Expected version 1 for %s, got %d", key, version)
		}
	}
}

// TestForEach tests ForEach functionality
func TestForEach(t *testing.T) {
	db := makeDB()

	// Add test data
	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for key, value := range testData {
		entity := &database.DataEntity{Data: value}
		db.PutEntity(key, entity)
	}

	// Count entries
	count := 0
	db.ForEach(func(key string, data *database.DataEntity, expiration *time.Time) bool {
		count++
		// Verify data
		expectedValue, exists := testData[key]
		if !exists {
			t.Errorf("Unexpected key: %s", key)
		}
		if data.Data != expectedValue {
			t.Errorf("Expected %s, got %v", expectedValue, data.Data)
		}
		// Verify no expiration
		if expiration != nil {
			t.Errorf("Expected no expiration for %s", key)
		}
		return true
	})

	if count != len(testData) {
		t.Errorf("Expected %d entries, got %d", len(testData), count)
	}
}

// TestForEachWithExpiration tests ForEach with TTL
func TestForEachWithExpiration(t *testing.T) {
	db := makeDB()

	key := "test_key"
	entity := &database.DataEntity{Data: "test_value"}
	db.PutEntity(key, entity)

	expireTime := time.Now().Add(10 * time.Second)
	db.Expire(key, expireTime)

	// Verify expiration is returned
	found := false
	db.ForEach(func(k string, data *database.DataEntity, expiration *time.Time) bool {
		if k == key {
			found = true
			if expiration == nil {
				t.Error("Expected expiration to be set")
			} else {
				// Allow some tolerance for time comparison
				diff := expiration.Sub(expireTime)
				if diff < -time.Millisecond || diff > time.Millisecond {
					t.Errorf("Expiration time mismatch")
				}
			}
		}
		return true
	})

	if !found {
		t.Error("Key not found in ForEach")
	}
}

// TestForEachEarlyExit tests ForEach early exit
func TestForEachEarlyExit(t *testing.T) {
	db := makeDB()

	// Add multiple keys
	for i := 0; i < 10; i++ {
		key := "key" + string(rune('0'+i))
		entity := &database.DataEntity{Data: i}
		db.PutEntity(key, entity)
	}

	// Stop after 5 iterations
	count := 0
	db.ForEach(func(key string, data *database.DataEntity, expiration *time.Time) bool {
		count++
		return count < 5
	})

	if count != 5 {
		t.Errorf("Expected 5 iterations, got %d", count)
	}
}

// TestGetEntityExpired tests getting expired entity
func TestGetEntityExpired(t *testing.T) {
	db := makeDB()
	key := "test_key"
	entity := &database.DataEntity{Data: "test_value"}

	db.PutEntity(key, entity)

	// Set expiration in the past
	expireTime := time.Now().Add(-1 * time.Second)
	db.Expire(key, expireTime)

	// GetEntity should return false for expired key
	_, exists := db.GetEntity(key)
	if exists {
		t.Error("GetEntity should return false for expired key")
	}
}

// TestValidateArity tests arity validation
func TestValidateArity(t *testing.T) {
	// Exact arity
	cmdArgs := [][]byte{[]byte("SET"), []byte("key"), []byte("value")}
	if !validateArity(3, cmdArgs) {
		t.Error("Expected true for exact arity")
	}
	if validateArity(2, cmdArgs) {
		t.Error("Expected false for wrong arity")
	}

	// Minimum arity (negative)
	cmdArgs = [][]byte{[]byte("DEL"), []byte("key1"), []byte("key2")}
	if !validateArity(-2, cmdArgs) {
		t.Error("Expected true for minimum arity")
	}
	cmdArgs = [][]byte{[]byte("DEL")}
	if validateArity(-2, cmdArgs) {
		t.Error("Expected false for insufficient arity")
	}
}

// TestConcurrentAccess tests concurrent access to DB
func TestConcurrentAccess(t *testing.T) {
	db := makeDB()
	done := make(chan bool)
	numGoroutines := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			key := "key" + string(rune('0'+id%10))
			entity := &database.DataEntity{Data: id}
			db.PutEntity(key, entity)
			done <- true
		}(i)
	}

	// Wait for all writes
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify data integrity
	count := 0
	db.ForEach(func(key string, data *database.DataEntity, expiration *time.Time) bool {
		count++
		return true
	})

	if count == 0 {
		t.Error("Expected some data after concurrent writes")
	}
}

// TestExecNormalCommand tests command execution
func TestExecNormalCommand(t *testing.T) {
	db := makeDB()

	// Test SET command
	cmdLine := [][]byte{[]byte("SET"), []byte("key1"), []byte("value1")}
	reply := db.execNormalCommand(cmdLine)
	if reply == nil {
		t.Error("Expected non-nil reply")
	}

	// Test GET command
	cmdLine = [][]byte{[]byte("GET"), []byte("key1")}
	reply = db.execNormalCommand(cmdLine)
	if reply == nil {
		t.Error("Expected non-nil reply")
	}

	// Test unknown command
	cmdLine = [][]byte{[]byte("UNKNOWN"), []byte("arg")}
	reply = db.execNormalCommand(cmdLine)
	if reply == nil {
		t.Error("Expected error reply for unknown command")
	}
}

// TestExec tests Exec with transaction commands
func TestExec(t *testing.T) {
	db := makeDB()
	conn := connection.NewFakeConn()

	// Test MULTI command
	cmdLine := [][]byte{[]byte("MULTI")}
	reply := db.Exec(conn, cmdLine)
	if reply == nil {
		t.Error("Expected non-nil reply for MULTI")
	}

	// Verify connection is in multi state
	if !conn.InMultiState() {
		t.Error("Connection should be in multi state after MULTI command")
	}

	// Test DISCARD command
	cmdLine = [][]byte{[]byte("DISCARD")}
	reply = db.Exec(conn, cmdLine)
	if reply == nil {
		t.Error("Expected non-nil reply for DISCARD")
	}

	// Verify connection is no longer in multi state
	if conn.InMultiState() {
		t.Error("Connection should not be in multi state after DISCARD")
	}
}

// TestMultipleExpiration tests multiple keys with different expiration times
func TestMultipleExpiration(t *testing.T) {
	db := makeDB()

	// Add keys with different expiration times
	db.PutEntity("key1", &database.DataEntity{Data: "value1"})
	db.Expire("key1", time.Now().Add(50*time.Millisecond))

	db.PutEntity("key2", &database.DataEntity{Data: "value2"})
	db.Expire("key2", time.Now().Add(150*time.Millisecond))

	db.PutEntity("key3", &database.DataEntity{Data: "value3"})
	// key3 has no expiration

	// Wait for key1 to expire
	time.Sleep(100 * time.Millisecond)

	_, exists := db.GetEntity("key1")
	if exists {
		t.Error("key1 should be expired")
	}

	_, exists = db.GetEntity("key2")
	if !exists {
		t.Error("key2 should still exist")
	}

	_, exists = db.GetEntity("key3")
	if !exists {
		t.Error("key3 should still exist")
	}

	// Wait for key2 to expire
	time.Sleep(100 * time.Millisecond)

	_, exists = db.GetEntity("key2")
	if exists {
		t.Error("key2 should be expired")
	}

	_, exists = db.GetEntity("key3")
	if !exists {
		t.Error("key3 should still exist (no expiration)")
	}
}
