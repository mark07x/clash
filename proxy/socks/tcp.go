package socks

import (
	"github.com/mark07x/clash/bridge"
	"io"
	"io/ioutil"
	"net"

	adapters "github.com/mark07x/clash/adapters/inbound"
	"github.com/mark07x/clash/component/socks5"
	C "github.com/mark07x/clash/constant"
	"github.com/mark07x/clash/log"
	authStore "github.com/mark07x/clash/proxy/auth"
	"github.com/mark07x/clash/tunnel"
)

type SockListener struct {
	net.Listener
	address string
	closed  bool
}

func NewSocksProxy(addr string) (*SockListener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	sl := &SockListener{l, addr, false}
	go func() {
		log.Infoln("SOCKS proxy listening at: %s", addr)
		bridge.Func.On("SOCKS START")
		for {
			c, err := l.Accept()
			if err != nil {
				if sl.closed {
					break
				}
				continue
			}
			go handleSocks(c)
		}
	}()

	return sl, nil
}

func (l *SockListener) Close() {
	l.closed = true
	l.Listener.Close()
}

func (l *SockListener) Address() string {
	return l.address
}

func handleSocks(conn net.Conn) {
	target, command, err := socks5.ServerHandshake(conn, authStore.Authenticator())
	if err != nil {
		conn.Close()
		return
	}
	conn.(*net.TCPConn).SetKeepAlive(true)
	if command == socks5.CmdUDPAssociate {
		defer conn.Close()
		io.Copy(ioutil.Discard, conn)
		return
	}
	tunnel.Add(adapters.NewSocket(target, conn, C.SOCKS))
}
