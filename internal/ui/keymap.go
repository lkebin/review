package ui

import tea "github.com/charmbracelet/bubbletea"

// Action represents a semantic UI action.
type Action int

const (
	ActionNone Action = iota
	ActionQuit
	ActionHelp
	ActionCursorUp
	ActionCursorDown
	ActionEnter
	ActionTop
	ActionBottom
	ActionHalfPageDown
	ActionHalfPageUp
	ActionPageDown
	ActionPageUp
	ActionNextHunk
	ActionPrevHunk
	ActionFocusLeft
	ActionFocusRight
	ActionFocusToggle
	ActionGrowPanel
	ActionShrinkPanel
)

// keyState represents the state machine state.
type keyState int

const (
	stateNormal keyState = iota
	stateG
)

// KeyMapper translates key events into semantic Actions using a prefix key state machine.
type KeyMapper struct {
	state keyState
}

// NewKeyMapper creates a KeyMapper in Normal state.
func NewKeyMapper() *KeyMapper {
	return &KeyMapper{state: stateNormal}
}

// Reset returns the state machine to Normal.
func (km *KeyMapper) Reset() {
	km.state = stateNormal
}

// HandleKey processes a key event and returns the corresponding Action.
func (km *KeyMapper) HandleKey(msg tea.KeyMsg, focus FocusType) Action {
	switch km.state {
	case stateG:
		km.state = stateNormal
		if msg.String() == "g" {
			return ActionTop
		}
		return ActionNone
	default:
		return km.handleNormal(msg, focus)
	}
}

func (km *KeyMapper) handleNormal(msg tea.KeyMsg, focus FocusType) Action {
	s := msg.String()

	// Global keys (work in any focus)
	switch s {
	case "q", "ctrl+c":
		return ActionQuit
	case "?":
		return ActionHelp
	case "g":
		km.state = stateG
		return ActionNone
	case "G":
		return ActionBottom
	case "tab":
		return ActionFocusToggle
	case ">":
		return ActionGrowPanel
	case "<":
		return ActionShrinkPanel
	case "j", "down":
		return ActionCursorDown
	case "k", "up":
		return ActionCursorUp
	case "ctrl+f":
		return ActionPageDown
	case "ctrl+b":
		return ActionPageUp
	}

	// Focus-specific keys
	if focus == FocusList {
		switch s {
		case "enter":
			return ActionEnter
		}
	} else {
		switch s {
		case "ctrl+d":
			return ActionHalfPageDown
		case "ctrl+u":
			return ActionHalfPageUp
		case "n":
			return ActionNextHunk
		case "N":
			return ActionPrevHunk
		}
	}

	return ActionNone
}

