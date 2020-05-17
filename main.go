package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gomodule/redigo/redis"
	"golang.zx2c4.com/wireguard/wgctrl"

	"github.com/bob620/bakaguard/config"
	"github.com/bob620/bakaguard/guard"
	"github.com/bob620/bakaguard/ws"
	"github.com/bob620/bakaguard/ws/state"
)

func main() {
	wg, err := wgctrl.New()
	if err != nil {
		log.Fatal("unable to connect to wireguard")
	}

	conf := config.LoadConfiguration()

	devices, err := wg.Devices()
	if err != nil {
		log.Fatal("unable to access wireguard")
	}

	fmt.Printf("Found %d Wireguard devices\n", len(devices))

	_, err = wg.Device(conf.Interface.Name)
	if os.IsNotExist(err) {
		log.Fatal("Unable to find ", conf.Interface.Name)
		//		fmt.Println("Making interface...")
		//		privateKey, _ := wgtypes.GeneratePrivateKey()
		//		wg.ConfigureDevice(conf.Interface.Name, wgtypes.Config{
		//			PrivateKey:   &privateKey,
		//			ListenPort:   nil,
		//			FirewallMark: nil,
		//			ReplacePeers: false,
		//			Peers:        nil,
		//		})
	}

	fmt.Println("Wireguard set up successfully")

	redisConn, err := redis.Dial("tcp", fmt.Sprintf(":%d", conf.Redis.Port),
		redis.DialDatabase(conf.Redis.Database),
	)
	if err != nil {
		log.Fatal("unable to connect to redis database")
	}

	fmt.Println("Redis connected")

	guard := guard.CreateGuard(conf, wg, redisConn)

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		connState := state.InitializeConnState(*conf.Websocket)

		socket := ws.CreateWs(guard, connState)
		socket.Handler(writer, request)
	})

	fmt.Println("Websocket listening on", conf.Websocket.Port)
	_ = http.ListenAndServe(fmt.Sprintf(":%d", conf.Websocket.Port), nil)
}
