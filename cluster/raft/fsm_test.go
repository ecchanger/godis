package raft

import (
	"bytes"
	"encoding/json"
	"io"
	"sort"
	"testing"

	"github.com/hashicorp/raft"
)

// TestFSMInit tests FSM initialization
func TestFSMInit(t *testing.T) {
	fsm := &FSM{
		Node2Slot:    make(map[string][]uint32),
		Slot2Node:    make(map[uint32]string),
		Migratings:   make(map[string]*MigratingTask),
		MasterSlaves: make(map[string]*MasterSlave),
		SlaveMasters: make(map[string]string),
		Failovers:    make(map[string]*FailoverTask),
	}

	if fsm.Node2Slot == nil {
		t.Error("Node2Slot is nil")
	}
	if fsm.Slot2Node == nil {
		t.Error("Slot2Node is nil")
	}
	if fsm.Migratings == nil {
		t.Error("Migratings is nil")
	}
	if fsm.MasterSlaves == nil {
		t.Error("MasterSlaves is nil")
	}
}

// TestApplySeedStart tests seed start event
func TestApplySeedStart(t *testing.T) {
	fsm := &FSM{
		Node2Slot:    make(map[string][]uint32),
		Slot2Node:    make(map[uint32]string),
		Migratings:   make(map[string]*MigratingTask),
		MasterSlaves: make(map[string]*MasterSlave),
		SlaveMasters: make(map[string]string),
		Failovers:    make(map[string]*FailoverTask),
	}

	entry := &LogEntry{
		Event: EventSeedStart,
		InitTask: &InitTask{
			Leader:    "node1",
			SlotCount: 16384,
		},
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}

	log := &raft.Log{
		Data: data,
	}

	fsm.Apply(log)

	// Verify all slots are assigned to node1
	if len(fsm.Node2Slot["node1"]) != 16384 {
		t.Errorf("Expected 16384 slots for node1, got %d", len(fsm.Node2Slot["node1"]))
	}

	// Verify slot to node mapping
	for i := 0; i < 16384; i++ {
		if fsm.Slot2Node[uint32(i)] != "node1" {
			t.Errorf("Slot %d should map to node1", i)
		}
	}

	// Verify node1 is added as master
	if fsm.MasterSlaves["node1"] == nil {
		t.Error("node1 should be added as master")
	}
}

// TestApplyStartMigrate tests start migrate event
func TestApplyStartMigrate(t *testing.T) {
	fsm := &FSM{
		Node2Slot:    make(map[string][]uint32),
		Slot2Node:    make(map[uint32]string),
		Migratings:   make(map[string]*MigratingTask),
		MasterSlaves: make(map[string]*MasterSlave),
		SlaveMasters: make(map[string]string),
		Failovers:    make(map[string]*FailoverTask),
	}

	task := &MigratingTask{
		ID:         "task1",
		SrcNode:    "node1",
		TargetNode: "node2",
		Slots:      []uint32{0, 1, 2},
	}

	entry := &LogEntry{
		Event:         EventStartMigrate,
		MigratingTask: task,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}

	log := &raft.Log{
		Data: data,
	}

	fsm.Apply(log)

	// Verify task is added to migratings
	if fsm.Migratings["task1"] == nil {
		t.Error("Task should be added to migratings")
	}

	if fsm.Migratings["task1"].ID != "task1" {
		t.Errorf("Expected task ID task1, got %s", fsm.Migratings["task1"].ID)
	}
}

