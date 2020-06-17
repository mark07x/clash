package redir

import (
	"github.com/gofrs/uuid"
	"github.com/mark07x/clash/bridge"
	"net"

	"github.com/mark07x/clash/adapters/inbound"
	C "github.com/mark07x/clash/constant"
	"github.com/mark07x/clash/log"
	"github.com/mark07x/clash/tunnel"
)

type RedirListener struct {
	net.Listener
	address string
	closed  bool
}

func NewRedirProxy(addr string) (*RedirListener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	rl := &RedirListener{l, addr, false}

	go func() {
		log.Infoln("Redir proxy listening at: %s", addr)
		bridge.Func.On("REDIR START")
		for {
			c, err := l.Accept()
			if err != nil {
				if rl.closed {
					break
				}
				continue
			}
			id := tunnel.SharedToken.MakeToken()
			go handleRedir(c, id)
		}
	}()

	return rl, nil
}

func (l *RedirListener) Close() {
	l.closed = true
	l.Listener.Close()
}

func (l *RedirListener) Address() string {
	return l.address
}

func handleRedir(conn net.Conn, id uuid.UUID) {
	target, err := parserPacket(conn)
	if err != nil {
		conn.Close()
		tunnel.SharedToken.ReleaseToken(id)
		return
	}
	conn.(*net.TCPConn).SetKeepAlive(true)
	tunnel.Add(inbound.NewSocket(target, conn, C.REDIR, id))
}
