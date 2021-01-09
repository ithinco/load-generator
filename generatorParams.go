package main

import (
	"errors"
	"load-generator/lib"
	"strings"
	"time"
)

type NewLoadGeneratorParams struct {
	Caller lib.Caller
	PPS uint64
	ProcessingDurationNS time.Duration
	TimeoutNS time.Duration
	ResultChan chan *lib.CallResult
}

func (receiver *NewLoadGeneratorParams) Check() error  {
	var errMsgs []string
	if receiver.Caller == nil {
		errMsgs = append(errMsgs, "Invalid caller!")
	}

	if receiver.ResultChan == nil {
		errMsgs = append(errMsgs, "Invalid resultChan!")
	}

	if receiver.PPS == 0 {
		errMsgs = append(errMsgs, "Invalid pps!")
	}

	if receiver.ProcessingDurationNS == 0 {
		errMsgs = append(errMsgs, "Invalid processingDurationNS!")
	}

	if receiver.TimeoutNS == 0 {
		errMsgs = append(errMsgs, "Invalid timeoutNS!")
	}

	if len(errMsgs) > 0 {
		return errors.New(strings.Join(errMsgs, " "))
	}

	return nil
}
