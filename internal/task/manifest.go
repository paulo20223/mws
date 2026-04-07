package task

import (
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	cfg "github.com/paulo20223/mws/internal/config"
)

// Manifest represents a task's manifest.toml file.
type Manifest struct {
	Task  TaskInfo    `toml:"task"`
	Repos []RepoRef   `toml:"repos"`
	Hooks HooksConfig `toml:"hooks"`
}

type TaskInfo struct {
	Name        string    `toml:"name"`
	Description string    `toml:"description"`
	Created     time.Time `toml:"created"`
	Base        string    `toml:"base"`
	Copy        string    `toml:"copy"`
}

type RepoRef struct {
	Path   string `toml:"path"`
	Branch string `toml:"branch"`
}

type HooksConfig struct {
	OnActivate   string `toml:"on_activate"`
	OnDeactivate string `toml:"on_deactivate"`
	PostCreate   string `toml:"post_create"`
}

// ManifestPath returns the path to a task's manifest file.
func ManifestPath(base, taskName string) string {
	return filepath.Join(cfg.MWSPath(base), cfg.TasksDir, taskName, "manifest.toml")
}

// LoadManifest reads a task's manifest.
func LoadManifest(base, taskName string) (Manifest, error) {
	var m Manifest
	path := ManifestPath(base, taskName)
	data, err := os.ReadFile(path)
	if err != nil {
		return m, err
	}
	err = toml.Unmarshal(data, &m)
	return m, err
}

// SaveManifest writes a task's manifest.
func SaveManifest(base string, m Manifest) error {
	dir := filepath.Dir(ManifestPath(base, m.Task.Name))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(ManifestPath(base, m.Task.Name))
	if err != nil {
		return err
	}
	defer f.Close()

	enc := toml.NewEncoder(f)
	return enc.Encode(m)
}

// DeleteManifest removes a task's manifest directory.
func DeleteManifest(base, taskName string) error {
	dir := filepath.Join(cfg.MWSPath(base), cfg.TasksDir, taskName)
	return os.RemoveAll(dir)
}
