package main

import (
	"fmt"
	"log"

	"github.com/bob620/baka-rpc-go/parameters"
	"github.com/bob620/baka-rpc-go/rpc"
	"github.com/gorilla/websocket"
)

func main() {
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()

	rpcClient := rpc.CreateBakaRpc(rpc.MakeSocketReaderChan(conn), rpc.MakeSocketWriterChan(conn))

	data, rpcErr := rpcClient.CallMethod(nil, "admin.auth", parameters.NewParametersByName([]parameters.Param{&parameters.StringParam{Default: "", Name: "password"}}))
	if rpcErr != nil {
		log.Fatal(rpcErr)
	}
	fmt.Println(data)

	data, rpcErr = rpcClient.CallMethod(nil, "admin.auth", parameters.NewParametersByName([]parameters.Param{}))
	if rpcErr != nil {
		log.Fatal(rpcErr)
	}
	fmt.Println(data)

}