// TestApplyFinishMigrate tests finish migrate event
func TestApplyFinishMigrate(t *testing.T) {
	fsm := &FSM{
		Node2Slot:    make(map[string][]uint32),
		Slot2Node:    make(map[uint32]string),
		Migratings:   make(map[string]*MigratingTask),
		MasterSlaves: make(map[string]*MasterSlave),
		SlaveMasters: make(map[string]string),
		Failovers:    make(map[string]*FailoverTask),
	}

	// Setup initial state
	fsm.Node2Slot["node1"] = []uint32{0, 1, 2, 3, 4}
	fsm.Node2Slot["node2"] = []uint32{}
	for _, slot := range []uint32{0, 1, 2, 3, 4} {
		fsm.Slot2Node[slot] = "node1"
	}

	task := &MigratingTask{
		ID:         "task1",
		SrcNode:    "node1",
		TargetNode: "node2",
		Slots:      []uint32{0, 1, 2},
	}
	fsm.Migratings["task1"] = task

	entry := &LogEntry{
		Event:         EventFinishMigrate,
		MigratingTask: task,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}

	log := &raft.Log{
		Data: data,
	}

	fsm.Apply(log)

	// Verify task is removed from migratings
	if fsm.Migratings["task1"] != nil {
		t.Error("Task should be removed from migratings")
	}

	// Verify slots are moved to target node
	for _, slot := range []uint32{0, 1, 2} {
		if fsm.Slot2Node[slot] != "node2" {
			t.Errorf("Slot %d should be assigned to node2", slot)
		}
	}

	// Verify node2 has the slots
	if len(fsm.Node2Slot["node2"]) != 3 {
		t.Errorf("node2 should have 3 slots, got %d", len(fsm.Node2Slot["node2"]))
	}

	// Verify node1 slots are removed
	if len(fsm.Node2Slot["node1"]) != 2 {
		t.Errorf("node1 should have 2 slots left, got %d", len(fsm.Node2Slot["node1"]))
	}
}

// TestApplyJoin tests join event
func TestApplyJoin(t *testing.T) {
	fsm := &FSM{
		Node2Slot:    make(map[string][]uint32),
		Slot2Node:    make(map[uint32]string),
		Migratings:   make(map[string]*MigratingTask),
		MasterSlaves: make(map[string]*MasterSlave),
		SlaveMasters: make(map[string]string),
		Failovers:    make(map[string]*FailoverTask),
	}

	// Add master first
	fsm.MasterSlaves["node1"] = &MasterSlave{
		MasterId: "node1",
		Slaves:   []string{},
	}

	// Join as slave
	entry := &LogEntry{
		Event: EventJoin,
		JoinTask: &JoinTask{
			NodeId: "node2",
			Master: "node1",
		},
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}

	log := &raft.Log{
		Data: data,
	}

	fsm.Apply(log)

	// Verify node2 is added as slave of node1
	if fsm.SlaveMasters["node2"] != "node1" {
		t.Errorf("node2 should be slave of node1, got %s", fsm.SlaveMasters["node2"])
	}

	// Verify node1's slaves include node2
	found := false
	for _, slave := range fsm.MasterSlaves["node1"].Slaves {
		if slave == "node2" {
			found = true
			break
		}
	}
	if !found {
		t.Error("node2 should be in node1's slaves")
	}
}

// TestApplyFailover tests failover event
func TestApplyFailover(t *testing.T) {
	fsm := &FSM{
		Node2Slot:    make(map[string][]uint32),
		Slot2Node:    make(map[uint32]string),
		Migratings:   make(map[string]*MigratingTask),
		MasterSlaves: make(map[string]*MasterSlave),
		SlaveMasters: make(map[string]string),
		Failovers:    make(map[string]*FailoverTask),
	}

	// Setup initial master-slave relationship
	fsm.MasterSlaves["node1"] = &MasterSlave{
		MasterId: "node1",
		Slaves:   []string{"node2", "node3"},
	}
	fsm.SlaveMasters["node2"] = "node1"
	fsm.SlaveMasters["node3"] = "node1"

	// Setup slots
	fsm.Node2Slot["node1"] = []uint32{0, 1, 2}
	for _, slot := range []uint32{0, 1, 2} {
		fsm.Slot2Node[slot] = "node1"
	}

	// Create failover task
	task := &FailoverTask{
		ID:          "failover1",
		OldMasterId: "node1",
		NewMasterId: "node2",
	}
	fsm.Failovers["failover1"] = task

	entry := &LogEntry{
		Event:        EventFinishFailover,
		FailoverTask: task,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}

	log := &raft.Log{
		Data: data,
	}

	fsm.Apply(log)

	// Verify node2 is now master
	if fsm.MasterSlaves["node2"] == nil {
		t.Error("node2 should be master")
	}

	// Verify node1 is now slave of node2
	if fsm.SlaveMasters["node1"] != "node2" {
		t.Errorf("node1 should be slave of node2, got %s", fsm.SlaveMasters["node1"])
	}

	// Verify node3 is now slave of node2
	if fsm.SlaveMasters["node3"] != "node2" {
		t.Errorf("node3 should be slave of node2, got %s", fsm.SlaveMasters["node3"])
	}

	// Verify slots are transferred to node2
	for _, slot := range []uint32{0, 1, 2} {
		if fsm.Slot2Node[slot] != "node2" {
			t.Errorf("Slot %d should be assigned to node2", slot)
		}
	}

	// Verify failover task is removed
	if fsm.Failovers["failover1"] != nil {
		t.Error("Failover task should be removed")
	}
}

