package config

import (
	"encoding/json"
	"log"
	"os"
)

const configLocation = "./config/config.json"

type Redis struct {
	Port     int `json:"port"`
	Database int `json:"db"`
}

type WSUsers struct {
	Password string              `json:"password"`
	Groups   map[string][]string `json:"groups"`
}

type Network struct {
	IP   string  `json:"ip"`
	Mask [4]byte `json:"mask"`
}

type WSGroup struct {
	Description string  `json:"description"`
	Network     Network `json:"network"`
}

type Websocket struct {
	Port          int                `json:"port"`
	AdminPassword string             `json:"adminPassword"`
	Users         map[string]WSUsers `json:"users"`
	Groups        map[string]WSGroup `json:"groups"`
}

type Interface struct {
	Name string `json:"name"`
}

type StorageType struct {
	Key  string `json:"key"`
	Type string `json:"type"`
}

type Config struct {
	Interface *Interface     `json:"interface"`
	Websocket *Websocket     `json:"ws"`
	Redis     *Redis         `json:"redis"`
	Storage   []*StorageType `json:"storage"`
}

func LoadConfiguration() (conf Config) {
	configFile, err := os.Open(configLocation)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&conf)
	return
}

func GetDefaultOf(storageType string) string {
	switch storageType {
	case "string":
		return ""
	case "bool":
		return "false"
	default:
		return ""
	}
}
