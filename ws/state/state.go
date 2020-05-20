package state

import (
	"net"

	"github.com/bob620/bakaguard/config"
)

type State struct {
	config     config.Websocket
	authScopes map[string]map[string]struct{}
	hasAdmin   bool
}

type Group struct {
	Description string
	Network     net.IPNet
}

func InitializeConnState(config config.Websocket) *State {
	return &State{
		config:     config,
		authScopes: map[string]map[string]struct{}{"auth.admin": {"*": {}}, "auth.user": {"*": {}}},
		hasAdmin:   false,
	}
}

func (state *State) GetGroupSettings(groupName string) *Group {
	group := state.config.Groups[groupName]

	return &Group{
		Description: group.Description,
		Network: net.IPNet{
			IP:   net.ParseIP(group.Network.IP),
			Mask: net.IPv4Mask(group.Network.Mask[0], group.Network.Mask[1], group.Network.Mask[2], group.Network.Mask[3]),
		},
	}
}

func (state *State) HasAdminAuth() bool {
	return state.hasAdmin
}

func (state *State) GetScopeGroups(scope string) map[string]struct{} {
	if state.HasAdminAuth() {
		return map[string]struct{}{"*": {}}
	}

	groups, _ := state.authScopes[scope]
	return groups
}

func (state *State) TryAdminPassword(password string) bool {
	pass := state.config.AdminPassword == password
	if pass {
		state.hasAdmin = true
	}

	return pass
}

func (state *State) TryUserLogin(username, password string) bool {
	user, _ := state.config.Users[username]
	if user.Password == password {
		for group, scopes := range user.Groups {
			for _, scope := range scopes {
				state.authScopes[scope][group] = struct{}{}
			}
		}
		return true
	}

	return false
}
