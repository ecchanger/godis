package consistenthash

import (
	"hash/crc32"
	"strconv"
	"testing"
)

func TestNew(t *testing.T) {
	// Test with default hash function
	m := New(3, nil)
	if m.replicas != 3 {
		t.Errorf("Expected replicas to be 3, got %d", m.replicas)
	}
	if m.hashFunc == nil {
		t.Error("Hash function should be set to default")
	}
	if len(m.keys) != 0 {
		t.Error("Keys should be empty initially")
	}
	if len(m.hashMap) != 0 {
		t.Error("Hash map should be empty initially")
	}

	// Test with custom hash function
	customHash := func(data []byte) uint32 {
		return crc32.ChecksumIEEE(data)
	}
	m2 := New(5, customHash)
	if m2.replicas != 5 {
		t.Errorf("Expected replicas to be 5, got %d", m2.replicas)
	}
}

func TestIsEmpty(t *testing.T) {
	m := New(3, nil)
	
	// Should be empty initially
	if !m.IsEmpty() {
		t.Error("Map should be empty initially")
	}
	
	// Should not be empty after adding nodes
	m.AddNode("a")
	if m.IsEmpty() {
		t.Error("Map should not be empty after adding nodes")
	}
}

func TestAddNode(t *testing.T) {
	m := New(3, nil)
	
	// Test adding single node
	m.AddNode("a")
	if len(m.keys) != 3 {
		t.Errorf("Expected 3 keys for 1 node with 3 replicas, got %d", len(m.keys))
	}
	if len(m.hashMap) != 3 {
		t.Errorf("Expected 3 entries in hash map, got %d", len(m.hashMap))
	}
	
	// Test adding multiple nodes
	m.AddNode("b", "c")
	if len(m.keys) != 9 {
		t.Errorf("Expected 9 keys for 3 nodes with 3 replicas, got %d", len(m.keys))
	}
	if len(m.hashMap) != 9 {
		t.Errorf("Expected 9 entries in hash map, got %d", len(m.hashMap))
	}
	
	// Test that keys are sorted
	for i := 1; i < len(m.keys); i++ {
		if m.keys[i-1] >= m.keys[i] {
			t.Error("Keys should be sorted in ascending order")
			break
		}
	}
}

func TestAddNodeEmptyKey(t *testing.T) {
	m := New(3, nil)
	
	// Test adding empty keys
	m.AddNode("", "a", "", "b")
	// Should only add non-empty keys
	if len(m.keys) != 6 { // 2 nodes * 3 replicas
		t.Errorf("Expected 6 keys for 2 non-empty nodes, got %d", len(m.keys))
	}
}

func TestPickNodeEmptyMap(t *testing.T) {
	m := New(3, nil)
	
	// Should return empty string for empty map
	result := m.PickNode("any_key")
	if result != "" {
		t.Errorf("Expected empty string for empty map, got %s", result)
	}
}

func TestPickNode(t *testing.T) {
	m := New(3, nil)
	m.AddNode("a", "b", "c")
	
	// Test that PickNode returns one of the added nodes
	for _, key := range []string{"test1", "test2", "test3"} {
		node := m.PickNode(key)
		if node != "a" && node != "b" && node != "c" {
			t.Errorf("PickNode should return one of the added nodes, got %s", node)
		}
	}
}

func TestPickNodeConsistency(t *testing.T) {
	m := New(3, nil)
	m.AddNode("a", "b", "c")
	
	// Same key should always return same node
	key := "test_key"
	node1 := m.PickNode(key)
	node2 := m.PickNode(key)
	if node1 != node2 {
		t.Errorf("Same key should return same node, got %s and %s", node1, node2)
	}
}

