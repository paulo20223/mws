package task

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	cfg "github.com/pavel/mws/internal/config"
	"github.com/pavel/mws/internal/git"
	"github.com/pavel/mws/internal/workspace"
)

// Manager handles task CRUD operations.
type Manager struct {
	ws *workspace.Workspace
}

// NewManager creates a new task manager.
func NewManager(ws *workspace.Workspace) *Manager {
	return &Manager{ws: ws}
}

// CreateOptions holds parameters for creating a new task.
type CreateOptions struct {
	Name        string
	Description string
	Branch      string   // branch to create in matching repos
	RepoGlobs   []string // glob patterns to match repos (empty = all)
}

// Create creates a new task: copies workspace, switches branches, writes manifest.
func (m *Manager) Create(opts CreateOptions) (*Manifest, error) {
	copyPath := cfg.TaskCopyPath(m.ws.Base, opts.Name)

	// Check if task already exists
	if _, err := os.Stat(copyPath); err == nil {
		return nil, fmt.Errorf("task %q already exists at %s", opts.Name, copyPath)
	}
	if _, err := os.Stat(ManifestPath(m.ws.Base, opts.Name)); err == nil {
		return nil, fmt.Errorf("task %q manifest already exists", opts.Name)
	}

	// Step 1: Full copy of workspace
	fmt.Printf("  Copying workspace to %s ...\n", copyPath)
	if err := workspace.CopyWorkspace(m.ws.Base, copyPath, m.ws.Config.Copy.Exclude); err != nil {
		// Cleanup on failure
		os.RemoveAll(copyPath)
		return nil, fmt.Errorf("copying workspace: %w", err)
	}

	// Step 2: Discover git repos in the copy
	repos, err := git.DiscoverRepos(copyPath)
	if err != nil {
		return nil, fmt.Errorf("discovering repos: %w", err)
	}

	// Step 3: Filter repos by glob patterns and switch branches
	var manifestRepos []RepoRef
	for _, repo := range repos {
		if !matchesGlobs(repo.RelPath, opts.RepoGlobs) {
			continue
		}

		branch := opts.Branch
		if branch == "" {
			// No branch specified, keep current branch
			manifestRepos = append(manifestRepos, RepoRef{
				Path:   repo.RelPath,
				Branch: repo.Branch,
			})
			continue
		}

		fmt.Printf("  Switching %s -> %s ...\n", repo.RelPath, branch)
		if err := git.CheckoutBranch(repo.AbsPath, branch); err != nil {
			fmt.Printf("  Warning: failed to switch %s to %s: %v\n", repo.RelPath, branch, err)
			manifestRepos = append(manifestRepos, RepoRef{
				Path:   repo.RelPath,
				Branch: repo.Branch,
			})
		} else {
			manifestRepos = append(manifestRepos, RepoRef{
				Path:   repo.RelPath,
				Branch: branch,
			})
		}
	}

	// Step 4: Write task marker file
	markerPath := filepath.Join(copyPath, cfg.TaskMarker)
	if err := os.WriteFile(markerPath, []byte(opts.Name), 0644); err != nil {
		return nil, fmt.Errorf("writing task marker: %w", err)
	}

	// Step 5: Save manifest
	manifest := Manifest{
		Task: TaskInfo{
			Name:        opts.Name,
			Description: opts.Description,
			Created:     time.Now(),
			Base:        m.ws.Base,
			Copy:        copyPath,
		},
		Repos: manifestRepos,
	}

	if err := SaveManifest(m.ws.Base, manifest); err != nil {
		return nil, fmt.Errorf("saving manifest: %w", err)
	}

	return &manifest, nil
}

// List returns all tasks.
func (m *Manager) List() ([]Manifest, error) {
	tasksDir := filepath.Join(cfg.MWSPath(m.ws.Base), cfg.TasksDir)
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var tasks []Manifest
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		manifest, err := LoadManifest(m.ws.Base, e.Name())
		if err != nil {
			continue // skip broken manifests
		}
		tasks = append(tasks, manifest)
	}

	return tasks, nil
}

// Delete removes a task (copy directory + manifest).
func (m *Manager) Delete(name string, force bool) error {
	manifest, err := LoadManifest(m.ws.Base, name)
	if err != nil {
		return fmt.Errorf("task %q not found: %w", name, err)
	}

	copyPath := manifest.Task.Copy
	if copyPath == "" {
		copyPath = cfg.TaskCopyPath(m.ws.Base, name)
	}

	// Check for uncommitted changes unless --force
	if !force {
		repos, _ := git.DiscoverRepos(copyPath)
		for _, repo := range repos {
			if git.HasUncommitted(repo.AbsPath) {
				return fmt.Errorf("repo %s has uncommitted changes (use --force to override)", repo.RelPath)
			}
		}
	}

	// Remove the copy directory
	fmt.Printf("  Removing %s ...\n", copyPath)
	if err := os.RemoveAll(copyPath); err != nil {
		return fmt.Errorf("removing copy: %w", err)
	}

	// Remove manifest
	if err := DeleteManifest(m.ws.Base, name); err != nil {
		return fmt.Errorf("removing manifest: %w", err)
	}

	return nil
}

// Status returns status information about a task.
type TaskStatus struct {
	Manifest  Manifest
	CopyPath  string
	Exists    bool
	RepoStats []RepoStatusInfo
}

type RepoStatusInfo struct {
	RelPath string
	Status  git.RepoStatus
}

// GetStatus returns the current status of a task.
func (m *Manager) GetStatus(name string) (*TaskStatus, error) {
	manifest, err := LoadManifest(m.ws.Base, name)
	if err != nil {
		return nil, fmt.Errorf("task %q not found: %w", name, err)
	}

	copyPath := manifest.Task.Copy
	if copyPath == "" {
		copyPath = cfg.TaskCopyPath(m.ws.Base, name)
	}

	status := &TaskStatus{
		Manifest: manifest,
		CopyPath: copyPath,
	}

	if _, err := os.Stat(copyPath); err == nil {
		status.Exists = true

		// Get status of each repo in the copy
		for _, repoRef := range manifest.Repos {
			repoPath := filepath.Join(copyPath, repoRef.Path)
			repoStatus := git.GetStatus(repoPath)
			status.RepoStats = append(status.RepoStats, RepoStatusInfo{
				RelPath: repoRef.Path,
				Status:  repoStatus,
			})
		}
	}

	return status, nil
}

// matchesGlobs checks if a path matches any of the given glob patterns.
// Empty globs means "match all".
func matchesGlobs(relPath string, globs []string) bool {
	if len(globs) == 0 {
		return true
	}
	for _, g := range globs {
		if git.MatchGlob(relPath, g) {
			return true
		}
	}
	return false
}