// TestPickNode tests PickNode functionality
func TestPickNode(t *testing.T) {
	fsm := &FSM{
		Node2Slot:    make(map[string][]uint32),
		Slot2Node:    make(map[uint32]string),
		Migratings:   make(map[string]*MigratingTask),
		MasterSlaves: make(map[string]*MasterSlave),
		SlaveMasters: make(map[string]string),
		Failovers:    make(map[string]*FailoverTask),
	}

	fsm.Slot2Node[0] = "node1"
	fsm.Slot2Node[1] = "node2"

	node := fsm.PickNode(0)
	if node != "node1" {
		t.Errorf("Expected node1, got %s", node)
	}

	node = fsm.PickNode(1)
	if node != "node2" {
		t.Errorf("Expected node2, got %s", node)
	}
}

// TestGetMigratingTask tests GetMigratingTask
func TestGetMigratingTask(t *testing.T) {
	fsm := &FSM{
		Node2Slot:    make(map[string][]uint32),
		Slot2Node:    make(map[uint32]string),
		Migratings:   make(map[string]*MigratingTask),
		MasterSlaves: make(map[string]*MasterSlave),
		SlaveMasters: make(map[string]string),
		Failovers:    make(map[string]*FailoverTask),
	}

	task := &MigratingTask{
		ID:         "task1",
		SrcNode:    "node1",
		TargetNode: "node2",
		Slots:      []uint32{0, 1, 2},
	}
	fsm.Migratings["task1"] = task

	retrieved := fsm.GetMigratingTask("task1")
	if retrieved == nil {
		t.Error("Task should be found")
	}
	if retrieved.ID != "task1" {
		t.Errorf("Expected task1, got %s", retrieved.ID)
	}

	retrieved = fsm.GetMigratingTask("nonexistent")
	if retrieved != nil {
		t.Error("Non-existent task should return nil")
	}
}

// TestAddSlots tests addSlots functionality
func TestAddSlots(t *testing.T) {
	fsm := &FSM{
		Node2Slot:    make(map[string][]uint32),
		Slot2Node:    make(map[uint32]string),
		Migratings:   make(map[string]*MigratingTask),
		MasterSlaves: make(map[string]*MasterSlave),
		SlaveMasters: make(map[string]string),
		Failovers:    make(map[string]*FailoverTask),
	}

	fsm.Node2Slot["node1"] = []uint32{}
	slots := []uint32{5, 3, 1, 7}

	fsm.addSlots("node1", slots)

	// Verify all slots are added
	if len(fsm.Node2Slot["node1"]) != 4 {
		t.Errorf("Expected 4 slots, got %d", len(fsm.Node2Slot["node1"]))
	}

	// Verify slots are sorted
	if !sort.SliceIsSorted(fsm.Node2Slot["node1"], func(i, j int) bool {
		return fsm.Node2Slot["node1"][i] < fsm.Node2Slot["node1"][j]
	}) {
		t.Error("Slots should be sorted")
	}

	// Verify slot to node mapping
	for _, slot := range slots {
		if fsm.Slot2Node[slot] != "node1" {
			t.Errorf("Slot %d should map to node1", slot)
		}
	}

	// Test adding duplicate slots
	fsm.addSlots("node1", []uint32{3, 5})
	if len(fsm.Node2Slot["node1"]) != 4 {
		t.Errorf("Duplicate slots should not be added, expected 4, got %d", len(fsm.Node2Slot["node1"]))
	}
}

