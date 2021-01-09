package main

import (
	"context"
	"fmt"
	"load-generator/lib"
	"math"
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

func (receiver *loadGenerator) Start() bool  {
	return false
}

func (receiver *loadGenerator) Stop() bool  {
	return false
}

func (receiver *loadGenerator) Status() uint8 {
	return 0
}

func (receiver *loadGenerator) CallCount() uint64  {
	return 0
}

func (receiver *loadGenerator) init() error {
	total := uint64(int64(receiver.timeoutDurationNS)/int64(1e9/receiver.processingDurationNS)+1)
	if total > math.MaxUint64 {
		total = math.MaxUint64
	}

	receiver.concurrency = total
	tickets, err := lib.NewGoroutinePoolTickets(receiver.concurrency)
	if err != nil {
		return err
	}
	receiver.ticketsImpl = tickets
	return nil
}

type Generator interface {
	Start() bool
	Stop() bool
	Status() uint8
	CallCount() uint64
}

// NewLoadGenerator ...
func NewLoadGenerator(
	params NewLoadGeneratorParams,
	) (Generator, error) {
	if err := params.Check(); err != nil {
		return nil, err
	}
	
	gen := &loadGenerator{
		callerImpl: params.Caller,
		pps: params.PPS,
		processingDurationNS: params.ProcessingDurationNS,
		timeoutDurationNS: params.TimeoutNS,
		resultChan: params.ResultChan,
		status: STATUS_INIT,
	}

	if err := gen.init(); err != nil {
		return nil, err
	}

	return gen, nil
}

func main()  {
	fmt.Println("Hello")
}