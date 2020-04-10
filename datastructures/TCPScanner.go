package TCPScanner

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/cheggaaa/pb/v3"
	"go.uber.org/zap"
)

type TCPScanner struct {
	Host         string
	PortRange    [][]int
	Headers      map[int][]string
	Concurrency  int
	Timeout      time.Duration
	ShowProgress bool
}

func (t *TCPScanner) SetHost(host string) {
	t.Host = host
	t.Headers = make(map[int][]string)
}

func (t *TCPScanner) SetTimeout(millisecond int) {
	t.Timeout = time.Millisecond * time.Duration(millisecond)
}

func (t *TCPScanner) AddPortRange(startPort, stopPort int) {
	t.PortRange = append(t.PortRange, []int{startPort, stopPort})
}

// GetUlimitValue return the current and max value for ulimit
func getUlimitValue() (uint64, uint64) {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	return rLimit.Cur, rLimit.Max
}

func (t *TCPScanner) Scan() {

	var wg sync.WaitGroup

	ulimitCurr, _ := getUlimitValue()
	if uint64(t.Concurrency) >= ulimitCurr {
		t.Concurrency = int(float64(ulimitCurr) * 0.7)
		zap.S().Warnf("Provided a thread factor greater than current ulimit size, setting at MAX [%d] requests\n", t.Concurrency)
	}
	semaphore := make(chan struct{}, t.Concurrency)

	var bar *pb.ProgressBar
	// create and start new bar
	if t.ShowProgress {
		bar = pb.Full.Start(getTotalPortCount(t.PortRange))
	}

	for _, ports := range t.PortRange {
		for j := ports[0]; j < ports[1]; j++ {
			wg.Add(1)
			go func(j int) {
				semaphore <- struct{}{}
				if t.IsOpen(j) {
					zap.S().Debugf("Open %d\n", j)
				}
				func() { <-semaphore }()
				wg.Done()
				if t.ShowProgress {
					bar.Increment()
				}
			}(j)
		}
	}
	wg.Wait()
	if t.ShowProgress {
		bar.Finish()
	}

	zap.S().Infof("Open port: %+v\n", t.Headers)
}

func (t *TCPScanner) IsOpen(port int) bool {
	var tcpAddr *net.TCPAddr
	var err error
	var conn net.Conn

	if tcpAddr, err = net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", t.Host, port)); err != nil {
		zap.S().Panic(err)
	}
	if conn, err = net.DialTimeout("tcp", tcpAddr.String(), t.Timeout); err != nil {
		if !strings.Contains(err.Error(), "connect: connection refused") && !strings.Contains(err.Error(), "i/o timeout") {
			zap.S().Panic(err)
		}
		return false
	}
	conn.Close()
	t.Headers[port] = t.getHeaders(port)
	return true
}

func (t *TCPScanner) getHeaders(port int) []string {
	var resp *http.Response
	var err error
	var headers []string
	if resp, err = http.Get(fmt.Sprintf("http://%s:%d", t.Host, port)); err != nil {
		if strings.Contains(err.Error(), "http: server gave HTTP response to HTTPS client") {
			if resp, err = http.Get(fmt.Sprintf("https://%s:%d", t.Host, port)); err != nil {
				zap.S().Info(err)
				return nil
			}
		} else {
			zap.S().Info(err)
			return nil
		}
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		if k == "Server" {
			for i := range v {
				headers = append(headers, v[i])
			}
		}
	}
	return headers
}

func getTotalPortCount(ports [][]int) int {
	var count int
	for _, ports := range ports {
		count += (ports[1] - ports[0])
	}
	return count
}
