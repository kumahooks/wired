package action

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
)

// Compile-time assertions that every exported action satisfies the Action contract.
var (
	_ Action = NoAction{}
	_ Action = QuitAction{}
	_ Action = OpenWhichKeyAction{}
	_ Action = ReloadConfigAction{}
	_ Action = ProceedFromInitAction{}
	_ Action = ActionCommand{Command: nil}
)

// closureComponent is a minimal Component whose HandleMessage delegates to a
// closure, so tests can seed the returned Action without a mock framework.
type closureComponent struct {
	handle func(tea.Msg) Action
}

func (component closureComponent) HandleMessage(msg tea.Msg) Action {
	return component.handle(msg)
}

var _ Component = closureComponent{handle: nil}

func TestClosureComponentReturnsSeededAction(t *testing.T) {
	t.Parallel()

	want := ProceedFromInitAction{}
	component := closureComponent{handle: func(_ tea.Msg) Action { return want }}

	got := component.HandleMessage(nil)
	assert.IsType(t, want, got)
}

func TestClosureComponentSatisfiesInterface(t *testing.T) {
	t.Parallel()

	var component Component = closureComponent{handle: func(_ tea.Msg) Action { return NoAction{} }}

	got := component.HandleMessage(nil)
	assert.IsType(t, NoAction{}, got)
}
