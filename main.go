package main

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"load-generator/helpers"
	"load-generator/lib"
	"math"
	"sync/atomic"
	"time"
)

const (
	STATUS_INIT     uint32 = 0
	STATUS_STARTING uint32 = 1
	STATUS_STARTED  uint32 = 2
	STATUS_STOPPING uint32 = 3
	STATUS_STOPPED  uint32 = 4

	CALL_STATUS_INIT    = 0
	CALL_STATUS_DONE    = 1
	CALL_STATUS_TIMEOUT = 2
)

type loadGenerator struct {
	status uint32

	ctx           context.Context
	ctxCancelFunc context.CancelFunc

	pps                  uint64 // payloads per second
	processingDurationNS time.Duration
	timeoutDurationNS    time.Duration

	concurrency uint64
	callCount   uint64

	resultChan chan *lib.CallResult

	ticketsImpl lib.GoroutinePoolTickets

	callerImpl lib.Caller
}

func (receiver *loadGenerator) init() error {
	helpers.Logger.Info("Initializing loadGenerator...")
	interval := 1e9 / receiver.pps
	if interval == 0 {
		helpers.Logger.Info("Set interval to default value 10")
		interval = 10
	}

	total := uint64(int64(receiver.timeoutDurationNS)/int64(interval) + 1)
	if total > math.MaxUint64 {
		helpers.Logger.Info("Set concurrency to MaxUint64")
		total = math.MaxUint64
	}

	receiver.concurrency = total
	tickets, err := lib.NewGoroutinePoolTickets(receiver.concurrency)
	if err != nil {
		helpers.Logger.Error("Create NewGoroutinePoolTickets", zap.String("err", err.Error()))
		return err
	}
	receiver.ticketsImpl = tickets
	helpers.Logger.Info("loadGenerator inited")
	return nil
}

func (receiver *loadGenerator) callOne(rawReq *lib.RawRequest) *lib.RawResponse {
	atomic.AddUint64(&receiver.callCount, 1)
	if rawReq == nil {
		helpers.Logger.Warn("rawReq is nil")
		return &lib.RawResponse{ID: -1, Err: errors.New("Invalid raw request")}
	}
	var rawResp *lib.RawResponse

	startTime := time.Now().UnixNano()
	resp, err := receiver.callerImpl.Call(rawReq.Req, receiver.timeoutDurationNS)
	endTime := time.Now().UnixNano()
	duration := time.Duration(startTime - endTime)
	if err != nil {
		errMsg := fmt.Sprintf("Sync CallOne Error: %s.", err)
		rawResp = &lib.RawResponse{
			ID:     rawReq.ID,
			Err:    errors.New(errMsg),
			Elapse: duration,
		}
	} else {
		rawResp = &lib.RawResponse{
			ID:     rawReq.ID,
			Resp:   resp,
			Elapse: duration,
		}
	}

	return rawResp
}

func (receiver *loadGenerator) asyncCall() {
	receiver.ticketsImpl.Take()
	go func() {
		defer func() {
			receiver.ticketsImpl.PutBack()
		}()
		rawReq := receiver.callerImpl.BuildReq()
		var callStatus uint32
		timer := time.AfterFunc(receiver.timeoutDurationNS, func() {
			if !atomic.CompareAndSwapUint32(&callStatus, CALL_STATUS_INIT, CALL_STATUS_TIMEOUT) {
				return
			}
			result := &lib.CallResult{
				ID:     rawReq.ID,
				Req:    rawReq,
				Code:   lib.RET_CODE_WARNING_TIMEOUT,
				Msg:    fmt.Sprintf("Timeout! Expected < %v", receiver.timeoutDurationNS),
				Elapse: receiver.timeoutDurationNS,
			}
			receiver.sendResult(result)
		})
		resp := receiver.callOne(&rawReq)
		if !atomic.CompareAndSwapUint32(&callStatus, CALL_STATUS_INIT, CALL_STATUS_DONE) {
			return
		}
		timer.Stop()
		var result *lib.CallResult
		if resp.Err != nil {
			result = &lib.CallResult{
				ID:     resp.ID,
				Req:    rawReq,
				Code:   lib.RET_CODE_ERR_CALL,
				Msg:    resp.Err.Error(),
				Elapse: resp.Elapse,
			}
		} else {
			result = receiver.callerImpl.CheckResp(rawReq, *resp)
			result.Elapse = resp.Elapse
		}
		receiver.sendResult(result)
	}()
}

