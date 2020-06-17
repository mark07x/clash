package http

import (
	"bufio"
	"encoding/base64"
	"github.com/gofrs/uuid"
	"github.com/mark07x/clash/bridge"
	"net"
	"net/http"
	"strings"
	"time"

	adapters "github.com/mark07x/clash/adapters/inbound"
	"github.com/mark07x/clash/common/cache"
	"github.com/mark07x/clash/component/auth"
	"github.com/mark07x/clash/log"
	authStore "github.com/mark07x/clash/proxy/auth"
	"github.com/mark07x/clash/tunnel"
)

type HttpListener struct {
	net.Listener
	address string
	closed  bool
	cache   *cache.Cache
}

func NewHttpProxy(addr string) (*HttpListener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	hl := &HttpListener{l, addr, false, cache.New(30 * time.Second)}

	go func() {
		log.Infoln("HTTP proxy listening at: %s", addr)
		bridge.Func.On("HTTP START")

		for {
			c, err := hl.Accept()
			if err != nil {
				if hl.closed {
					break
				}
				continue
			}
			id := tunnel.SharedToken.MakeToken()
			go handleConn(c, hl.cache, id)
		}
	}()

	return hl, nil
}

func (l *HttpListener) Close() {
	l.closed = true
	l.Listener.Close()
}

func (l *HttpListener) Address() string {
	return l.address
}

func canActivate(loginStr string, authenticator auth.Authenticator, cache *cache.Cache) (ret bool) {
	if result := cache.Get(loginStr); result != nil {
		ret = result.(bool)
	}
	loginData, err := base64.StdEncoding.DecodeString(loginStr)
	login := strings.Split(string(loginData), ":")
	ret = err == nil && len(login) == 2 && authenticator.Verify(login[0], login[1])

	cache.Put(loginStr, ret, time.Minute)
	return
}

func handleConn(conn net.Conn, cache *cache.Cache, id uuid.UUID) {
	br := bufio.NewReader(conn)
	request, err := http.ReadRequest(br)
	if err != nil || request.URL.Host == "" {
		conn.Close()
		tunnel.SharedToken.ReleaseToken(id)
		return
	}

	authenticator := authStore.Authenticator()
	if authenticator != nil {
		if authStrings := strings.Split(request.Header.Get("Proxy-Authorization"), " "); len(authStrings) != 2 {
			_, err = conn.Write([]byte("HTTP/1.1 407 Proxy Authentication Required\r\nProxy-Authenticate: Basic\r\n\r\n"))
			conn.Close()
			tunnel.SharedToken.ReleaseToken(id)
			return
		} else if !canActivate(authStrings[1], authenticator, cache) {
			conn.Write([]byte("HTTP/1.1 403 Forbidden\r\n\r\n"))
			log.Infoln("Auth failed from %s", conn.RemoteAddr().String())
			conn.Close()
			tunnel.SharedToken.ReleaseToken(id)
			return
		}
	}

	if request.Method == http.MethodConnect {
		_, err := conn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
		if err != nil {
			tunnel.SharedToken.ReleaseToken(id)
			return
		}
		tunnel.Add(adapters.NewHTTPS(request, conn, id))
		return
	}

	tunnel.Add(adapters.NewHTTP(request, conn, id))
}
