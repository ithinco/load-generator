package main

import (
	"context"
	"fmt"
	"load-generator/lib"
	"time"
)

const (
	STATUS_INIT uint8 = 0
	STATUS_STARTING uint8 = 1
	STATUS_STARTED uint8 = 2
	STATUS_STOPPING uint8 = 3
	STATUS_STOPPED uint8 = 4
)

type loadGenerator struct {
	status uint8

	ctx context.Context
	ctxCancelFunc context.CancelFunc

	pps uint64 // payloads per second
	processingDurationNS time.Duration
	timeoutDurationNS time.Duration

	concurrency uint64
	callCount uint64

	resultChan chan *lib.CallResult

	ticketsImpl lib.GoroutinePoolTickets

	callerImpl lib.Caller
}

func main()  {
	fmt.Println("Hello")
}