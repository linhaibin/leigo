package lei

import (
	l4g "base/log4go"
)

const (
	PACK_HEAD_LEN = 20
)

const (
	UINT32_BYTE_LEN = 4
)

type PackHead struct {
	Len uint32
	Cmd uint32
	Uid uint64
	Sid uint32
}

//big endian，大端解码编码，高位数据位于数组的低索引位置。
func DecodeUint32(data []byte) uint32 {
	return (uint32(data[0]) << 24) | (uint32(data[1]) << 16) | (uint32(data[2]) << 8) | (uint32(data[3]))
}

func EncodeUint32(n uint32, data []byte) {
	data[0] = byte((n >> 24) & 0xFF)
	data[1] = byte((n >> 16) & 0xFF)
	data[2] = byte((n >> 8) & 0xFF)
	data[3] = byte(n & 0xFF)
}

func DecodeUint64(data []byte) uint64 {
	return (uint64(data[0]) << 56) | (uint64(data[1]) << 48) | (uint64(data[2]) << 40) | (uint64(data[3]) << 32) | (uint64(data[4]) << 24) | (uint64(data[5]) << 16) | (uint64(data[6]) << 8) | uint64(data[7])
}

func EncodeUint64(n uint64, data []byte) {
	data[0] = byte((n >> 56) & 0xFF)
	data[1] = byte((n >> 48) & 0xFF)
	data[2] = byte((n >> 40) & 0xFF)
	data[3] = byte((n >> 32) & 0xFF)
	data[4] = byte((n >> 24) & 0xFF)
	data[5] = byte((n >> 16) & 0xFF)
	data[6] = byte((n >> 8) & 0xFF)
	data[7] = byte(n & 0xFF)
}

func EncodePackHead(buf []byte, ph *PackHead) bool {
	if len(buf) < PACK_HEAD_LEN {
		l4g.Error("[EncodePackHead] error. len(buff)=%v, PACK_HEAD_LEN=%v", len(buf), PACK_HEAD_LEN)
		return false
	}
	EncodeUint32(ph.Len, buf[0:4])
	EncodeUint32(ph.Cmd, buf[4:8])
	EncodeUint64(ph.Uid, buf[8:16])
	EncodeUint32(ph.Sid, buf[16:])
	return true
}

func DecodePackHead(buf []byte) (*PackHead, bool) {
	if len(buf) < PACK_HEAD_LEN {
		return nil, false
	}
	ph := &PackHead{
		Len: DecodeUint32(buf[:4]),
		Cmd: DecodeUint32(buf[4:8]),
		Uid: DecodeUint64(buf[8:16]),
		Sid: DecodeUint32(buf[16:]),
	}
	return ph, true
}
