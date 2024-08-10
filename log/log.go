package log

import (
	"fmt"
	"os"

	"github.com/lijinglin3/clash/common/observable"

	"github.com/sirupsen/logrus"
)

var (
	logCh  = make(chan any)
	source = observable.NewObservable(logCh)
	level  = INFO
)

func init() {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
}

type Event struct {
	LogLevel Level
	Payload  string
}

func (e *Event) Type() string {
	return e.LogLevel.String()
}

func Infoln(format string, v ...any) {
	event := newLog(INFO, format, v...)
	logCh <- event
	log(event)
}

func Warnln(format string, v ...any) {
	event := newLog(WARNING, format, v...)
	logCh <- event
	log(event)
}

func Errorln(format string, v ...any) {
	event := newLog(ERROR, format, v...)
	logCh <- event
	log(event)
}

func Debugln(format string, v ...any) {
	event := newLog(DEBUG, format, v...)
	logCh <- event
	log(event)
}

func Fatalln(format string, v ...any) {
	logrus.Fatalf(format, v...)
}

func Subscribe() observable.Subscription {
	sub, _ := source.Subscribe()
	return sub
}

func UnSubscribe(sub observable.Subscription) {
	source.UnSubscribe(sub)
}

func GetLevel() Level {
	return level
}

func SetLevel(newLevel Level) {
	level = newLevel
}

func log(data Event) {
	if data.LogLevel < level {
		return
	}

	switch data.LogLevel {
	case INFO:
		logrus.Infoln(data.Payload)
	case WARNING:
		logrus.Warnln(data.Payload)
	case ERROR:
		logrus.Errorln(data.Payload)
	case DEBUG:
		logrus.Debugln(data.Payload)
	case SILENT:
		return
	}
}

func newLog(logLevel Level, format string, v ...any) Event {
	return Event{
		LogLevel: logLevel,
		Payload:  fmt.Sprintf(format, v...),
	}
}
