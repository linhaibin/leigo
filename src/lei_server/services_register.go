package main

import (
	l4g "base/log4go"
	"github.com/golang/protobuf/proto"
	"lei"
	cs "lei_protocol"
)

var g_command = lei.NewCommandM()

func InitCsCommand() {
	g_command.Register(uint32(cs.ID_ID_C2S_Login), Login_C2S_Login)
	g_command.Register(uint32(cs.ID_ID_C2S_Chat), Chat_C2S_Chat)
}

func Login_C2S_Login(ses lei.Sessioner, ph *lei.PackHead, data []byte) bool {
	session, ok := ses.(*ChatSession)
	if !ok {
		l4g.Error("[Login_C2S_Login] Can not transform to ChatSession ! type:%+v", ses)
		return false
	}
	recv := &cs.C2S_Login{}
	if err := proto.Unmarshal(data, recv); err != nil {
		l4g.Error("[Login_C2S_Login] Can not unmarshal ! cmd:%v, uid:%v, sid:%v", ph.Cmd, ph.Uid, ph.Sid)
		return false
	}

	//l4g.Debug("[Login_C2S_Login] msg:%v", recv)
	id := recv.GetId()
	session.Id = id
	session.Server.Add(id, session)

	sendPh := &lei.PackHead{
		Cmd: uint32(cs.ID_ID_S2C_Login),
		Uid: uint64(id),
		Sid: uint32(id),
	}
	sendMsg := &cs.S2C_Login{
		Ret: proto.Uint32(uint32(cs.RET_RET_OK)),
	}
	session.Send(sendPh, sendMsg)
	//l4g.Debug("[Login_C2S_Login] done ! id:%v", id)
	return true
}

func Chat_C2S_Chat(ses lei.Sessioner, ph *lei.PackHead, data []byte) bool {
	session, ok := ses.(*ChatSession)
	if !ok {
		l4g.Error("[Chat_C2S_Chat] Can not transform to ChatSession ! type:%+v", ses)
		return false
	}
	recv := &cs.C2S_Chat{}
	if err := proto.Unmarshal(data, recv); err != nil {
		l4g.Error("[Chat_C2S_Chat] Can not unmarshal ! cmd:%v, uid:%v, sid:%v", ph.Cmd, ph.Uid, ph.Sid)
		return false
	}

	//l4g.Debug("[Chat_C2S_Chat] processing msg:%v", recv)

	tp := recv.GetType()
	toId := recv.GetToId()

	sendPh := &lei.PackHead{
		Cmd: uint32(cs.ID_ID_S2C_Chat),
		Uid: uint64(toId),
		Sid: toId,
	}
	sendMsg := &cs.S2C_Chat{
		Type:   recv.Type,
		FromId: recv.FromId,
		ToId:   recv.ToId,
		Msg:    recv.Msg,
	}

	if tp == uint32(1) {
		session.Server.SendToSession(toId, sendPh, sendMsg)
	} else {
		session.Server.Broadcast(sendPh, sendMsg)
	}
	//l4g.Debug("[Chat_C2S_Chat] done !")
	return true
}
