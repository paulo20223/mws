package workspace

import (
	"fmt"
	"os"
	"path/filepath"

	cfg "github.com/pavel/mws/internal/config"
	"github.com/pavel/mws/internal/git"
)

// Workspace represents an initialized mws workspace.
type Workspace struct {
	Base   string
	Config cfg.Config
}

// Init initializes a new mws workspace in the given directory.
func Init(base string) (*Workspace, error) {
	absBase, err := filepath.Abs(base)
	if err != nil {
		return nil, err
	}

	mwsDir := cfg.MWSPath(absBase)
	tasksDir := filepath.Join(mwsDir, cfg.TasksDir)

	// Create directory structure
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		return nil, fmt.Errorf("creating .mws directory: %w", err)
	}

	// Create default config
	config := cfg.DefaultConfig(absBase)
	if err := cfg.Save(config); err != nil {
		return nil, fmt.Errorf("saving config: %w", err)
	}

	return &Workspace{
		Base:   absBase,
		Config: config,
	}, nil
}

// Open opens an existing mws workspace.
func Open(base string) (*Workspace, error) {
	absBase, err := filepath.Abs(base)
	if err != nil {
		return nil, err
	}

	config, err := cfg.Load(absBase)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	return &Workspace{
		Base:   absBase,
		Config: config,
	}, nil
}

// DiscoverRepos finds all git repos in the workspace.
func (w *Workspace) DiscoverRepos() ([]git.RepoInfo, error) {
	return git.DiscoverRepos(w.Base)
}

// IsInitialized checks if the directory has been initialized with mws.
func IsInitialized(base string) bool {
	_, err := os.Stat(cfg.ConfigPath(base))
	return err == nil
}

// ResolveWorkspace finds the workspace from cwd.
// Returns the workspace and optionally the task name if cwd is inside a task copy.
func ResolveWorkspace() (ws *Workspace, taskName string, err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, "", fmt.Errorf("getting working directory: %w", err)
	}

	base, task := cfg.FindBaseOrTask(cwd)
	if base == "" {
		return nil, "", fmt.Errorf("not inside an mws workspace (run 'mws init' first)")
	}

	ws, err = Open(base)
	if err != nil {
		return nil, "", err
	}

	return ws, task, nil
}
