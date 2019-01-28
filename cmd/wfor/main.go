package main

import (
	"flag"
	"os"

	"github.com/pbaettig/waitfor/internal/app"

	log "github.com/sirupsen/logrus"
)

func main() {
	// channel for the waits to put their results into
	result := make(chan bool)

	// concat all different waits to one slice
	var waits []app.WaitCondition
	for _, pc := range pathConditions {
		waits = append(waits, pc)
	}
	for _, hc := range httpConditions {
		waits = append(waits, hc)
	}
	for _, tc := range tcpConditions {
		waits = append(waits, tc)
	}
	for _, uc := range udpConditions {
		waits = append(waits, uc)
	}

	// start all waits
	for _, w := range waits {
		go app.Wait(w, intervalFlag, timeoutFlag, result)
	}

	if orFlag {
		// Wait for any wait conditions to complete successfully
		log.Debug("Waiting for one condition to succeed")
		for i := 0; i < len(waits); i++ {
			c := <-result
			if c {
				goto exec
			}
		}
		log.Fatal("All conditions have timed out")
	} else if andFlag {
		// Wait for all WaitConditions to complete successfully
		log.Debug("Waiting for all conditions to succeed.")
		for i := 0; i < len(waits); i++ {
			c := <-result
			if !c {
				log.Fatal("A condition has timed out")
			}
		}
		goto exec
	}

exec:
	if len(flag.Args()) > 0 {
		app.ExecWithEnv(flag.Args())
	} else {
		os.Exit(0)
	}
}
