package ui

import (
	"testing"

	"github.com/lkebin/review/internal/diff"
)

func TestComputeInlineDiffSimple(t *testing.T) {
	old := "hello"
	new := "hallo"
	oldSpans, newSpans := ComputeInlineDiff(old, new)

	if len(oldSpans) != 1 {
		t.Fatalf("oldSpans count = %d, want 1", len(oldSpans))
	}
	if len(newSpans) != 1 {
		t.Fatalf("newSpans count = %d, want 1", len(newSpans))
	}

	if oldSpans[0].Start != 1 || oldSpans[0].End != 2 {
		t.Errorf("oldSpan = [%d,%d), want [1,2)", oldSpans[0].Start, oldSpans[0].End)
	}
	if newSpans[0].Start != 1 || newSpans[0].End != 2 {
		t.Errorf("newSpan = [%d,%d), want [1,2)", newSpans[0].Start, newSpans[0].End)
	}
}

func TestComputeInlineDiffPrefixInsertion(t *testing.T) {
	old := "println()"
	new := "fmt.Println()"
	oldSpans, newSpans := ComputeInlineDiff(old, new)

	if len(oldSpans) != 1 || len(newSpans) != 1 {
		t.Fatalf("spans = old:%d new:%d, want 1 each", len(oldSpans), len(newSpans))
	}
	if oldSpans[0].Start != 0 || oldSpans[0].End != 1 {
		t.Errorf("oldSpan = [%d,%d), want [0,1)", oldSpans[0].Start, oldSpans[0].End)
	}
	if newSpans[0].Start != 0 || newSpans[0].End != 5 {
		t.Errorf("newSpan = [%d,%d), want [0,5)", newSpans[0].Start, newSpans[0].End)
	}
}

func TestComputeInlineDiffIdentical(t *testing.T) {
	oldSpans, newSpans := ComputeInlineDiff("same", "same")
	if len(oldSpans) != 0 || len(newSpans) != 0 {
		t.Errorf("identical strings should have no spans")
	}
}

func TestComputeInlineDiffCompletelyDifferent(t *testing.T) {
	oldSpans, newSpans := ComputeInlineDiff("abc", "xyz")
	if len(oldSpans) != 1 {
		t.Fatalf("oldSpans = %d, want 1", len(oldSpans))
	}
	if oldSpans[0].Start != 0 || oldSpans[0].End != 3 {
		t.Errorf("oldSpan = [%d,%d), want [0,3)", oldSpans[0].Start, oldSpans[0].End)
	}
	if newSpans[0].Start != 0 || newSpans[0].End != 3 {
		t.Errorf("newSpan = [%d,%d), want [0,3)", newSpans[0].Start, newSpans[0].End)
	}
}

func TestPairDiffLines(t *testing.T) {
	lines := []diff.Line{
		{Type: diff.LineContext, Content: " context"},
		{Type: diff.LineRemoved, Content: "-old1"},
		{Type: diff.LineRemoved, Content: "-old2"},
		{Type: diff.LineAdded, Content: "+new1"},
		{Type: diff.LineAdded, Content: "+new2"},
		{Type: diff.LineContext, Content: " context"},
	}

	pairs := PairDiffLines(lines)
	if len(pairs) != 2 {
		t.Fatalf("pair count = %d, want 2", len(pairs))
	}
	if pairs[0].OldIdx != 1 || pairs[0].NewIdx != 3 {
		t.Errorf("pair[0] = {%d,%d}, want {1,3}", pairs[0].OldIdx, pairs[0].NewIdx)
	}
	if pairs[1].OldIdx != 2 || pairs[1].NewIdx != 4 {
		t.Errorf("pair[1] = {%d,%d}, want {2,4}", pairs[1].OldIdx, pairs[1].NewIdx)
	}
}

func TestPairDiffLinesUneven(t *testing.T) {
	lines := []diff.Line{
		{Type: diff.LineRemoved, Content: "-old"},
		{Type: diff.LineAdded, Content: "+new1"},
		{Type: diff.LineAdded, Content: "+new2"},
	}
	pairs := PairDiffLines(lines)
	if len(pairs) != 1 {
		t.Fatalf("pair count = %d, want 1", len(pairs))
	}
	if pairs[0].OldIdx != 0 || pairs[0].NewIdx != 1 {
		t.Errorf("pair[0] = {%d,%d}, want {0,1}", pairs[0].OldIdx, pairs[0].NewIdx)
	}
}