// TestRemoveSlots tests removeSlots functionality
func TestRemoveSlots(t *testing.T) {
	fsm := &FSM{
		Node2Slot:    make(map[string][]uint32),
		Slot2Node:    make(map[uint32]string),
		Migratings:   make(map[string]*MigratingTask),
		MasterSlaves: make(map[string]*MasterSlave),
		SlaveMasters: make(map[string]string),
		Failovers:    make(map[string]*FailoverTask),
	}

	fsm.Node2Slot["node1"] = []uint32{1, 3, 5, 7, 9}
	for _, slot := range []uint32{1, 3, 5, 7, 9} {
		fsm.Slot2Node[slot] = "node1"
	}

	fsm.removeSlots("node1", []uint32{3, 7})

	// Verify slots are removed
	if len(fsm.Node2Slot["node1"]) != 3 {
		t.Errorf("Expected 3 slots, got %d", len(fsm.Node2Slot["node1"]))
	}

	// Verify correct slots remain
	expected := []uint32{1, 5, 9}
	for i, slot := range expected {
		if fsm.Node2Slot["node1"][i] != slot {
			t.Errorf("Expected slot %d at index %d, got %d", slot, i, fsm.Node2Slot["node1"][i])
		}
	}

	// Verify slot to node mapping is removed
	if _, exists := fsm.Slot2Node[3]; exists {
		t.Error("Slot 3 should be removed from Slot2Node")
	}
	if _, exists := fsm.Slot2Node[7]; exists {
		t.Error("Slot 7 should be removed from Slot2Node")
	}
}

// TestGetMaster tests GetMaster functionality
func TestGetMaster(t *testing.T) {
	fsm := &FSM{
		Node2Slot:    make(map[string][]uint32),
		Slot2Node:    make(map[uint32]string),
		Migratings:   make(map[string]*MigratingTask),
		MasterSlaves: make(map[string]*MasterSlave),
		SlaveMasters: make(map[string]string),
		Failovers:    make(map[string]*FailoverTask),
	}

	fsm.SlaveMasters["node2"] = "node1"

	master := fsm.GetMaster("node2")
	if master != "node1" {
		t.Errorf("Expected node1, got %s", master)
	}

	// Test master node (should return empty string)
	master = fsm.GetMaster("node1")
	if master != "" {
		t.Errorf("Expected empty string for master, got %s", master)
	}
}

// TestSnapshot tests snapshot creation
func TestSnapshot(t *testing.T) {
	fsm := &FSM{
		Node2Slot:    make(map[string][]uint32),
		Slot2Node:    make(map[uint32]string),
		Migratings:   make(map[string]*MigratingTask),
		MasterSlaves: make(map[string]*MasterSlave),
		SlaveMasters: make(map[string]string),
		Failovers:    make(map[string]*FailoverTask),
	}

	// Setup state
	fsm.Slot2Node[0] = "node1"
	fsm.Slot2Node[1] = "node2"
	fsm.Migratings["task1"] = &MigratingTask{ID: "task1"}
	fsm.MasterSlaves["node1"] = &MasterSlave{MasterId: "node1"}

	snapshot, err := fsm.Snapshot()
	if err != nil {
		t.Fatal(err)
	}

	fsmSnapshot, ok := snapshot.(*FSMSnapshot)
	if !ok {
		t.Fatal("Expected FSMSnapshot")
	}

	// Verify snapshot contains correct data
	if len(fsmSnapshot.Slot2Node) != 2 {
		t.Errorf("Expected 2 slots in snapshot, got %d", len(fsmSnapshot.Slot2Node))
	}

	if len(fsmSnapshot.Migratings) != 1 {
		t.Errorf("Expected 1 migrating task in snapshot, got %d", len(fsmSnapshot.Migratings))
	}

	if len(fsmSnapshot.MasterSlaves) != 1 {
		t.Errorf("Expected 1 master in snapshot, got %d", len(fsmSnapshot.MasterSlaves))
	}
}

