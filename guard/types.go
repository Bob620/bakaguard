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
	Name        string
	Description string
	PublicKey   string
}

type Peer struct {
	Uuid          string `json:"uuid"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	PublicKey     string
	AllowedIPs    []net.IPNet `json:"allowedIPs"`
	KeepAlive     time.Duration
	LastHandshake time.Time
}
