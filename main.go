package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dolmen-go/contextio"
	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
	"golang.org/x/net/ipv4"
)

const (
	// Name will be the name of the tunnel
	Name = "tun0"

	// IPAddr will be the IP address added to the interface
	IPAddr = "192.168.9.10/24"
)

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

	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = Name

	// Create a tunnel
	inf, err := water.New(config)
	if err != nil {
		log.Fatalf("error while creating a tun interface: %v", err)
	}
	defer inf.Close()

	link, err := netlink.LinkByName(Name)
	if err != nil {
		log.Fatalf("error while getting link: %v", err)
	}
	addr, _ := netlink.ParseAddr(IPAddr)
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

	// Read packets from the interface
	r := contextio.NewReader(ctx, inf)
	packet := make([]byte, 1500)
	for {
		n, err := r.Read(packet)
		if err != nil {
			log.Printf("error reading from the interface: %v", err)
			break
		}

		hdr, err := ipv4.ParseHeader(packet[:n])
		if err != nil {
			log.Printf("error while parsing ip header: %v", err)
			continue
		}

		log.Printf("got packet: %+v", hdr)
	}

	<-done
	log.Println("Exiting....")
}
