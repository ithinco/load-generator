package helper

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"load-generator/lib"
	"math/rand"
	"net"
	"time"
)

const (
	DELIM = '\n' // 分隔符。
)

// ServerRequest ...
type ServerRequest struct {
	ID       int64
	Operands []int
	Operator string
}

// ServerResponse ...
type ServerResponse struct {
	ID      int64
	Formula string
	Result  int
	Err     error
}

var operators = []string{"+", "-", "*", "/"}

type tcpCallerClient struct {
	addr string
}

// NewTCPCallerClient ...
func NewTCPCallerClient(addr string) lib.Caller {
	return &tcpCallerClient{addr: addr}
}

func (receiver *tcpCallerClient) BuildReq() lib.RawRequest {
	id := time.Now().UnixNano()
	req := &ServerRequest{
		ID:       id,
		Operator: operators[rand.Int31n(100)%4],
		Operands: []int{
			int(rand.Int31n(1000) + 1),
			int(rand.Int31n(1000) + 1),
		},
	}

	jsonBytes, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}

	return lib.RawRequest{
		ID:  id,
		Req: jsonBytes,
	}
}

func (receiver *tcpCallerClient) Call(req []byte, timeoutNS time.Duration) ([]byte, error) {
	conn, err := net.DialTimeout("tcp", receiver.addr, timeoutNS)
	if err != nil {
		return nil, err
	}

	_, err = Write(conn, req, DELIM)
	if err != nil {
		return nil, err
	}

	return Read(conn, DELIM)
}

func (receiver *tcpCallerClient) CheckResp(req lib.RawRequest, resp lib.RawResponse) *lib.CallResult {
	var commResult lib.CallResult
	commResult.ID = resp.ID
	commResult.Req = req
	commResult.Resp = resp
	var sreq ServerRequest
	err := json.Unmarshal(req.Req, &sreq)
	if err != nil {
		commResult.Code = lib.RET_CODE_FATAL_CALL
		commResult.Msg =
			fmt.Sprintf("Incorrectly formatted Req: %s!\n", string(req.Req))
		return &commResult
	}
	var sresp ServerResponse
	err = json.Unmarshal(resp.Resp, &sresp)
	if err != nil {
		commResult.Code = lib.RET_CODE_ERR_RESPONSE
		commResult.Msg =
			fmt.Sprintf("Incorrectly formatted Resp: %s!\n", string(resp.Resp))
		return &commResult
	}
	if sresp.ID != sreq.ID {
		commResult.Code = lib.RET_CODE_ERR_RESPONSE
		commResult.Msg =
			fmt.Sprintf("Inconsistent raw id! (%d != %d)\n", req.ID, resp.ID)
		return &commResult
	}
	if sresp.Err != nil {
		commResult.Code = lib.RET_CODE_ERR_CALLEE
		commResult.Msg =
			fmt.Sprintf("Abnormal server: %s!\n", sresp.Err)
		return &commResult
	}
	if sresp.Result != Operation(sreq.Operands, sreq.Operator) {
		commResult.Code = lib.RET_CODE_ERR_RESPONSE
		commResult.Msg =
			fmt.Sprintf(
				"Incorrect result: %s!\n",
				GenFormula(sreq.Operands, sreq.Operator, sresp.Result, false))
		return &commResult
	}
	commResult.Code = lib.RET_CODE_SUCCESS
	commResult.Msg = fmt.Sprintf("Success. (%s)", sresp.Formula)
	return &commResult
}

// Read ...
func Read(conn net.Conn, delim byte) ([]byte, error) {
	readBytes := make([]byte, 1)
	var buffer bytes.Buffer
	for {
		_, err := conn.Read(readBytes)
		if err != nil {
			return nil, err
		}
		readByte := readBytes[0]
		if readByte == delim {
			break
		}
		buffer.WriteByte(readByte)
	}
	return buffer.Bytes(), nil
}

// Write ...
func Write(conn net.Conn, content []byte, delim byte) (int, error) {
	writer := bufio.NewWriter(conn)
	n, err := writer.Write(content)
	if err == nil {
		writer.WriteByte(delim)
	}
	if err == nil {
		err = writer.Flush()
	}
	return n, err
}
