package tunnel

import (
	"github.com/gofrs/uuid"
	"sync"
	"time"
)

var DefaultManager *Manager
var lock sync.Mutex
var SharedToken = Token{
	token: make(chan struct{}, 100),
}
const ConnectionNumber = 9
const Difference = 1
type Token struct {
	token chan struct{}
	use sync.Map
}
func (t *Token) MakeToken() uuid.UUID {
	<- t.token
	id, _ := uuid.NewV4()
	t.use.Store(id, struct{}{})
	return id
}
func (t *Token) PushToken() {
	t.token <- struct{}{}
}
func (t *Token) ReleaseToken(id uuid.UUID) {
	lock.Lock()
	if _, ok := t.use.Load(id); ok {
		t.use.Delete(id)
		lock.Unlock()
		t.token <- struct{}{}
	} else {
		lock.Unlock()
	}
}

func init() {
	DefaultManager = &Manager{
		upload:   make(chan int64),
		download: make(chan int64),
	}
	DefaultManager.handle()
	for i := 0; i < ConnectionNumber; i++ {
		SharedToken.PushToken()
	}
}

type Manager struct {
	connections   sync.Map
	upload        chan int64
	download      chan int64
	uploadTemp    int64
	downloadTemp  int64
	uploadBlip    int64
	downloadBlip  int64
	uploadTotal   int64
	downloadTotal int64
}

func (m *Manager) Join(c tracker) {
	m.connections.Store(c.GetTokenID(), c)
	go DefaultManager.Trim()
}

func (m *Manager) Leave(c tracker) {
	m.connections.Delete(c.GetTokenID())
}

func (m *Manager) Upload() chan<- int64 {
	return m.upload
}

func (m *Manager) Download() chan<- int64 {
	return m.download
}

func (m *Manager) Now() (up int64, down int64) {
	return m.uploadBlip, m.downloadBlip
}

func (m *Manager) Trim() {
	var size = 0
	var dieTime = time.Now().Add(time.Hour)
	var dieTracker tracker = nil
	m.connections.Range(func(key, value interface{}) bool {
		if tcpTracker, ok := value.(*tcpTracker); ok {
			size++
			time := tcpTracker.GetKeepAlive()
			if time.Before(dieTime) {
				dieTracker = tcpTracker
				dieTime = time
			}
		}
		return true
	})
	if size > ConnectionNumber - Difference {
		dieTracker.Close()
	}
}

func (m *Manager) Snapshot() *Snapshot {
	connections := []tracker{}
	m.connections.Range(func(key, value interface{}) bool {
		connections = append(connections, value.(tracker))
		return true
	})

	return &Snapshot{
		UploadTotal:   m.uploadTotal,
		DownloadTotal: m.downloadTotal,
		Connections:   connections,
	}
}

func (m *Manager) ResetStatistic() {
	m.uploadTemp = 0
	m.uploadBlip = 0
	m.uploadTotal = 0
	m.downloadTemp = 0
	m.downloadBlip = 0
	m.downloadTotal = 0
}

func (m *Manager) handle() {
	go m.handleCh(m.upload, &m.uploadTemp, &m.uploadBlip, &m.uploadTotal)
	go m.handleCh(m.download, &m.downloadTemp, &m.downloadBlip, &m.downloadTotal)
}

func (m *Manager) handleCh(ch <-chan int64, temp *int64, blip *int64, total *int64) {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case n := <-ch:
			*temp += n
			*total += n
		case <-ticker.C:
			*blip = *temp
			*temp = 0
		}
	}
}

type Snapshot struct {
	DownloadTotal int64     `json:"downloadTotal"`
	UploadTotal   int64     `json:"uploadTotal"`
	Connections   []tracker `json:"connections"`
}
