package log

import (
	"fmt"
	"github.com/mark07x/clash/bridge"
	"os"

	"github.com/mark07x/clash/common/observable"

	log "github.com/sirupsen/logrus"
)

var (
	logCh  = make(chan interface{})
	source = observable.NewObservable(logCh)
	level  = INFO
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

type Event struct {
	LogLevel LogLevel
	Payload  string
}

func (e *Event) Type() string {
	return e.LogLevel.String()
}

func Infoln(format string, v ...interface{}) {
	event := newLog(INFO, format, v...)
	logCh <- event
	print(event)
}

func Warnln(format string, v ...interface{}) {
	event := newLog(WARNING, format, v...)
	logCh <- event
	print(event)
}

func Errorln(format string, v ...interface{}) {
	event := newLog(ERROR, format, v...)
	logCh <- event
	print(event)
}

func Debugln(format string, v ...interface{}) {
	event := newLog(DEBUG, format, v...)
	logCh <- event
	print(event)
}

func Fatalln(format string, v ...interface{}) {
	bridge.Func.Log(fmt.Sprintf(format, v...), "fatal")
}

func Subscribe() observable.Subscription {
	sub, _ := source.Subscribe()
	return sub
}

func UnSubscribe(sub observable.Subscription) {
	source.UnSubscribe(sub)
	return
}

func Level() LogLevel {
	return level
}

func SetLevel(newLevel LogLevel) {
	level = newLevel
}

func print(data *Event) {
	if data.LogLevel < level {
		return
	}

	switch data.LogLevel {
	case INFO:
		bridge.Func.Log(data.Payload, "info")
	case WARNING:
		bridge.Func.Log(data.Payload, "warning")
	case ERROR:
		bridge.Func.Log(data.Payload, "error")
	case DEBUG:
		bridge.Func.Log(data.Payload, "debug")
	}
}

func newLog(logLevel LogLevel, format string, v ...interface{}) *Event {
	return &Event{
		LogLevel: logLevel,
		Payload:  fmt.Sprintf(format, v...),
	}
}