// mockSnapshotSink implements raft.SnapshotSink for testing
type mockSnapshotSink struct {
	buffer *bytes.Buffer
}

func (m *mockSnapshotSink) Write(b []byte) (int, error) {
	return m.buffer.Write(b)
}

func (m *mockSnapshotSink) Close() error {
	return nil
}

func (m *mockSnapshotSink) ID() string {
	return "mock"
}

func (m *mockSnapshotSink) Cancel() error {
	return nil
}

// TestSnapshotPersist tests snapshot persistence
func TestSnapshotPersist(t *testing.T) {
	snapshot := &FSMSnapshot{
		Slot2Node: map[uint32]string{
			0: "node1",
			1: "node2",
		},
		Migratings: map[string]*MigratingTask{
			"task1": {ID: "task1"},
		},
		MasterSlaves: map[string]*MasterSlave{
			"node1": {MasterId: "node1"},
		},
	}

	sink := &mockSnapshotSink{
		buffer: &bytes.Buffer{},
	}

	err := snapshot.Persist(sink)
	if err != nil {
		t.Fatal(err)
	}

	// Verify data was written
	if sink.buffer.Len() == 0 {
		t.Error("No data written to sink")
	}

	// Verify data is valid JSON
	var restored FSMSnapshot
	err = json.Unmarshal(sink.buffer.Bytes(), &restored)
	if err != nil {
		t.Fatal(err)
	}

	if len(restored.Slot2Node) != 2 {
		t.Errorf("Expected 2 slots in restored snapshot, got %d", len(restored.Slot2Node))
	}
}

// TestRestore tests snapshot restoration
func TestRestore(t *testing.T) {
	fsm := &FSM{
		Node2Slot:    make(map[string][]uint32),
		Slot2Node:    make(map[uint32]string),
		Migratings:   make(map[string]*MigratingTask),
		MasterSlaves: make(map[string]*MasterSlave),
		SlaveMasters: make(map[string]string),
		Failovers:    make(map[string]*FailoverTask),
	}

	snapshot := &FSMSnapshot{
		Slot2Node: map[uint32]string{
			0: "node1",
			1: "node1",
			2: "node2",
		},
		Migratings: map[string]*MigratingTask{
			"task1": {ID: "task1"},
		},
		MasterSlaves: map[string]*MasterSlave{
			"node1": {MasterId: "node1", Slaves: []string{"node3"}},
			"node2": {MasterId: "node2"},
		},
	}

	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}

	reader := io.NopCloser(bytes.NewReader(data))
	err = fsm.Restore(reader)
	if err != nil {
		t.Fatal(err)
	}

	// Verify Slot2Node is restored
	if len(fsm.Slot2Node) != 3 {
		t.Errorf("Expected 3 slots, got %d", len(fsm.Slot2Node))
	}

	// Verify Node2Slot is rebuilt
	if len(fsm.Node2Slot["node1"]) != 2 {
		t.Errorf("Expected 2 slots for node1, got %d", len(fsm.Node2Slot["node1"]))
	}
	if len(fsm.Node2Slot["node2"]) != 1 {
		t.Errorf("Expected 1 slot for node2, got %d", len(fsm.Node2Slot["node2"]))
	}

	// Verify Migratings is restored
	if len(fsm.Migratings) != 1 {
		t.Errorf("Expected 1 migrating task, got %d", len(fsm.Migratings))
	}

	// Verify MasterSlaves is restored
	if len(fsm.MasterSlaves) != 2 {
		t.Errorf("Expected 2 masters, got %d", len(fsm.MasterSlaves))
	}

	// Verify SlaveMasters is rebuilt
	if fsm.SlaveMasters["node3"] != "node1" {
		t.Errorf("Expected node3 to be slave of node1, got %s", fsm.SlaveMasters["node3"])
	}
}

