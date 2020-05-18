package ws

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bob620/baka-rpc-go/parameters"
	"github.com/bob620/baka-rpc-go/rpc"
	"github.com/gorilla/websocket"

	"github.com/bob620/bakaguard/guard"
	"github.com/bob620/bakaguard/ws/state"
)

var upgrader = websocket.Upgrader{}
var rpcClient = rpc.CreateBakaRpc(nil, nil)

type WS struct {
	client *rpc.BakaRpc
}

func CreateWs(guard *guard.Guard, state *state.State) WS {
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
		},
		func(params map[string]parameters.Param) (returnMessage json.RawMessage, err error) {
			if !state.HasAdminAuth() {
				return nil, fmt.Errorf("please authenticate as admin")
			}

			uuid, _ := params["uuid"].(*parameters.StringParam).GetString()
			name, _ := params["name"].(*parameters.StringParam).GetString()
			desc, _ := params["description"].(*parameters.StringParam).GetString()

			peer, err := guard.GetWgPeer(uuid)
			if err != nil {
				return nil, err
			}

			if name != "" {
				peer.Name = name
			}

			if desc != "" {
				peer.Description = desc
			}

			err = guard.UpdatePeer(peer)
			if err != nil {
				return nil, err
			}
			return json.Marshal(peer)
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
