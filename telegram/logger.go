package telegram

import (
	"fmt"
	"log"
)

// Logger defines an interface for logs
type Logger interface {
	// logging facilities
	Debug(...interface{})
	Debugf(string, ...interface{})
	Debugw(string, ...interface{})
	Error(...interface{})
	Errorf(string, ...interface{})
	Errorw(string, ...interface{})
	Info(...interface{})
	Infof(string, ...interface{})
	Infow(string, ...interface{})
}

type defaultLogger struct {
	*log.Logger
}

func (d *defaultLogger) Debug(v ...interface{}) {
	d.Logger.Print(v...)
}

func (d *defaultLogger) Debugf(format string, v ...interface{}) {
	d.Logger.Printf(format, v...)
}

func (d *defaultLogger) Debugw(base string, keysAndValues ...interface{}) {
	msg := base
	if msg == "" && len(keysAndValues) > 0 {
		msg = fmt.Sprint(keysAndValues...)
	} else if msg != "" && len(keysAndValues) > 0 {
		msg = fmt.Sprintf(base, keysAndValues...)
	}
	fmt.Println(msg)
}

func (d *defaultLogger) Error(v ...interface{}) {
	d.Logger.Print(v...)
}

func (d *defaultLogger) Errorf(format string, v ...interface{}) {
	d.Logger.Printf(format, v...)
}

func (d *defaultLogger) Errorw(base string, keysAndValues ...interface{}) {
	d.Debugw(base, keysAndValues...)
}

func (d *defaultLogger) Info(v ...interface{}) {
	d.Logger.Print(v...)
}

func (d *defaultLogger) Infof(format string, v ...interface{}) {
	d.Logger.Printf(format, v...)
}

func (d *defaultLogger) Infow(base string, keysAndValues ...interface{}) {
	d.Debugw(base, keysAndValues...)
}
