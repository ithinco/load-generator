package helper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"net"
	"strconv"
	"sync/atomic"
)

type tcpServer struct {
	listener net.Listener
	active   uint32
}

// NewTCPServer ...
func NewTCPServer() *tcpServer {
	return &tcpServer{}
}

func (receiver *tcpServer) init(addr string) error {
	if !atomic.CompareAndSwapUint32(&receiver.active, 0, 1) {
		return nil
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		atomic.StoreUint32(&receiver.active, 0)
		return err
	}
	receiver.listener = ln
	return nil
}

func (receiver *tcpServer) Listen(addr string) error {
	err := receiver.init(addr)
	if err != nil {
		return err
	}
	go func() {
		for {
			if atomic.LoadUint32(&receiver.active) != 1 {
				break
			}
			conn, err := receiver.listener.Accept()
			if err != nil {
				if atomic.LoadUint32(&receiver.active) == 1 {
					Logger.Error("Server", zap.String("Request Acception Error", err.Error()))
				} else {
					Logger.Warn("Server: Broken acception because of closed network connection.")
				}
				continue
			}
			go reqHandler(conn)
		}
	}()
	return nil
}

func reqHandler(conn net.Conn) {
	var errMsg string
	var sresp ServerResponse
	req, err := Read(conn, DELIM)
	if err != nil {
		errMsg = fmt.Sprintf("Server: Req Read Error: %s", err.Error())
	} else {
		var sreq ServerRequest
		err = json.Unmarshal(req, &sreq)
		if err != nil {
			errMsg = fmt.Sprintf("Server: Req Unmarshal Error: %s", err)
		} else {
			sresp.ID = sreq.ID
			sresp.Result = Operation(sreq.Operands, sreq.Operator)
			sresp.Formula =
				GenFormula(sreq.Operands, sreq.Operator, sresp.Result, true)
		}
	}
	if errMsg != "" {
		sresp.Err = errors.New(errMsg)
	}
	bytes, err := json.Marshal(sresp)
	if err != nil {
		Logger.Error("Server: Resp Marshal", zap.String("err", err.Error()))
	}
	_, err = Write(conn, bytes, DELIM)
	if err != nil {
		Logger.Error("Server: Resp Write", zap.String("err", err.Error()))
	}
}

func (receiver *tcpServer) Close() bool {
	if !atomic.CompareAndSwapUint32(&receiver.active, 1, 0) {
		return false
	}
	receiver.listener.Close()
	return true
}

// Operation ...
func Operation(operands []int, operator string) int {
	var result int
	switch {
	case operator == "+":
		for _, v := range operands {
			if result == 0 {
				result = v
			} else {
				result += v
			}
		}
	case operator == "-":
		for _, v := range operands {
			if result == 0 {
				result = v
			} else {
				result -= v
			}
		}
	case operator == "*":
		for _, v := range operands {
			if result == 0 {
				result = v
			} else {
				result *= v
			}
		}
	case operator == "/":
		for _, v := range operands {
			if result == 0 {
				result = v
			} else {
				result /= v
			}
		}
	}
	return result
}

// GenFormula ...
func GenFormula(operands []int, operator string, result int, equal bool) string {
	var buff bytes.Buffer
	n := len(operands)
	for i := 0; i < n; i++ {
		if i > 0 {
			buff.WriteString(" ")
			buff.WriteString(operator)
			buff.WriteString(" ")
		}

		buff.WriteString(strconv.Itoa(operands[i]))
	}
	if equal {
		buff.WriteString(" = ")
	} else {
		buff.WriteString(" != ")
	}
	buff.WriteString(strconv.Itoa(result))
	return buff.String()
}
