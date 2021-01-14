package main

import (
	"go.uber.org/zap"
	"load-generator/helper"
	"load-generator/lib"
	"testing"
	"time"
)

// printDetail 代表是否打印详细结果。
var printDetail = false

func TestStart(t *testing.T) {

	// 初始化服务器。
	server := helper.NewTCPServer()
	defer server.Close()
	serverAddr := "127.0.0.1:8080"
	t.Logf("Startup TCP server(%s)...\n", serverAddr)
	err := server.Listen(serverAddr)
	if err != nil {
		t.Fatalf("TCP Server startup failing! (addr=%s)!\n", serverAddr)
		t.FailNow()
	}

	// 初始化载荷发生器。
	pset := NewLoadGeneratorParams{
		Caller:               helper.NewTCPCallerClient(serverAddr),
		TimeoutNS:            50 * time.Millisecond,
		PPS:                  uint64(1000),
		ProcessingDurationNS: 10 * time.Second,
		ResultChan:           make(chan *lib.CallResult, 50),
	}
	t.Logf("Initialize load generator (timeoutNS=%v, pps=%d, durationNS=%v)...",
		pset.TimeoutNS, pset.PPS, pset.ProcessingDurationNS)
	gen, err := NewLoadGenerator(pset)
	if err != nil {
		t.Fatalf("Load generator initialization failing: %s\n",
			err)
		t.FailNow()
	}

	// 开始！
	helper.Logger.Info("Start load generator...")
	gen.Start()

	// 显示结果。
	countMap := make(map[lib.RetCode]int)
	for r := range pset.ResultChan {
		countMap[r.Code] = countMap[r.Code] + 1
		if printDetail {
			t.Logf("Result: ID=%d, Code=%d, Msg=%s, Elapse=%v.\n",
				r.ID, r.Code, r.Msg, r.Elapse)
		}
	}

	var total int
	helper.Logger.Info("RetCode Count:")
	for k, v := range countMap {
		codePlain := lib.GetRetCodePlain(k)
		t.Logf("  Code plain: %s (%d), Count: %d.\n",
			codePlain, k, v)
		total += v
	}

	successCount := countMap[lib.RET_CODE_SUCCESS]
	pps := float64(successCount) / float64(pset.ProcessingDurationNS/1e9)
	helper.Logger.Info("Result", zap.Int("tasks", total), zap.Uint64("Loads per second", pset.PPS), zap.Float64("Treatments per second", pps))
}

func TestStop(t *testing.T) {

	// 初始化服务器。
	server := helper.NewTCPServer()
	defer server.Close()
	serverAddr := "127.0.0.1:8081"
	t.Logf("Startup TCP server(%s)...\n", serverAddr)
	err := server.Listen(serverAddr)
	if err != nil {
		t.Fatalf("TCP Server startup failing! (addr=%s)!\n", serverAddr)
		t.FailNow()
	}

	// 初始化载荷发生器。
	pset := NewLoadGeneratorParams{
		Caller:               helper.NewTCPCallerClient(serverAddr),
		TimeoutNS:            50 * time.Millisecond,
		PPS:                  uint64(1000),
		ProcessingDurationNS: 10 * time.Second,
		ResultChan:           make(chan *lib.CallResult, 50),
	}
	t.Logf("Initialize load generator (timeoutNS=%v, lps=%d, durationNS=%v)...",
		pset.TimeoutNS, pset.PPS, pset.ProcessingDurationNS)
	gen, err := NewLoadGenerator(pset)
	if err != nil {
		t.Fatalf("Load generator initialization failing: %s.\n",
			err)
		t.FailNow()
	}

	// 开始！
	helper.Logger.Info("Start load generator...")
	gen.Start()
	timeoutNS := 2 * time.Second
	time.AfterFunc(timeoutNS, func() {
		gen.Stop()
	})

	// 显示调用结果。
	countMap := make(map[lib.RetCode]int)
	count := 0
	for r := range pset.ResultChan {
		countMap[r.Code] = countMap[r.Code] + 1
		if printDetail {
			t.Logf("Result: ID=%d, Code=%d, Msg=%s, Elapse=%v.\n",
				r.ID, r.Code, r.Msg, r.Elapse)
		}
		count++
	}

	var total int
	helper.Logger.Info("RetCode Count:")
	for k, v := range countMap {
		codePlain := lib.GetRetCodePlain(k)
		t.Logf("  Code plain: %s (%d), Count: %d.\n",
			codePlain, k, v)
		total += v
	}

	successCount := countMap[lib.RET_CODE_SUCCESS]
	pps := float64(successCount) / float64(timeoutNS/1e9)
	helper.Logger.Info("Result", zap.Int("tasks", total), zap.Uint64("Loads per second", pset.PPS), zap.Float64("Treatments per second", pps))
}
