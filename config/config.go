package config

import (
	"encoding/json"
	"log"
	"os"
)

const configLocation = "./config/config.json"

type Config struct {
	Interface struct {
		Name string `json:"name"`
	} `json:"interface"`
	Websocket struct {
		Port          int    `json:"port"`
		AdminPassword string `json:"adminPassword"`
		UserPassword  string `json:"userPassword"`
	} `json:"ws"`
	Redis struct {
		Port     int `json:"port"`
		Database int `json:"db"`
	} `json:"redis"`
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
