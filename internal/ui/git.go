package ui

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// getFiles retrieves the list of changed files
func getFiles(opts Options) ([]FileInfo, error) {
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

	// Parse output
	var files []FileInfo
	lines := strings.Split(out.String(), "\n")
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

	return files, nil
}

// getDiff retrieves the diff for a specific file
func getDiff(opts Options, file string) ([]DiffLine, error) {
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
	if err := cmd.Run(); err != nil {
		// git diff returns exit code 1 when there are differences
		// but still outputs the diff, so we ignore the error if we have output
		if out.Len() == 0 {
			return nil, fmt.Errorf("git diff failed: %w", err)
		}
	}

	return parseDiffOutput(out.String()), nil
}

// parseDiffOutput parses git diff output into DiffLines
func parseDiffOutput(output string) []DiffLine {
	var lines []DiffLine
	oldLine := 0
	newLine := 0
	inHunk := false

	for _, line := range strings.Split(output, "\n") {
		if line == "" {
			continue
		}

		// Check for hunk header: @@ -old_start,old_count +new_start,new_count @@
		if strings.HasPrefix(line, "@@") {
			inHunk = true
			lines = append(lines, DiffLine{
				Type:    LineHunkHeader,
				Content: line,
			})

			// Parse line numbers from hunk header
			// Format: @@ -start,count +start,count @@
			var oldStart, newStart int
			fmt.Sscanf(line, "@@ -%d,%*d +%d,%*d @@", &oldStart, &newStart)
			oldLine = oldStart
			newLine = newStart
			continue
		}

		if !inHunk {
			// Skip header lines (---, +++, diff --git, etc.)
			continue
		}

		if len(line) == 0 {
			continue
		}

		switch line[0] {
		case '+':
			lines = append(lines, DiffLine{
				Type:      LineAdded,
				NewLineNo: newLine,
				Content:   line,
			})
			newLine++

		case '-':
			lines = append(lines, DiffLine{
				Type:      LineRemoved,
				OldLineNo: oldLine,
				Content:   line,
			})
			oldLine++

		case ' ':
			lines = append(lines, DiffLine{
				Type:      LineContext,
				OldLineNo: oldLine,
				NewLineNo: newLine,
				Content:   line,
			})
			oldLine++
			newLine++
		}
	}

	return lines
}
