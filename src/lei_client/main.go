package main

import (
	"math/rand"
	"runtime"
	"sync"
	"time"
	//"fmt"

	"lei"

	l4g "base/log4go"
)

const (
	MAX_CLIENT_NUM uint32 = 10000
)

func init() {
	InitCsCommand()
}

//var loginWg sync.WaitGroup
//var loginChan chan bool = make(chan bool)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	var wg sync.WaitGroup
	conf := &lei.Config{
		Address:           "127.0.0.1:7233",
		MaxReadMsgSize:    65536,
		ReadMsgQueueSize:  10240,
		ReadTimeOut:       600,
		MaxWriteMsgSize:   65536,
		WriteMsgQueueSize: 10240,
		WriteTimeOut:      600,
	}
	l4g.LoadConfiguration("config_log.xml")
	for n := MAX_CLIENT_NUM; n > 0; n -= 1 {
		//l4g.Info("forloop %v", n)
		wg.Add(1)
		//loginWg.Add(1)
		go func(id uint32) {
			defer wg.Done()

			r := rand.New(rand.NewSource(time.Now().UnixNano()))

			//delay := r.Intn(1 << 7) //伪随机时间点登陆
			//l4g.Info("client %v wait for creating for %vs", id, delay)
			//time.Sleep(time.Second * time.Duration(delay))

			client := NewChatSession(id)
			l4g.Info("client %v connecting...", id)
			if !lei.TcpClientServe(client, conf) {
				l4g.Error("client %v TCPClientServefailed.", id)
				return
			}
			l4g.Info("client %v created.", id)
			client.Login()
			defer client.Stop()
			l4g.Info("client %v logined.", id)

			//<-loginChan
			l4g.Trace("[main] client %v start to chat.", id)

			silence := false
			chated := 0
			for {
				if client.IsClosed() {
					break
				}
				//client.Write([]byte("hello world !"))
				//l4g.Trace("[main] client %v RandChat.", id)
				if !silence {
					client.RandChat()
					chated += 1
					if ttt := r.Intn(100); ttt < 0 {
						l4g.Trace("[main] client %v is silenced. chated:%v, silence:%v, ttt:%v", id, chated, silence, ttt)
						silence = true
					}
				}

				time.Sleep(time.Second)
				sleep := r.Intn(3)
				time.Sleep(time.Second * time.Duration(sleep))

				//if i := r.Intn(100); i < 0 {
				//	l4g.Trace("[main] client %v break.", id)
				//	l4g.Trace("[main] client %v chated:%v", id, chated)
				//	break
				//}
			}
		}(uint32(n))
	}

	//loginWg.Wait()
	//close(loginChan)

	l4g.Trace("[main] waiting...")
	wg.Wait()
	return
}
