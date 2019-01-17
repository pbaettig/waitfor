package app

import (
	"errors"
	"flag"
	"os"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

// WaitCondition represents any condition that can be waited for
type WaitCondition interface {
	Fulfilled() bool
}

// PathWaitCondition is used for waiting on a path to exist
type PathWaitCondition struct {
	Path string
}

// Fulfilled returns true if the PathWaitCondition.Path exists, false otherwise
func (pwc PathWaitCondition) Fulfilled() bool {
	if _, err := os.Stat(pwc.Path); err == nil {
		log.Debugf("%s exists.", pwc.Path)
		return true
	} else if os.IsNotExist(err) {
		log.Debugf("%s does not exist.", pwc.Path)
		return false
	} else {
		// Schrodinger: file may or may not exist. See err for details.
		return false
	}
}

// ExecWithEnv runs args with the same environment as the parent process
func ExecWithEnv(args []string) {
	log.Debugf("Starting %s", args[0])
	argv := make([]string, 0)
	if len(args) > 1 {
		log.Debugf("Args: %v", argv)
		argv = flag.Args()[1:]
	}

	err := syscall.Exec(args[0], argv, os.Environ())
	if err != nil {
		log.Fatalf("Unable to start %s: %s", args[0], err)
	}
}

// Wait waits until the passed WaitCondition is fulfilled.
// It returns an error if it timed out
func Wait(wc WaitCondition, interval, timeout time.Duration) error {
	maxIterations := int(timeout / interval)
	i := 0
	for {
		loopStart := time.Now()
		if wc.Fulfilled() {
			break
		}
		if i >= maxIterations {
			goto timedOut
		}

		loopDuration := time.Now().Sub(loopStart)
		if loopDuration < interval {
			sleepFor := interval - loopDuration
			log.Debugf("Sleeping for %s", sleepFor)
			time.Sleep(sleepFor)
		}
		log.Debug("Waiting")

		i++
	}
	return nil

timedOut:
	return errors.New("wait timed out")
}
