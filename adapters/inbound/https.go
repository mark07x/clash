package inbound

import (
	"github.com/gofrs/uuid"
	"net"
	"net/http"

	C "github.com/mark07x/clash/constant"
)

// NewHTTPS is HTTPAdapter generator
func NewHTTPS(request *http.Request, conn net.Conn, id uuid.UUID) *SocketAdapter {
	metadata := parseHTTPAddr(request)
	metadata.Type = C.HTTPCONNECT
	if ip, port, err := parseAddr(conn.RemoteAddr().String()); err == nil {
		metadata.SrcIP = ip
		metadata.SrcPort = port
	}
	return &SocketAdapter{
		metadata: metadata,
		Conn:     conn,
		ID:       id,
	}
}
