package app

import (
	"errors"
	"flag"
	"fmt"
	"net"
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
	Check() error
}

// PathWait is used for waiting on a path to exist
type PathWait struct {
	Path string
}

func (pwc PathWait) String() string {
	return fmt.Sprintf("PathWait:%s", pwc.Path)
}

// Fulfilled returns true if the PathWaitCondition.Path exists, false otherwise
func (pwc PathWait) Check() error {
	if _, err := os.Stat(pwc.Path); err == nil {
		return nil
	} else {
		return err
	}
}

type HTTPWait struct {
	URL                   string
	AcceptableStatusCodes []int
}

func (hwc HTTPWait) String() string {
	return fmt.Sprintf("HTTPWait:%s", hwc.URL)
}

func (hwc HTTPWait) Check() error {
	resp, err := http.Get(hwc.URL)
	if err != nil {
		return err
	}
	resp.Body.Close()
	for _, s := range hwc.AcceptableStatusCodes {
		if resp.StatusCode == s {
			return nil
		}
	}
	return errors.New(fmt.Sprintf("received status %d is not accepted", resp.StatusCode))
}

type TCPWait struct {
	HostPort       string
	ConnectTimeout time.Duration
}

func (tcw TCPWait) String() string {
	return fmt.Sprintf("TCPWait:%s", tcw.HostPort)
}
func (tcw TCPWait) Check() error {
	conn, err := net.DialTimeout("tcp", tcw.HostPort, tcw.ConnectTimeout)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

type UDPWait struct {
	HostPort    string
	ReadTimeout time.Duration
}

func (udw UDPWait) String() string {
	return fmt.Sprintf("UDPWait:%s", udw.HostPort)
}
func (udw UDPWait) Check() error {
	conn, cerr := net.Dial("udp", udw.HostPort)
	conn.SetDeadline(time.Now().Add(udw.ReadTimeout))
	if cerr != nil {
		return cerr
	}
	fmt.Fprintf(conn, "HELLO\r\n\r\n")
	rbuf := make([]byte, 8)
	_, rerr := conn.Read(rbuf)
	if rerr != nil {
		return rerr
	}
	conn.Close()
	return nil
}

// Wait waits until the passed WaitCondition is fulfilled.
// It returns an error if it timed out
func Wait(wc WaitCondition, interval, timeout time.Duration, result chan<- bool) {
	maxIterations := int(timeout / interval)
	i := 0

	for {

		loopStart := time.Now()
		if err := wc.Check(); err != nil {
			log.WithFields(log.Fields{
				"wait":  wc.String(),
				"error": err.Error(),
			}).Debug("Check failed")

		} else {
			log.WithField("wait", wc.String()).Debugf("Check succeeded")
			result <- true
			return
		}

		if i >= maxIterations {
			log.WithField("wait", wc.String()).Debugf("Timed out")
			result <- false
			return
		}

		loopDuration := time.Now().Sub(loopStart)
		if loopDuration < interval {
			sleepFor := interval - loopDuration
			time.Sleep(sleepFor)
		}

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
