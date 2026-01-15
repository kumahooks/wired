// Package core initializes bubbletea's model and the program's core state
package core

type CoreState struct {
	IsStarting bool
}

type CoreModel struct {
	CoreState *CoreState
}

func InitializeCoreState() *CoreState {
	return &CoreState{
		IsStarting: true,
	}
}

func NewCoreModel() *CoreModel {
	return &CoreModel{
		CoreState: InitializeCoreState(),
	}
}
