package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func key(s string) tea.KeyMsg {
	switch s {
	case "ctrl+w":
		return tea.KeyMsg{Type: tea.KeyCtrlW}
	case "ctrl+d":
		return tea.KeyMsg{Type: tea.KeyCtrlD}
	case "ctrl+u":
		return tea.KeyMsg{Type: tea.KeyCtrlU}
	case "ctrl+f":
		return tea.KeyMsg{Type: tea.KeyCtrlF}
	case "ctrl+b":
		return tea.KeyMsg{Type: tea.KeyCtrlB}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func TestKeyMapSimpleKeys(t *testing.T) {
	km := NewKeyMapper()

	cases := []struct {
		key    string
		focus  FocusType
		expect Action
	}{
		{"j", FocusList, ActionCursorDown},
		{"k", FocusList, ActionCursorUp},
		{"j", FocusDiff, ActionCursorDown},
		{"k", FocusDiff, ActionCursorUp},
		{"enter", FocusList, ActionEnter},
		{"G", FocusList, ActionBottom},
		{"G", FocusDiff, ActionBottom},
		{"n", FocusDiff, ActionNextHunk},
		{"N", FocusDiff, ActionPrevHunk},
		{"q", FocusList, ActionQuit},
		{"?", FocusList, ActionHelp},
		{"ctrl+d", FocusDiff, ActionHalfPageDown},
		{"ctrl+u", FocusDiff, ActionHalfPageUp},
		{"ctrl+f", FocusDiff, ActionPageDown},
		{"ctrl+b", FocusDiff, ActionPageUp},
	}

	for _, tc := range cases {
		t.Run(tc.key+"_"+focusName(tc.focus), func(t *testing.T) {
			km.Reset()
			action := km.HandleKey(key(tc.key), tc.focus)
			if action != tc.expect {
				t.Errorf("HandleKey(%q, %v) = %v, want %v", tc.key, tc.focus, action, tc.expect)
			}
		})
	}
}

func focusName(f FocusType) string {
	if f == FocusList {
		return "list"
	}
	return "diff"
}

func TestKeyMapCtrlWPrefix(t *testing.T) {
	km := NewKeyMapper()

	action := km.HandleKey(key("ctrl+w"), FocusDiff)
	if action != ActionNone {
		t.Fatalf("Ctrl+W alone = %v, want ActionNone", action)
	}

	action = km.HandleKey(key("h"), FocusDiff)
	if action != ActionFocusLeft {
		t.Errorf("Ctrl+W h = %v, want ActionFocusLeft", action)
	}

	km.Reset()
	action = km.HandleKey(key("ctrl+w"), FocusList)
	if action != ActionNone {
		t.Fatalf("Ctrl+W alone = %v, want ActionNone", action)
	}
	action = km.HandleKey(key("l"), FocusList)
	if action != ActionFocusRight {
		t.Errorf("Ctrl+W l = %v, want ActionFocusRight", action)
	}
}

func TestKeyMapCtrlWResize(t *testing.T) {
	km := NewKeyMapper()

	km.HandleKey(key("ctrl+w"), FocusList)
	action := km.HandleKey(key(">"), FocusList)
	if action != ActionGrowPanel {
		t.Errorf("Ctrl+W > = %v, want ActionGrowPanel", action)
	}

	km.HandleKey(key("ctrl+w"), FocusList)
	action = km.HandleKey(key("<"), FocusList)
	if action != ActionShrinkPanel {
		t.Errorf("Ctrl+W < = %v, want ActionShrinkPanel", action)
	}
}

func TestKeyMapGGPrefix(t *testing.T) {
	km := NewKeyMapper()

	action := km.HandleKey(key("g"), FocusDiff)
	if action != ActionNone {
		t.Fatalf("single g = %v, want ActionNone", action)
	}

	action = km.HandleKey(key("g"), FocusDiff)
	if action != ActionTop {
		t.Errorf("gg = %v, want ActionTop", action)
	}
}

func TestKeyMapPrefixInvalidFollowUp(t *testing.T) {
	km := NewKeyMapper()

	km.HandleKey(key("ctrl+w"), FocusDiff)
	action := km.HandleKey(key("x"), FocusDiff)
	if action != ActionNone {
		t.Errorf("Ctrl+W x = %v, want ActionNone", action)
	}

	action = km.HandleKey(key("j"), FocusDiff)
	if action != ActionCursorDown {
		t.Errorf("j after invalid prefix = %v, want ActionCursorDown", action)
	}
}
