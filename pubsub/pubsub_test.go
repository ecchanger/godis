package pubsub

import (
	"sync"
	"testing"

	"github.com/hdt3213/godis/redis/protocol"
)

// mockConnection implements redis.Connection for testing
type mockConnection struct {
	name         string
	channels     map[string]bool
	channelMutex sync.RWMutex
	writeBuffer  [][]byte
	writeMutex   sync.Mutex
}

func newMockConnection(name string) *mockConnection {
	return &mockConnection{
		name:        name,
		channels:    make(map[string]bool),
		writeBuffer: make([][]byte, 0),
	}
}

func (m *mockConnection) Write(b []byte) (int, error) {
	m.writeMutex.Lock()
	defer m.writeMutex.Unlock()
	m.writeBuffer = append(m.writeBuffer, b)
	return len(b), nil
}

func (m *mockConnection) Close() error {
	return nil
}

func (m *mockConnection) RemoteAddr() string {
	return "127.0.0.1:6379"
}

func (m *mockConnection) SetPassword(string) {}

func (m *mockConnection) GetPassword() string {
	return ""
}

func (m *mockConnection) Subscribe(channel string) {
	m.channelMutex.Lock()
	defer m.channelMutex.Unlock()
	m.channels[channel] = true
}

func (m *mockConnection) UnSubscribe(channel string) {
	m.channelMutex.Lock()
	defer m.channelMutex.Unlock()
	delete(m.channels, channel)
}

func (m *mockConnection) SubsCount() int {
	m.channelMutex.RLock()
	defer m.channelMutex.RUnlock()
	return len(m.channels)
}

func (m *mockConnection) GetChannels() []string {
	m.channelMutex.RLock()
	defer m.channelMutex.RUnlock()
	channels := make([]string, 0, len(m.channels))
	for ch := range m.channels {
		channels = append(channels, ch)
	}
	return channels
}

func (m *mockConnection) InMultiState() bool {
	return false
}

func (m *mockConnection) SetMultiState(bool) {}

func (m *mockConnection) GetQueuedCmdLine() [][][]byte {
	return nil
}

func (m *mockConnection) EnqueueCmd([][]byte) {}

func (m *mockConnection) ClearQueuedCmds() {}

func (m *mockConnection) GetWatching() map[string]uint32 {
	return nil
}

func (m *mockConnection) AddTxError(err error) {}

func (m *mockConnection) GetTxErrors() []error {
	return nil
}

func (m *mockConnection) GetDBIndex() int {
	return 0
}

func (m *mockConnection) SelectDB(int) {}

func (m *mockConnection) SetSlave() {}

func (m *mockConnection) IsSlave() bool {
	return false
}

func (m *mockConnection) SetMaster() {}

func (m *mockConnection) IsMaster() bool {
	return false
}

func (m *mockConnection) Name() string {
	return m.name
}

func (m *mockConnection) getWriteCount() int {
	m.writeMutex.Lock()
	defer m.writeMutex.Unlock()
	return len(m.writeBuffer)
}

func (m *mockConnection) clearWriteBuffer() {
	m.writeMutex.Lock()
	defer m.writeMutex.Unlock()
	m.writeBuffer = make([][]byte, 0)
}

// TestSubscribe tests basic subscribe functionality
func TestSubscribe(t *testing.T) {
	hub := MakeHub()
	conn := newMockConnection("client1")

	// Subscribe to a channel
	args := [][]byte{[]byte("channel1")}
	reply := Subscribe(hub, conn, args)

	// Check reply type
	if _, ok := reply.(*protocol.NoReply); !ok {
		t.Errorf("Expected NoReply, got %T", reply)
	}

	// Check if client is subscribed
	if conn.SubsCount() != 1 {
		t.Errorf("Expected 1 subscription, got %d", conn.SubsCount())
	}

	channels := conn.GetChannels()
	if len(channels) != 1 || channels[0] != "channel1" {
		t.Errorf("Expected channel1, got %v", channels)
	}

	// Check if hub has the subscriber
	raw, ok := hub.subs.Get("channel1")
	if !ok {
		t.Error("Channel not found in hub")
	}
	if raw == nil {
		t.Error("Subscribers list is nil")
	}
}

// TestMultipleSubscribe tests subscribing to multiple channels
func TestMultipleSubscribe(t *testing.T) {
	hub := MakeHub()
	conn := newMockConnection("client1")

	// Subscribe to multiple channels
	args := [][]byte{[]byte("channel1"), []byte("channel2"), []byte("channel3")}
	Subscribe(hub, conn, args)

	// Check subscription count
	if conn.SubsCount() != 3 {
		t.Errorf("Expected 3 subscriptions, got %d", conn.SubsCount())
	}

	// Check all channels are subscribed
	channels := conn.GetChannels()
	channelMap := make(map[string]bool)
	for _, ch := range channels {
		channelMap[ch] = true
	}

	if !channelMap["channel1"] || !channelMap["channel2"] || !channelMap["channel3"] {
		t.Errorf("Missing expected channels, got %v", channels)
	}
}

