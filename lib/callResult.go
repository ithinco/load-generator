package lib

import "time"

type RetCode int

type CallResult struct {
	ID uint64
	Req RawRequest
	Resp RawResponse
	Code RetCode
	Msg string
	Elapse time.Duration
}