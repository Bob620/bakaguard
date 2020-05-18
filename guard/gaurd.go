package guard

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
	uuid "github.com/nu7hatch/gouuid"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/bob620/bakaguard/config"
)

const redisRoot = "bakaguard"
const redisPeer = "peers"
const peerSearchPublicKey = "search:publicKey"

func CreateRedisPeer(publicKey string, name string, description string) *RedisPeer {
	id, _ := uuid.NewV4()

	return &RedisPeer{
		Uuid:        id.String(),
		Name:        name,
		Description: description,
		PublicKey:   publicKey,
	}
}

func CreateGuard(conf config.Config, wg *wgctrl.Client, redisConn redis.Conn) *Guard {
	return &Guard{
		config:    conf,
		wg:        wg,
		redisConn: redisConn,
	}
}

func (guard *Guard) CleanupPeers() error {
	device, err := guard.wg.Device(guard.config.Interface.Name)
	if err != nil {
		return err
	}

	peerList, err := guard.GetRedisPeerMap()
	if err != nil {
		return err
	}

	for _, key := range device.Peers {
		keyString := key.PublicKey.String()
		uuid := peerList[keyString]
		if uuid != "" {
			delete(peerList, keyString)
		} else {
			_ = guard.SetRedisPeer(CreateRedisPeer(
				keyString,
				"",
				"",
			))
		}
	}

	for peerKey, peerUuid := range peerList {
		_ = guard.DeleteRedisPeer(peerUuid, peerKey)
	}

	return nil
}

func (guard *Guard) UpdatePeer(peer *Peer) (err error) {
	if peer.Uuid == "" {
		return fmt.Errorf("no uuid provided")
	}

	publicKey, err := wgtypes.ParseKey(peer.PublicKey)
	if err != nil {
		return fmt.Errorf("unable to verify public key")
	}

	err = guard.wg.ConfigureDevice(guard.config.Interface.Name, wgtypes.Config{
		PrivateKey:   nil,
		ListenPort:   nil,
		FirewallMark: nil,
		ReplacePeers: false,
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey:                   publicKey,
				Remove:                      false,
				UpdateOnly:                  true,
				PresharedKey:                nil,
				Endpoint:                    nil,
				PersistentKeepaliveInterval: &peer.KeepAlive,
				ReplaceAllowedIPs:           true,
				AllowedIPs:                  peer.AllowedIPs,
			},
		},
	})

	if err != nil {
		return fmt.Errorf("unable to update peer configuration")
	}

	err = guard.SetRedisPeer(&RedisPeer{
		Uuid:        peer.Uuid,
		Name:        peer.Name,
		Description: peer.Description,
		PublicKey:   peer.PublicKey,
	})
	if err != nil {
		return fmt.Errorf("unable to update peer configuration")
	}

	return
}

func (guard *Guard) GetRedisPeerMap() (peers map[string]string, err error) {
	peers = make(map[string]string)

	keys, err := redis.Strings(guard.redisConn.Do("smembers", fmt.Sprintf("%s:%s", redisRoot, peerSearchPublicKey)))
	if err != nil {
		return
	}

	for _, key := range keys {
		uuid, _ := redis.String(guard.redisConn.Do("get", fmt.Sprintf("%s:%s:%s", redisRoot, peerSearchPublicKey, key)))
		peers[key] = uuid
	}

	return
}

func (guard *Guard) DeleteRedisPeer(uuid string, publicKey string) error {
	_, err := guard.redisConn.Do("srem", fmt.Sprintf("%s:%s", redisRoot, redisPeer), uuid)
	if err != nil {
		return err
	}

	_, err = guard.redisConn.Do("srem", fmt.Sprintf("%s:%s", redisRoot, peerSearchPublicKey), publicKey)
	if err != nil {
		return err
	}

	_, err = guard.redisConn.Do("del", fmt.Sprintf("%s:%s:%s:uuid", redisRoot, redisPeer, uuid))
	if err != nil {
		return err
	}

	_, err = guard.redisConn.Do("del", fmt.Sprintf("%s:%s:%s:name", redisRoot, redisPeer, uuid))
	if err != nil {
		return err
	}

	_, err = guard.redisConn.Do("del", fmt.Sprintf("%s:%s:%s:desc", redisRoot, redisPeer, uuid))
	if err != nil {
		return err
	}

	_, err = guard.redisConn.Do("del", fmt.Sprintf("%s:%s:%s:publicKey", redisRoot, redisPeer, uuid))
	if err != nil {
		return err
	}

	_, err = guard.redisConn.Do("del", fmt.Sprintf("%s:%s:%s", redisRoot, peerSearchPublicKey, publicKey))
	if err != nil {
		return err
	}

	return nil
}

