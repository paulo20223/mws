package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	cfg "github.com/paulo20223/mws/internal/config"
	"github.com/paulo20223/mws/internal/task"
	"github.com/paulo20223/mws/internal/workspace"
)

var (
	bold   = color.New(color.Bold)
	green  = color.New(color.FgGreen)
	yellow = color.New(color.FgYellow)
	red    = color.New(color.FgRed)
	cyan   = color.New(color.FgCyan)
	dim    = color.New(color.Faint)
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "mws",
		Short: "Multi-task workspace manager",
		Long:  "Manage multiple concurrent development tasks across repositories with full workspace isolation.",
	}

	rootCmd.AddCommand(
		initCmd(),
		createCmd(),
		listCmd(),
		deleteCmd(),
		statusCmd(),
		cdCmd(),
		openCmd(),
		baseCmd(),
		syncCmd(),
		shellInitCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// ── mws init ─────────────────────────────────────────────────────────────────

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize mws in current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, _ := os.Getwd()

			if workspace.IsInitialized(cwd) {
				yellow.Println("Workspace already initialized.")
				return nil
			}

			ws, err := workspace.Init(cwd)
			if err != nil {
				return err
			}

			green.Printf("Initialized mws workspace in %s\n\n", ws.Base)

			// Discover and display repos
			repos, err := ws.DiscoverRepos()
			if err != nil {
				return err
			}

			if len(repos) == 0 {
				dim.Println("No git repositories found.")
			} else {
				fmt.Printf("Found %d git repositories:\n\n", len(repos))
				tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintf(tw, "  REPO\tBRANCH\n")
				fmt.Fprintf(tw, "  ────\t──────\n")
				for _, r := range repos {
					fmt.Fprintf(tw, "  %s\t%s\n", r.RelPath, r.Branch)
				}
				tw.Flush()
			}

			fmt.Printf("\nConfig: %s\n", cfg.ConfigPath(ws.Base))
			fmt.Printf("\nExclude patterns:\n")
			for _, ex := range ws.Config.Copy.Exclude {
				dim.Printf("  - %s\n", ex)
			}

			return nil
		},
	}
}

// ── mws create ───────────────────────────────────────────────────────────────

func createCmd() *cobra.Command {
	var (
		branch      string
		repos       string
		description string
	)

	cmd := &cobra.Command{
		Use:   "create <task-name>",
		Short: "Create a new task (full workspace copy)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Validate task name
			if strings.Contains(name, "/") || strings.Contains(name, " ") || strings.Contains(name, "..") {
				return fmt.Errorf("invalid task name %q (no spaces, slashes, or '..')", name)
			}

			ws, _, err := workspace.ResolveWorkspace()
			if err != nil {
				return err
			}

			mgr := task.NewManager(ws)

			var repoGlobs []string
			if repos != "" {
				repoGlobs = strings.Split(repos, ",")
			}

			bold.Printf("Creating task: %s\n\n", name)

			manifest, err := mgr.Create(task.CreateOptions{
				Name:        name,
				Description: description,
				Branch:      branch,
				RepoGlobs:   repoGlobs,
			})
			if err != nil {
				return err
			}

			// Show result
			fmt.Println()
			green.Printf("Task %q created successfully!\n\n", name)

			fmt.Printf("  Path:   %s\n", manifest.Task.Copy)
			fmt.Printf("  Repos:  %d\n", len(manifest.Repos))
			if branch != "" {
				fmt.Printf("  Branch: %s\n", branch)
			}

			// Show size
			size, err := workspace.DirSize(manifest.Task.Copy)
			if err == nil {
				fmt.Printf("  Size:   %s\n", workspace.FormatSize(size))
			}

			fmt.Printf("\nTo work on this task:\n")
			cyan.Printf("  cd %s\n", manifest.Task.Copy)

			return nil
		},
	}

	cmd.Flags().StringVarP(&branch, "branch", "b", "", "Git branch to create/switch to in matching repos")
	cmd.Flags().StringVarP(&repos, "repos", "r", "", "Comma-separated glob patterns to filter repos (e.g. 'superApp*,vkAuto')")
	cmd.Flags().StringVarP(&description, "desc", "d", "", "Task description")

	return cmd
}

