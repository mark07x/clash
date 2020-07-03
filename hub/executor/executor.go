package executor

import (
	"fmt"
	"github.com/mark07x/clash/bridge"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mark07x/clash/adapters/provider"
	"github.com/mark07x/clash/component/auth"
	"github.com/mark07x/clash/component/dialer"
	trie "github.com/mark07x/clash/component/domain-trie"
	"github.com/mark07x/clash/component/resolver"
	"github.com/mark07x/clash/config"
	C "github.com/mark07x/clash/constant"
	"github.com/mark07x/clash/dns"
	"github.com/mark07x/clash/log"
	P "github.com/mark07x/clash/proxy"
	authStore "github.com/mark07x/clash/proxy/auth"
	"github.com/mark07x/clash/tunnel"
)

// forward compatibility before 1.0
func readRawConfig(path string) ([]byte, error) {
	data, err := ioutil.ReadFile(path)
	if err == nil && len(data) != 0 {
		return data, nil
	}

	if filepath.Ext(path) != ".yaml" {
		return nil, err
	}

	path = path[:len(path)-5] + ".yml"
	if _, fallbackErr := os.Stat(path); fallbackErr == nil {
		return ioutil.ReadFile(path)
	}

	return data, err
}

func readConfig(path string) ([]byte, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}
	data, err := readRawConfig(path)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("Configuration file %s is empty", path)
	}

	return data, err
}

// Parse config with default config path
func Parse() (*config.Config, error) {
	return ParseWithPath(C.Path.Config())
}

// ParseWithPath parse config with custom config path
func ParseWithPath(path string) (*config.Config, error) {
	buf, err := readConfig(path)
	if err != nil {
		return nil, err
	}

	return ParseWithBytes(buf)
}

// ParseWithBytes config with buffer
func ParseWithBytes(buf []byte) (*config.Config, error) {
	return config.Parse(buf)
}

// ApplyConfig dispatch configure to all parts
func ApplyConfig(cfg *config.Config, force bool) {
	updateUsers(cfg.Users)
	if force {
		updateGeneral(cfg.General)
	}
	updateProxies(cfg.Proxies, cfg.Providers)
	updateRules(cfg.Rules)
	updateDNS(cfg.DNS)
	updateHosts(cfg.Hosts)
	updateExperimental(cfg)
}

func GetGeneral() *config.General {
	ports := P.GetPorts()
	authenticator := []string{}
	if auth := authStore.Authenticator(); auth != nil {
		authenticator = auth.Users()
	}

	general := &config.General{
		Port:           ports.Port,
		SocksPort:      ports.SocksPort,
		RedirPort:      ports.RedirPort,
		Authentication: authenticator,
		AllowLan:       P.AllowLan(),
		BindAddress:    P.BindAddress(),
		Mode:           tunnel.Mode(),
		LogLevel:       log.Level(),
	}

	return general
}

func updateExperimental(c *config.Config) {
	cfg := c.Experimental

	tunnel.UpdateExperimental(cfg.IgnoreResolveFail)
	if cfg.Interface != "" && c.DNS.Enable {
		dialer.DialHook = dialer.DialerWithInterface(cfg.Interface)
		dialer.ListenPacketHook = dialer.ListenPacketWithInterface(cfg.Interface)
	} else {
		dialer.DialHook = nil
		dialer.ListenPacketHook = nil
	}
}

func updateDNS(c *config.DNS) {
	if c.Enable == false {
		resolver.DefaultResolver = nil
		tunnel.SetResolver(nil)
		dns.ReCreateServer("", nil)
		return
	}
	r := dns.New(dns.Config{
		Main:         c.NameServer,
		Fallback:     c.Fallback,
		IPv6:         c.IPv6,
		EnhancedMode: c.EnhancedMode,
		Pool:         c.FakeIPRange,
		FallbackFilter: dns.FallbackFilter{
			GeoIP:  c.FallbackFilter.GeoIP,
			IPCIDR: c.FallbackFilter.IPCIDR,
		},
		Default: c.DefaultNameserver,
	})
	resolver.DefaultResolver = r
	tunnel.SetResolver(r)
	if err := dns.ReCreateServer(c.Listen, r); err != nil {
		log.Errorln("Start DNS server error: %s", err.Error())
		return
	}

	if c.Listen != "" {
		log.Infoln("DNS server listening at: %s", c.Listen)
		bridge.Func.On("DNS START")
	}
}

func updateHosts(tree *trie.Trie) {
	resolver.DefaultHosts = tree
}

func updateProxies(proxies map[string]C.Proxy, providers map[string]provider.ProxyProvider) {
	tunnel.UpdateProxies(proxies, providers)
}

func updateRules(rules []C.Rule) {
	tunnel.UpdateRules(rules)
}

func updateGeneral(general *config.General) {
	log.SetLevel(general.LogLevel)
	tunnel.SetMode(general.Mode)

	allowLan := general.AllowLan
	P.SetAllowLan(allowLan)

	bindAddress := general.BindAddress
	P.SetBindAddress(bindAddress)

	if err := P.ReCreateHTTP(general.Port); err != nil {
		log.Errorln("Start HTTP server error: %s", err.Error())
	}

	if err := P.ReCreateSocks(general.SocksPort); err != nil {
		log.Errorln("Start SOCKS5 server error: %s", err.Error())
	}

	if err := P.ReCreateRedir(general.RedirPort); err != nil {
		log.Errorln("Start Redir server error: %s", err.Error())
	}
}

func updateUsers(users []auth.AuthUser) {
	authenticator := auth.NewAuthenticator(users)
	authStore.SetAuthenticator(authenticator)
	if authenticator != nil {
		log.Infoln("Authentication of local server updated")
	}
}
