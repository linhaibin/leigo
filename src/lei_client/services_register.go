package main

import (
	l4g "base/log4go"
	"github.com/golang/protobuf/proto"
	"lei"
	cs "lei_protocol"
)

var g_command = lei.NewCommandM()

func InitCsCommand() {
	g_command.Register(uint32(cs.ID_ID_S2C_Login), Login_S2C_Login)
	g_command.Register(uint32(cs.ID_ID_S2C_Chat), Chat_S2C_Chat)
}

func Login_S2C_Login(ses lei.Sessioner, ph *lei.PackHead, data []byte) bool {
	session, ok := ses.(*ChatSession)
	if !ok {
		l4g.Error("Can not transform to ChatSession ! type:%+v", ses)
		return false
	}
	recv := &cs.S2C_Login{}
	if err := proto.Unmarshal(data, recv); err != nil {
		l4g.Error("Can not unmarshal ! cmd:%v, uid:%v, sid:%v", ph.Cmd, ph.Uid, ph.Sid)
		return false
	}

	if recv.GetRet() != uint32(cs.RET_RET_OK) {
		l4g.Error("Login_S2C_Login login failed ! ret:%v", recv.GetRet())
		return true
	}

	session.Logined()
	//l4g.Debug("Login_S2C_Login done !")
	return true
}

func Chat_S2C_Chat(ses lei.Sessioner, ph *lei.PackHead, data []byte) bool {
	/*
		_, ok := ses.(*ChatSession)
		if !ok {
			l4g.Error("Can not transform to ChatSession ! type:%+v", ses)
			return false
		}
		recv := &cs.S2C_Chat{}
		if err := proto.Unmarshal(data, recv); err != nil {
			l4g.Error("Can not unmarshal ! cmd:%v, uid:%v, sid:%v", ph.Cmd, ph.Uid, ph.Sid)
			return false
		}

		tp := recv.GetType()
		//fromId := recv.GetFromId()
		//toId := recv.GetToId()
		//content := recv.GetMsg()
		if uint32(1) == tp {
			//l4g.Info("Session %v receive Single chat. from:%v, to:%v, cont:%v", session.Id, fromId, toId, content)
		} else {
			//l4g.Info("Session %v receive Group chat from %v:%v", session.Id, fromId, content)
		}
		return true
	*/
	session, ok := ses.(*ChatSession)
	if !ok {
		l4g.Error("Can not transform to ChatSession ! type:%+v", ses)
		return false
	}
	recv := &cs.S2C_Chat{}
	if err := proto.Unmarshal(data, recv); err != nil {
		l4g.Error("Can not unmarshal ! cmd:%v, uid:%v, sid:%v", ph.Cmd, ph.Uid, ph.Sid)
		return false
	}

	tp := recv.GetType()
	fromId := recv.GetFromId()
	toId := recv.GetToId()
	content := recv.GetMsg()
	if uint32(1) == tp {
		l4g.Info("C:%v red. f:%v, t:%v", session.Id, fromId, toId)
	} else {
		l4g.Info("Session %v receive Group chat from %v:%v", session.Id, fromId, content)
	}
	return true
}
