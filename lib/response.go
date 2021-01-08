package lib

import "time"

type RawResponse struct {
	ID uint64
	Resp []byte
	Err error
	Elapse time.Duration
}