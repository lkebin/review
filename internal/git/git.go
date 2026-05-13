package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// FileInfo represents a file in the review
type FileInfo struct {
	Status string // A, M, D, R, C
	Name   string
}

// Options contains git operation options
type Options struct {
	Target       string
	Staged       bool
	ContextLines int
}

// GetFiles retrieves the list of changed files
func GetFiles(opts Options) ([]FileInfo, error) {
	// Check if we're in a git repo
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("not in a git repository")
	}

	// Build diff command
	args := []string{"diff", "--name-status", fmt.Sprintf("-U%d", opts.ContextLines)}
	if opts.Staged {
		args = append(args, "--cached")
	} else if opts.Target != "" {
		args = append(args, opts.Target)
	}

	cmd = exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}

	return parseFileStatus(out.String()), nil
}

func parseFileStatus(output string) []FileInfo {
	var files []FileInfo
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 2 {
			continue
		}
		status := parts[0]
		name := parts[1]
		// Renamed (R<score>) and copied (C<score>) have 3 fields: status, old_path, new_path.
		// Normalize status to a single letter and use the new path as the canonical name.
		if len(status) > 1 && (status[0] == 'R' || status[0] == 'C') && len(parts) == 3 {
			status = string(status[0])
			name = parts[2]
		}
		files = append(files, FileInfo{Status: status, Name: name})
	}
	return files
}

// GetFileStats retrieves added/removed line counts for all changed files in one git call.
// Returns a map from filename to [2]int{added, removed}. Binary files are omitted.
func GetFileStats(opts Options) (map[string][2]int, error) {
	args := []string{"diff", "--numstat"}
	if opts.Staged {
		args = append(args, "--cached")
	} else if opts.Target != "" {
		args = append(args, opts.Target)
	}

	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &bytes.Buffer{}
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git diff --numstat failed: %w", err)
	}
	return parseNumstat(out.String()), nil
}

func parseNumstat(output string) map[string][2]int {
	result := make(map[string][2]int)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 3 {
			continue
		}
		added, err1 := strconv.Atoi(parts[0])
		removed, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			continue // binary files show "-"
		}
		result[parts[2]] = [2]int{added, removed}
	}
	return result
}

// GetDiff retrieves the diff for a specific file
func GetDiff(opts Options, file string) (string, error) {
	args := []string{"diff", "--no-color", fmt.Sprintf("-U%d", opts.ContextLines)}
	if opts.Staged {
		args = append(args, "--cached")
	} else if opts.Target != "" {
		args = append(args, opts.Target)
	}
	args = append(args, "--", file)

	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &bytes.Buffer{}

	if err := cmd.Run(); err != nil {
		// git diff returns exit code 1 when there are differences
		// but still outputs the diff, so we ignore the error if we have output
		if out.Len() == 0 {
			return "", fmt.Errorf("git diff failed: %w", err)
		}
	}

	return out.String(), nil
}
