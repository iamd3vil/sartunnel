package main

import (
	"log"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
)

type cfgTunnel struct {
	InfName      string `koanf:"interface"`
	IPRange      string `koanf:"ip_range"`
	PrivateKey   string `koanf:"private_key"`
	LocalAddress string `koanf:"local_address"`
}

type cfgPeer struct {
	RemoteAddress string `koanf:"remote_address"`
	PublicKey     string `koanf:"public_key"`
}

// Config contains all config for running sartunnel
type Config struct {
	Tunnel cfgTunnel
	Peer   cfgPeer
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
	k.Unmarshal("peer", &cfg.Peer)
}
