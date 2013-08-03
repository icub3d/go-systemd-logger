// Copyright 2013 Joshua Marsh <joshua@themarshians.com>. All rights
// reserved. Use of this source code is governed by the MIT license
// which can be found in the LICENSE file.

// Package sysdlog provides a simple interface to systemd's logging
// service. It connects via '/dev/log' and can include priority
// messages. It also implements the io.Writer interface so that it can
// be used as the default logger.
package sysdlog

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

// Severity is a standard linux logging severity. They represent that
// script prefixes used by systemd to label a logs severity.
type Severity string

const (
	LOG_EMERG   Severity = "<0>"
	LOG_ALERT   Severity = "<1>"
	LOG_CRIT    Severity = "<2>"
	LOG_ERR     Severity = "<3>"
	LOG_WARNING Severity = "<4>"
	LOG_NOTICE  Severity = "<5>"
	LOG_INFO    Severity = "<6>"
	LOG_DEBUG   Severity = "<7>"
)

// Sysdlog is a connection to the systemd logger.
type Sysdlog struct {
	prefix string

	conn net.Conn
	mu   sync.Mutex
}

// New creates a new Sysdlog. All messages sent to this logger will
// have the given prefix. The systemd logger is not a fan of prefixes
// that have the form "prefix: " or "[prefix]: ". If you are using a
// prefix like that it will likely be stripped off in the
// log. Instead, you can try something like "<prefix> " or "[prefix]
// ".
func New(prefix string) (*Sysdlog, error) {
	sdl := &Sysdlog{
		prefix: prefix,
	}

	if err := sdl.connect(); err != nil {
		return nil, err
	}

	return sdl, nil
}

// NewLogger creates a log.Logger whose output is written to a systemd
// logger with the given flag.
func NewLogger(flags int) (*log.Logger, error) {
	s, err := New("")
	if err != nil {
		return nil, err
	}

	return log.New(s, "", flags), nil
}

// Close closes the open connection to the systemd logger.
func (sdl *Sysdlog) Close() {
	sdl.conn.Close()
}

// Write writes the given bytes to the logger using the severity
// LOG_ERR.
func (sdl *Sysdlog) Write(b []byte) (int, error) {
	return sdl.writeRetry(LOG_ERR, string(b))
}

// Emerg logs a message with severity LOG_EMERG.
func (sdl *Sysdlog) Emerg(m string) error {
	_, err := sdl.writeRetry(LOG_EMERG, m)
	return err
}

// Alert logs a message with severity LOG_ALERT.
func (sdl *Sysdlog) Alert(m string) error {
	_, err := sdl.writeRetry(LOG_ALERT, m)
	return err
}

// Crit logs a message with severity LOG_CRIT.
func (sdl *Sysdlog) Crit(m string) error {
	_, err := sdl.writeRetry(LOG_CRIT, m)
	return err
}

// Err logs a message with severity LOG_ERR.
func (sdl *Sysdlog) Err(m string) error {
	_, err := sdl.writeRetry(LOG_ERR, m)
	return err
}

// Warning logs a message with severity LOG_WARNING.
func (sdl *Sysdlog) Warning(m string) error {
	_, err := sdl.writeRetry(LOG_WARNING, m)
	return err
}

// Notice logs a message with severity LOG_NOTICE.
func (sdl *Sysdlog) Notice(m string) error {
	_, err := sdl.writeRetry(LOG_NOTICE, m)
	return err
}

// Info logs a message with severity LOG_INFO.
func (sdl *Sysdlog) Info(m string) error {
	_, err := sdl.writeRetry(LOG_INFO, m)
	return err
}

// Debug logs a message with severity LOG_DEBUG.
func (sdl *Sysdlog) Debug(m string) error {
	_, err := sdl.writeRetry(LOG_DEBUG, m)
	return err
}

// Emergf logs a message with severity LOG_EMERG.
func (sdl *Sysdlog) Emergf(format string, v ...interface{}) error {
	_, err := sdl.writeRetry(LOG_EMERG, fmt.Sprintf(format, v...))
	return err
}

// Alertf logs a message with severity LOG_ALERT.
func (sdl *Sysdlog) Alertf(format string, v ...interface{}) error {
	_, err := sdl.writeRetry(LOG_ALERT, fmt.Sprintf(format, v...))
	return err
}

// Critf logs a message with severity LOG_CRIT.
func (sdl *Sysdlog) Critf(format string, v ...interface{}) error {
	_, err := sdl.writeRetry(LOG_CRIT, fmt.Sprintf(format, v...))
	return err
}

// Errf logs a message with severity LOG_ERR.
func (sdl *Sysdlog) Errf(format string, v ...interface{}) error {
	_, err := sdl.writeRetry(LOG_ERR, fmt.Sprintf(format, v...))
	return err
}

// Warningf logs a message with severity LOG_WARNING.
func (sdl *Sysdlog) Warningf(format string, v ...interface{}) error {
	_, err := sdl.writeRetry(LOG_WARNING, fmt.Sprintf(format, v...))
	return err
}

// Noticef logs a message with severity LOG_NOTICE.
func (sdl *Sysdlog) Noticef(format string, v ...interface{}) error {
	_, err := sdl.writeRetry(LOG_NOTICE, fmt.Sprintf(format, v...))
	return err
}

// Infof logs a message with severity LOG_INFO.
func (sdl *Sysdlog) Infof(format string, v ...interface{}) error {
	_, err := sdl.writeRetry(LOG_INFO, fmt.Sprintf(format, v...))
	return err
}

// Debugf logs a message with severity LOG_DEBUG.
func (sdl *Sysdlog) Debugf(format string, v ...interface{}) error {
	_, err := sdl.writeRetry(LOG_DEBUG, fmt.Sprintf(format, v...))
	return err
}

// writeRetry attempts to write the given log message and is capable
// of reconnecting to a closed connection.
func (sdl *Sysdlog) writeRetry(s Severity, m string) (int, error) {
	sdl.mu.Lock()
	defer sdl.mu.Unlock()

	// Try a write if we have a connection.
	if sdl.conn != nil {
		if n, err := sdl.write(s, m); err == nil {
			return n, err
		}
	}

	// If we have no connection or the write above failed, try to
	// connect again.
	if err := sdl.connect(); err != nil {
		return 0, err
	}

	// Try the write again after a reconnect.
	return sdl.write(s, m)
}

func (sdl *Sysdlog) write(s Severity, m string) (int, error) {

	nl := ""
	if !strings.HasSuffix(m, "\n") {
		nl = "\n"
	}

	fmt.Println("prefix:", sdl.prefix)
	fmt.Fprintf(sdl.conn, "%s %s%s%s", s, sdl.prefix, m, nl)
	return len(m), nil
}

// connect is a helper function that does the dialing to the logger.
func (sdl *Sysdlog) connect() error {
	conn, err := net.Dial("unixgram", "/dev/log")
	if err != nil {
		return err
	}

	sdl.conn = conn

	return nil
}
