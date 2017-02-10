package lei

import (
	//"fmt"
	"bufio"
	"io"
	"net"
	//"sync"
	"errors"
	"time"

	"counter"

	l4g "base/log4go"
	"github.com/golang/protobuf/proto"
)

const (
	WRITE_COUNTER_NAME string = "writecounter"
	READ_COUNTER_NAME  string = "readcounter"
)

const (
	WRITE_TIME_OUT = 10 * time.Second
	READ_TIME_OUT  = 60 * time.Second
)

func init() {
	counter.AddCounter("writecounter", 1024*1024, 0)
	counter.AddCounter("readcounter", 1024*1024, 0)
}

type conn struct {
	netConn net.Conn
	broker  *Broker

	msgLengthBuf []byte

	writeTimeOut time.Duration
	readTimeOut  time.Duration
}

func newconn(cn net.Conn, b *Broker) *conn {
	c := &conn{
		netConn:      cn,
		broker:       b,
		msgLengthBuf: make([]byte, UINT32_BYTE_LEN),
	}
	if c.broker.conf.ReadTimeOut > 0 {
		c.readTimeOut = time.Duration(c.broker.conf.ReadTimeOut) * time.Second
	} else {
		c.readTimeOut = READ_TIME_OUT
	}
	if c.broker.conf.WriteTimeOut > 0 {
		c.writeTimeOut = time.Duration(c.broker.conf.WriteTimeOut) * time.Second
	} else {
		c.writeTimeOut = WRITE_TIME_OUT
	}
	return c
}

func (this *conn) close() {
	this.broker = nil
}

/*
func (this *conn) writeLoop() {
	maxSize := uint32(this.broker.conf.MaxWriteMsgSize)
	write_buff := make([]byte, maxSize)
FOR1:
	for {
		select {
		case msg := <-this.broker.WriteMsgQueue:
			{
				len, ok := Marshal(msg.PH, msg.Info, write_buff, maxSize)
				if !ok {
					l4g.Error("[conn]writeLoop Marshal failed !")
					break
				}
				if !this.write(write_buff[:len]) {
					break FOR1
				}
				counter.Count(WRITE_COUNTER_NAME, 1)
			}
		case <-this.broker.CloseChan:
			{
				l4g.Info("broker is closed !")
				this.broker.wg.Done()
				return
			}
		}
	}
	l4g.Info("stop the broker in writeLoop !")
	this.broker.Stop()
	this.broker.wg.Done()
}
*/

func (this *conn) writeLoop() {
	maxSize := uint32(this.broker.conf.MaxWriteMsgSize)
	write_buff := make([]byte, maxSize)
	msg_buff := make([]byte, maxSize)
FOR1:
	for {
		select {
		case msg := <-this.broker.WriteMsgQueue:
			{
				data, len, ok := Marshal(msg.PH, msg.Info, msg_buff, maxSize)
				if !ok {
					l4g.Error("[conn]writeLoop Marshal failed !")
					break
				}
				copy(write_buff, data)
				index := len
				count := int64(1)

				for more := true; more; {
					select {
					case msg := <-this.broker.WriteMsgQueue:
						{
							data, len, ok := Marshal(msg.PH, msg.Info, msg_buff, maxSize)
							if ok {
								if index+len > maxSize {
									if !this.write(write_buff[:index]) {
										break FOR1
									}
									index = 0
									counter.Count(WRITE_COUNTER_NAME, count)
									count = 0
								}
								copy(write_buff[index:], data)
								index += len
								count += 1
							}
						}
					case <-this.broker.CloseChan:
						{
							l4g.Info("broker is closed !")
							this.broker.wg.Done()
							return
						}
					default:
						more = false
					}
				}
				if !this.write(write_buff[:index]) {
					break FOR1
				}
				counter.Count(WRITE_COUNTER_NAME, count)
			}
		case <-this.broker.CloseChan:
			{
				l4g.Info("broker is closed !")
				this.broker.wg.Done()
				return
			}
		}
	}
	l4g.Info("stop the broker in writeLoop !")
	this.broker.Stop()
	this.broker.wg.Done()
}

func (this *conn) write(buf []byte) bool {
	wn := len(buf)
	for wn > 0 {
		n, err := this.netConn.Write(buf)
		if err != nil {
			l4g.Info("Write error ! err:%v", err)
			return false
		}
		wn -= n
		buf = buf[n:]
	}
	return true
}

