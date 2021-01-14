package main

import (
	"errors"
	"go.uber.org/zap"
	"load-generator/helper"
	"load-generator/lib"
	"strings"
	"time"
)

type NewLoadGeneratorParams struct {
	Caller               lib.Caller
	PPS                  uint64
	ProcessingDurationNS time.Duration
	TimeoutNS            time.Duration
	ResultChan           chan *lib.CallResult
}

func (receiver *NewLoadGeneratorParams) Check() error {
	helper.Logger.Info("Start Checking NewLoadGeneratorParams...")
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
		errMsg := strings.Join(errMsgs, " ")
		helper.Logger.Info("Didn't Pass Params Check", zap.String("err", errMsg))
		return errors.New(errMsg)
	}
	helper.Logger.Info("Passed Params Check", zap.Uint64("pps", receiver.PPS), zap.Duration("timeoutNS", receiver.TimeoutNS), zap.Duration("processingDurationNS", receiver.ProcessingDurationNS))
	return nil
}
