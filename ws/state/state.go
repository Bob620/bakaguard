package state

import (
	"github.com/bob620/bakaguard/config"
)

type State struct {
	config    config.Websocket
	authScope int
}

func InitializeConnState(config config.Websocket) *State {
	return &State{
		config:    config,
		authScope: 0,
	}
}

func (state *State) HasAdminAuth() bool {
	return state.authScope >= 2
}

func (state *State) HasUserAuth() bool {
	return state.authScope >= 1
}

func (state *State) TryAdminPassword(password string) bool {
	pass := state.config.AdminPassword == password
	if pass {
		state.authScope = 2
	}

	return pass
}

func (state *State) TryUserPassword(password string) bool {
	pass := state.config.UserPassword == password
	if pass {
		state.authScope = 1
	}

	return pass
}
