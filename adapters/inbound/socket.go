package inbound

import (
	"github.com/gofrs/uuid"
	"net"

	"github.com/mark07x/clash/component/socks5"
	C "github.com/mark07x/clash/constant"
)

// SocketAdapter is a adapter for socks and redir connection
type SocketAdapter struct {
	net.Conn
	metadata *C.Metadata
	ID uuid.UUID
}

// Metadata return destination metadata
func (s *SocketAdapter) Metadata() *C.Metadata {
	return s.metadata
}

func (s *SocketAdapter) GetTokenID() uuid.UUID {
	return s.ID
}

// NewSocket is SocketAdapter generator
func NewSocket(target socks5.Addr, conn net.Conn, source C.Type, id uuid.UUID) *SocketAdapter {
	metadata := parseSocksAddr(target)
	metadata.NetWork = C.TCP
	metadata.Type = source
	if ip, port, err := parseAddr(conn.RemoteAddr().String()); err == nil {
		metadata.SrcIP = ip
		metadata.SrcPort = port
	}

	return &SocketAdapter{
		Conn:     conn,
		metadata: metadata,
		ID: id,
	}
}
