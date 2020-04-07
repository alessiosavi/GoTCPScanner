package TCPScanner

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"syscall"
	"time"
)

type TCPScanner struct {
	Host        string
	PortRange   [][]int
	Concurrency int
	Timeout     time.Duration
}

func (t *TCPScanner) SetHost(host string) {
	t.Host = host
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
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		fmt.Printf("Error Getting Rlimit: %s\n", err)
	}
	fmt.Printf("Current Ulimit: %d\n", rLimit.Cur)
	return rLimit.Cur, rLimit.Max
}

func (t *TCPScanner) Scan() {

	var wg sync.WaitGroup

	ulimitCurr, _ := getUlimitValue()
	if uint64(t.Concurrency) >= ulimitCurr {
		t.Concurrency = int(float64(ulimitCurr) * 0.7)
		fmt.Printf("Provided a thread factor greater than current ulimit size, setting at MAX [%d] requests\n", t.Concurrency)
	}
	semaphore := make(chan struct{}, t.Concurrency)
	for _, ports := range t.PortRange {
		for j := ports[0]; j < ports[1]; j++ {
			wg.Add(1)
			go func(j int) {
				semaphore <- struct{}{}
				if t.IsOpen(j) {
					fmt.Printf("Open %d\n", j)
				}
				func() { <-semaphore }()
				wg.Done()
			}(j)
		}
	}
	wg.Wait()
}

func (t *TCPScanner) IsOpen(port int) bool {
	var tcpAddr *net.TCPAddr
	var err error
	var conn net.Conn

	if tcpAddr, err = net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", t.Host, port)); err != nil {
		fmt.Println(err)
		return false
	}
	if conn, err = net.DialTimeout("tcp", tcpAddr.String(), t.Timeout); err != nil {
		if !strings.Contains(err.Error(), "connect: connection refused") {
			fmt.Println(err)
		}
		return false
	}
	conn.Close()

	return true
}
