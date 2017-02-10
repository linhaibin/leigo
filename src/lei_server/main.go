package main

import (
	//"os"
	"runtime"
	//"runtime/pprof"
	//"fmt"
	//"math/rand"
	//"net"
	//"sync"
	//"time"

	"lei"
	//l4g "base/log4go"
	//"base/radix/redis"
)

func init() {
	InitCsCommand()
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	//var f *os.File
	//f, _ = os.Create("profile_file.cpu")
	//pprof.StartCPUProfile(f)
	//defer pprof.StopCPUProfile()

	conf := &lei.Config{
		Address:           ":2333",
		MaxReadMsgSize:    65536,
		ReadMsgQueueSize:  10240,
		ReadTimeOut:       600,
		MaxWriteMsgSize:   65536,
		WriteMsgQueueSize: 10240,
		WriteTimeOut:      600,
	}
	s := NewChatServer()
	//go s.HeartBeat()
	go s.BroadcastLoop()
	lei.TcpServe(s, conf)
}