// TestDuplicateSubscribe tests subscribing to the same channel twice
func TestDuplicateSubscribe(t *testing.T) {
	hub := MakeHub()
	conn := newMockConnection("client1")

	// Subscribe to a channel twice
	args := [][]byte{[]byte("channel1")}
	Subscribe(hub, conn, args)
	Subscribe(hub, conn, args)

	// Should still have only 1 subscription
	if conn.SubsCount() != 1 {
		t.Errorf("Expected 1 subscription, got %d", conn.SubsCount())
	}
}

// TestUnSubscribe tests basic unsubscribe functionality
func TestUnSubscribe(t *testing.T) {
	hub := MakeHub()
	conn := newMockConnection("client1")

	// Subscribe first
	subscribeArgs := [][]byte{[]byte("channel1"), []byte("channel2")}
	Subscribe(hub, conn, subscribeArgs)

	// Unsubscribe from one channel
	unsubscribeArgs := [][]byte{[]byte("channel1")}
	reply := UnSubscribe(hub, conn, unsubscribeArgs)

	// Check reply type
	if _, ok := reply.(*protocol.NoReply); !ok {
		t.Errorf("Expected NoReply, got %T", reply)
	}

	// Check subscription count
	if conn.SubsCount() != 1 {
		t.Errorf("Expected 1 subscription, got %d", conn.SubsCount())
	}

	// Check remaining channel
	channels := conn.GetChannels()
	if len(channels) != 1 || channels[0] != "channel2" {
		t.Errorf("Expected channel2, got %v", channels)
	}

	// Check if channel1 is removed from hub
	raw, ok := hub.subs.Get("channel1")
	if ok && raw != nil {
		t.Error("channel1 should be removed from hub")
	}
}

// TestUnSubscribeAll tests unsubscribing from all channels
func TestUnSubscribeAll(t *testing.T) {
	hub := MakeHub()
	conn := newMockConnection("client1")

	// Subscribe to multiple channels
	subscribeArgs := [][]byte{[]byte("channel1"), []byte("channel2"), []byte("channel3")}
	Subscribe(hub, conn, subscribeArgs)

	// Unsubscribe from all
	UnsubscribeAll(hub, conn)

	// Check subscription count
	if conn.SubsCount() != 0 {
		t.Errorf("Expected 0 subscriptions, got %d", conn.SubsCount())
	}

	// Check all channels are removed from hub
	for _, channel := range []string{"channel1", "channel2", "channel3"} {
		raw, ok := hub.subs.Get(channel)
		if ok && raw != nil {
			t.Errorf("channel %s should be removed from hub", channel)
		}
	}
}

// TestUnSubscribeEmpty tests unsubscribing with no arguments
func TestUnSubscribeEmpty(t *testing.T) {
	hub := MakeHub()
	conn := newMockConnection("client1")

	// Subscribe to channels
	subscribeArgs := [][]byte{[]byte("channel1"), []byte("channel2")}
	Subscribe(hub, conn, subscribeArgs)

	// Unsubscribe with no arguments (should unsubscribe from all)
	reply := UnSubscribe(hub, conn, [][]byte{})

	// Check reply type
	if _, ok := reply.(*protocol.NoReply); !ok {
		t.Errorf("Expected NoReply, got %T", reply)
	}

	// Check subscription count
	if conn.SubsCount() != 0 {
		t.Errorf("Expected 0 subscriptions, got %d", conn.SubsCount())
	}
}

// TestPublish tests basic publish functionality
func TestPublish(t *testing.T) {
	hub := MakeHub()
	conn1 := newMockConnection("client1")
	conn2 := newMockConnection("client2")

	// Subscribe both clients to channel1
	args := [][]byte{[]byte("channel1")}
	Subscribe(hub, conn1, args)
	Subscribe(hub, conn2, args)

	// Clear write buffers
	conn1.clearWriteBuffer()
	conn2.clearWriteBuffer()

	// Publish a message
	publishArgs := [][]byte{[]byte("channel1"), []byte("hello")}
	reply := Publish(hub, publishArgs)

	// Check reply (should be number of subscribers)
	intReply, ok := reply.(*protocol.IntReply)
	if !ok {
		t.Errorf("Expected IntReply, got %T", reply)
	}
	if intReply.Code != 2 {
		t.Errorf("Expected 2 subscribers, got %d", intReply.Code)
	}

	// Check both clients received the message
	if conn1.getWriteCount() != 1 {
		t.Errorf("Client1 expected 1 message, got %d", conn1.getWriteCount())
	}
	if conn2.getWriteCount() != 1 {
		t.Errorf("Client2 expected 1 message, got %d", conn2.getWriteCount())
	}
}

// TestPublishNoSubscribers tests publishing to a channel with no subscribers
func TestPublishNoSubscribers(t *testing.T) {
	hub := MakeHub()

	// Publish to a channel with no subscribers
	publishArgs := [][]byte{[]byte("channel1"), []byte("hello")}
	reply := Publish(hub, publishArgs)

	// Check reply (should be 0)
	intReply, ok := reply.(*protocol.IntReply)
	if !ok {
		t.Errorf("Expected IntReply, got %T", reply)
	}
	if intReply.Code != 0 {
		t.Errorf("Expected 0 subscribers, got %d", intReply.Code)
	}
}

