package tpkt

import (
	"bytes"
	"fmt"
	"github.com/chuckpreslar/emission"
	"github.com/icodeface/grdp/core"
)

// take idea from https://github.com/Madnikulin50/gordp

/**
 * Type of tpkt packet
 * Fastpath is use to shortcut RDP stack
 * @see http://msdn.microsoft.com/en-us/library/cc240621.aspx
 * @see http://msdn.microsoft.com/en-us/library/cc240589.aspx
 */
type TpktAction byte

const (
	FASTPATH_ACTION_FASTPATH TpktAction = 0x0
	FASTPATH_ACTION_X224                = 0x3
)

/**
 * TPKT layer of rdp stack
 */
type TPKT struct {
	emission.Emitter
	Conn    *core.SocketLayer
	secFlag byte
}

func New(s *core.SocketLayer) *TPKT {
	t := &TPKT{*emission.NewEmitter(), s, 0}
	core.StartReadBytes(2, s, t.recvHeader)
	return t
}

func (t *TPKT) Read(b []byte) (n int, err error) {
	return t.Conn.Read(b)
}

func (t *TPKT) Write(data []byte) (n int, err error) {
	buff := &bytes.Buffer{}
	core.WriteUInt8(FASTPATH_ACTION_X224, buff)
	core.WriteUInt8(0, buff)
	core.WriteUInt16BE(uint16(len(data)+4), buff)
	buff.Write(data)
	fmt.Println("tpkt Write", buff.Bytes())
	return t.Conn.Write(buff.Bytes())
}

func (t *TPKT) Close() error {
	return t.Conn.Close()
}

func (t *TPKT) recvHeader(s []byte, err error) {
	fmt.Println("tpkt recvHeader", s, err)
	if err != nil {
		t.Emit("error", err)
		return
	}
	version := s[0]
	if version == FASTPATH_ACTION_X224 {
		fmt.Println("tptk recvHeader FASTPATH_ACTION_X224, wait for recvExtendedHeader")
		core.StartReadBytes(2, t.Conn, t.recvExtendedHeader)
	} else {
		t.secFlag = (version >> 6) & 0x3
		length := int(s[1])
		if length&0x80 != 0 {
			core.StartReadBytes(1, t.Conn, func(s []byte, err error) {
				t.recvExtendedFastPathHeader(s, length, err)
			})
		} else {
			core.StartReadBytes(length-2, t.Conn, t.recvFastPath)
		}
	}
}

func (t *TPKT) recvExtendedHeader(s []byte, err error) {
	fmt.Println("tpkt recvExtendedHeader", s, err)
	if err != nil {
		return
	}
	r := bytes.NewReader(s)
	size, _ := core.ReadUint16BE(r)
	fmt.Println("tpkt wait recvData")
	core.StartReadBytes(int(size-4), t.Conn, t.recvData)
}

func (t *TPKT) recvData(s []byte, err error) {
	fmt.Println("tpkt recvData", s, err)
	if err != nil {
		return
	}
	fmt.Println("tpkt emit data")
	t.Emit("data", s)
	fmt.Println("tpkt wait recvHeader")
	core.StartReadBytes(2, t.Conn, t.recvHeader)
}

func (t *TPKT) recvExtendedFastPathHeader(s []byte, length int, err error) {
	fmt.Println("tpkt recvExtendedFastPathHeader", s, length, err)

}

func (t *TPKT) recvFastPath(s []byte, err error) {
	fmt.Println("tpkt recvFastPath", s, err)
}
