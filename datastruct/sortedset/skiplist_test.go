package sortedset

import (
	"testing"
)

func TestRandomLevel(t *testing.T) {
	m := make(map[int16]int)
	for i := 0; i < 10000; i++ {
		level := randomLevel()
		m[level]++
	}
	for i := 0; i <= maxLevel; i++ {
		t.Logf("level %d, count %d", i, m[int16(i)])
	}
	
	// Verify that levels are within valid range
	for level := range m {
		if level < 1 || level > maxLevel {
			t.Errorf("Invalid level generated: %d", level)
		}
	}
	
	// Level 1 should have the most nodes (roughly half of all nodes)
	if m[1] == 0 {
		t.Error("Level 1 should have some nodes")
	}
}

func TestMakeNode(t *testing.T) {
	level := int16(3)
	score := 10.5
	member := "test_member"
	
	node := makeNode(level, score, member)
	
	if node.Score != score {
		t.Errorf("Expected score %f, got %f", score, node.Score)
	}
	if node.Member != member {
		t.Errorf("Expected member %s, got %s", member, node.Member)
	}
	if len(node.level) != int(level) {
		t.Errorf("Expected %d levels, got %d", level, len(node.level))
	}
	
	// Check that all levels are initialized
	for i, l := range node.level {
		if l == nil {
			t.Errorf("Level %d should not be nil", i)
		}
	}
}

func TestMakeSkiplist(t *testing.T) {
	sl := makeSkiplist()
	
	if sl.level != 1 {
		t.Errorf("Expected initial level 1, got %d", sl.level)
	}
	if sl.length != 0 {
		t.Errorf("Expected initial length 0, got %d", sl.length)
	}
	if sl.header == nil {
		t.Error("Header should not be nil")
	}
	if sl.tail != nil {
		t.Error("Tail should be nil initially")
	}
	if len(sl.header.level) != maxLevel {
		t.Errorf("Header should have %d levels, got %d", maxLevel, len(sl.header.level))
	}
}

func TestSkiplistInsertSingle(t *testing.T) {
	sl := makeSkiplist()
	
	member := "test"
	score := 10.0
	
	node := sl.insert(member, score)
	if node == nil {
		t.Error("Insert should return a node")
	}
	if node.Member != member {
		t.Errorf("Expected member %s, got %s", member, node.Member)
	}
	if node.Score != score {
		t.Errorf("Expected score %f, got %f", score, node.Score)
	}
	if sl.length != 1 {
		t.Errorf("Expected length 1, got %d", sl.length)
	}
	if sl.tail != node {
		t.Error("Tail should point to the inserted node")
	}
}

func TestSkiplistInsertMultiple(t *testing.T) {
	sl := makeSkiplist()
	
	// Insert nodes with different scores
	members := []string{"a", "b", "c", "d"}
	scores := []float64{1.0, 2.0, 3.0, 4.0}
	
	for i, member := range members {
		sl.insert(member, scores[i])
	}
	
	if sl.length != int64(len(members)) {
		t.Errorf("Expected length %d, got %d", len(members), sl.length)
	}
	
	// Check that nodes are in order
	current := sl.header.level[0].forward
	for i := 0; i < len(members); i++ {
		if current == nil {
			t.Errorf("Expected node at position %d", i)
			break
		}
		if current.Member != members[i] {
			t.Errorf("Expected member %s at position %d, got %s", members[i], i, current.Member)
		}
		if current.Score != scores[i] {
			t.Errorf("Expected score %f at position %d, got %f", scores[i], i, current.Score)
		}
		current = current.level[0].forward
	}
}

func TestSkiplistInsertSameScore(t *testing.T) {
	sl := makeSkiplist()
	
	score := 10.0
	members := []string{"a", "b", "c"}
	
	for _, member := range members {
		sl.insert(member, score)
	}
	
	// With same score, should be ordered by member name
	current := sl.header.level[0].forward
	for i, expectedMember := range members {
		if current == nil {
			t.Errorf("Expected node at position %d", i)
			break
		}
		if current.Member != expectedMember {
			t.Errorf("Expected member %s at position %d, got %s", expectedMember, i, current.Member)
		}
		current = current.level[0].forward
	}
}

func TestSkiplistRemove(t *testing.T) {
	sl := makeSkiplist()
	
	// Insert some nodes
	members := []string{"a", "b", "c"}
	scores := []float64{1.0, 2.0, 3.0}
	
	for i, member := range members {
		sl.insert(member, scores[i])
	}
	
	// Remove middle node
	removed := sl.remove("b", 2.0)
	if !removed {
		t.Error("Should have removed node 'b'")
	}
	if sl.length != 2 {
		t.Errorf("Expected length 2 after removal, got %d", sl.length)
	}
	
	// Check remaining nodes
	current := sl.header.level[0].forward
	expectedMembers := []string{"a", "c"}
	for i, expectedMember := range expectedMembers {
		if current == nil {
			t.Errorf("Expected node at position %d", i)
			break
		}
		if current.Member != expectedMember {
			t.Errorf("Expected member %s at position %d, got %s", expectedMember, i, current.Member)
		}
		current = current.level[0].forward
	}
}