func TestGetPartitionKey(t *testing.T) {
	// Test key without hash tag
	key1 := "simple_key"
	result1 := getPartitionKey(key1)
	if result1 != key1 {
		t.Errorf("Expected %s, got %s", key1, result1)
	}
	
	// Test key with hash tag
	key2 := "user:123{abc}"
	result2 := getPartitionKey(key2)
	if result2 != "abc" {
		t.Errorf("Expected 'abc', got %s", result2)
	}
	
	// Test key with incomplete hash tag (no closing brace)
	key3 := "user:123{abc"
	result3 := getPartitionKey(key3)
	if result3 != key3 {
		t.Errorf("Expected %s, got %s", key3, result3)
	}
	
	// Test key with empty hash tag
	key4 := "user:123{}"
	result4 := getPartitionKey(key4)
	if result4 != key4 {
		t.Errorf("Expected %s, got %s", key4, result4)
	}
	
	// Test key with no opening brace
	key5 := "user:123abc}"
	result5 := getPartitionKey(key5)
	if result5 != key5 {
		t.Errorf("Expected %s, got %s", key5, result5)
	}
}

func TestHashTagSupport(t *testing.T) {
	m := New(3, nil)
	m.AddNode("a", "b", "c")
	
	// Keys with same hash tag should go to same node
	key1 := "user:123{abc}"
	key2 := "data:456{abc}"
	
	node1 := m.PickNode(key1)
	node2 := m.PickNode(key2)
	
	if node1 != node2 {
		t.Errorf("Keys with same hash tag should go to same node, got %s and %s", node1, node2)
	}
}

func TestHash(t *testing.T) {
	m := New(3, nil)
	m.AddNode("a", "b", "c", "d")
	if m.PickNode("zxc") != "a" {
		t.Error("wrong answer")
	}
	if m.PickNode("123{abc}") != "b" {
		t.Error("wrong answer")
	}
	if m.PickNode("abc") != "b" {
		t.Error("wrong answer")
	}
}

func TestDistribution(t *testing.T) {
	m := New(100, nil) // Use more replicas for better distribution
	nodes := []string{"node1", "node2", "node3", "node4"}
	m.AddNode(nodes...)
	
	// Count distribution of 1000 keys
	counts := make(map[string]int)
	for i := 0; i < 1000; i++ {
		key := "key" + strconv.Itoa(i)
		node := m.PickNode(key)
		counts[node]++
	}
	
	// Each node should get some keys (not a strict test, just ensuring basic distribution)
	for _, node := range nodes {
		if counts[node] == 0 {
			t.Errorf("Node %s got no keys", node)
		}
	}
	
	// No node should get more than 50% of keys (rough distribution check)
	for node, count := range counts {
		if count > 500 {
			t.Errorf("Node %s got too many keys: %d", node, count)
		}
	}
}

func TestCustomHashFunction(t *testing.T) {
	// Simple hash function that always returns the same value
	constantHash := func(data []byte) uint32 {
		return 42
	}
	
	m := New(3, constantHash)
	m.AddNode("a", "b", "c")
	
	// With constant hash, all keys should go to the same node
	node1 := m.PickNode("key1")
	node2 := m.PickNode("key2")
	node3 := m.PickNode("key3")
	
	if node1 != node2 || node2 != node3 {
		t.Error("With constant hash function, all keys should go to the same node")
	}
}

func TestAddNodeAfterPick(t *testing.T) {
	m := New(3, nil)
	m.AddNode("a", "b")
	
	// Pick a node for a key
	key := "test_key"
	originalNode := m.PickNode(key)
	
	// Add more nodes
	m.AddNode("c", "d")
	
	// The key might go to a different node now (which is expected in consistent hashing)
	// This test just ensures the system still works
	newNode := m.PickNode(key)
	if newNode == "" {
		t.Error("PickNode should still return a valid node after adding more nodes")
	}
	
	// Verify the node is one of the valid nodes
	validNodes := map[string]bool{"a": true, "b": true, "c": true, "d": true}
	if !validNodes[newNode] {
		t.Errorf("PickNode returned invalid node: %s", newNode)
	}
	
	t.Logf("Key %s: %s -> %s (after adding nodes)", key, originalNode, newNode)
}
