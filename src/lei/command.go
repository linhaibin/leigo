package lei

import (
	l4g "base/log4go"
)

type Services func(Sessioner, *PackHead, []byte) bool

type CommandM struct {
	cmds map[uint32]Services
}

func NewCommandM() *CommandM {
	return &CommandM{
		cmds: make(map[uint32]Services),
	}
}

func (this *CommandM) Register(cmd uint32, s Services) {
	this.cmds[cmd] = s
}

func (this *CommandM) Dispatch(session Sessioner, ph *PackHead, data []byte) bool {
	if service, ok := this.cmds[ph.Cmd]; ok {
		return service(session, ph, data)
	}
	l4g.Error("[CommandM]Dispatch failed ! cmd:%v, sid:%v, uid:%v", ph.Cmd, ph.Sid, ph.Uid)
	return false
}
