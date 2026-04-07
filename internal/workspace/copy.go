package workspace

import (
	"fmt"
	"os/exec"
	"strings"
)

// CopyWorkspace performs a full copy of the workspace using rsync.
// It applies exclude patterns from config.
func CopyWorkspace(src, dst string, excludes []string) error {
	args := []string{
		"-a",
	}

	for _, ex := range excludes {
		args = append(args, "--exclude", ex)
	}

	// Exclude any existing task copies (directories matching base--*)
	// We use the pattern to skip sibling task directories
	args = append(args, "--exclude", "mws/") // exclude the mws tool itself

	// Ensure trailing slash on src for rsync directory semantics
	if !strings.HasSuffix(src, "/") {
		src += "/"
	}

	args = append(args, src, dst)

	cmd := exec.Command("rsync", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rsync failed: %w\n%s", err, string(out))
	}

	return nil
}

// SyncFromBase syncs changes from base to a task copy (non-destructive).
// Only updates files that are newer in base. Does not delete files in dst.
// Excludes git repos (they have their own branches).
func SyncFromBase(src, dst string, excludes []string, gitRepoPaths []string, dryRun bool) (string, error) {
	args := []string{
		"-av",      // archive + verbose (shows transferred files)
		"--update", // only update files that are newer in source
	}

	if dryRun {
		args = append(args, "--dry-run")
	}

	// Exclude the mws tool directory
	args = append(args, "--exclude", "mws/")

	for _, ex := range excludes {
		args = append(args, "--exclude", ex)
	}

	// Exclude git repo directories (they have their own branches)
	for _, repoPath := range gitRepoPaths {
		args = append(args, "--exclude", repoPath+"/")
	}

	// Ensure trailing slashes
	if !strings.HasSuffix(src, "/") {
		src += "/"
	}
	if !strings.HasSuffix(dst, "/") {
		dst += "/"
	}

	args = append(args, src, dst)

	cmd := exec.Command("rsync", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("rsync sync failed: %w\n%s", err, string(out))
	}

	return string(out), nil
}

// DirSize returns the size of a directory in bytes using `du`.
func DirSize(path string) (int64, error) {
	cmd := exec.Command("du", "-sk", path)
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	parts := strings.Fields(string(out))
	if len(parts) < 1 {
		return 0, fmt.Errorf("unexpected du output")
	}

	var size int64
	fmt.Sscanf(parts[0], "%d", &size)
	return size * 1024, nil // du -sk gives KB
}

// FormatSize formats bytes into human-readable string.
func FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
