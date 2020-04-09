package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/dolmen-go/contextio"
	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
	"golang.org/x/net/ipv4"
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

	// Start a goroutine to listen to UDP packets
	if shouldStartServer() {
		startServer(ctx, inf)
	} else {
		log.Printf("Starting a client")
		startClient(ctx, inf)
	}

	<-done
	log.Println("Exiting....")
}

func startServer(ctx context.Context, inf *water.Interface) {
	s, err := NewServer(cfg.Peers.LocalAddress)
	if err != nil {
		log.Fatalf("error while starting UDP server: %v", err)
	}

	var rAddr *net.UDPAddr

	if cfg.Peers.RemoteAddress != "" {
		rAddr, err = net.ResolveUDPAddr("udp", cfg.Peers.RemoteAddress)
		if err != nil {
			log.Fatalf("invalid remote address, error: %v", err)
		}
	}

	// Read from the UDP Socket
	go func() {
		packet := make([]byte, 1500)
		w := contextio.NewWriter(ctx, inf)
		for {
			n, addr, err := s.Read(ctx, packet)
			if err != nil {
				log.Fatalf("error while reading from the UDP socket: %v", err)
			}

			// Parse the IP packet and find out source
			hdr, err := ipv4.ParseHeader(packet[:n])
			if err != nil {
				continue
			}

			// Save the client's UDP address so that we can reply back to the client
			s.SetClientAddr(hdr.Src.String(), addr)

			// Write to interface
			_, err = w.Write(packet[:n])
			if err != nil {
				log.Fatalf("error while writing to the interface: %v", err)
			}
		}
	}()

	// Read from interface and write back to socket
	go func() {
		packet := make([]byte, 1500)
		r := contextio.NewReader(ctx, inf)
		for {
			n, err := r.Read(packet)
			if err != nil {
				log.Fatalf("error while reading from the interface: %v", err)
			}

			// Parse the packet and find out the destination
			hdr, err := ipv4.ParseHeader(packet[:n])
			if err != nil {
				continue
			}

			// Get the UDP address for this client
			addr, ok := s.GetClientAddr(hdr.Dst.String())
			if !ok {
				// If remote address exists, send a packet there.
				if cfg.Peers.RemoteAddress != "" {
					_, err = s.Write(ctx, rAddr, packet[:n])
					if err != nil {
						log.Printf("error sending data to remote address: %v", err)
					}
				}
				continue
			}

			n, err = s.Write(ctx, addr, packet[:n])
			if err != nil {
				log.Printf("error sending data to client: %v", err)
				continue
			}
		}
	}()
}

func startClient(ctx context.Context, inf *water.Interface) {
	c, err := NewClient(cfg.Peers.RemoteAddress)
	if err != nil {
		log.Fatalf("error starting the client: %v", err)
	}

	// Read from the UDP socket
	go func() {
		packet := make([]byte, 1500)
		w := contextio.NewWriter(ctx, inf)
		for {
			n, err := c.Read(ctx, packet)
			if err != nil {
				log.Printf("error while reading from client UDP socket: %v", err)
				continue
			}

			// Write to interface
			_, err = w.Write(packet[:n])
			if err != nil {
				log.Fatalf("error while writing to the interface: %v", err)
			}
		}
	}()

	go func() {
		packet := make([]byte, 1500)
		r := contextio.NewReader(ctx, inf)
		for {
			n, err := r.Read(packet)
			if err != nil {
				log.Fatalf("error while reading from interface: %v", err)
			}

			// Write to socket
			_, err = c.Write(ctx, packet[:n])
			if err != nil {
				log.Printf("error while writing to client socket: %v", err)
				continue
			}
		}
	}()
}

func shouldStartServer() bool {
	if cfg.Peers.LocalAddress != "" {
		return true
	}

	return false
}
