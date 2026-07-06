package testagent

import (
	"fmt"
	"sync"
)

// Registry holds all agents and coordinates start/stop.
type Registry struct {
	mu     sync.RWMutex
	agents []*Agent
	cfg    Config
}

// NewRegistry creates agents (all stopped) for count personas.
func NewRegistry(count int, cfg Config) (*Registry, error) {
	personas := DefaultPersonas()
	if count > len(personas) {
		count = len(personas)
	}
	r := &Registry{cfg: cfg}
	for i := 1; i <= count; i++ {
		a, err := NewAgent(i, personas[i-1], cfg)
		if err != nil {
			return nil, fmt.Errorf("agent %d: %w", i, err)
		}
		r.agents = append(r.agents, a)
	}
	return r, nil
}

// Agents returns all agents.
func (r *Registry) Agents() []*Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Agent, len(r.agents))
	copy(out, r.agents)
	return out
}

// AgentByIndex returns agent by 1-based index.
func (r *Registry) AgentByIndex(index int) (*Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.agents {
		if a.Index == index {
			return a, nil
		}
	}
	return nil, fmt.Errorf("agent %d not found", index)
}

// Snapshots returns UI snapshots for all agents.
func (r *Registry) Snapshots() []Snapshot {
	agents := r.Agents()
	out := make([]Snapshot, len(agents))
	for i, a := range agents {
		out[i] = a.Snapshot()
	}
	return out
}

// Start starts one agent by index.
func (r *Registry) Start(index int) error {
	a, err := r.AgentByIndex(index)
	if err != nil {
		return err
	}
	return a.Start(r.cfg)
}

// Stop stops one agent by index.
func (r *Registry) Stop(index int) {
	a, err := r.AgentByIndex(index)
	if err != nil {
		return
	}
	a.Stop()
}

// StartAll starts every non-running agent.
func (r *Registry) StartAll() {
	for _, a := range r.Agents() {
		if a.Status() != StatusRunning {
			_ = a.Start(r.cfg)
		}
	}
}

// StopAll stops all running agents.
func (r *Registry) StopAll() {
	for _, a := range r.Agents() {
		a.Stop()
	}
}

// WaitAll waits for all agents to finish.
func (r *Registry) WaitAll() {
	for _, a := range r.Agents() {
		a.Wait()
	}
}
