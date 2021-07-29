package zmodem

type Frame struct {
	ZHeader Header
	Subs    []SubPacket
}

type SubPacket struct {
}

type Header struct {
	frameType byte
	frameData [4]byte
}

const (
	ZF0 = 3 /* First flags byte */
	ZF1 = 2
	ZF2 = 1
	ZF3 = 0


	ZP0 = 0 /* Low order 8 bits of position */
	ZP1 = 1
	ZP2 = 2
	ZP3 = 3 /* High order 8 bits of file position */

	CRC = 0xff
)
