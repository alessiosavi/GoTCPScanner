package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	stringutils "github.com/alessiosavi/GoGPUtils/string"
	datastructures "github.com/alessiosavi/GoTCPScanner/datastructures"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Created so that multiple inputs can be accecpted
type portRangeFlag [][]int

func (i *portRangeFlag) String() string {
	var sb strings.Builder
	for _, c := range *i {
		sb.WriteString(fmt.Sprintf("%v", c))
	}
	return sb.String()
}

func (i *portRangeFlag) Set(value string) error {
	var startPort, stopPort int
	var err error
	value = strings.TrimSpace(value)
	a := strings.Split(value, "-")
	if len(a) != 2 {
		panic("Error during insert of the input: " + value)
	}
	if startPort, err = strconv.Atoi(a[0]); err != nil {
		panic(err)
	}
	if stopPort, err = strconv.Atoi(a[1]); err != nil {
		panic(err)
	}
	if stopPort < startPort {
		panic("Stop port is greater than start port")
	}

	*i = append(*i, []int{startPort, stopPort})
	return nil
}

func main() {

	var tcpScanner datastructures.TCPScanner
	var myFlags portRangeFlag

	loggerMgr := initZapLog()
	zap.ReplaceGlobals(loggerMgr)
	defer loggerMgr.Sync() // flushes buffer, if any
	log := loggerMgr.Sugar()

	host := flag.String("host", "localhost", "Set the ip/hostname of the target")
	port := flag.Int("port", -1, "Single port to scan")
	flag.Var(&myFlags, "ports", "start and stop port separated by -")
	flag.Parse()

	if stringutils.IsBlank(*host) {
		panic("Empty hostname string")
	}
	tcpScanner.SetHost(*host)

	if *port == -1 && len(myFlags) == 0 {
		panic("You need to specify the start and stop port (--ports) or a single port (--port)")
	}

	if *port != -1 && len(myFlags) != 0 {
		panic("You need to specify only one parameter related to port")
	}

	for i := range myFlags {
		tcpScanner.AddPortRange(myFlags[i][0], myFlags[i][1])
	}

	log.Infof("Starting scan target [%s] in port rage {%v}\n", tcpScanner.Host, tcpScanner.PortRange)
	tcpScanner.SetTimeout(3000)
	tcpScanner.Concurrency = 300

	tcpScanner.Scan()
}

func initZapLog() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	logger, _ := config.Build()
	return logger
}
