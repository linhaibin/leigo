package lei

import (
	l4g "base/log4go"
	//"fmt"
	"net"
	"sync"
	"sync/atomic"
	//"time"
)

const (
	StateInit         int32 = 1
	StateDisconnected int32 = 2
	StateConnected    int32 = 3
)

type Sessioner interface {
	Init(*Broker)
	ProcessData(*PackHead, []byte) bool
	Close()
}

type Broker struct {
	cn      *conn
	session Sessioner
	conf    *Config

	ReadMsgQueue  chan []byte
	WriteMsgQueue chan *Message

	CloseChan chan struct{}

	State int32

	wg sync.WaitGroup
}

func newBroker(conf *Config, ses Sessioner) *Broker {
	b := &Broker{
		conf:      conf,
		session:   ses,
		CloseChan: make(chan struct{}),
		State:     StateInit,
	}
	return b
}

func (this *Broker) Connect() bool {
	conn, err := net.Dial("tcp", this.conf.Address)
	if err != nil {
		l4g.Info("connect failed ! err:%v", err)
		return false
	}

	return this.serve(conn)
}

func (this *Broker) serve(conn net.Conn) bool {
	if atomic.LoadInt32(&this.State) == StateConnected {
		l4g.Info("Broker is already serving !")
		return false
	}
	//conn.SetKeepAlive(true)
	this.cn = newconn(conn, this)
	this.WriteMsgQueue = make(chan *Message, this.conf.WriteMsgQueueSize)
	if this.conf.ReadMsgQueueSize > 0 {
		this.ReadMsgQueue = make(chan []byte, this.conf.ReadMsgQueueSize)
	}
	//this.cn.cn.SetReadDeadline(time.Duration(this.conf.ReadTimeOut) * time.Second)
	//this.cn.cn.SetWriteDeadline(time.Duration(this.conf.WriteTimeOut) * time.Second)

	atomic.StoreInt32(&this.State, StateConnected)

	this.wg.Add(1)
	go this.cn.writeLoop()
	this.wg.Add(1)
	go this.cn.readLoop()
	//this.wg.Add(1)
	//go this.processLoop()

	this.session.Init(this)
	return true
}

/*
func (this *Broker) processLoop() {
	defer this.wg.Done()
	if this.conf.ReadMsgQueueSize <= 0 {
		l4g.Info("ReadMsgQueueSize == 0, will not launch the processLoop.")
		return
	}
	for {
		select {
		case buf := <-this.ReadMsgQueue:
			if ph, ok := DecodePackHead(buf); ok {
				this.session.ProcessData(ph, buf[PACK_HEAD_LEN:])
			}
		case <-this.CloseChan:
			return
		}
	}
}
*/

func (this *Broker) transmitOrProcessMsg(buf []byte) {
	if this.conf.ReadMsgQueueSize > 0 {
		select {
		case this.ReadMsgQueue <- buf:
		case <-this.CloseChan:
			return
		}
	} else {
		if ph, ok := DecodePackHead(buf); ok {
			this.session.ProcessData(ph, buf[PACK_HEAD_LEN:])
		}
	}
}

func (this *Broker) Stop() {
	if !atomic.CompareAndSwapInt32(&this.State, StateConnected, StateDisconnected) {
		return
	}
	close(this.CloseChan)
	this.cn.netConn.Close()
	go func() {
		l4g.Info("wg is waiting !")
		this.wg.Wait()
		this.cn.close()
		this.session.Close()
		l4g.Info("[Broker] closed !")
		this.session = nil
		atomic.StoreInt32(&this.State, StateInit)
	}()
}

func (this *Broker) Write(ph *PackHead, info interface{}) bool {
	//如果这里的WriteMsgQueue堵住了，Write是在处理事务的goroutine中调用的，
	//如果这里使用同步，那么处理事务就会被这里堵住。所以这里不用for。
	select {
	case this.WriteMsgQueue <- &Message{ph, info}:
		return true
	case <-this.CloseChan:
		l4g.Error("session is closed: ph:%v, info:%v", ph, info)
		return false
	}
}

func (this *Broker) AddWaitGroup() {
	this.wg.Add(1)
}

func (this *Broker) DecWaitGroup() {
	this.wg.Done()
}

type Message struct {
	PH   *PackHead
	Info interface{}
}
