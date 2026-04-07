package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	MWSDir     = ".mws"
	ConfigFile = "config.toml"
	ActiveFile = "active"
	TasksDir   = "tasks"
	TaskMarker = ".mws-task"
)

// Config represents the global mws configuration.
type Config struct {
	Workspace WorkspaceConfig `toml:"workspace"`
	Copy      CopyConfig      `toml:"copy"`
	Defaults  DefaultsConfig  `toml:"defaults"`
}

type WorkspaceConfig struct {
	Base   string `toml:"base"`
	Editor string `toml:"editor"`
}

type CopyConfig struct {
	Exclude       []string `toml:"exclude"`
	PostCopyHooks []string `toml:"post_copy_hooks"`
}

type DefaultsConfig struct {
	StashOnBranchSwitch bool `toml:"stash_on_branch_switch"`
}

// DefaultConfig returns a config with sensible defaults.
func DefaultConfig(base string) Config {
	return Config{
		Workspace: WorkspaceConfig{
			Base:   base,
			Editor: "code",
		},
		Copy: CopyConfig{
			Exclude: []string{
				"node_modules/",
				".venv/",
				"__pycache__/",
				".next/",
				"dist/",
				"build/",
				".mws/",
				"*.pyc",
				".DS_Store",
				"plans/",
				".git/lfs/",
				"vendor/",
				"target/",
			},
			PostCopyHooks: []string{},
		},
		Defaults: DefaultsConfig{
			StashOnBranchSwitch: true,
		},
	}
}

// MWSPath returns the path to .mws/ directory inside the base workspace.
func MWSPath(base string) string {
	return filepath.Join(base, MWSDir)
}

// ConfigPath returns the full path to config.toml.
func ConfigPath(base string) string {
	return filepath.Join(MWSPath(base), ConfigFile)
}

// Load reads config.toml from the given base workspace.
func Load(base string) (Config, error) {
	cfg := DefaultConfig(base)
	path := ConfigPath(base)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	// Ensure base is always set
	if cfg.Workspace.Base == "" {
		cfg.Workspace.Base = base
	}

	return cfg, nil
}

// Save writes the config to config.toml.
func Save(cfg Config) error {
	path := ConfigPath(cfg.Workspace.Base)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := toml.NewEncoder(f)
	return enc.Encode(cfg)
}

// FindBase walks up from the given directory looking for .mws/ directory.
// Returns the base workspace path or empty string if not found.
func FindBase(from string) string {
	dir, _ := filepath.Abs(from)
	for {
		if _, err := os.Stat(filepath.Join(dir, MWSDir, ConfigFile)); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// FindBaseOrTask finds the base workspace, even if cwd is inside a task copy.
// It checks for .mws-task marker and reads the base from the manifest.
func FindBaseOrTask(from string) (base string, taskName string) {
	dir, _ := filepath.Abs(from)

	// First check if we're inside a task copy (has .mws-task marker)
	cur := dir
	for {
		markerPath := filepath.Join(cur, TaskMarker)
		if data, err := os.ReadFile(markerPath); err == nil {
			taskName = string(data)
			// The base is derived from the copy path by removing the --taskname suffix
			// But safer: read it from the manifest in the base .mws/
			// For now, find base by removing suffix
			base = taskCopyToBase(cur, taskName)
			return
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}

	// Not in a task copy, look for .mws/ directly
	base = FindBase(from)
	return base, ""
}

// taskCopyToBase converts a task copy path back to base path.
// e.g., /Users/pavel/test--excel-refactor -> /Users/pavel/test
func taskCopyToBase(copyPath, taskName string) string {
	suffix := "--" + taskName
	if len(copyPath) > len(suffix) && copyPath[len(copyPath)-len(suffix):] == suffix {
		return copyPath[:len(copyPath)-len(suffix)]
	}
	return ""
}

// TaskCopyPath returns the path for a task's workspace copy.
func TaskCopyPath(base, taskName string) string {
	return base + "--" + taskName
}
