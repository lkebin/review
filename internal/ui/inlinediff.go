// internal/ui/inlinediff.go
package ui

import (
	"github.com/lkebin/review/internal/diff"
)

// InlineSpan marks a byte range in a line's content that was changed.
type InlineSpan struct {
	Start int
	End   int
}

// LinePair associates a removed line with an added line by index.
type LinePair struct {
	OldIdx int
	NewIdx int
}

// ComputeInlineDiff finds the changed character spans between two strings.
// Uses common-prefix/suffix approach: finds matching head and tail, marks middle as changed.
func ComputeInlineDiff(oldStr, newStr string) ([]InlineSpan, []InlineSpan) {
	oldRunes := []rune(oldStr)
	newRunes := []rune(newStr)

	prefixLen := 0
	for prefixLen < len(oldRunes) && prefixLen < len(newRunes) &&
		oldRunes[prefixLen] == newRunes[prefixLen] {
		prefixLen++
	}

	suffixLen := 0
	for suffixLen < len(oldRunes)-prefixLen && suffixLen < len(newRunes)-prefixLen &&
		oldRunes[len(oldRunes)-1-suffixLen] == newRunes[len(newRunes)-1-suffixLen] {
		suffixLen++
	}

	oldPrefixBytes := len(string(oldRunes[:prefixLen]))
	oldSuffixBytes := len(string(oldRunes[len(oldRunes)-suffixLen:]))
	newPrefixBytes := len(string(newRunes[:prefixLen]))
	newSuffixBytes := len(string(newRunes[len(newRunes)-suffixLen:]))

	oldStart := oldPrefixBytes
	oldEnd := len(oldStr) - oldSuffixBytes
	newStart := newPrefixBytes
	newEnd := len(newStr) - newSuffixBytes

	if oldStart >= oldEnd && newStart >= newEnd {
		return nil, nil
	}

	var oldSpans, newSpans []InlineSpan
	if oldStart < oldEnd {
		oldSpans = []InlineSpan{{Start: oldStart, End: oldEnd}}
	}
	if newStart < newEnd {
		newSpans = []InlineSpan{{Start: newStart, End: newEnd}}
	}
	return oldSpans, newSpans
}

// PairDiffLines pairs consecutive removed lines with consecutive added lines.
// Within a block of [-lines][+lines], they are paired 1:1 in order.
// Excess lines on either side are left unpaired.
func PairDiffLines(lines []diff.Line) []LinePair {
	var pairs []LinePair
	i := 0
	for i < len(lines) {
		remStart := i
		for i < len(lines) && lines[i].Type == diff.LineRemoved {
			i++
		}
		remEnd := i

		addStart := i
		for i < len(lines) && lines[i].Type == diff.LineAdded {
			i++
		}
		addEnd := i

		remCount := remEnd - remStart
		addCount := addEnd - addStart
		pairCount := remCount
		if addCount < pairCount {
			pairCount = addCount
		}
		for j := 0; j < pairCount; j++ {
			pairs = append(pairs, LinePair{
				OldIdx: remStart + j,
				NewIdx: addStart + j,
			})
		}

		if remStart == remEnd && addStart == addEnd {
			i++
		}
	}
	return pairs
}
