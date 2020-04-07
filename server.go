package main

import (
	"context"
	"net"
	"sync"
)

// Server is the UDP server
type Server struct {
	sync.Mutex
	Addr    string
	clients map[string]*net.UDPAddr
	conn    *net.UDPConn
}

// NewServer starts a new UDP server and returns it
func NewServer(addr string) (*Server, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	return &Server{
		Addr:    addr,
		clients: make(map[string]*net.UDPAddr),
		conn:    conn,
	}, nil
}

// Read reads packets from the UDP socket and returns the UDP address from where the packet came
func (s *Server) Read(ctx context.Context, data []byte) (int, *net.UDPAddr, error) {
	if err := ctx.Err(); err != nil {
		return 0, nil, err
	}
	return s.conn.ReadFromUDP(data)
}

// Write writes packets to the given UDP address
func (s *Server) Write(ctx context.Context, addr *net.UDPAddr, data []byte) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	return s.conn.WriteTo(data, addr)
}

// SetClientAddr stores client's UDP address against the source
func (s *Server) SetClientAddr(src string, addr *net.UDPAddr) {
	s.Lock()
	defer s.Unlock()
	s.clients[src] = addr
}

// GetClientAddr returns the Client's UDP address according to the destination
func (s *Server) GetClientAddr(dst string) (*net.UDPAddr, bool) {
	s.Lock()
	defer s.Unlock()
	addr, ok := s.clients[dst]
	return addr, ok
}
