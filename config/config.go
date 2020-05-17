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

type Websocket struct {
	Port          int    `json:"port"`
	AdminPassword string `json:"adminPassword"`
	UserPassword  string `json:"userPassword"`
}

type Interface struct {
	Name string `json:"name"`
}

type Config struct {
	Interface *Interface `json:"interface"`
	Websocket *Websocket `json:"ws"`
	Redis     *Redis     `json:"redis"`
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
