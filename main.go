package clash

import (
	"github.com/mark07x/clash/bridge"
	"github.com/mark07x/clash/config"
	"github.com/mark07x/clash/constant"
	C "github.com/mark07x/clash/constant"
	"github.com/mark07x/clash/hub"
	"github.com/mark07x/clash/hub/executor"
	"github.com/mark07x/clash/log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
)

type BridgeFunctions interface {
	Print(str string)
	Fatal(str string)
	Log(str string, level string)
	On(name string)
}
func InitBridge(fun BridgeFunctions) {
	bridge.Func = fun
}

func Main(homeDir string, configFile string, externalUI string, externalController string, secret string, version bool, testConfig bool) {
	bridge.Func.Print("iClash core is started")
	if version {
		bridge.Printf("Clash %s %s %s %s\n", C.Version, runtime.GOOS, runtime.GOARCH, C.BuildTime)
		return
	}

	if homeDir != "" {
		if !filepath.IsAbs(homeDir) {
			currentDir, _ := os.Getwd()
			homeDir = filepath.Join(currentDir, homeDir)
		}
		C.SetHomeDir(homeDir)
	}

	if configFile != "" {
		if !filepath.IsAbs(configFile) {
			currentDir, _ := os.Getwd()
			configFile = filepath.Join(currentDir, configFile)
		}
		C.SetConfig(configFile)
	} else {
		configFile := filepath.Join(C.Path.HomeDir(), C.Path.Config())
		C.SetConfig(configFile)
	}

	if err := config.Init(C.Path.HomeDir()); err != nil {
		log.Fatalln("Initial configuration directory error: %s", err.Error())
	}

	if testConfig {
		if _, err := executor.Parse(); err != nil {
			log.Errorln(err.Error())
			bridge.Printf("configuration file %s test failed\n", constant.Path.Config())
			os.Exit(1)
		}
		bridge.Printf("configuration file %s test is successful\n", constant.Path.Config())
		return
	}

	var options []hub.Option
	if externalUI != "" {
		options = append(options, hub.WithExternalUI(externalUI))
	}
	if externalController != "" {
		options = append(options, hub.WithExternalController(externalController))
	}
	if secret != "" {
		options = append(options, hub.WithSecret(secret))
	}

	if err := hub.Parse(options...); err != nil {
		log.Fatalln("Parse config error: %s", err.Error())
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
