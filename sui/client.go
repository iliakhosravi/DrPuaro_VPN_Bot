package sui

import (
	"time"

	"github.com/google/uuid"
	"github.com/nrednav/cuid2"
)

const (
	ONE_GB = 1073741824
	ONE_MB = 1048576
)

type Link map[string]string
type Config map[string]map[string]any

var configNames = []string{"mixed", "socks", "http", "shadowsocks", "shadowsocks16", "tuic", "hysteria2"}

type Client struct {
	ID       *uint  `json:"id"`
	Enable   bool   `json:"enable"`
	Name     string `json:"name"`
	Inbounds []uint `json:"inbounds"`
	Links    []Link `json:"links"`
	Volume   int64  `json:"volume"`
	Expiry   int64  `json:"expiry"`
	Down     int64  `json:"down"`
	Up       int64  `json:"up"`
	Desc     string `json:"desc"`
	Group    string `json:"group"`
	Config   Config `json:"config"`
}

func InitClient(name string, inbounds []uint, volume int64, expiry time.Time, desc string, group string) *Client {
	return &Client{
		Enable:   true,
		Name:     name,
		Inbounds: inbounds,
		Links:    []Link{},
		Volume:   volume,
		Expiry:   expiry.Unix(),
		Down:     0,
		Up:       0,
		Desc:     desc,
		Group:    group,
		Config:   InitConfig(name),
	}
}

func InitConfig(name string) Config {
	password := cuid2.Generate()
	uid := uuid.NewString()
	config := Config{
		"mixed": {
			"username": name,
			"password": password,
		},
		"socks": {
			"username": name,
			"password": password,
		},
		"http": {
			"username": name,
			"password": password,
		},
		"shadowsocks": {
			"name":     name,
			"password": password,
		},
		"shadowsocks16": {
			"name":     name,
			"password": password,
		},
		"shadowtls": {
			"name":     name,
			"password": password,
		},
		"vmess": {
			"name":    name,
			"uuid":    uid,
			"alterId": 0,
		},
		"vless": {
			"name": name,
			"uuid": uid,
			"flow": "xtls-rprx-vision",
		},
		"trojan": {
			"name":     name,
			"password": password,
		},
		"naive": {
			"name":     name,
			"password": password,
		},
		"hysteria": {
			"name":     name,
			"auth_str": password,
		},
		"tuic": {
			"name":     name,
			"uuid":     uid,
			"password": password,
		},
		"hysteria2": {
			"name":     name,
			"password": password,
		},
	}
	return config
}

func (cl Client) RemainedTraffic() int {
	return int(cl.Volume-(cl.Down+cl.Up)) / ONE_MB
}
