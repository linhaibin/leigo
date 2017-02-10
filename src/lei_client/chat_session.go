package main

import (
	"math/rand"
	"time"
	//"fmt"
	"sync"

	l4g "base/log4go"
	"lei"
	cs "lei_protocol"

	"github.com/golang/protobuf/proto"
)

type ChatSession struct {
	Id uint32
	B  *lei.Broker
	wg sync.WaitGroup

	r *rand.Rand
}

func NewChatSession(id uint32) *ChatSession {
	//l4g.Trace("new session id:%v", id)
	return &ChatSession{
		Id: id,
		B:  nil,
		r:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (this *ChatSession) ProcessData(ph *lei.PackHead, buf []byte) bool {
	return g_command.Dispatch(this, ph, buf)
}

func (this *ChatSession) Init(b *lei.Broker) {
	this.B = b
}

func (this *ChatSession) Close() {
	l4g.Info("********Client session %d is exited.", this.Id)
}

func (this *ChatSession) Send(ph *lei.PackHead, info interface{}) {
	this.B.Write(ph, info)
}

func (this *ChatSession) Stop() {
	this.B.Stop()
}

func (this *ChatSession) ChatSingle(toId uint32, content string) {
	ph := &lei.PackHead{
		Cmd: uint32(cs.ID_ID_C2S_Chat),
		Uid: uint64(this.Id),
	}
	msg := &cs.C2S_Chat{
		Type:   proto.Uint32(uint32(1)),
		FromId: proto.Uint32(this.Id),
		ToId:   proto.Uint32(toId),
		Msg:    proto.String(content),
	}
	l4g.Info("C:%v f:%v, t:%v", this.Id, this.Id, toId)
	this.Send(ph, msg)
}

func (this *ChatSession) ChatGroup(content string) {
	ph := &lei.PackHead{
		Cmd: uint32(cs.ID_ID_C2S_Chat),
		Uid: uint64(this.Id),
	}
	msg := &cs.C2S_Chat{
		Type:   proto.Uint32(uint32(2)),
		FromId: proto.Uint32(this.Id),
		ToId:   proto.Uint32(0),
		Msg:    proto.String(content),
	}
	this.Send(ph, msg)
}

func (this *ChatSession) RandChat() {
	//tp := uint32(this.r.Int31n(100)) % 2
	tp := uint32(1)
	content := "HelloWorld!"
	if tp == 1 {
		this.ChatSingle(uint32(this.r.Int31n(int32(MAX_CLIENT_NUM))+1), content)
	} else {
		this.ChatGroup(content)
	}
}

func (this *ChatSession) Login() {
	this.wg.Add(1)
	ph := &lei.PackHead{
		Cmd: uint32(cs.ID_ID_C2S_Login),
	}
	msg := &cs.C2S_Login{
		Id: proto.Uint32(this.Id),
	}
	this.Send(ph, msg)
	this.wg.Wait()
}

func (this *ChatSession) Logined() {
	this.wg.Done()
	//loginWg.Done()
}

func (this *ChatSession) IsClosed() bool {
	select {
	case <-this.B.CloseChan:
		return true
	default:
		return false
	}
	return false
}
