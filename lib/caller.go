package lib

import (
	"time"
)

type Caller interface {
	BuildReq() RawRequest
	Call(req []byte, timeoutNS time.Duration) ([]byte, error)
	CheckResp(req RawRequest, resp RawResponse) *CallResult
}
