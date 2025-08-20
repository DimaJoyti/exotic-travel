package langgraph

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// StateKey represents a key in the state
type StateKey string

// State represents the current state of a graph execution
type State struct {
	ID          string                 `json:"id"`
	GraphID     string                 `json:"graph_id"`
	UserID      string                 `json:"user_id,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	Data        map[string]interface{} `json:"data"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Version     int                    `json:"version"`
	mutex       sync.RWMutex           `json:"-"`
}

// NewState creates a new state instance
func NewState(id, graphID string) *State {
	now := time.Now()
	return &State{
		ID:        id,
		GraphID:   graphID,
		Data:      make(map[string]interface{}),
		Metadata:  make(map[string]interface{}),
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}
}

// Get retrieves a value from the state
func (s *State) Get(key string) (interface{}, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	value, exists := s.Data[key]
	return value, exists
}

// GetString retrieves a string value from the state
func (s *State) GetString(key string) (string, bool) {
	value, exists := s.Get(key)
	if !exists {
		return "", false
	}
	
	str, ok := value.(string)
	return str, ok
}

// GetInt retrieves an int value from the state
func (s *State) GetInt(key string) (int, bool) {
	value, exists := s.Get(key)
	if !exists {
		return 0, false
	}
	
	switch v := value.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

// GetBool retrieves a bool value from the state
func (s *State) GetBool(key string) (bool, bool) {
	value, exists := s.Get(key)
	if !exists {
		return false, false
	}
	
	b, ok := value.(bool)
	return b, ok
}

// GetSlice retrieves a slice value from the state
func (s *State) GetSlice(key string) ([]interface{}, bool) {
	value, exists := s.Get(key)
	if !exists {
		return nil, false
	}
	
	slice, ok := value.([]interface{})
	return slice, ok
}

// GetMap retrieves a map value from the state
func (s *State) GetMap(key string) (map[string]interface{}, bool) {
	value, exists := s.Get(key)
	if !exists {
		return nil, false
	}
	
	m, ok := value.(map[string]interface{})
	return m, ok
}

// Set stores a value in the state
func (s *State) Set(key string, value interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.Data[key] = value
	s.UpdatedAt = time.Now()
	s.Version++
}

// SetMultiple stores multiple values in the state atomically
func (s *State) SetMultiple(values map[string]interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	for key, value := range values {
		s.Data[key] = value
	}
	s.UpdatedAt = time.Now()
	s.Version++
}

// Delete removes a value from the state
func (s *State) Delete(key string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	delete(s.Data, key)
	s.UpdatedAt = time.Now()
	s.Version++
}

// Has checks if a key exists in the state
func (s *State) Has(key string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	_, exists := s.Data[key]
	return exists
}

// Keys returns all keys in the state
func (s *State) Keys() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	keys := make([]string, 0, len(s.Data))
	for key := range s.Data {
		keys = append(keys, key)
	}
	return keys
}

// Size returns the number of items in the state
func (s *State) Size() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	return len(s.Data)
}

// Clone creates a deep copy of the state
func (s *State) Clone() *State {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	clone := &State{
		ID:        s.ID,
		GraphID:   s.GraphID,
		UserID:    s.UserID,
		SessionID: s.SessionID,
		Data:      make(map[string]interface{}),
		Metadata:  make(map[string]interface{}),
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
		Version:   s.Version,
	}
	
	// Deep copy data
	for key, value := range s.Data {
		clone.Data[key] = deepCopy(value)
	}
	
	// Deep copy metadata
	for key, value := range s.Metadata {
		clone.Metadata[key] = deepCopy(value)
	}
	
	return clone
}

// ToJSON serializes the state to JSON
func (s *State) ToJSON() ([]byte, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	return json.Marshal(s)
}

// FromJSON deserializes the state from JSON
func (s *State) FromJSON(data []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	return json.Unmarshal(data, s)
}

// SetMetadata sets metadata for the state
func (s *State) SetMetadata(key string, value interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.Metadata[key] = value
	s.UpdatedAt = time.Now()
}

