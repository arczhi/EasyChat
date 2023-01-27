package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var Cfg = &Config{}

type Config struct {
	MYSQL     Mysql           `json:"mysql"`
	REDIS     Redis           `json:"redis"`
	AES       Aes             `json:"aes"`
	HOSTNAME  string          `json:"hostname"`
	WEBSOCKET WebsocketConfig `json:"websocket"`
}

type Mysql struct {
	Addr     string `json:"addr"`
	Username string `json:"username"`
	Password string `json:"password"`
	Db       string `json:"db"`
}

type Redis struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	Db       string `json:"db"`
}

type Aes struct {
	Key string `json:"key"`
}

type WebsocketConfig struct {
	ConnectionTimeout int `json:"connection_timeout"`
}

func init() {
	Cfg = GetConfig()
}

func GetConfig() *Config {
	f, err := os.Open("./config/.cfg.json")
	if err != nil {
		panic(err)
	}
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	cfg := &Config{}
	if err := json.Unmarshal(buf, &cfg); err != nil {
		panic(err)
	}
	// fmt.Println(*cfg)
	return cfg
}
