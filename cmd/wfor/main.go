package main

import (
	"flag"

	"github.com/pbaettig/waitfor/internal/app"

	log "github.com/sirupsen/logrus"
)

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
