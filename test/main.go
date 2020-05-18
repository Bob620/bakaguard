package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/bob620/baka-rpc-go/parameters"
	"github.com/bob620/baka-rpc-go/rpc"
	"github.com/gorilla/websocket"

	"github.com/bob620/bakaguard/guard"
	"github.com/bob620/bakaguard/ws"
)

func main() {
	conn, _, err := websocket.DefaultDialer.Dial("ws://134.84.31.138:6065", nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()

	rpcClient := rpc.CreateBakaRpc(rpc.MakeSocketReaderChan(conn), rpc.MakeSocketWriterChan(conn))

	data, rpcErr := rpcClient.CallMethod(nil, "auth.admin", parameters.NewParametersByName([]parameters.Param{&parameters.StringParam{Default: "", Name: "password"}}))
	if rpcErr != nil {
		log.Fatal(rpcErr)
	}

	auth := ws.Auth{}
	_ = json.Unmarshal(*data, &auth)
	fmt.Println(auth)

	data, rpcErr = rpcClient.CallMethod(nil, "peers.all", parameters.NewParametersByName([]parameters.Param{}))
	if rpcErr != nil {
		log.Fatal(rpcErr)
	}

	peers := map[string]guard.Peer{}
	_ = json.Unmarshal(*data, &peers)
	fmt.Println(peers)

	_, rpcErr = rpcClient.CallMethod(nil, "peers.update", parameters.NewParametersByName([]parameters.Param{
		&parameters.StringParam{
			Name:    "uuid",
			Default: "b64ac4ab-fbac-43d5-4070-c2a0b886b2b8",
		}, &parameters.StringParam{
			Name:    "name",
			Default: "test",
		}, &parameters.StringParam{
			Name:    "description",
			Default: "ahhhhh",
		},
	}))
	if rpcErr != nil {
		log.Fatal(rpcErr)
	}

	data, rpcErr = rpcClient.CallMethod(nil, "peers.all", parameters.NewParametersByName([]parameters.Param{}))
	if rpcErr != nil {
		log.Fatal(rpcErr)
	}

	_ = json.Unmarshal(*data, &peers)
	fmt.Println(peers)
	fmt.Println("")
}