// ── mws list ─────────────────────────────────────────────────────────────────

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			ws, currentTask, err := workspace.ResolveWorkspace()
			if err != nil {
				return err
			}

			mgr := task.NewManager(ws)
			tasks, err := mgr.List()
			if err != nil {
				return err
			}

			if len(tasks) == 0 {
				dim.Println("No tasks yet. Create one with: mws create <task-name>")
				return nil
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintf(tw, "  TASK\tREPOS\tBRANCHES\tSIZE\tPATH\n")
			fmt.Fprintf(tw, "  ────\t─────\t────────\t────\t────\n")

			for _, t := range tasks {
				copyPath := t.Task.Copy
				if copyPath == "" {
					copyPath = cfg.TaskCopyPath(ws.Base, t.Task.Name)
				}

				// Check if copy exists
				exists := ""
				sizeStr := "-"
				if _, err := os.Stat(copyPath); err == nil {
					size, err := workspace.DirSize(copyPath)
					if err == nil {
						sizeStr = workspace.FormatSize(size)
					}
				} else {
					exists = " (missing)"
				}

				// Collect unique branches
				branches := uniqueBranches(t.Repos)

				// Mark current task
				marker := " "
				name := t.Task.Name
				if name == currentTask {
					marker = "*"
					name = green.Sprint(name)
				}

				fmt.Fprintf(tw, "%s %s\t%d\t%s\t%s\t%s%s\n",
					marker, name, len(t.Repos), branches, sizeStr, copyPath, exists)
			}
			tw.Flush()

			return nil
		},
	}
}

// ── mws delete ───────────────────────────────────────────────────────────────

func deleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "delete <task-name>",
		Aliases: []string{"rm"},
		Short:   "Delete a task (removes workspace copy)",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			ws, _, err := workspace.ResolveWorkspace()
			if err != nil {
				return err
			}

			mgr := task.NewManager(ws)

			if !force {
				fmt.Printf("Delete task %q and all its files? [y/N]: ", name)
				var answer string
				fmt.Scanln(&answer)
				if strings.ToLower(answer) != "y" {
					fmt.Println("Cancelled.")
					return nil
				}
			}

			bold.Printf("Deleting task: %s\n\n", name)

			if err := mgr.Delete(name, force); err != nil {
				return err
			}

			green.Printf("\nTask %q deleted.\n", name)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force delete even with uncommitted changes")

	return cmd
}

// ── mws status ───────────────────────────────────────────────────────────────

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status [task-name]",
		Short: "Show status of a task (or current task if in a task copy)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ws, currentTask, err := workspace.ResolveWorkspace()
			if err != nil {
				return err
			}

			mgr := task.NewManager(ws)

			var taskName string
			if len(args) > 0 {
				taskName = args[0]
			} else if currentTask != "" {
				taskName = currentTask
			} else {
				return fmt.Errorf("specify a task name, or run from inside a task copy")
			}

			status, err := mgr.GetStatus(taskName)
			if err != nil {
				return err
			}

			bold.Printf("Task: %s\n", status.Manifest.Task.Name)
			if status.Manifest.Task.Description != "" {
				dim.Printf("      %s\n", status.Manifest.Task.Description)
			}
			fmt.Printf("Path: %s\n", status.CopyPath)
			if !status.Exists {
				red.Println("Status: MISSING (workspace copy not found)")
				return nil
			}
			fmt.Println()

			if len(status.RepoStats) == 0 {
				dim.Println("No repos tracked.")
			} else {
				tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintf(tw, "  REPO\tBRANCH\tSTATUS\n")
				fmt.Fprintf(tw, "  ────\t──────\t──────\n")
				for _, rs := range status.RepoStats {
					statusStr := rs.Status.StatusString()
					if rs.Status.Clean {
						statusStr = green.Sprint(statusStr)
					} else {
						statusStr = yellow.Sprint(statusStr)
					}
					fmt.Fprintf(tw, "  %s\t%s\t%s\n", rs.RelPath, rs.Status.Branch, statusStr)
				}
				tw.Flush()
			}

			return nil
		},
	}
}

// ── mws cd ───────────────────────────────────────────────────────────────────

func cdCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cd <task-name>",
		Short: "Print the path to a task's workspace (use with: cd $(mws cd <task>))",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			ws, _, err := workspace.ResolveWorkspace()
			if err != nil {
				return err
			}

			copyPath := cfg.TaskCopyPath(ws.Base, name)
			if _, err := os.Stat(copyPath); err != nil {
				return fmt.Errorf("task %q not found at %s", name, copyPath)
			}

			// Print just the path (for shell integration)
			fmt.Print(copyPath)
			return nil
		},
	}
}

// ── mws open ─────────────────────────────────────────────────────────────────

func openCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "open <task-name>",
		Short: "Open a task's workspace in the configured editor",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			ws, _, err := workspace.ResolveWorkspace()
			if err != nil {
				return err
			}

			copyPath := cfg.TaskCopyPath(ws.Base, name)
			if _, err := os.Stat(copyPath); err != nil {
				return fmt.Errorf("task %q not found at %s", name, copyPath)
			}

			editor := ws.Config.Workspace.Editor
			if editor == "" {
				editor = "code"
			}

			fmt.Printf("Opening %s in %s ...\n", name, editor)
			editorCmd := exec.Command(editor, copyPath)
			return editorCmd.Start()
		},
	}
}

// ── mws base ─────────────────────────────────────────────────────────────────

func baseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "base",
		Short: "Print the base workspace path (use with: cd $(mws base))",
		RunE: func(cmd *cobra.Command, args []string) error {
			ws, _, err := workspace.ResolveWorkspace()
			if err != nil {
				return err
			}
			fmt.Print(ws.Base)
			return nil
		},
	}
}

// ── mws sync ─────────────────────────────────────────────────────────────────

func syncCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "sync <task-name>",
		Short: "Sync non-repo files from base workspace to task copy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			ws, _, err := workspace.ResolveWorkspace()
			if err != nil {
				return err
			}

			manifest, err := task.LoadManifest(ws.Base, name)
			if err != nil {
				return fmt.Errorf("task %q not found: %w", name, err)
			}

			copyPath := manifest.Task.Copy
			if copyPath == "" {
				copyPath = cfg.TaskCopyPath(ws.Base, name)
			}

			// Collect git repo paths to exclude from sync
			var repoPaths []string
			for _, r := range manifest.Repos {
				repoPaths = append(repoPaths, r.Path)
			}

			if dryRun {
				bold.Println("Dry run — no changes will be made:")
				fmt.Println()
			} else {
				bold.Printf("Syncing base -> %s\n\n", name)
			}

			output, err := workspace.SyncFromBase(ws.Base, copyPath, ws.Config.Copy.Exclude, repoPaths, dryRun)
			if err != nil {
				return err
			}

			if strings.TrimSpace(output) == "" {
				green.Println("Already up to date.")
			} else {
				fmt.Println(output)
				if !dryRun {
					green.Println("Sync complete.")
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be synced without making changes")

	return cmd
}

// ── helpers ──────────────────────────────────────────────────────────────────

func uniqueBranches(repos []task.RepoRef) string {
	seen := make(map[string]bool)
	var branches []string
	for _, r := range repos {
		if !seen[r.Branch] {
			seen[r.Branch] = true
			branches = append(branches, r.Branch)
		}
	}
	if len(branches) == 0 {
		return "-"
	}
	result := strings.Join(branches, ", ")
	if len(result) > 40 {
		result = result[:37] + "..."
	}
	return result
}

// ── mws shell-init ──────────────────────────────────────────────────────────

const shellInitCode = `# mws shell integration — managed by 'mws shell-init'

# quick navigation: mcd <task-name>
mcd() { cd "$(mws cd "$1")" }

# return to base workspace
mbase() { cd "$(mws base)" }

# show current mws task in prompt
_mws_prompt() {
  local f="$PWD"
  while [[ "$f" != "/" ]]; do
    if [[ -f "$f/.mws-task" ]]; then
      printf "[mws:%s] " "$(cat "$f/.mws-task")"
      return
    fi
    f="$(dirname "$f")"
  done
}

# inject into PROMPT if not already present
if [[ "$PROMPT" != *'$(_mws_prompt)'* ]]; then
  PROMPT='$(_mws_prompt)'"$PROMPT"
fi
`

func shellInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "shell-init",
		Short: "Output shell integration code (add to ~/.zshrc: eval \"$(mws shell-init)\")",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			os.Stdout.WriteString(shellInitCode)
		},
	}
}
