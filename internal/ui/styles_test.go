package ui

import "testing"

func TestDefaultThemeHasAllColors(t *testing.T) {
	th := DefaultTheme()

	checks := []struct {
		name  string
		color string
	}{
		{"AddedBg", th.AddedBg},
		{"RemovedBg", th.RemovedBg},
		{"AddedCursorBg", th.AddedCursorBg},
		{"RemovedCursorBg", th.RemovedCursorBg},
		{"CursorBg", th.CursorBg},
		{"LineNoBg", th.LineNoBg},
		{"LineNoFg", th.LineNoFg},
		{"SepFg", th.SepFg},
		{"HunkFg", th.HunkFg},
		{"NormalFg", th.NormalFg},
		{"StatusBarBg", th.StatusBarBg},
		{"StatusBarFg", th.StatusBarFg},
		{"FileSelectedBg", th.FileSelectedBg},
		{"FileSelectedFg", th.FileSelectedFg},
		{"InlineAddBg", th.InlineAddBg},
		{"InlineDelBg", th.InlineDelBg},
	}
	for _, c := range checks {
		if c.color == "" {
			t.Errorf("DefaultTheme().%s is empty", c.name)
		}
	}
}

func TestStatusColors(t *testing.T) {
	th := DefaultTheme()
	for _, status := range []string{"M", "A", "D", "R", "C"} {
		if th.StatusColor(status) == "" {
			t.Errorf("StatusColor(%q) returned empty", status)
		}
	}
}
