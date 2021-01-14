package lib

import "time"

type RetCode int

const (
	RET_CODE_SUCCESS         RetCode = 0
	RET_CODE_WARNING_TIMEOUT RetCode = 1001
	RET_CODE_ERR_CALL        RetCode = 2001
	RET_CODE_ERR_RESPONSE    RetCode = 2002
	RET_CODE_ERR_CALLEE      RetCode = 2003
	RET_CODE_FATAL_CALL      RetCode = 3001
)

type CallResult struct {
	ID     int64
	Req    RawRequest
	Resp   RawResponse
	Code   RetCode
	Msg    string
	Elapse time.Duration
}

// GetRetCodePlain ...
func GetRetCodePlain(code RetCode) string {
	var codePlain string
	switch code {
	case RET_CODE_SUCCESS:
		codePlain = "Success"
	case RET_CODE_WARNING_TIMEOUT:
		codePlain = "Call Timeout Warning"
	case RET_CODE_ERR_CALL:
		codePlain = "Call Error"
	case RET_CODE_ERR_RESPONSE:
		codePlain = "Response Error"
	case RET_CODE_ERR_CALLEE:
		codePlain = "Callee Error"
	case RET_CODE_FATAL_CALL:
		codePlain = "Call Fatal Error"
	default:
		codePlain = "Unknown result code"
	}
	return codePlain
}
