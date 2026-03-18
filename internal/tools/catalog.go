package tools

import (
	"fmt"
	"sync"

	"github.com/tianrking/ClawRemove/internal/model"
)

// Registry manages provider tools with proper registration and lookup.
type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// NewRegistry creates a new tool registry.
func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool)}
}

// Register adds a tool to the registry.
// Returns error if a tool with the same ID already exists.
func (r *Registry) Register(tool Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	info := tool.Info()
	if !info.ReadOnly {
		return fmt.Errorf("tool %q must be read-only", info.ID)
	}
	if _, exists := r.tools[info.ID]; exists {
		return fmt.Errorf("tool %q already registered", info.ID)
	}
	r.tools[info.ID] = tool
	return nil
}

// Get retrieves a tool by ID.
func (r *Registry) Get(id string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[id]
	return t, ok
}

// List returns all registered tools.
func (r *Registry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		result = append(result, t)
	}
	return result
}

// Catalog returns provider tool metadata from capabilities.
func Catalog(capabilities model.ProviderCapabilities) []model.ProviderTool {
	return capabilities.Tools
}

// BuildRegistry creates a registry from a slice of tools.
func BuildRegistry(tools []Tool) *Registry {
	r := NewRegistry()
	for _, t := range tools {
		_ = r.Register(t)
	}
	return r
}
