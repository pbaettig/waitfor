package main

import (
	"flag"
	"time"

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

var (
	httpConditions            conditionList
	pathConditions            conditionList
	tcpConditions             conditionList
	udpConditions             conditionList
	timeoutFlag, intervalFlag time.Duration
	andFlag                   bool
	orFlag                    bool
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
	flag.Parse()

	if andFlag && orFlag {
		log.Fatal("Cannot specify both -and / -or")
	}
	if !andFlag && !orFlag {
		andFlag = true
	}
	log.SetLevel(log.DebugLevel)
}