func TestSkiplistRemoveNonexistent(t *testing.T) {
	sl := makeSkiplist()
	
	sl.insert("a", 1.0)
	
	// Try to remove non-existent node
	removed := sl.remove("b", 2.0)
	if removed {
		t.Error("Should not have removed non-existent node")
	}
	if sl.length != 1 {
		t.Errorf("Length should remain 1, got %d", sl.length)
	}
}

func TestSkiplistGetRank(t *testing.T) {
	sl := makeSkiplist()
	
	members := []string{"a", "b", "c", "d"}
	scores := []float64{1.0, 2.0, 3.0, 4.0}
	
	for i, member := range members {
		sl.insert(member, scores[i])
	}
	
	// Test rank retrieval (1-based)
	for i, member := range members {
		rank := sl.getRank(member, scores[i])
		expectedRank := int64(i + 1)
		if rank != expectedRank {
			t.Errorf("Expected rank %d for member %s, got %d", expectedRank, member, rank)
		}
	}
	
	// Test non-existent member
	rank := sl.getRank("z", 10.0)
	if rank != 0 {
		t.Errorf("Expected rank 0 for non-existent member, got %d", rank)
	}
}

func TestSkiplistGetByRank(t *testing.T) {
	sl := makeSkiplist()
	
	members := []string{"a", "b", "c", "d"}
	scores := []float64{1.0, 2.0, 3.0, 4.0}
	
	for i, member := range members {
		sl.insert(member, scores[i])
	}
	
	// Test retrieval by rank (1-based)
	for i, expectedMember := range members {
		rank := int64(i + 1)
		node := sl.getByRank(rank)
		if node == nil {
			t.Errorf("Expected node at rank %d", rank)
			continue
		}
		if node.Member != expectedMember {
			t.Errorf("Expected member %s at rank %d, got %s", expectedMember, rank, node.Member)
		}
	}
	
	// Test invalid ranks
	node := sl.getByRank(0)
	if node != nil && node != sl.header {
		t.Error("Rank 0 should return nil or header")
	}
	
	node = sl.getByRank(int64(len(members) + 1))
	if node != nil {
		t.Error("Expected nil for rank beyond length")
	}
}

func TestSkiplistHasInRange(t *testing.T) {
	sl := makeSkiplist()
	
	// Insert some nodes
	members := []string{"a", "b", "c", "d"}
	scores := []float64{1.0, 2.0, 3.0, 4.0}
	
	for i, member := range members {
		sl.insert(member, scores[i])
	}
	
	// Test range that includes some elements
	min := ScoreBorder{Value: 1.5, Exclude: false}
	max := ScoreBorder{Value: 3.5, Exclude: false}
	
	hasInRange := sl.hasInRange(&min, &max)
	if !hasInRange {
		t.Error("Should have elements in range [1.5, 3.5]")
	}
	
	// Test range that includes no elements
	min = ScoreBorder{Value: 5.0, Exclude: false}
	max = ScoreBorder{Value: 6.0, Exclude: false}
	
	hasInRange = sl.hasInRange(&min, &max)
	if hasInRange {
		t.Error("Should not have elements in range [5.0, 6.0]")
	}
	
	// Test invalid range (min > max)
	min = ScoreBorder{Value: 4.0, Exclude: false}
	max = ScoreBorder{Value: 1.0, Exclude: false}
	
	hasInRange = sl.hasInRange(&min, &max)
	if hasInRange {
		t.Error("Should not have elements in invalid range [4.0, 1.0]")
	}
}

func TestSkiplistGetFirstInRange(t *testing.T) {
	sl := makeSkiplist()
	
	// Insert some nodes
	members := []string{"a", "b", "c", "d"}
	scores := []float64{1.0, 2.0, 3.0, 4.0}
	
	for i, member := range members {
		sl.insert(member, scores[i])
	}
	
	// Test getting first in range [1.5, 3.5]
	min := ScoreBorder{Value: 1.5, Exclude: false}
	max := ScoreBorder{Value: 3.5, Exclude: false}
	
	node := sl.getFirstInRange(&min, &max)
	if node == nil {
		t.Error("Should find first node in range")
	} else if node.Member != "b" {
		t.Errorf("Expected first node to be 'b', got %s", node.Member)
	}
	
	// Test range with no elements
	min = ScoreBorder{Value: 5.0, Exclude: false}
	max = ScoreBorder{Value: 6.0, Exclude: false}
	
	node = sl.getFirstInRange(&min, &max)
	if node != nil {
		t.Error("Should not find node in empty range")
	}
}

func TestSkiplistLargeDataset(t *testing.T) {
	sl := makeSkiplist()
	
	// Insert many nodes
	count := 1000
	for i := 0; i < count; i++ {
		member := string(rune('a' + (i % 26))) + string(rune('a' + ((i/26) % 26)))
		score := float64(i)
		sl.insert(member, score)
	}
	
	if sl.length != int64(count) {
		t.Errorf("Expected length %d, got %d", count, sl.length)
	}
	
	// Test that we can find all nodes by rank
	for i := 1; i <= count; i++ {
		node := sl.getByRank(int64(i))
		if node == nil {
			t.Errorf("Should find node at rank %d", i)
		}
	}
	
	// Test that nodes are in correct order
	current := sl.header.level[0].forward
	prev_score := -1.0
	for i := 0; i < count && current != nil; i++ {
		if current.Score < prev_score {
			t.Errorf("Nodes not in correct order at position %d", i)
			break
		}
		prev_score = current.Score
		current = current.level[0].forward
	}
}
