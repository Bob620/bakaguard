package ws

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/bob620/baka-rpc-go/parameters"
	"github.com/bob620/baka-rpc-go/rpc"
	"github.com/gorilla/websocket"

	Guard "github.com/bob620/bakaguard/guard"
	"github.com/bob620/bakaguard/ws/state"
)

var upgrader = websocket.Upgrader{}
var rpcClient = rpc.CreateBakaRpc(nil, nil)

type WS struct {
	client *rpc.BakaRpc
}

func CreateWs(guard *Guard.Guard, state *state.State) WS {
	ws := WS{rpcClient}

	rpcClient.RegisterMethod(
		"auth.admin",
		[]parameters.Param{
			&parameters.StringParam{Name: "password", Required: true},
		},
		func(params map[string]parameters.Param) (json.RawMessage, error) {
			if state.HasAdminAuth() {
				return json.Marshal(Auth{true})
			}

			password, _ := params["password"].(*parameters.StringParam).GetString()

			return json.Marshal(Auth{state.TryAdminPassword(password)})
		})

	rpcClient.RegisterMethod(
		"auth.user",
		[]parameters.Param{
			&parameters.StringParam{Name: "username", Required: true},
			&parameters.StringParam{Name: "password", Required: true},
		},
		func(params map[string]parameters.Param) (json.RawMessage, error) {
			username, _ := params["username"].(*parameters.StringParam).GetString()
			password, _ := params["password"].(*parameters.StringParam).GetString()

			return json.Marshal(Auth{state.TryUserLogin(username, password)})
		})

	rpcClient.RegisterMethod(
		"peers.get",
		[]parameters.Param{
			&parameters.StringParam{Name: "uuid", Required: true},
		},
		func(params map[string]parameters.Param) (json.RawMessage, error) {
			validGroups := state.GetScopeGroups("peers.get")

			if len(validGroups) == 0 {
				return nil, fmt.Errorf("please authenticate")
			}

			uuid, _ := params["uuid"].(*parameters.StringParam).GetString()
			peer, err := guard.GetWgPeer(uuid)
			if err != nil {
				return nil, err
			}

			_, ok := validGroups[peer.Group]
			_, adminOk := validGroups["*"]

			if ok || adminOk {
				return json.Marshal(peer)
			}
			return nil, fmt.Errorf("peer not found")
		})

	rpcClient.RegisterMethod(
		"peers.all",
		[]parameters.Param{},
		func(params map[string]parameters.Param) (json.RawMessage, error) {
			validGroups := state.GetScopeGroups("peers.all")

			if len(validGroups) == 0 {
				return nil, fmt.Errorf("please authenticate")
			}

			_, ok := validGroups["*"]
			if ok {
				peers, err := guard.GetPeers()
				if err != nil {
					return nil, err
				}
				return json.Marshal(peers)
			}

			peers := make(map[string]*Guard.Peer, len(validGroups)*3)
			for group, _ := range validGroups {
				groupPeers, err := guard.GetGroupPeers(group)
				if err != nil {
					continue
				}

				for _, peer := range groupPeers {
					peers[peer.Uuid] = peer
				}
			}

			return json.Marshal(peers)
		})

	rpcClient.RegisterMethod(
		"peers.getGroup",
		[]parameters.Param{
			&parameters.StringParam{Name: "group", Required: true},
		},
		func(params map[string]parameters.Param) (json.RawMessage, error) {
			validGroups := state.GetScopeGroups("peers.getGroup")

			if len(validGroups) == 0 {
				return nil, fmt.Errorf("please authenticate")
			}

			group, _ := params["group"].(*parameters.StringParam).GetString()
			_, ok := validGroups[group]
			_, adminOk := validGroups["*"]

			if !ok && !adminOk {
				return json.Marshal(map[string]*Guard.Peer{})
			}

			peers, err := guard.GetGroupPeers(group)
			if err != nil {
				return nil, err
			}
			return json.Marshal(peers)
		})

	rpcClient.RegisterMethod(
		"peers.update",
		[]parameters.Param{
			&parameters.StringParam{Name: "uuid", Required: true},
			&parameters.StringParam{Name: "name"},
			&parameters.StringParam{Name: "description"},
			&parameters.StringParam{Name: "keepAlive", Default: "-1s"},
			&InterfaceParam{Name: "storage", Default: map[string]interface{}{}},
			&IPNetParam{Name: "allowedIPs", Default: []net.IPNet{}},
		},
		func(params map[string]parameters.Param) (json.RawMessage, error) {
			validGroups := state.GetScopeGroups("peers.update")

			if len(validGroups) == 0 {
				return nil, fmt.Errorf("please authenticate")
			}

			uuid, _ := params["uuid"].(*parameters.StringParam).GetString()
			name, _ := params["name"].(*parameters.StringParam).GetString()
			desc, _ := params["description"].(*parameters.StringParam).GetString()
			keepAlive, _ := params["keepAlive"].(*parameters.StringParam).GetString()
			allowedIPs, _ := params["allowedIPs"].(*IPNetParam).GetIPNet()
			storage, _ := params["storage"].(*InterfaceParam).GetInterface()

			fmt.Println(uuid)
			fmt.Println(name)
			fmt.Println(desc)
			fmt.Println(keepAlive)
			fmt.Println(allowedIPs)
			fmt.Println(storage)

			fmt.Println("Looking up peer")

			peer, err := guard.GetWgPeer(uuid)
			if err != nil {
				fmt.Print("1: ")
				fmt.Println(err)
				return nil, err
			}

			_, ok := validGroups[peer.Group]
			_, adminOk := validGroups["*"]

			if !ok && !adminOk {
				fmt.Println("peer not found")
				return nil, fmt.Errorf("peer not found")
			}

			if name != "" {
				peer.Name = name
			}

			if desc != "" {
				peer.Description = desc
			}

			if keepAlive != "-1s" {
				peer.KeepAlive, _ = time.ParseDuration(keepAlive)
			}

			if len(allowedIPs) > 0 {
				peer.AllowedIPs = allowedIPs
			}

			if len(storage) > 0 {
				peer.Storage = guard.FormatUpdateStorage(peer.Storage, storage)
			}

			fmt.Println("Attempting to update peer...")
			err = guard.UpdatePeer(peer)
			if err != nil {
				fmt.Print("2: ")
				fmt.Println(err)
				return nil, err
			}
			fmt.Println("Peer updated")
			return json.Marshal(peer)
		})

	rpcClient.RegisterMethod(
		"peers.add",
		[]parameters.Param{
			&parameters.StringParam{Name: "publicKey", Required: true},
			&parameters.StringParam{Name: "group", Required: true},
			&parameters.StringParam{Name: "name"},
			&parameters.StringParam{Name: "description"},
			&parameters.StringParam{Name: "keepAlive", Default: "-1s"},
			&InterfaceParam{Name: "storage", Default: map[string]interface{}{}},
			&IPNetParam{Name: "allowedIPs", Default: []net.IPNet{}},
		},
		func(params map[string]parameters.Param) (json.RawMessage, error) {
			validGroups := state.GetScopeGroups("peers.add")

			if len(validGroups) == 0 {
				return nil, fmt.Errorf("please authenticate")
			}

			publicKey, _ := params["publicKey"].(*parameters.StringParam).GetString()
			group, _ := params["group"].(*parameters.StringParam).GetString()
			name, _ := params["name"].(*parameters.StringParam).GetString()
			desc, _ := params["description"].(*parameters.StringParam).GetString()
			keepAlive, _ := params["keepAlive"].(*parameters.StringParam).GetString()
			allowedIPs, _ := params["allowedIPs"].(*IPNetParam).GetIPNet()
			storage, _ := params["storage"].(*InterfaceParam).GetInterface()

			_, ok := validGroups[group]
			_, adminOk := validGroups["*"]

			if !ok && !adminOk {
				return nil, fmt.Errorf("please authenticate")
			}

			keepAliveDuration := time.Duration(0)

			if keepAlive != "-1s" {
				keepAliveDuration, _ = time.ParseDuration(keepAlive)
			}

			peer := Guard.CreatePeer(
				publicKey,
				group,
				name,
				desc,
				keepAliveDuration,
				append(allowedIPs, state.GetGroupSettings(group).Network),
				guard.FormatUpdateStorage(nil, storage),
			)

			err := guard.SetPeer(peer)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			return json.Marshal(peer)
		})

	rpcClient.RegisterMethod(
		"peers.delete",
		[]parameters.Param{
			&parameters.StringParam{Name: "uuid", Required: true},
		},
		func(params map[string]parameters.Param) (json.RawMessage, error) {
			validGroups := state.GetScopeGroups("peers.delete")

			if len(validGroups) == 0 {
				return nil, fmt.Errorf("please authenticate")
			}

			uuid, _ := params["uuid"].(*parameters.StringParam).GetString()
			peer, err := guard.GetWgPeer(uuid)
			if err != nil {
				return json.Marshal([]byte(`{"done":true}`))
			}

			_, ok := validGroups[peer.Group]
			_, adminOk := validGroups["*"]

			if !ok && !adminOk {
				return nil, fmt.Errorf("please authenticate")
			}

			err = guard.DeletePeer(uuid)
			if err != nil {
				return nil, err
			}

			return json.Marshal([]byte(`{"done":true}`))
		})

	return ws
}

func (ws *WS) Handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	defer conn.Close()
	ws.client.UseChannels(rpc.MakeSocketReaderChan(conn), rpc.MakeSocketWriterChan(conn))
}
