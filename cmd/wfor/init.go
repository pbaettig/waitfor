package main

import (
	"flag"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

func init() {
	flag.Var(&httpConditions, "http", "Wait for `url` to respond with 2xx")
	flag.Var(&pathConditions, "path", "Wait for `path` to exist")
	flag.Var(&tcpConditions, "tcp", "Wait for `host:port` to accept connection")
	flag.Var(&udpConditions, "udp", "Wait for `host:port` to respond with at least 1 byte")
	flag.DurationVar(&timeoutFlag, "timeout", 5*time.Minute, "max `duration` to wait for")
	flag.DurationVar(&intervalFlag, "interval", 10*time.Second, "`duration` between checks")
	flag.BoolVar(&andFlag, "and", false, "AND all conditions (default)")
	flag.BoolVar(&orFlag, "or", false, "OR all conditions")
	flag.BoolVar(&debugFlag, "debug", false, "enable verbose logging")
	flag.BoolVar(&versionFlag, "version", false, "print version and exit")
	flag.Parse()

	if versionFlag {
		fmt.Printf("wfor %s (%s) built on %s\n", Tag, Commit, BuildDate)
	}

	if andFlag && orFlag {
		log.Fatal("Cannot specify both -and / -or")
	}
	if !andFlag && !orFlag {
		andFlag = true
	}
	if intervalFlag >= timeoutFlag {
		log.Fatal("-timeout has to be bigger than -interval")
	}
	if int(timeoutFlag)%int(intervalFlag) != 0 {
		log.Fatal("-interval has to divide -timeout evenly")
	}
	if intervalFlag < time.Second {
		log.Fatal("-interval cannot be smaller than 1 second")
	}

	if debugFlag {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

}
