package main

import (
	"log"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
)

type cfgTunnel struct {
	InfName string `koanf:"interface"`
	IPRange string `koanf:"ip_range"`
}

type cfgPeers struct {
	LocalAddress  string `koanf:"local_address"`
	RemoteAddress string `koanf:"remote_address"`
}

// Config contains all config for running sartunnel
type Config struct {
	Tunnel cfgTunnel
	Peers  cfgPeers
}

var (
	k   = koanf.New(".")
	cfg Config
)

func initConfig() {
	if err := k.Load(file.Provider("config.toml"), toml.Parser()); err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	k.Unmarshal("tunnel", &cfg.Tunnel)
	k.Unmarshal("peers", &cfg.Peers)
}