// TestWithReadLock tests WithReadLock functionality
func TestWithReadLock(t *testing.T) {
	fsm := &FSM{
		Node2Slot:    make(map[string][]uint32),
		Slot2Node:    make(map[uint32]string),
		Migratings:   make(map[string]*MigratingTask),
		MasterSlaves: make(map[string]*MasterSlave),
		SlaveMasters: make(map[string]string),
		Failovers:    make(map[string]*FailoverTask),
	}

	fsm.Slot2Node[0] = "node1"

	var node string
	fsm.WithReadLock(func(f *FSM) {
		node = f.Slot2Node[0]
	})

	if node != "node1" {
		t.Errorf("Expected node1, got %s", node)
	}
}

// TestAddNode tests addNode functionality
func TestAddNode(t *testing.T) {
	fsm := &FSM{
		Node2Slot:    make(map[string][]uint32),
		Slot2Node:    make(map[uint32]string),
		Migratings:   make(map[string]*MigratingTask),
		MasterSlaves: make(map[string]*MasterSlave),
		SlaveMasters: make(map[string]string),
		Failovers:    make(map[string]*FailoverTask),
	}

	// Add master node
	err := fsm.addNode("node1", "")
	if err != nil {
		t.Fatal(err)
	}

	if fsm.MasterSlaves["node1"] == nil {
		t.Error("node1 should be added as master")
	}

	// Add slave node
	err = fsm.addNode("node2", "node1")
	if err != nil {
		t.Fatal(err)
	}

	if fsm.SlaveMasters["node2"] != "node1" {
		t.Errorf("node2 should be slave of node1, got %s", fsm.SlaveMasters["node2"])
	}

	// Verify node1's slaves include node2
	found := false
	for _, slave := range fsm.MasterSlaves["node1"].Slaves {
		if slave == "node2" {
			found = true
			break
		}
	}
	if !found {
		t.Error("node2 should be in node1's slaves")
	}

	// Test adding slave to non-existent master
	err = fsm.addNode("node3", "nonexistent")
	if err == nil {
		t.Error("Should return error when adding slave to non-existent master")
	}
}

// TestFailover tests failover functionality
func TestFailover(t *testing.T) {
	fsm := &FSM{
		Node2Slot:    make(map[string][]uint32),
		Slot2Node:    make(map[uint32]string),
		Migratings:   make(map[string]*MigratingTask),
		MasterSlaves: make(map[string]*MasterSlave),
		SlaveMasters: make(map[string]string),
		Failovers:    make(map[string]*FailoverTask),
	}

	// Setup master-slave relationship
	fsm.MasterSlaves["node1"] = &MasterSlave{
		MasterId: "node1",
		Slaves:   []string{"node2", "node3", "node4"},
	}
	fsm.SlaveMasters["node2"] = "node1"
	fsm.SlaveMasters["node3"] = "node1"
	fsm.SlaveMasters["node4"] = "node1"

	// Failover from node1 to node2
	fsm.failover("node1", "node2")

	// Verify node2 is now master
	if fsm.MasterSlaves["node2"] == nil {
		t.Error("node2 should be master")
	}

	// Verify node1 is now slave of node2
	if fsm.SlaveMasters["node1"] != "node2" {
		t.Errorf("node1 should be slave of node2, got %s", fsm.SlaveMasters["node1"])
	}

	// Verify node3 and node4 are still slaves but of node2 now
	if fsm.SlaveMasters["node3"] != "node2" {
		t.Errorf("node3 should be slave of node2, got %s", fsm.SlaveMasters["node3"])
	}
	if fsm.SlaveMasters["node4"] != "node2" {
		t.Errorf("node4 should be slave of node2, got %s", fsm.SlaveMasters["node4"])
	}

	// Verify node2's slaves include node1, node3, node4
	expectedSlaves := map[string]bool{"node1": true, "node3": true, "node4": true}
	for _, slave := range fsm.MasterSlaves["node2"].Slaves {
		if !expectedSlaves[slave] {
			t.Errorf("Unexpected slave %s", slave)
		}
		delete(expectedSlaves, slave)
	}
	if len(expectedSlaves) != 0 {
		t.Errorf("Missing expected slaves: %v", expectedSlaves)
	}

	// Verify node1 is removed as master
	if fsm.MasterSlaves["node1"] != nil {
		t.Error("node1 should no longer be a master")
	}
}
