package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/pbaettig/waitfor/internal/app"

	log "github.com/sirupsen/logrus"
)

type conditionList []string

func (i *conditionList) String() string {
	return "String"
}

func (i *conditionList) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type httpWaits []app.HTTPWait

func (i *httpWaits) String() string {
	return "String"
}

func (i *httpWaits) Set(value string) error {
	var url string
	if !strings.HasPrefix(value, "http") {
		url = fmt.Sprintf("http://%s", value)
	} else {
		url = value
	}
	*i = append(*i, app.HTTPWait{URL: url, AcceptableStatusCodes: []int{201, 202, 203, 204}})
	return nil
}

type pathWaits []app.PathWait

func (i *pathWaits) String() string {
	return "String"
}

func (i *pathWaits) Set(value string) error {
	*i = append(*i, app.PathWait{Path: value})
	return nil
}

type tcpWaits []app.TCPWait

func (i *tcpWaits) String() string {
	return "String"
}

func (i *tcpWaits) Set(value string) error {
	*i = append(*i, app.TCPWait{HostPort: value, ConnectTimeout: 500 * time.Millisecond})
	return nil
}

type udpWaits []app.UDPWait

func (i *udpWaits) String() string {
	return "String"
}

func (i *udpWaits) Set(value string) error {
	*i = append(*i, app.UDPWait{HostPort: value, ReadTimeout: 500 * time.Millisecond})
	return nil
}

var (
	httpConditions            httpWaits
	pathConditions            pathWaits
	tcpConditions             tcpWaits
	udpConditions             udpWaits
	timeoutFlag, intervalFlag time.Duration
	andFlag                   bool
	orFlag                    bool
	debugFlag                 bool
)

func init() {
	flag.Var(&httpConditions, "http", "usage")
	flag.Var(&pathConditions, "path", "usage")
	flag.Var(&tcpConditions, "tcp", "usage")
	flag.Var(&udpConditions, "udp", "usage")
	flag.DurationVar(&timeoutFlag, "timeout", 5*time.Minute, "usage")
	flag.DurationVar(&intervalFlag, "interval", 10*time.Second, "usage")
	flag.BoolVar(&andFlag, "and", false, "usage")
	flag.BoolVar(&orFlag, "or", false, "usage")
	flag.BoolVar(&debugFlag, "debug", false, "usage")
	flag.Parse()

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

	if debugFlag {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

}
