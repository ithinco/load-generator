package lib

import "time"

type RawResponse struct {
	ID int64
	Resp []byte
	Err error
	Elapse time.Duration
}