func (this *conn) readLoop() {
	reader := bufio.NewReader(this.netConn)
	for {
		this.netConn.SetReadDeadline(time.Now().Add(this.readTimeOut))
		buf, err := this.read(reader)
		if err != nil {
			l4g.Error("[Conn] read error ! readLoop will exit. err:%v", err)
			break
		}
		this.broker.transmitOrProcessMsg(buf)
		counter.Count(READ_COUNTER_NAME, 1)
	}
	l4g.Info("stop the broker in readLoop !")
	this.broker.Stop()
	this.broker.wg.Done()
}

func (this *conn) read(reader io.Reader) ([]byte, error) {
	if _, err := io.ReadFull(reader, this.msgLengthBuf); err != nil {
		l4g.Error("[Conn] msgLength ReadFull failed ! err:%v, (local,remote):(%s,%s)", err, this.netConn.LocalAddr().String(), this.netConn.RemoteAddr().String())
		return nil, err
	}
	length := DecodeUint32(this.msgLengthBuf)
	if length > uint32(this.broker.conf.MaxReadMsgSize) {
		l4g.Error("[Conn]length:%v > maxmsgsize:%v", length, this.broker.conf.MaxReadMsgSize)
		return nil, errors.New("read overflow !")
	}
	buf := make([]byte, length)
	copy(buf, this.msgLengthBuf)
	if _, err := io.ReadFull(reader, buf[UINT32_BYTE_LEN:]); err != nil {
		l4g.Error("[Conn] body ReadFull failed ! err:%v, (local,remote):(%s,%s)", err, this.netConn.LocalAddr().String(), this.netConn.RemoteAddr().String())
		return nil, err
	}
	return buf, nil
}

/*
func Marshal(ph *PackHead, info interface{}, buf []byte, maxSize uint32) (length uint32, ok bool) {
	length = 0
	switch v := info.(type) {
	case []byte:
		{
			var data []byte = v
			length = uint32(len(data))
			copy(buf[PACK_HEAD_LEN:], data)
		}
	case proto.Message:
		{
			//MarshalWithBytes会将序列化后的数据，进行data_buff=append(data_buff, data...)
			if data, err := proto.MarshalWithBytes(v, buf[PACK_HEAD_LEN:PACK_HEAD_LEN]); err == nil {
				length = uint32(len(data))
			} else {
				l4g.Error("[Conn] Marshal error. cmd:%v, sid:%v, uid:%v", ph.Cmd, ph.Sid, ph.Uid)
				return 0, false
			}
		}
	default:
		{
			l4g.Error("[Conn] Marshal unknown info type. cmd:%v, sid:%v, uid:%v", ph.Cmd, ph.Sid, ph.Uid)
			return 0, false
		}
	}
	if PACK_HEAD_LEN+length > maxSize {
		l4g.Error("[Conn] Marshal overflow. cmd:%v, sid:%v, uid:%v, size:%v, maxSize:%v", ph.Cmd, ph.Sid, ph.Uid, PACK_HEAD_LEN+length, maxSize)
		return 0, false
	}
	length += PACK_HEAD_LEN
	ph.Len = length
	EncodePackHead(buf[:PACK_HEAD_LEN], ph)
	return length, true
}
*/

func Marshal(ph *PackHead, info interface{}, buf []byte, maxSize uint32) (data []byte, length uint32, ok bool) {
	data = nil
	length = 0
	ok = false
	switch v := info.(type) {
	case []byte:
		{
			//var data []byte = v
			data = v
			length = uint32(len(data))
			copy(buf[PACK_HEAD_LEN:], data)
		}
	case proto.Message:
		{
			//MarshalWithBytes会将序列化后的数据，进行data_buff=append(data_buff, data...)
			var err error = nil
			if data, err = proto.MarshalWithBytes(v, buf[PACK_HEAD_LEN:PACK_HEAD_LEN]); err == nil {
				length = uint32(len(data))
			} else {
				l4g.Error("[Conn] Marshal error. cmd:%v, sid:%v, uid:%v", ph.Cmd, ph.Sid, ph.Uid)
				return
			}
		}
	default:
		{
			l4g.Error("[Conn] Marshal unknown info type. cmd:%v, sid:%v, uid:%v", ph.Cmd, ph.Sid, ph.Uid)
			return
		}
	}
	if PACK_HEAD_LEN+length > maxSize {
		l4g.Error("[Conn] Marshal overflow. cmd:%v, sid:%v, uid:%v, size:%v, maxSize:%v", ph.Cmd, ph.Sid, ph.Uid, PACK_HEAD_LEN+length, maxSize)
		return
	}
	length += PACK_HEAD_LEN
	ph.Len = length
	EncodePackHead(buf[:PACK_HEAD_LEN], ph)
	ok = true
	data = buf[0:length]
	//l4g.Debug("Marshal length:%v", length)
	return
}
