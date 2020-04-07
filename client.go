package main

import (
	"context"
	"net"
)

// Client defines a remote connection
type Client struct {
	Addr string
	conn *net.UDPConn
}

// NewClient returns a new client connected to remote address
func NewClient(addr string) (*Client, error) {
	rAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, rAddr)
	if err != nil {
		return nil, err
	}

	return &Client{
		Addr: addr,
		conn: conn,
	}, nil
}

// Write writes data to remote
func (p *Client) Write(ctx context.Context, data []byte) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	n, err := p.conn.Write(data)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// Read receives data from remote and writes to data
func (p *Client) Read(ctx context.Context, data []byte) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	n, _, err := p.conn.ReadFromUDP(data)
	if err != nil {
		return 0, err
	}
	return n, nil
}
