package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RepoInfo holds information about a discovered git repository.
type RepoInfo struct {
	// Path relative to the workspace root.
	RelPath string
	// Absolute path.
	AbsPath string
	// Current branch name.
	Branch string
}

// DiscoverRepos finds all git repositories within the given directory (depth 1-2).
func DiscoverRepos(root string) ([]RepoInfo, error) {
	var repos []RepoInfo
	seen := make(map[string]bool)

	// Walk depth 1 and 2
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("reading directory %s: %w", root, err)
	}

	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		dirPath := filepath.Join(root, e.Name())

		// Check depth 1: is this dir itself a git repo?
		if isGitRepo(dirPath) && !seen[dirPath] {
			seen[dirPath] = true
			branch := currentBranch(dirPath)
			repos = append(repos, RepoInfo{
				RelPath: e.Name(),
				AbsPath: dirPath,
				Branch:  branch,
			})
			continue
		}

		// Check depth 2: subdirectories of this dir
		subEntries, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}
		for _, se := range subEntries {
			if !se.IsDir() || strings.HasPrefix(se.Name(), ".") {
				continue
			}
			subPath := filepath.Join(dirPath, se.Name())
			if isGitRepo(subPath) && !seen[subPath] {
				seen[subPath] = true
				branch := currentBranch(subPath)
				relPath := filepath.Join(e.Name(), se.Name())
				repos = append(repos, RepoInfo{
					RelPath: relPath,
					AbsPath: subPath,
					Branch:  branch,
				})
			}
		}
	}

	return repos, nil
}

// isGitRepo checks if a directory contains a .git directory.
func isGitRepo(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	if err != nil {
		return false
	}
	return info.IsDir()
}

// currentBranch returns the current branch of a git repo.
func currentBranch(repoPath string) string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

// CheckoutBranch switches to the given branch, creating it if needed.
func CheckoutBranch(repoPath, branch string) error {
	// Try to checkout existing branch first
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = repoPath
	if err := cmd.Run(); err == nil {
		return nil
	}

	// Branch doesn't exist, create it
	cmd = exec.Command("git", "checkout", "-b", branch)
	cmd.Dir = repoPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git checkout -b %s in %s: %s", branch, repoPath, string(out))
	}
	return nil
}

// Status returns a human-readable status of the git repo.
type RepoStatus struct {
	Branch    string
	Modified  int
	Staged    int
	Untracked int
	Clean     bool
}

// GetStatus returns the status of a git repository.
func GetStatus(repoPath string) RepoStatus {
	status := RepoStatus{
		Branch: currentBranch(repoPath),
	}

	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return status
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		status.Clean = true
		return status
	}

	for _, line := range lines {
		if len(line) < 2 {
			continue
		}
		x, y := line[0], line[1]
		if x == '?' && y == '?' {
			status.Untracked++
		} else if x != ' ' && x != '?' {
			status.Staged++
		}
		if y != ' ' && y != '?' {
			status.Modified++
		}
	}

	return status
}

// StatusString returns a human-readable status string.
func (s RepoStatus) StatusString() string {
	if s.Clean {
		return "clean"
	}
	var parts []string
	if s.Modified > 0 {
		parts = append(parts, fmt.Sprintf("%d modified", s.Modified))
	}
	if s.Staged > 0 {
		parts = append(parts, fmt.Sprintf("%d staged", s.Staged))
	}
	if s.Untracked > 0 {
		parts = append(parts, fmt.Sprintf("%d untracked", s.Untracked))
	}
	return strings.Join(parts, ", ")
}

// HasUncommitted returns true if repo has uncommitted changes.
func HasUncommitted(repoPath string) bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

// HasUnpushed returns true if repo has commits not pushed to remote.
func HasUnpushed(repoPath string) bool {
	cmd := exec.Command("git", "log", "--oneline", "@{upstream}..HEAD")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		// No upstream configured, consider as "unpushed"
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

// MatchGlob checks if a repo relative path matches a glob pattern.
func MatchGlob(relPath, pattern string) bool {
	matched, err := filepath.Match(pattern, relPath)
	if err != nil {
		return false
	}
	return matched
}
