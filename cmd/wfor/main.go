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

func main() {
	log.Debugf("-http: %s\n", httpConditions)
	log.Debugf("-path: %s\n", pathConditions)
	log.Debugf("-tcp: %s\n", tcpConditions)
	log.Debugf("-udp: %s\n", udpConditions)
	log.Debugf("-timeout: %s\n", timeoutFlag)
	log.Debugf("-interval: %s\n", intervalFlag)

	waitConditions := make([]app.WaitCondition, 0)
	for _, path := range pathConditions {
		waitConditions = append(waitConditions, app.PathWait{Path: path})
	}
	for _, url := range httpConditions {
		waitConditions = append(waitConditions, app.HTTPWait{URL: url, AcceptableStatusCodes: []int{200, 202, 203}})
	}

	result := make(chan bool)
	for _, wc := range waitConditions {
		go app.Wait(wc, intervalFlag, timeoutFlag, result)
	}

	if orFlag {
		// Wait for any WaitCondition to complete successfully
		log.Debug("Waiting for one condition to succeed.")
		for i := 0; i < len(waitConditions); i++ {
			c := <-result
			if c {
				goto exec
			}
		}
		log.Fatal("All WaitConditions have timed out!")
	} else if andFlag {
		// Wait for all WaitConditions to complete successfully
		log.Debug("Waiting for all conditions to succeed.")
		for i := 0; i < len(waitConditions); i++ {
			c := <-result
			if !c {
				log.Fatal("A WaitCondition has timed out.")
			}
		}
		goto exec
	}

exec:
	if len(flag.Args()) > 0 {
		app.ExecWithEnv(flag.Args())
	}

}
