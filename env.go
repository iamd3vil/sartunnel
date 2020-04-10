package main

import (
	"context"
	"crypto"
	"log"
	"net"

	"github.com/aead/ecdh"
	"github.com/dolmen-go/contextio"
	"github.com/songgao/water"
	"golang.org/x/net/ipv4"
)

// Env contains all the environment the app needs
type Env struct {
	client       *Client
	server       *Server
	inf          *water.Interface
	curve        ecdh.KeyExchange
	localPrivKey crypto.PrivateKey
	peerPubKey   crypto.PublicKey
	key          []byte
}

// NewEnv returns a Env instance with a curve intialized
func NewEnv(inf *water.Interface) (*Env, error) {
	localPrivKey, err := DecodePrivateKey(cfg.Tunnel.PrivateKey)
	if err != nil {
		return nil, err
	}
	peerPublicKey, err := DecodePublicKey(cfg.Peer.PublicKey)
	if err != nil {
		return nil, err
	}

	curve := ecdh.X25519()

	if err = curve.Check(peerPublicKey); err != nil {
		return nil, err
	}

	// Compute secret
	key := curve.ComputeSecret(localPrivKey, peerPublicKey)

	return &Env{
		inf:          inf,
		curve:        ecdh.X25519(),
		localPrivKey: localPrivKey,
		peerPubKey:   peerPublicKey,
		key:          key,
	}, nil
}

func (env *Env) startServer(ctx context.Context) {
	s, err := NewServer(cfg.Tunnel.LocalAddress)
	if err != nil {
		log.Fatalf("error while starting UDP server: %v", err)
	}

	var rAddr *net.UDPAddr

	if cfg.Peer.RemoteAddress != "" {
		rAddr, err = net.ResolveUDPAddr("udp", cfg.Peer.RemoteAddress)
		if err != nil {
			log.Fatalf("invalid remote address, error: %v", err)
		}
	}

	// Read from the UDP Socket
	go func() {
		packet := make([]byte, 1500)
		w := contextio.NewWriter(ctx, env.inf)
		for {
			n, addr, err := s.Read(ctx, packet)
			if err != nil {
				log.Fatalf("error while reading from the UDP socket: %v", err)
			}

			// Decode the packet
			hdr, data, err := DecodePacket(packet[:n])
			if err != nil {
				log.Printf("error while reading from client UDP socket: %v", err)
				continue
			}

			if hdr.MessageType == MTypeHeartbeat {
				continue
			}

			// Decrypt the packet
			if hdr.MessageType == MTypeData {
				dec, err := Decrypt(env.key, data)
				if err != nil {
					log.Printf("error while decrypting packet: %v", err)
					continue
				}

				// Parse the IP packet and find out source
				hdr, err := ipv4.ParseHeader(dec)
				if err != nil {
					log.Printf("got an invalid packet which is not an ip packet.Error: %v", err)
					continue
				}

				// Save the client's UDP address so that we can reply back to the client
				s.SetClientAddr(hdr.Src.String(), addr)

				// Write to interface
				_, err = w.Write(dec)
				if err != nil {
					log.Fatalf("error while writing to the interface: %v", err)
				}
			}

		}
	}()

	// Read from interface and write back to socket
	go func() {
		packet := make([]byte, 1500)
		r := contextio.NewReader(ctx, env.inf)
		for {
			n, err := r.Read(packet)
			if err != nil {
				log.Fatalf("error while reading from the interface: %v", err)
			}

			// Parse the packet and find out the destination
			hdr, err := ipv4.ParseHeader(packet[:n])
			if err != nil {
				log.Printf("error while parsing ip packet from interface: %v", err)
				continue
			}

			// Write to socket
			encrypted, err := Encrypt(env.key, packet[:n])
			if err != nil {
				log.Printf("error while encrypting packets")
				continue
			}
			data := MakePacket(MTypeData, encrypted)

			// Get the UDP address for this client
			addr, ok := s.GetClientAddr(hdr.Dst.String())
			if !ok {
				// If remote address exists, send a packet there.
				if rAddr != nil {
					_, err = s.Write(ctx, rAddr, data)
					if err != nil {
						log.Printf("error sending data to remote address: %v", err)
					}
				}
				continue
			}

			n, err = s.Write(ctx, addr, data)
			if err != nil {
				log.Printf("error sending data to client: %v", err)
				continue
			}
		}
	}()
}

func (env *Env) startClient(ctx context.Context) {
	c, err := NewClient(cfg.Peer.RemoteAddress)
	if err != nil {
		log.Fatalf("error starting the client: %v", err)
	}

	// Read from the UDP socket
	go func() {
		packet := make([]byte, 1500)
		w := contextio.NewWriter(ctx, env.inf)
		for {
			n, err := c.Read(ctx, packet)
			if err != nil {
				log.Printf("error while reading from client UDP socket: %v", err)
				continue
			}

			// Decode the packet
			hdr, data, err := DecodePacket(packet[:n])
			if err != nil {
				log.Printf("error while reading from client UDP socket: %v", err)
				continue
			}

			if hdr.MessageType == MTypeHeartbeat {
				continue
			}

			// Decrypt the packet
			if hdr.MessageType == MTypeData {
				dec, err := Decrypt(env.key, data)
				if err != nil {
					log.Printf("error while decrypting packet: %v", err)
					continue
				}

				// Write to interface
				_, err = w.Write(dec)
				if err != nil {
					log.Fatalf("error while writing to the interface: %v", err)
				}
			}
		}
	}()

	go func() {
		packet := make([]byte, 1500)
		r := contextio.NewReader(ctx, env.inf)
		for {
			n, err := r.Read(packet)
			if err != nil {
				log.Fatalf("error while reading from interface: %v", err)
			}

			// Write to socket
			encrypted, err := Encrypt(env.key, packet[:n])
			if err != nil {
				log.Printf("error while encrypting packets")
				continue
			}

			data := MakePacket(MTypeData, encrypted)
			_, err = c.Write(ctx, data)
			if err != nil {
				log.Printf("error while writing to client socket: %v", err)
				continue
			}
		}
	}()
}

func shouldStartServer() bool {
	if cfg.Tunnel.LocalAddress != "" {
		return true
	}

	return false
}
