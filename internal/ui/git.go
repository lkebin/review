package ui

import (
	"github.com/kbliu/review/internal/diff"
	"github.com/kbliu/review/internal/git"
)

// getFiles retrieves the list of changed files
func getFiles(opts Options) ([]FileInfo, error) {
	gitOpts := git.Options{
		Target:       opts.Target,
		Staged:       opts.Staged,
		ContextLines: opts.ContextLines,
	}

	files, err := git.GetFiles(gitOpts)
	if err != nil {
		return nil, err
	}

	// Convert to UI FileInfo
	result := make([]FileInfo, len(files))
	for i, f := range files {
		result[i] = FileInfo{
			Status: f.Status,
			Name:   f.Name,
		}
	}
	return result, nil
}

// getDiff retrieves the diff for a specific file
func getDiff(opts Options, file string) ([]DiffLine, error) {
	gitOpts := git.Options{
		Target:       opts.Target,
		Staged:       opts.Staged,
		ContextLines: opts.ContextLines,
	}

	content, err := git.GetDiff(gitOpts, file)
	if err != nil {
		return nil, err
	}

	lines := diff.Parse(content)

	// Convert to UI DiffLine
	result := make([]DiffLine, len(lines))
	for i, l := range lines {
		result[i] = DiffLine{
			Type:      LineType(l.Type),
			OldLineNo: l.OldLineNo,
			NewLineNo: l.NewLineNo,
			Content:   l.Content,
		}
	}
	return result, nil
}
