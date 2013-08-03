Introduction
============

Package sysdlog is compatible with the log package but also
provides the ability to send more specialized logs to systemd. While
it is possible to use log/syslog, it tends to log a lot of duplication
information to systemd because systemd alread includes information
like the timestamp, host, executable, and pid. Using this library
eliminates that extra data.

Usage
=====
```go
package main

import (
	"fmt"
	"github.com/icub3d/go-systemd-logger/sysdlog"
	"log"
)

func main() {
	// Create an instance of Sysdlog
	sdl, err := sysdlog.New("[prefix] ")
	if err != nil {
		fmt.Println("opening systemd log:", err)
		return
	}
		
	// Set the default loggers output to Sysdlog and remove any flags
	// so it doesn't include additonal data.
	log.SetOutput(sdl)
	log.SetFlags(0)
	
	// The default logger uses LOG_ERR.
	log.Println("This is a message coming the the default logger.")
	
	// Your application can use the actual Sysdlog instance though
	// to log different levels.
	sdl.Emerg("This is an emergency!")
	sdl.Alert("This is an alert!")
	sdl.Crit("This is critical!")
	sdl.Err("This is an error!")
	sdl.Warning("This is a warning!")
	sdl.Notice("This is a notice!")
	sdl.Info("This is some info!")
	sdl.Debug("This is some debug!")
}
```
