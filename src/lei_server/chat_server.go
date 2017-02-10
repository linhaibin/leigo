package main

import (
	//"fmt"
	"sync"
	"time"

	"lei"
	cs "lei_protocol"

	l4g "base/log4go"
)

type ChatServer struct {
	AuthedCount   uint32
	UnauthedCount uint32
	MapSessions   map[uint32]*ChatSession
	BroadcastChan chan *lei.Message
	sync.RWMutex
}

func NewChatServer() *ChatServer {
	return &ChatServer{
		AuthedCount:   uint32(0),
		UnauthedCount: uint32(0),
		MapSessions:   make(map[uint32]*ChatSession),
		BroadcastChan: make(chan *lei.Message, 1024*10),
	}
}

func (this *ChatServer) NewSession() lei.Sessioner {
	this.UnauthedCount += 1
	s := &ChatSession{
		Id:     uint32(0),
		Server: this,
	}
	return s
}

func (this *ChatServer) Add(id uint32, session *ChatSession) {
	this.Lock()
	this.AuthedCount += 1
	this.MapSessions[id] = session
	//l4g.Info("Add session:%v", id)
	if this.AuthedCount%100 == 0 {
		l4g.Info("Session count:%v", this.AuthedCount)
	}
	this.Unlock()
}

func (this *ChatServer) Remove(id uint32) {
	//l4g.Info("Remove session:%v", id)
	this.Lock()
	defer this.Unlock()
	if uint32(0) == id {
		this.UnauthedCount -= 1
	} else {
		this.AuthedCount -= 1
		delete(this.MapSessions, id)
	}
}

func (this *ChatServer) Close() {
	//l4g.Info("The ChatServer is closing...")
}

func (this *ChatServer) CountSession() int {
	return len(this.MapSessions)
}

func (this *ChatServer) HeartBeat() {
	if len(this.MapSessions) > 0 {
		for _, session := range this.MapSessions {
			ph := &lei.PackHead{
				Cmd: uint32(cs.ID_ID_S2C_HeartBeat),
				Uid: 0,
				Sid: 0,
			}
			msg := &cs.S2C_HeartBeat{}
			session.Send(ph, msg)
		}
	}
	time.AfterFunc(time.Second*3, this.HeartBeat)
}

func (this *ChatServer) BroadcastLoop() {
	for {
		select {
		case msg := <-this.BroadcastChan:
			{
				for _, ses := range this.MapSessions {
					ses.Send(msg.PH, msg.Info)
				}
			}
		}
	}
}

func (this *ChatServer) SendToSession(id uint32, ph *lei.PackHead, info interface{}) {
	this.RLock()
	defer this.RUnlock()
	if session, ok := this.MapSessions[uint32(id)]; ok {
		session.Send(ph, info)
	}
}

func (this *ChatServer) Broadcast(ph *lei.PackHead, info interface{}) {
	select {
	case this.BroadcastChan <- &lei.Message{ph, info}:
	}
}

type ChatSession struct {
	Id     uint32
	B      *lei.Broker
	Server *ChatServer
}

func (this *ChatSession) ProcessData(ph *lei.PackHead, buf []byte) bool {
	return g_command.Dispatch(this, ph, buf)
}

func (this *ChatSession) Init(b *lei.Broker) {
	this.B = b
}

func (this *ChatSession) Close() {
	this.Server.Remove(this.Id)
	//l4g.Info("********Server session %d is exited. current session num:%d", this.Id, this.Server.CountSession())
	this.Server = nil
}

func (this *ChatSession) Send(ph *lei.PackHead, info interface{}) {
	this.B.Write(ph, info)
}
