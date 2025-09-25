package repository

import (
	"fmt"
	"strconv"

	"github.com/hdt3213/godis/database"
	"github.com/hdt3213/godis/interface/redis"
	"github.com/hdt3213/godis/redis/protocol"
)

// GodisClientAdapter adapts Godis database to implement GodisClient interface
type GodisClientAdapter struct {
	server *database.Server
	conn   redis.Connection
}

// NewGodisClientAdapter creates a new adapter for Godis
func NewGodisClientAdapter(server *database.Server) *GodisClientAdapter {
	// Create a dummy connection for internal operations
	conn := &dummyConnection{}
	return &GodisClientAdapter{
		server: server,
		conn:   conn,
	}
}

// dummyConnection implements redis.Connection interface for internal use
type dummyConnection struct{}

func (d *dummyConnection) Write([]byte) (int, error)             { return 0, nil }
func (d *dummyConnection) GetDBIndex() int                       { return 0 }
func (d *dummyConnection) SelectDB(int)                          {}
func (d *dummyConnection) Close() error                          { return nil }
func (d *dummyConnection) SetPassword(string)                    {}
func (d *dummyConnection) GetPassword() string                   { return "" }
func (d *dummyConnection) Subscribe(string)                      {}
func (d *dummyConnection) UnSubscribe(string)                    {}
func (d *dummyConnection) SubsCount() int                        { return 0 }
func (d *dummyConnection) GetChannels() []string                 { return nil }
func (d *dummyConnection) InMultiState() bool                    { return false }
func (d *dummyConnection) SetMultiState(bool)                    {}
func (d *dummyConnection) GetQueuedCmdLine() [][][]byte          { return nil }
func (d *dummyConnection) EnqueueCmd([][]byte)                   {}
func (d *dummyConnection) ClearQueuedCmds()                      {}
func (d *dummyConnection) GetWatching() map[string]uint32        { return nil }
func (d *dummyConnection) AddTxError(error)                      {}
func (d *dummyConnection) GetTxErrors() []error                  { return nil }

func (d *dummyConnection) RemoteAddr() string                    { return "127.0.0.1:0" }
func (d *dummyConnection) SetSlave()                             {}
func (d *dummyConnection) IsSlave() bool                         { return false }
func (d *dummyConnection) SetMaster()                            {}
func (d *dummyConnection) IsMaster() bool                        { return false }
func (d *dummyConnection) Name() string                          { return "dummy" }

// Set implements GodisClient.Set
func (g *GodisClientAdapter) Set(key string, value interface{}) error {
	valueStr := fmt.Sprintf("%v", value)
	cmdLine := [][]byte{
		[]byte("SET"),
		[]byte(key),
		[]byte(valueStr),
	}
	
	result := g.server.Exec(g.conn, cmdLine)
	if result.ToBytes() == nil {
		return fmt.Errorf("failed to set key %s", key)
	}
	return nil
}

// Get implements GodisClient.Get
func (g *GodisClientAdapter) Get(key string) (interface{}, error) {
	cmdLine := [][]byte{
		[]byte("GET"),
		[]byte(key),
	}
	
	result := g.server.Exec(g.conn, cmdLine)
	if result == nil {
		return nil, fmt.Errorf("key %s not found", key)
	}
	
	// Handle different reply types
	switch r := result.(type) {
	case *protocol.BulkReply:
		if r.Arg == nil {
			return nil, fmt.Errorf("key %s not found", key)
		}
		return string(r.Arg), nil
	case *protocol.StatusReply:
		return r.Status, nil
	case *protocol.IntReply:
		return strconv.FormatInt(r.Code, 10), nil
	default:
		return nil, fmt.Errorf("unexpected reply type for key %s", key)
	}
}

// Del implements GodisClient.Del
func (g *GodisClientAdapter) Del(keys ...string) error {
	cmdLine := make([][]byte, len(keys)+1)
	cmdLine[0] = []byte("DEL")
	for i, key := range keys {
		cmdLine[i+1] = []byte(key)
	}
	
	result := g.server.Exec(g.conn, cmdLine)
	if result == nil {
		return fmt.Errorf("failed to delete keys")
	}
	return nil
}

// Exists implements GodisClient.Exists
func (g *GodisClientAdapter) Exists(key string) (bool, error) {
	cmdLine := [][]byte{
		[]byte("EXISTS"),
		[]byte(key),
	}
	
	result := g.server.Exec(g.conn, cmdLine)
	if result == nil {
		return false, fmt.Errorf("failed to check existence of key %s", key)
	}
	
	switch r := result.(type) {
	case *protocol.IntReply:
		return r.Code > 0, nil
	default:
		return false, fmt.Errorf("unexpected reply type for EXISTS")
	}
}

