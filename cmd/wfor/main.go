package main

import (
	"flag"
	"time"

	"github.com/pbaettig/waitfor/internal/app"

	log "github.com/sirupsen/logrus"
)

type args struct {
	http, path, tcp, udp string
	timeout              time.Duration
}

var (
	httpFlag                  string
	pathFlag                  string
	tcpFlag                   string
	udpFlag                   string
	timeoutFlag, intervalFlag time.Duration
)

func init() {
	flag.StringVar(&httpFlag, "http", "", "usage")
	flag.StringVar(&pathFlag, "path", "", "usage")
	flag.StringVar(&tcpFlag, "tcp", "", "usage")
	flag.StringVar(&udpFlag, "udp", "", "usage")
	flag.DurationVar(&timeoutFlag, "timeout", 5*time.Minute, "usage")
	flag.DurationVar(&intervalFlag, "interval", 10*time.Second, "usage")
	flag.Parse()

	log.SetLevel(log.DebugLevel)
}

func main() {
	log.Debugf("-http: %s\n", httpFlag)
	log.Debugf("-path: %s\n", pathFlag)
	log.Debugf("-tcp: %s\n", tcpFlag)
	log.Debugf("-udp: %s\n", udpFlag)
	log.Debugf("-timeout: %s\n", timeoutFlag)
	log.Debugf("-interval: %s\n", intervalFlag)

	if pathFlag != "" {
		f := app.PathWaitCondition{pathFlag}
		err := app.Wait(f, intervalFlag, timeoutFlag)
		if err != nil {
			log.Fatal(err)
		}
	}

	if len(flag.Args()) > 0 {
		app.ExecWithEnv(flag.Args())
	}
}
