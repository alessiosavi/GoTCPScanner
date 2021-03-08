package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	stringutils "github.com/alessiosavi/GoGPUtils/string"
	datastructures "github.com/alessiosavi/GoTCPScanner/datastructures"
	"github.com/aws/aws-lambda-go/lambda"
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

type InputRequest struct {
	Ports       [][]int `json:"ports"`
	Host        string  `json:"host"`
	Port        int     `json:"port"`
	Concurrency int     `json:"concurrency"`
	Timeout     int     `json:"timeout"`
}

func HandleRequest(ctx context.Context, in InputRequest) (string, error) {
	var tcpScanner datastructures.TCPScanner

	if in.Port == 0 && len(in.Ports) == 0 {
		return "", errors.New("port(s) not set")
	}

	if in.Port != 0 && len(in.Ports) != 0 {
		return "", errors.New("set only port or ports")
	}
	if in.Port != 0 {
		tcpScanner.PortRange = append(tcpScanner.PortRange, []int{0, in.Port})
	} else {
		tcpScanner.PortRange = in.Ports
	}
	if in.Concurrency < 1 {
		in.Concurrency = 1
	}

	tcpScanner.Concurrency = in.Concurrency
	tcpScanner.ShowProgress = false
	tcpScanner.SetTimeout(in.Timeout)
	tcpScanner.SetHost(in.Host)
	log := initZapLog()
	defer log.Sync()
	tcpScanner.Scan()
	result, _ := json.Marshal(tcpScanner)
	return fmt.Sprintf("%s\n", string(result)), nil
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

func consoleInput() datastructures.TCPScanner {
	var myFlags portRangeFlag

	host := flag.String("host", "localhost", "Set the ip/hostname of the target")
	port := flag.Int("port", -1, "Single port to scan")
	concurrency := flag.Int("thread", 8, "Number of concurrent thread for scan the target")
	timeout := flag.Int("timeout", 2000, "Number to millisecond to wait before raise a timeout excpetion")

	flag.Var(&myFlags, "ports", "start and stop port separated by -")
	flag.Parse()

	if stringutils.IsBlank(*host) {
		panic("Empty hostname string")
	}

	if *port == -1 && len(myFlags) == 0 {
		panic("You need to specify the start and stop port (--ports) or a single port (--port)")
	}

	if *port != -1 && len(myFlags) != 0 {
		panic("You need to specify only one parameter related to the port to scan")
	}

	var ports [][]int
	if *port != -1 {
		ports = append(ports, []int{-1, *port})
	} else {
		for i := range myFlags {
			ports = append(ports, []int{myFlags[i][0], myFlags[i][1]})
		}
	}

	tcpScanner := datastructures.TCPScanner{PortRange: ports, Concurrency: *concurrency, ShowProgress: true}
	tcpScanner.SetHost(*host)
	tcpScanner.SetTimeout(*timeout)

	return tcpScanner

}

func main() {

	log.SetFlags(log.Ldate | log.Lshortfile | log.LstdFlags | log.Lmicroseconds)

	// Run the lambda
	if stringutils.IsBlank(os.Getenv("console")) {
		lambda.Start(HandleRequest)
	} else { // Run in console
		var tcpScanner datastructures.TCPScanner = consoleInput()
		log := initZapLog()
		defer log.Sync()

		log.Infof("Starting scan target [%s] in port rage {%v}\n", tcpScanner.Host, tcpScanner.PortRange)
		tcpScanner.Scan()
		result, _ := json.Marshal(tcpScanner)

		fmt.Printf("%s\n", string(result))
	}
}

func initZapLog() *zap.SugaredLogger {
	config := zap.NewDevelopmentConfig()
	// config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	logger, _ := config.Build()

	zap.ReplaceGlobals(logger)
	return logger.Sugar()
}
