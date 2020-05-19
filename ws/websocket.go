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
			&parameters.StringParam{Name: "password", IsRequired: true},
		}, func(params map[string]parameters.Param) (returnMessage json.RawMessage, err error) {
			if state.HasAdminAuth() {
				return json.Marshal(Auth{true})
			}

			password, _ := params["password"].(*parameters.StringParam).GetString()
			if state.TryAdminPassword(password) {
				returnMessage, err = json.Marshal(Auth{true})
			} else {
				returnMessage, err = json.Marshal(Auth{false})
			}

			return returnMessage, err
		})

	rpcClient.RegisterMethod(
		"auth.user",
		[]parameters.Param{
			&parameters.StringParam{Name: "password", IsRequired: true},
		}, func(params map[string]parameters.Param) (returnMessage json.RawMessage, err error) {
			if state.HasUserAuth() {
				return json.Marshal(Auth{true})
			}

			password, _ := params["password"].(*parameters.StringParam).GetString()
			if state.TryUserPassword(password) {
				returnMessage, err = json.Marshal(Auth{true})
			} else {
				returnMessage, err = json.Marshal(Auth{false})
			}

			return returnMessage, err
		})

	rpcClient.RegisterMethod(
		"peers.get",
		[]parameters.Param{
			&parameters.StringParam{Name: "uuid", IsRequired: true},
		}, func(params map[string]parameters.Param) (returnMessage json.RawMessage, err error) {
			if !state.HasAdminAuth() {
				return nil, fmt.Errorf("please authenticate as admin")
			}

			uuid, _ := params["uuid"].(*parameters.StringParam).GetString()
			peer, err := guard.GetWgPeer(uuid)
			if err != nil {
				return nil, err
			}
			return json.Marshal(peer)
		})

	rpcClient.RegisterMethod(
		"peers.all",
		[]parameters.Param{},
		func(params map[string]parameters.Param) (returnMessage json.RawMessage, err error) {
			if !state.HasAdminAuth() {
				return nil, fmt.Errorf("please authenticate as admin")
			}

			peers, err := guard.GetPeers()
			if err != nil {
				return nil, err
			}
			return json.Marshal(peers)
		})

	rpcClient.RegisterMethod(
		"peers.update",
		[]parameters.Param{
			&parameters.StringParam{Name: "uuid", IsRequired: true},
			&parameters.StringParam{Name: "name"},
			&parameters.StringParam{Name: "description"},
			&parameters.StringParam{Name: "keepAlive", Default: "-1s"},
			&InterfaceParam{Name: "storage", Default: map[string]interface{}{}},
			&IPNetParam{Name: "allowedIPs", Default: []net.IPNet{}},
		},
		func(params map[string]parameters.Param) (returnMessage json.RawMessage, err error) {
			if !state.HasAdminAuth() {
				return nil, fmt.Errorf("please authenticate as admin")
			}

			uuid, _ := params["uuid"].(*parameters.StringParam).GetString()
			name, _ := params["name"].(*parameters.StringParam).GetString()
			desc, _ := params["description"].(*parameters.StringParam).GetString()
			keepAlive, _ := params["keepAlive"].(*parameters.StringParam).GetString()
			allowedIPs, _ := params["allowedIPs"].(*IPNetParam).GetIPNet()
			storage, _ := params["storage"].(*InterfaceParam).GetInterface()

			peer, err := guard.GetWgPeer(uuid)
			if err != nil {
				fmt.Println(err)
				return nil, err
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

			err = guard.UpdatePeer(peer)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			return json.Marshal(peer)
		})

	rpcClient.RegisterMethod(
		"peers.add",
		[]parameters.Param{
			&parameters.StringParam{Name: "publicKey", IsRequired: true},
			&parameters.StringParam{Name: "name"},
			&parameters.StringParam{Name: "description"},
			&parameters.StringParam{Name: "keepAlive", Default: "-1s"},
			&InterfaceParam{Name: "storage", Default: map[string]interface{}{}},
			&IPNetParam{Name: "allowedIPs", Default: []net.IPNet{}},
		},
		func(params map[string]parameters.Param) (returnMessage json.RawMessage, err error) {
			if !state.HasAdminAuth() {
				return nil, fmt.Errorf("please authenticate as admin")
			}

			publicKey, _ := params["publicKey"].(*parameters.StringParam).GetString()
			name, _ := params["name"].(*parameters.StringParam).GetString()
			desc, _ := params["description"].(*parameters.StringParam).GetString()
			keepAlive, _ := params["keepAlive"].(*parameters.StringParam).GetString()
			allowedIPs, _ := params["allowedIPs"].(*IPNetParam).GetIPNet()
			storage, _ := params["storage"].(*InterfaceParam).GetInterface()

			keepAliveDuration := time.Duration(0)

			if keepAlive != "-1s" {
				keepAliveDuration, _ = time.ParseDuration(keepAlive)
			}

			peer := Guard.CreatePeer(publicKey, name, desc, keepAliveDuration, allowedIPs, guard.FormatUpdateStorage(nil, storage))

			err = guard.SetPeer(peer)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			return json.Marshal(peer)
		})

	rpcClient.RegisterMethod(
		"peers.delete",
		[]parameters.Param{
			&parameters.StringParam{Name: "uuid", IsRequired: true},
		},
		func(params map[string]parameters.Param) (returnMessage json.RawMessage, err error) {
			if !state.HasAdminAuth() {
				return nil, fmt.Errorf("please authenticate as admin")
			}

			uuid, _ := params["uuid"].(*parameters.StringParam).GetString()

			err = guard.DeletePeer(uuid)
			if err != nil {
				return nil, err
			}

			peers, err := guard.GetPeers()
			if err != nil {
				return nil, err
			}

			return json.Marshal(peers)
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
