package tunnel

import (
	"net"
	"time"

	C "github.com/mark07x/clash/constant"
	"github.com/gofrs/uuid"
)

type tracker interface {
	GetTokenID() uuid.UUID
	Close() error
	GetKeepAlive() time.Time
	GetStartTime() time.Time
}

type trackerInfo struct {
	UUID          uuid.UUID   `json:"id"`
	Metadata      *C.Metadata `json:"metadata"`
	UploadTotal   int64       `json:"upload"`
	DownloadTotal int64       `json:"download"`
	Start         time.Time   `json:"start"`
	KeepAlive	  time.Time
	Chain         C.Chain     `json:"chains"`
	Rule          string      `json:"rule"`
}

type tcpTracker struct {
	C.Conn `json:"-"`
	*trackerInfo
	manager *Manager
}

func (tt *tcpTracker) GetTokenID() uuid.UUID {
	return tt.UUID
}

func (tt *tcpTracker) GetKeepAlive() time.Time {
	return tt.KeepAlive
}

func (tt *tcpTracker) GetStartTime() time.Time {
	return tt.Start
}

func (tt *tcpTracker) Read(b []byte) (int, error) {
	n, err := tt.Conn.Read(b)
	download := int64(n)
	tt.manager.Download() <- download
	tt.DownloadTotal += download
	tt.KeepAlive = time.Now()
	return n, err
}

func (tt *tcpTracker) Write(b []byte) (int, error) {
	n, err := tt.Conn.Write(b)
	upload := int64(n)
	tt.manager.Upload() <- upload
	tt.UploadTotal += upload
	tt.KeepAlive = time.Now()
	return n, err
}

func (tt *tcpTracker) Close() error {
	tt.manager.Leave(tt)
	SharedToken.ReleaseToken(tt.UUID)
	return tt.Conn.Close()
}

func newTCPTracker(conn C.Conn, manager *Manager, metadata *C.Metadata, rule C.Rule, id uuid.UUID) *tcpTracker {
	ruleType := ""
	if rule != nil {
		ruleType = rule.RuleType().String()
	}

	t := &tcpTracker{
		Conn:    conn,
		manager: manager,
		trackerInfo: &trackerInfo{
			Start:    time.Now(),
			KeepAlive:    time.Now(),
			Metadata: metadata,
			Chain:    conn.Chains(),
			Rule:     ruleType,
			UUID:     id,
		},
	}

	manager.Join(t)
	return t
}

type udpTracker struct {
	C.PacketConn `json:"-"`
	*trackerInfo
	manager *Manager
}

func (ut *udpTracker) GetTokenID() uuid.UUID {
	return ut.UUID
}

func (ut *udpTracker) GetKeepAlive() time.Time {
	return ut.KeepAlive
}

func (ut *udpTracker) GetStartTime() time.Time {
	return ut.Start
}

func (ut *udpTracker) ReadFrom(b []byte) (int, net.Addr, error) {
	n, addr, err := ut.PacketConn.ReadFrom(b)
	download := int64(n)
	ut.manager.Download() <- download
	ut.DownloadTotal += download
	ut.KeepAlive = time.Now()
	return n, addr, err
}

func (ut *udpTracker) WriteTo(b []byte, addr net.Addr) (int, error) {
	n, err := ut.PacketConn.WriteTo(b, addr)
	upload := int64(n)
	ut.manager.Upload() <- upload
	ut.UploadTotal += upload
	ut.KeepAlive = time.Now()
	return n, err
}

func (ut *udpTracker) WriteWithMetadata(p []byte, metadata *C.Metadata) (int, error) {
	n, err := ut.PacketConn.WriteWithMetadata(p, metadata)
	upload := int64(n)
	ut.manager.Upload() <- upload
	ut.UploadTotal += upload
	ut.KeepAlive = time.Now()
	return n, err
}

func (ut *udpTracker) Close() error {
	ut.manager.Leave(ut)
	SharedToken.ReleaseToken(ut.UUID)
	return ut.PacketConn.Close()
}

func newUDPTracker(conn C.PacketConn, manager *Manager, metadata *C.Metadata, rule C.Rule) *udpTracker {
	ruleType := ""
	if rule != nil {
		ruleType = rule.RuleType().String()
	}

	ut := &udpTracker{
		PacketConn: conn,
		manager:    manager,
		trackerInfo: &trackerInfo{
			Start:    time.Now(),
			KeepAlive:    time.Now(),
			Metadata: metadata,
			Chain:    conn.Chains(),
			Rule:     ruleType,
		},
	}

	manager.Join(ut)
	return ut
}