func (receiver *loadGenerator) sendResult(result *lib.CallResult) bool {
	if receiver.Status() != STATUS_STARTED {
		receiver.printIgnoredResult(result, "load generator stopped")
		return false
	}

	select {
	case receiver.resultChan <- result:
		return true
	default:
		receiver.printIgnoredResult(result, "result channel is full")
		return false
	}
	return true
}

func (receiver *loadGenerator) printIgnoredResult(result *lib.CallResult, cause string) {
	helpers.Logger.Info("Ignored result", zap.Int64("ID", result.ID), zap.Int("Code", int(result.Code)), zap.String("Msg", result.Msg), zap.Duration("Elapse", result.Elapse), zap.String("cause", cause))
}

func (receiver *loadGenerator) prepareToStop(err error) {
	helpers.Logger.Info("loadGenerator prepareToStop")
	atomic.CompareAndSwapUint32(&receiver.status, STATUS_STARTED, STATUS_STOPPING)
	close(receiver.resultChan)
	atomic.StoreUint32(&receiver.status, STATUS_STOPPED)
}

func (receiver *loadGenerator) genLoad(throttle <-chan time.Time) {
	helpers.Logger.Info("loadGenerator generating payloads...")
	for {
		select {
		case <-receiver.ctx.Done():
			receiver.prepareToStop(receiver.ctx.Err())
			return
		default:
		}
		receiver.asyncCall()
		if receiver.pps > 0 {
			select {
			case <-throttle:
			case <-receiver.ctx.Done():
				receiver.prepareToStop(receiver.ctx.Err())
				return
			}
		}
	}
}

func (receiver *loadGenerator) Start() bool {
	helpers.Logger.Info("loadGenerator Starting...")
	if !atomic.CompareAndSwapUint32(&receiver.status, STATUS_INIT, STATUS_STARTING) {
		if !atomic.CompareAndSwapUint32(&receiver.status, STATUS_STOPPED, STATUS_STARTING) {
			return false
		}
	}

	var throttle <-chan time.Time
	if receiver.pps > 0 {
		interval := time.Duration(1e9 / receiver.pps)
		throttle = time.Tick(interval)
	}

	receiver.ctx, receiver.ctxCancelFunc = context.WithTimeout(context.Background(), receiver.processingDurationNS)
	receiver.callCount = 0

	atomic.StoreUint32(&receiver.status, STATUS_STARTED)

	go func() {
		receiver.genLoad(throttle)
	}()

	helpers.Logger.Info("loadGenerator Started")
	return true
}

func (receiver *loadGenerator) Stop() bool {
	helpers.Logger.Info("loadGenerator Stopping...")
	if !atomic.CompareAndSwapUint32(&receiver.status, STATUS_STARTED, STATUS_STOPPING) {
		return false
	}
	receiver.ctxCancelFunc()
	for {
		if atomic.LoadUint32(&receiver.status) == STATUS_STOPPED {
			break
		}
		time.Sleep(time.Microsecond)
	}
	helpers.Logger.Info("loadGenerator Stopped")
	return true
}

func (receiver *loadGenerator) Status() uint32 {
	return atomic.LoadUint32(&receiver.status)
}

func (receiver *loadGenerator) CallCount() uint64 {
	return atomic.LoadUint64(&receiver.callCount)
}

type Generator interface {
	Start() bool
	Stop() bool
	Status() uint32
	CallCount() uint64
}

// NewLoadGenerator ...
func NewLoadGenerator(
	params NewLoadGeneratorParams,
) (Generator, error) {
	helpers.Logger.Info("Constructing NewLoadGenerator...")
	if err := params.Check(); err != nil {
		return nil, err
	}

	gen := &loadGenerator{
		callerImpl:           params.Caller,
		pps:                  params.PPS,
		processingDurationNS: params.ProcessingDurationNS,
		timeoutDurationNS:    params.TimeoutNS,
		resultChan:           params.ResultChan,
		status:               STATUS_INIT,
	}

	if err := gen.init(); err != nil {
		return nil, err
	}
	helpers.Logger.Info("NewLoadGenerator constructed")
	return gen, nil
}

func main() {
	fmt.Println("Hello")
}
