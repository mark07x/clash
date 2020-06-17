package inbound

import (
	"github.com/gofrs/uuid"
	"github.com/mark07x/clash/component/socks5"
	C "github.com/mark07x/clash/constant"
)

// PacketAdapter is a UDP Packet adapter for socks/redir/tun
type PacketAdapter struct {
	C.UDPPacket
	metadata *C.Metadata
	ID       uuid.UUID
}

// Metadata returns destination metadata
func (s *PacketAdapter) Metadata() *C.Metadata {
	return s.metadata
}

func (s *PacketAdapter) GetTokenID() uuid.UUID {
	return s.ID
}

// NewPacket is PacketAdapter generator
func NewPacket(target socks5.Addr, packet C.UDPPacket, source C.Type, id uuid.UUID) *PacketAdapter {
	metadata := parseSocksAddr(target)
	metadata.NetWork = C.UDP
	metadata.Type = source
	if ip, port, err := parseAddr(packet.LocalAddr().String()); err == nil {
		metadata.SrcIP = ip
		metadata.SrcPort = port
	}

	return &PacketAdapter{
		UDPPacket: packet,
		metadata:  metadata,
		ID:        id,
	}
}