// GetMetadata retrieves metadata from the state
func (s *State) GetMetadata(key string) (interface{}, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	value, exists := s.Metadata[key]
	return value, exists
}

// StateManager manages state persistence and retrieval
type StateManager interface {
	// SaveState persists a state
	SaveState(ctx context.Context, state *State) error
	
	// LoadState retrieves a state by ID
	LoadState(ctx context.Context, stateID string) (*State, error)
	
	// DeleteState removes a state
	DeleteState(ctx context.Context, stateID string) error
	
	// ListStates lists states with optional filters
	ListStates(ctx context.Context, filters map[string]interface{}) ([]*State, error)
}

// MemoryStateManager implements StateManager using in-memory storage
type MemoryStateManager struct {
	states map[string]*State
	mutex  sync.RWMutex
	tracer trace.Tracer
}

// NewMemoryStateManager creates a new in-memory state manager
func NewMemoryStateManager() *MemoryStateManager {
	return &MemoryStateManager{
		states: make(map[string]*State),
		tracer: otel.Tracer("langgraph.state_manager"),
	}
}

// SaveState saves a state to memory
func (m *MemoryStateManager) SaveState(ctx context.Context, state *State) error {
	ctx, span := m.tracer.Start(ctx, "memory_state_manager.save_state")
	defer span.End()
	
	span.SetAttributes(
		attribute.String("state.id", state.ID),
		attribute.String("state.graph_id", state.GraphID),
		attribute.Int("state.version", state.Version),
	)
	
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Clone the state to avoid external modifications
	m.states[state.ID] = state.Clone()
	
	return nil
}

// LoadState loads a state from memory
func (m *MemoryStateManager) LoadState(ctx context.Context, stateID string) (*State, error) {
	ctx, span := m.tracer.Start(ctx, "memory_state_manager.load_state")
	defer span.End()
	
	span.SetAttributes(attribute.String("state.id", stateID))
	
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	state, exists := m.states[stateID]
	if !exists {
		err := fmt.Errorf("state not found: %s", stateID)
		span.RecordError(err)
		return nil, err
	}
	
	// Return a clone to avoid external modifications
	return state.Clone(), nil
}

// DeleteState removes a state from memory
func (m *MemoryStateManager) DeleteState(ctx context.Context, stateID string) error {
	ctx, span := m.tracer.Start(ctx, "memory_state_manager.delete_state")
	defer span.End()
	
	span.SetAttributes(attribute.String("state.id", stateID))
	
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	delete(m.states, stateID)
	return nil
}

// ListStates lists all states in memory with optional filters
func (m *MemoryStateManager) ListStates(ctx context.Context, filters map[string]interface{}) ([]*State, error) {
	ctx, span := m.tracer.Start(ctx, "memory_state_manager.list_states")
	defer span.End()
	
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	var result []*State
	for _, state := range m.states {
		if matchesFilters(state, filters) {
			result = append(result, state.Clone())
		}
	}
	
	span.SetAttributes(attribute.Int("states.count", len(result)))
	return result, nil
}

// deepCopy creates a deep copy of an interface{}
func deepCopy(src interface{}) interface{} {
	switch v := src.(type) {
	case map[string]interface{}:
		dst := make(map[string]interface{})
		for key, value := range v {
			dst[key] = deepCopy(value)
		}
		return dst
	case []interface{}:
		dst := make([]interface{}, len(v))
		for i, value := range v {
			dst[i] = deepCopy(value)
		}
		return dst
	default:
		return v
	}
}

// matchesFilters checks if a state matches the given filters
func matchesFilters(state *State, filters map[string]interface{}) bool {
	for key, expectedValue := range filters {
		switch key {
		case "graph_id":
			if state.GraphID != expectedValue {
				return false
			}
		case "user_id":
			if state.UserID != expectedValue {
				return false
			}
		case "session_id":
			if state.SessionID != expectedValue {
				return false
			}
		default:
			// Check in state data
			actualValue, exists := state.Get(key)
			if !exists || actualValue != expectedValue {
				return false
			}
		}
	}
	return true
}