func (guard *Guard) SetRedisPeer(peer *RedisPeer) error {
	_, err := guard.redisConn.Do("set", fmt.Sprintf("%s:%s:%s:uuid", redisRoot, redisPeer, peer.Uuid), peer.Uuid)
	if err != nil {
		return err
	}

	_, err = guard.redisConn.Do("set", fmt.Sprintf("%s:%s:%s:name", redisRoot, redisPeer, peer.Uuid), peer.Name)
	if err != nil {
		return err
	}

	_, err = guard.redisConn.Do("set", fmt.Sprintf("%s:%s:%s:desc", redisRoot, redisPeer, peer.Uuid), peer.Description)
	if err != nil {
		return err
	}

	_, err = guard.redisConn.Do("set", fmt.Sprintf("%s:%s:%s:publicKey", redisRoot, redisPeer, peer.Uuid), peer.PublicKey)
	if err != nil {
		return err
	}

	_, err = guard.redisConn.Do("sadd", fmt.Sprintf("%s:%s", redisRoot, redisPeer), peer.Uuid)
	if err != nil {
		return err
	}

	_, err = guard.redisConn.Do("sadd", fmt.Sprintf("%s:%s", redisRoot, peerSearchPublicKey), peer.PublicKey)
	if err != nil {
		return err
	}

	_, err = guard.redisConn.Do("set", fmt.Sprintf("%s:%s:%s", redisRoot, peerSearchPublicKey, peer.PublicKey), peer.Uuid)
	if err != nil {
		return err
	}

	return nil
}

func (guard *Guard) GetPeers() (peers map[string]*Peer, err error) {
	peers = make(map[string]*Peer)

	peerMap, err := guard.GetRedisPeerMap()
	if err != nil {
		return nil, err
	}

	for _, peerUuid := range peerMap {
		peer, err := guard.GetWgPeer(peerUuid)
		if err == nil {
			peers[peerUuid] = peer
		}
	}

	return
}

func (guard *Guard) GetRedisPeer(id string) (*RedisPeer, error) {
	peer := RedisPeer{
		Uuid:        "",
		Name:        "",
		Description: "",
		PublicKey:   "",
	}
	data, err := redis.String(guard.redisConn.Do("get", fmt.Sprintf("%s:%s:%s:uuid", redisRoot, redisPeer, id)))
	if err != nil {
		return nil, err
	}
	peer.Uuid = data

	data, err = redis.String(guard.redisConn.Do("get", fmt.Sprintf("%s:%s:%s:name", redisRoot, redisPeer, id)))
	if err != nil {
		return nil, err
	}
	peer.Name = data

	data, err = redis.String(guard.redisConn.Do("get", fmt.Sprintf("%s:%s:%s:desc", redisRoot, redisPeer, id)))
	if err != nil {
		return nil, err
	}
	peer.Description = data

	data, err = redis.String(guard.redisConn.Do("get", fmt.Sprintf("%s:%s:%s:publicKey", redisRoot, redisPeer, id)))
	if err != nil {
		return nil, err
	}
	peer.PublicKey = data

	return &peer, nil
}

func (guard *Guard) GetWgPeer(id string) (*Peer, error) {
	device, err := guard.wg.Device(guard.config.Interface.Name)
	if err != nil {
		return nil, err
	}

	redisPeer, err := guard.GetRedisPeer(id)
	if err != nil || redisPeer.Uuid == "" {
		if err == nil {
			err = fmt.Errorf("unable to find peer")
		}
		return nil, err
	}
	redisKey, _ := wgtypes.ParseKey(redisPeer.PublicKey)

	for _, peer := range device.Peers {
		if peer.PublicKey == redisKey {
			return &Peer{
				Uuid:          redisPeer.Uuid,
				Name:          redisPeer.Name,
				Description:   redisPeer.Description,
				PublicKey:     redisPeer.PublicKey,
				AllowedIPs:    peer.AllowedIPs,
				KeepAlive:     peer.PersistentKeepaliveInterval,
				LastHandshake: peer.LastHandshakeTime,
			}, nil
		}
	}

	return nil, fmt.Errorf("unable to find peer")
}
