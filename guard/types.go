package guard

import (
	"net"
	"time"

	"github.com/gomodule/redigo/redis"
	"golang.zx2c4.com/wireguard/wgctrl"

	"github.com/bob620/bakaguard/config"
)

type Guard struct {
	config    config.Config
	wg        *wgctrl.Client
	redisConn redis.Conn
}

type RedisPeer struct {
	Uuid        string
	Group       string
	Name        string
	Description string
	PublicKey   string
	Storage     map[string]string
}

type Peer struct {
	Uuid           string `json:"uuid"`
	Group          string `json:"group"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	PublicKey      string
	AllowedIPs     []net.IPNet       `json:"allowedIPs"`
	KeepAlive      time.Duration     `json:"keepAlive"`
	LastHandshake  time.Time         `json:"lastSeen"`
	LastExternalIp net.UDPAddr       `json:"lastExternalIp"`
	Storage        map[string]string `json:"storage"`
}
