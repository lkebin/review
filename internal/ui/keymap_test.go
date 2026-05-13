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
		{"ctrl+f", FocusList, ActionPageDown},
		{"ctrl+b", FocusList, ActionPageUp},
		{"tab", FocusList, ActionFocusToggle},
		{"tab", FocusDiff, ActionFocusToggle},
		{">", FocusList, ActionGrowPanel},
		{"<", FocusDiff, ActionShrinkPanel},
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

func TestKeyMapFocusSwitch(t *testing.T) {
	km := NewKeyMapper()

	action := km.HandleKey(key("tab"), FocusList)
	if action != ActionFocusToggle {
		t.Errorf("tab = %v, want ActionFocusToggle", action)
	}

	km.Reset()
	action = km.HandleKey(key("tab"), FocusDiff)
	if action != ActionFocusToggle {
		t.Errorf("tab = %v, want ActionFocusToggle", action)
	}
}

func TestKeyMapResize(t *testing.T) {
	km := NewKeyMapper()

	action := km.HandleKey(key(">"), FocusList)
	if action != ActionGrowPanel {
		t.Errorf("> = %v, want ActionGrowPanel", action)
	}

	action = km.HandleKey(key("<"), FocusList)
	if action != ActionShrinkPanel {
		t.Errorf("< = %v, want ActionShrinkPanel", action)
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

	// g followed by non-g should produce ActionNone and reset state
	km.HandleKey(key("g"), FocusDiff)
	action := km.HandleKey(key("x"), FocusDiff)
	if action != ActionNone {
		t.Errorf("g x = %v, want ActionNone", action)
	}

	// Normal keys work after reset
	action = km.HandleKey(key("j"), FocusDiff)
	if action != ActionCursorDown {
		t.Errorf("j after invalid prefix = %v, want ActionCursorDown", action)
	}
}