// TestPublishInvalidArgs tests publish with invalid arguments
func TestPublishInvalidArgs(t *testing.T) {
	hub := MakeHub()

	// Publish with too few arguments
	publishArgs := [][]byte{[]byte("channel1")}
	reply := Publish(hub, publishArgs)

	// Check reply (should be error)
	_, ok := reply.(*protocol.ArgNumErrReply)
	if !ok {
		t.Errorf("Expected ArgNumErrReply, got %T", reply)
	}
}

// TestConcurrentSubscribe tests concurrent subscribe operations
func TestConcurrentSubscribe(t *testing.T) {
	hub := MakeHub()
	numClients := 100

	var wg sync.WaitGroup
	wg.Add(numClients)

	for i := 0; i < numClients; i++ {
		go func(id int) {
			defer wg.Done()
			conn := newMockConnection("client")
			args := [][]byte{[]byte("channel1")}
			Subscribe(hub, conn, args)
		}(i)
	}

	wg.Wait()

	// Check if all clients are subscribed
	raw, ok := hub.subs.Get("channel1")
	if !ok {
		t.Error("Channel not found in hub")
	}
	if raw == nil {
		t.Error("Subscribers list is nil")
	}
}

// TestMultipleClientsMultipleChannels tests complex scenario
func TestMultipleClientsMultipleChannels(t *testing.T) {
	hub := MakeHub()
	conn1 := newMockConnection("client1")
	conn2 := newMockConnection("client2")
	conn3 := newMockConnection("client3")

	// conn1 subscribes to channel1 and channel2
	Subscribe(hub, conn1, [][]byte{[]byte("channel1"), []byte("channel2")})

	// conn2 subscribes to channel1
	Subscribe(hub, conn2, [][]byte{[]byte("channel1")})

	// conn3 subscribes to channel2 and channel3
	Subscribe(hub, conn3, [][]byte{[]byte("channel2"), []byte("channel3")})

	// Clear write buffers
	conn1.clearWriteBuffer()
	conn2.clearWriteBuffer()
	conn3.clearWriteBuffer()

	// Publish to channel1
	reply := Publish(hub, [][]byte{[]byte("channel1"), []byte("msg1")})
	intReply, _ := reply.(*protocol.IntReply)
	if intReply.Code != 2 {
		t.Errorf("Expected 2 subscribers for channel1, got %d", intReply.Code)
	}

	// Check conn1 and conn2 received message, conn3 did not
	if conn1.getWriteCount() != 1 {
		t.Errorf("Client1 expected 1 message, got %d", conn1.getWriteCount())
	}
	if conn2.getWriteCount() != 1 {
		t.Errorf("Client2 expected 1 message, got %d", conn2.getWriteCount())
	}
	if conn3.getWriteCount() != 0 {
		t.Errorf("Client3 expected 0 messages, got %d", conn3.getWriteCount())
	}

	// Clear buffers
	conn1.clearWriteBuffer()
	conn2.clearWriteBuffer()
	conn3.clearWriteBuffer()

	// Publish to channel2
	reply = Publish(hub, [][]byte{[]byte("channel2"), []byte("msg2")})
	intReply, _ = reply.(*protocol.IntReply)
	if intReply.Code != 2 {
		t.Errorf("Expected 2 subscribers for channel2, got %d", intReply.Code)
	}

	// Check conn1 and conn3 received message, conn2 did not
	if conn1.getWriteCount() != 1 {
		t.Errorf("Client1 expected 1 message, got %d", conn1.getWriteCount())
	}
	if conn2.getWriteCount() != 0 {
		t.Errorf("Client2 expected 0 messages, got %d", conn2.getWriteCount())
	}
	if conn3.getWriteCount() != 1 {
		t.Errorf("Client3 expected 1 message, got %d", conn3.getWriteCount())
	}
}

// TestUnsubscribeAfterPublish tests unsubscribe after receiving messages
func TestUnsubscribeAfterPublish(t *testing.T) {
	hub := MakeHub()
	conn := newMockConnection("client1")

	// Subscribe to channel
	Subscribe(hub, conn, [][]byte{[]byte("channel1")})
	conn.clearWriteBuffer()

	// Publish message
	Publish(hub, [][]byte{[]byte("channel1"), []byte("hello")})

	if conn.getWriteCount() != 1 {
		t.Errorf("Expected 1 message, got %d", conn.getWriteCount())
	}

	// Unsubscribe
	UnSubscribe(hub, conn, [][]byte{[]byte("channel1")})

	// Publish again
	conn.clearWriteBuffer()
	Publish(hub, [][]byte{[]byte("channel1"), []byte("world")})

	// Should not receive the second message
	if conn.getWriteCount() != 0 {
		t.Errorf("Expected 0 messages after unsubscribe, got %d", conn.getWriteCount())
	}
}
