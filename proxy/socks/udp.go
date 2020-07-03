package socks

import (
	"github.com/gofrs/uuid"
	"net"

	adapters "github.com/mark07x/clash/adapters/inbound"
	"github.com/mark07x/clash/common/pool"
	"github.com/mark07x/clash/common/sockopt"
	"github.com/mark07x/clash/component/socks5"
	C "github.com/mark07x/clash/constant"
	"github.com/mark07x/clash/log"
	"github.com/mark07x/clash/tunnel"
)

type SockUDPListener struct {
	net.PacketConn
	address string
	closed  bool
}

func NewSocksUDPProxy(addr string) (*SockUDPListener, error) {
	l, err := net.ListenPacket("udp", addr)
	if err != nil {
		return nil, err
	}

	err = sockopt.UDPReuseaddr(l.(*net.UDPConn))
	if err != nil {
		log.Warnln("Failed to Reuse UDP Address: %s", err)
	}

	sl := &SockUDPListener{l, addr, false}
	go func() {
		for {
			buf := pool.Get(pool.RelayBufferSize)
			n, remoteAddr, err := l.ReadFrom(buf)
			if err != nil {
				pool.Put(buf)
				if sl.closed {
					break
				}
				continue
			}
			//id := tunnel.SharedToken.MakeToken()
			id, _ := uuid.NewV4()
			handleSocksUDP(l, buf[:n], remoteAddr, id)
		}
	}()

	return sl, nil
}

func (l *SockUDPListener) Close() error {
	l.closed = true
	return l.PacketConn.Close()
}

func (l *SockUDPListener) Address() string {
	return l.address
}

func handleSocksUDP(pc net.PacketConn, buf []byte, addr net.Addr, id uuid.UUID) {
	target, payload, err := socks5.DecodeUDPPacket(buf)
	if err != nil {
		// Unresolved UDP packet, return buffer to the pool
		pool.Put(buf)
		//tunnel.SharedToken.ReleaseToken(id)
		return
	}
	packet := &packet{
		pc:      pc,
		rAddr:   addr,
		payload: payload,
		bufRef:  buf,
	}
	tunnel.AddPacket(adapters.NewPacket(target, packet, C.SOCKS, id))
}
