package git

import (
	"bytes"
	"fmt"
	"os/exec"
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
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "\t", 2)
		if len(parts) >= 2 {
			files = append(files, FileInfo{
				Status: parts[0],
				Name:   parts[1],
			})
		}
	}
	return files
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
