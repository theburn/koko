package zmodem

// ASCII
const (
	SOH = 0x01
	STX = 0x02
	EOT = 0x04
	ENQ = 0x05
	ACK = 0x06
	LF  = 0x0a
	CR  = 0x0d
	//XON  = 0x11
	//XOFF = 0x13
	NAK = 0x15
	//CAN  = 0x18
)

// ZMODEM
const (

	//ZDLE     = 0x18 /* ctrl-x zmodem escape */

	ZDLEE    = 0x58 /* escaped ZDLE */
	ZBINR32  = 0x44 /* run length encoded binary frame (CRC32) */
	ZVBIN    = 0x61 /* binary frame indicator (CRC16) */
	ZVHEX    = 0x62 /* hex frame indicator */
	ZVBIN32  = 0x63 /* binary frame indicator (CRC32) */
	ZVBINR32 = 0x64 /* run length encoded binary frame (CRC32) */

	ZRESC = 0x7e /* run length encoding flag / escape character */

)

//FRAME TYPES
const (
	ZRQINIT    = 0x00 /* request receive init (s->r) */
	ZRINIT     = 0x01 /* receive init (r->s) */
	ZSINIT     = 0x02 /* send init sequence (optional) (s->r) */
	ZACK       = 0x03 /* ack to ZRQINIT ZRINIT or ZSINIT (s<->r) */
	ZFILE      = 0x04 /* file name (s->r) */
	ZSKIP      = 0x05 /* skip this file (r->s) */
	ZNAK       = 0x06 /* last packet was corrupted (?) */
	ZABORT     = 0x07 /* abort batch transfers (?) */
	ZFIN       = 0x08 /* finish session (s<->r) */
	ZRPOS      = 0x09 /* resume data transmission here (r->s) */
	ZDATA      = 0x0a /* data packet(s) follow (s->r) */
	ZEOF       = 0x0b /* end of file reached (s->r) */
	ZFERR      = 0x0c /* fatal read or write error detected (?) */
	ZCRC       = 0x0d /* request for file CRC and response (?) */
	ZCHALLENGE = 0x0e /* security challenge (r->s) */
	ZCOMPL     = 0x0f /* request is complete (?) */
	ZCAN       = 0x10 /* pseudo frame; other end cancelled session with 5* CAN */
	ZFREECNT   = 0x11 /* request free bytes on file system (s->r) */
	ZCOMMAND   = 0x12 /* issue command (s->r) */
	ZSTDERR    = 0x13 /* output data to stderr (??) */
)

//ZDLE SEQUENCES
const (

	//ZCRCE = 0x68 /* CRC next, frame ends, header packet follows */
	//ZCRCG = 0x69 /* CRC next, frame continues nonstop */
	//ZCRCQ = 0x6a /* CRC next, frame continuous, ZACK expected */
	//ZCRCW = 0x6b /* CRC next, ZACK expected, end of frame */

	ZRUB0 = 0x6c /* translate to rubout 0x7f */
	ZRUB1 = 0x6d /* translate to rubout 0xff */
)

//RECEIVER CAPABILITY FLAGS
const (
	CANFDX  = 0x01 /* Rx can send and receive true full duplex */
	CANOVIO = 0x02 /* Rx can receive data during disk I/O */
	CANBRK  = 0x04 /* Rx can send a break signal */
	CANCRY  = 0x08 /* Receiver can decrypt */
	CANLZW  = 0x10 /* Receiver can uncompress */
	CANFC32 = 0x20 /* Receiver can use 32 bit Frame Check */
	ESCCTL  = 0x40 /* Receiver expects ctl chars to be escaped    */
	ESC8    = 0x80 /* Receiver expects 8th bit to be escaped */
)

/*
zmodem 以 ZDLE 作为标志符号
Link Escape coding

5 个 连续的 CAN 中断 zmodem 会话 即
8 个连续的 CAN 则是正常的发送数据


zmodem 数据包 header--- 最大 16 bytes
			 data --- 最大 1024 bytes
*/

var (
	abortSession = []byte{CAN, CAN, CAN, CAN, CAN}
	sentSession  = []byte{CAN, CAN, CAN, CAN, CAN, CAN, CAN, CAN}

	set             = []byte{ZDLE, ACK}
	reset           = []byte{ZDLE, ENQ}
	Ignores         = []byte{ZDLE, ZCAN, 0x90, 0x11, 0x91, 0x13, 0x93}
	telnetIgnore1   = []byte{0x40, 0x0d, 0x8d}
	telnetIgnore2   = []byte{0xc0, 0x0d, 0x8d}
	receiverIgnores = []byte{0x11, 0x91, 0x13, 0x93}
)



const (
	zmoden = iota
)

/*

rz 相关

客户端发送 rz    服务端
			<---
*/
