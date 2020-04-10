package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aead/ecdh"
	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

func init() {
	initConfig()
}

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-sigs
		cancel()
		done <- true
	}()

	if len(os.Args) > 1 {
		if os.Args[1] == "genkey" {
			// First argument should be genkeys
			curve := ecdh.X25519()
			privKey, pubKey, err := GenerateKeys(curve)
			if err != nil {
				log.Fatalf("error while generating keys: %v", err)
			}

			fmt.Printf("Public Key: %s\nPrivateKey: %s\n", EncodePublicKey(pubKey), EncodePrivateKey(privKey))
			os.Exit(0)
		}
	}

	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = cfg.Tunnel.InfName

	// Create a tunnel
	inf, err := water.New(config)
	if err != nil {
		log.Fatalf("error while creating a tun interface: %v", err)
	}
	defer inf.Close()

	link, err := netlink.LinkByName(cfg.Tunnel.InfName)
	if err != nil {
		log.Fatalf("error while getting link: %v", err)
	}
	addr, _ := netlink.ParseAddr(cfg.Tunnel.IPRange)
	if err != nil {
		log.Fatalf("error parsing ip address: %v", err)
	}

	err = netlink.LinkSetMTU(link, 1300)
	if err != nil {
		log.Fatalf("error setting MTU: %v", err)
	}

	err = netlink.AddrAdd(link, addr)
	if err != nil {
		log.Fatalf("error assigning ip address: %v", err)
	}

	err = netlink.LinkSetUp(link)
	if err != nil {
		log.Fatalf("error bringing up the link: %v", err)
	}

	log.Printf("tunnel created with name: %s", inf.Name())

	// Create Env
	env, err := NewEnv(inf)
	if err != nil {
		log.Fatalf("error initializing app context: %v", err)
	}

	// Start a goroutine to listen to UDP packets
	if shouldStartServer() {
		env.startServer(ctx)
	} else {
		log.Printf("Starting a client")
		env.startClient(ctx)
	}

	<-done
	log.Println("Exiting....")
}
