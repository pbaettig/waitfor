package app

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

// WaitCondition represents any condition that can be waited for
type WaitCondition interface {
	String() string
	Fulfilled() bool
}

// PathWait is used for waiting on a path to exist
type PathWait struct {
	Path string
}

func (pwc PathWait) String() string {
	return fmt.Sprintf("PathWait:%s", pwc.Path)
}

// Fulfilled returns true if the PathWaitCondition.Path exists, false otherwise
func (pwc PathWait) Fulfilled() bool {
	if _, err := os.Stat(pwc.Path); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		// Schrodinger: file may or may not exist. See err for details.
		return false
	}
}

type HTTPWait struct {
	URL                   string
	AcceptableStatusCodes []int
}

func (hwc HTTPWait) String() string {
	return fmt.Sprintf("HTTPWait:%s", hwc.URL)
}

func (hwc HTTPWait) Fulfilled() bool {
	resp, err := http.Get(hwc.URL)
	if err != nil {
		return false
	}
	resp.Body.Close()
	for _, s := range hwc.AcceptableStatusCodes {
		if resp.StatusCode == s {
			return true
		}
	}
	return false
}

// Wait waits until the passed WaitCondition is fulfilled.
// It returns an error if it timed out
func Wait(wc WaitCondition, interval, timeout time.Duration, result chan<- bool) {
	maxIterations := int(timeout / interval)
	i := 0

	for {

		loopStart := time.Now()
		if wc.Fulfilled() {
			log.WithField("condition", wc.String()).Debugf("Check succeeded")
			result <- true
			return
		}
		log.WithField("condition", wc.String()).Debugf("Check failed")

		if i >= maxIterations {
			log.WithField("condition", wc.String()).Debugf("Timed out")
			result <- false
			return
		}

		loopDuration := time.Now().Sub(loopStart)
		if loopDuration < interval {
			sleepFor := interval - loopDuration
			//log.Debugf("Sleeping for %s", sleepFor)
			time.Sleep(sleepFor)
		}
		//log.Debug("Waiting")

		i++

	}
}

// ExecWithEnv runs args with the same environment as the parent process
func ExecWithEnv(args []string) {
	cmd, _ := exec.LookPath(args[0])

	argv := make([]string, 0)
	if len(args) > 1 {
		argv = flag.Args()[1:]
	}
	log.Debugf("Starting %s with args %q", cmd, argv)

	err := syscall.Exec(cmd, args, os.Environ())
	if err != nil {
		log.Fatalf("Unable to start %s: %s", cmd, err)
	}
}
