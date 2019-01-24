package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pbaettig/waitfor/internal/app"

	log "github.com/sirupsen/logrus"
)

const (
	Version = "0.2"
)

var (
	httpConditions            httpWaits
	httpCodes                 httpCodeList
	pathConditions            pathWaits
	tcpConditions             tcpWaits
	udpConditions             udpWaits
	timeoutFlag, intervalFlag time.Duration
	andFlag                   bool
	orFlag                    bool
	debugFlag                 bool
	versionFlag               bool
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
	*i = append(*i,
		app.HTTPWait{
			URL:                   url,
			AcceptableStatusCodes: []int{200, 201, 202, 203, 204, 205, 206, 207, 208, 226}, // Default value, can be overridden by -httpcodes
		})
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

type httpCodeList string

func (s *httpCodeList) String() string {
	return string(*s)
}

func (s *httpCodeList) Set(value string) error {
	*s = httpCodeList(value)
	return nil
}

func (hcl httpCodeList) Parse() ([]int, error) {
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

func init() {
	flag.Var(&httpConditions, "http", "Wait for `url` to respond with 2xx")
	flag.Var(&httpCodes, "httpcodes", "Comma-separated list of accepted HTTP status codes when using a -http check")
	flag.Var(&pathConditions, "path", "Wait for `path` to exist")
	flag.Var(&tcpConditions, "tcp", "Wait for `host:port` to accept connection")
	flag.Var(&udpConditions, "udp", "Wait for `host:port` to respond with at least 1 byte")
	flag.DurationVar(&timeoutFlag, "timeout", 5*time.Minute, "max `duration` to wait for")
	flag.DurationVar(&intervalFlag, "interval", 10*time.Second, "`duration` between checks")
	flag.BoolVar(&andFlag, "and", false, "AND all conditions (default)")
	flag.BoolVar(&orFlag, "or", false, "OR all conditions")
	flag.BoolVar(&debugFlag, "debug", false, "enable verbose logging")
	flag.BoolVar(&versionFlag, "version", false, "print version and exit")
	flag.Parse()

	// Check if -httpcodes was added to the cmdline
	if httpCodes != "" {
		if len(httpConditions) == 0 {
			log.Fatalln("Cannot specify -httpcodes when no -http checks are defined")
		}
		acs, err := httpCodes.Parse()
		if err != nil {
			log.Fatalln(err)
		}

		// Update all -http checks with the new
		for i := range httpConditions {
			httpConditions[i].AcceptableStatusCodes = acs
		}
	}

	if versionFlag {
		fmt.Printf("wfor v%s\n", Version)
	}

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
	if intervalFlag < time.Second {
		log.Fatal("-interval cannot be smaller than 1 second")
	}

	if debugFlag {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

}
