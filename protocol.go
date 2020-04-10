package main

import (
	"encoding/binary"
	"errors"
)

// Checksum should be present in every message
const Checksum uint32 = 255

// Message Types
const (
	MTypeData uint32 = iota
	MTypeHeartbeat
)

// Header is the packet header attached to the encrypted data
type Header struct {
	Checksum    uint32
	MessageType uint32
	MessageSize uint32
}

// MakePacket adds the header to the data and returns the packet
func MakePacket(mtype uint32, data []byte) []byte {
	packet := make([]byte, 12+len(data))
	binary.BigEndian.PutUint32(packet[0:4], Checksum)
	binary.BigEndian.PutUint32(packet[4:8], mtype)
	binary.BigEndian.PutUint32(packet[8:12], uint32(len(data)))
	copy(packet[12:], data)
	return packet
}

// DecodePacket decodes the packet and returns the header, data
func DecodePacket(packet []byte) (Header, []byte, error) {
	if len(packet) < 12 {
		return Header{}, []byte{}, errors.New("invalid packet")
	}

	checksum := binary.BigEndian.Uint32(packet[0:4])

	if checksum != Checksum {
		return Header{}, []byte{}, errors.New("invalid packet")
	}

	hdr := Header{
		Checksum:    checksum,
		MessageType: binary.BigEndian.Uint32(packet[4:8]),
		MessageSize: binary.BigEndian.Uint32(packet[8:12]),
	}

	if hdr.MessageSize != uint32(len(packet))-12 {
		return Header{}, []byte{}, errors.New("invalid packet")
	}

	if !(hdr.MessageType == MTypeData || hdr.MessageType == MTypeHeartbeat) {
		return Header{}, []byte{}, errors.New("invalid packet")
	}

	return hdr, packet[12:], nil
}