// SAdd implements GodisClient.SAdd
func (g *GodisClientAdapter) SAdd(key string, members ...interface{}) error {
	cmdLine := make([][]byte, len(members)+2)
	cmdLine[0] = []byte("SADD")
	cmdLine[1] = []byte(key)
	
	for i, member := range members {
		cmdLine[i+2] = []byte(fmt.Sprintf("%v", member))
	}
	
	result := g.server.Exec(g.conn, cmdLine)
	if result == nil {
		return fmt.Errorf("failed to add members to set %s", key)
	}
	return nil
}

// SRem implements GodisClient.SRem
func (g *GodisClientAdapter) SRem(key string, members ...interface{}) error {
	cmdLine := make([][]byte, len(members)+2)
	cmdLine[0] = []byte("SREM")
	cmdLine[1] = []byte(key)
	
	for i, member := range members {
		cmdLine[i+2] = []byte(fmt.Sprintf("%v", member))
	}
	
	result := g.server.Exec(g.conn, cmdLine)
	if result == nil {
		return fmt.Errorf("failed to remove members from set %s", key)
	}
	return nil
}

// SMembers implements GodisClient.SMembers
func (g *GodisClientAdapter) SMembers(key string) ([]string, error) {
	cmdLine := [][]byte{
		[]byte("SMEMBERS"),
		[]byte(key),
	}
	
	result := g.server.Exec(g.conn, cmdLine)
	if result == nil {
		return nil, fmt.Errorf("failed to get members of set %s", key)
	}
	
	switch r := result.(type) {
	case *protocol.MultiBulkReply:
		members := make([]string, len(r.Args))
		for i, arg := range r.Args {
			if arg != nil {
				members[i] = string(arg)
			}
		}
		return members, nil
	case *protocol.EmptyMultiBulkReply:
		// Return empty slice for empty sets
		return []string{}, nil
	default:
		return nil, fmt.Errorf("unexpected reply type for SMEMBERS")
	}
}

// HSet implements GodisClient.HSet
func (g *GodisClientAdapter) HSet(key string, field string, value interface{}) error {
	valueStr := fmt.Sprintf("%v", value)
	cmdLine := [][]byte{
		[]byte("HSET"),
		[]byte(key),
		[]byte(field),
		[]byte(valueStr),
	}
	
	result := g.server.Exec(g.conn, cmdLine)
	if result == nil {
		return fmt.Errorf("failed to set hash field %s in key %s", field, key)
	}
	return nil
}

// HGet implements GodisClient.HGet
func (g *GodisClientAdapter) HGet(key string, field string) (interface{}, error) {
	cmdLine := [][]byte{
		[]byte("HGET"),
		[]byte(key),
		[]byte(field),
	}
	
	result := g.server.Exec(g.conn, cmdLine)
	if result == nil {
		return nil, fmt.Errorf("failed to get hash field %s from key %s", field, key)
	}
	
	switch r := result.(type) {
	case *protocol.BulkReply:
		if r.Arg == nil {
			return nil, fmt.Errorf("field %s not found in hash %s", field, key)
		}
		return string(r.Arg), nil
	default:
		return nil, fmt.Errorf("unexpected reply type for HGET")
	}
}

// HGetAll implements GodisClient.HGetAll
func (g *GodisClientAdapter) HGetAll(key string) (map[string]interface{}, error) {
	cmdLine := [][]byte{
		[]byte("HGETALL"),
		[]byte(key),
	}
	
	result := g.server.Exec(g.conn, cmdLine)
	if result == nil {
		return nil, fmt.Errorf("failed to get all hash fields from key %s", key)
	}
	
	switch r := result.(type) {
	case *protocol.MultiBulkReply:
		if len(r.Args)%2 != 0 {
			return nil, fmt.Errorf("invalid HGETALL response")
		}
		
		fields := make(map[string]interface{})
		for i := 0; i < len(r.Args); i += 2 {
			if r.Args[i] != nil && r.Args[i+1] != nil {
				field := string(r.Args[i])
				value := string(r.Args[i+1])
				fields[field] = value
			}
		}
		return fields, nil
	default:
		return nil, fmt.Errorf("unexpected reply type for HGETALL")
	}
}

// HDel implements GodisClient.HDel
func (g *GodisClientAdapter) HDel(key string, fields ...string) error {
	cmdLine := make([][]byte, len(fields)+2)
	cmdLine[0] = []byte("HDEL")
	cmdLine[1] = []byte(key)
	
	for i, field := range fields {
		cmdLine[i+2] = []byte(field)
	}
	
	result := g.server.Exec(g.conn, cmdLine)
	if result == nil {
		return fmt.Errorf("failed to delete hash fields from key %s", key)
	}
	return nil
}