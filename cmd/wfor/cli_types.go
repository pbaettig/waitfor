package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pbaettig/waitfor/internal/app"
)

var (
	httpConditions            httpWaits
	pathConditions            pathWaits
	tcpConditions             tcpWaits
	udpConditions             udpWaits
	timeoutFlag, intervalFlag time.Duration
	andFlag                   bool
	orFlag                    bool
	debugFlag                 bool
	versionFlag               bool
)

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

func parseCodes(hcl string) ([]int, error) {
	codes := make([]int, 0)
	for _, s := range strings.Split(string(hcl), ",") {
		s = strings.TrimSpace(s)
		if len(s) == 3 {
			i, err := strconv.Atoi(s)
			if err != nil {
				return nil, fmt.Errorf("cannot convert %s to int (%s)", s, hcl)
			}
			codes = append(codes, i)
		} else if len(s) >= 7 && s[3] == '-' {
			// Range
			start, err := strconv.Atoi(s[:3])
			if err != nil {
				return nil, fmt.Errorf("cannot convert %s to int (%s)", s[:3], hcl)
			}

			end, err := strconv.Atoi(s[4:])
			if err != nil {
				return nil, fmt.Errorf("cannot convert %s to int (%s)", s[4:], hcl)
			}

			if start > end {
				return nil, fmt.Errorf("start %d is bigger than end %d", start, end)
			}
			for ; start <= end; start++ {
				codes = append(codes, start)
			}
		} else {
			return nil, fmt.Errorf("%s is not a valid HTTP status", s)
		}

	}
	return codes, nil
}

type httpWaits []app.HTTPWait

func (i *httpWaits) String() string {
	return "String"
}
func (i *httpWaits) Set(value string) error {
	split := strings.Split(value, "|")

	if len(split) == 0 {
		return fmt.Errorf("string is empty")
	}

	htw := new(app.HTTPWait)
	if strings.HasPrefix(split[0], "http") {
		htw.URL = split[0]
	} else {
		htw.URL = fmt.Sprintf("http://%s", split[0])
	}

	if len(split) >= 2 {
		cs, err := parseCodes(split[1])

		if err == nil && len(cs) > 0 {
			htw.AcceptableStatusCodes = cs
		}
	}
	if len(split) == 3 {
		if split[2] != "" {
			content, err := regexp.Compile(split[2])
			if err != nil {
				return err
			}
			htw.ContentMatch = content
		} else {
			return fmt.Errorf("content match regexp is emtpy")
		}

	}

	*i = append(*i, *htw)
	return nil
}
