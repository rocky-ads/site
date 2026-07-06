package testagent

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AgentState is persisted credentials for an agent.
type AgentState struct {
	Username string `json:"username"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

// LoadState reads agent state from path. Missing file returns nil, nil.
func LoadState(path string) (*AgentState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var s AgentState
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// SaveState writes agent state to path.
func SaveState(path string, s AgentState) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// NewPassword generates a random password meeting site requirements.
func NewPassword() (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// Site requires mixed case, digit, special char.
	return "Ag!" + hex.EncodeToString(b), nil
}

// StatePath returns the state file path for an agent index.
func StatePath(dir string, index int) string {
	return filepath.Join(dir, fmt.Sprintf("agent-%d.json", index))
}
