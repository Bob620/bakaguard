package ws

import (
	"encoding/json"
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
		"admin.auth",
		[]parameters.Param{
			&parameters.StringParam{Name: "password", IsRequired: true},
		}, func(params map[string]parameters.Param) (returnMessage json.RawMessage, err error) {
			if state.HasAdminAuth() {
				return []byte(`{"auth": true}`), nil
			}

			password, _ := params["password"].(*parameters.StringParam).GetString()
			if state.TryAdminPassword(password) {
				returnMessage = []byte(`{"auth": true}`)
			} else {
				returnMessage = []byte(`{"auth": false}`)
			}

			return returnMessage, err
		})

	rpcClient.RegisterMethod(
		"user.auth",
		[]parameters.Param{
			&parameters.StringParam{Name: "password", IsRequired: true},
		}, func(params map[string]parameters.Param) (returnMessage json.RawMessage, err error) {
			if state.HasUserAuth() {
				return []byte(`{"auth": true}`), nil
			}

			password, _ := params["password"].(*parameters.StringParam).GetString()
			if state.TryUserPassword(password) {
				returnMessage = []byte(`{"auth": true}`)
			} else {
				returnMessage = []byte(`{"auth": false}`)
			}

			return returnMessage, err
		})

	rpcClient.RegisterMethod(
		"peers.get",
		[]parameters.Param{
			&parameters.StringParam{Name: "uuid", IsRequired: true},
		}, func(params map[string]parameters.Param) (returnMessage json.RawMessage, err error) {
			if !state.HasAdminAuth() {
				return []byte(`{"error": "Please authenticate as admin"}`), nil
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
				return []byte(`{"error": "Please authenticate as admin"}`), nil
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
