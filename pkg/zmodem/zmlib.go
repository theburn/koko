package zmodem

import "bytes"

const (
	ZDLE     = 0x18 /* ctrl-x zmodem escape */
	XON      = 0x11
	XOFF     = 0x13
	XONHIGH  = 0x80 | XON
	XOFFHIGH = 0x80 | XOFF
	CAN      = 0x18
)

var (
	ABORTSEQUENCE  = []byte{CAN, CAN, CAN, CAN, CAN}
	CancelSequence = []byte{CAN, CAN, CAN, CAN, CAN, CAN, CAN, CAN,
		8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 0}
)

type Zmlib struct {
}

func (z Zmlib) StripIgnoredBytes(octets []byte) []byte {
	copyOctets := make([]byte, 0, len(octets))
	for i := len(octets) - 1; i >= 0; i-- {
		switch octets[i] {
		case XON, XOFFHIGH,
			XOFF, XONHIGH:
			continue
		default:
			copyOctets = append(copyOctets, octets[i])
		}
	}
	return copyOctets
}

func (z Zmlib) findSubArray(haystack, needle []byte) int {
	var (
		h, n int
	)
HAYSTACK:
	for h != -1 {
		h = bytes.IndexByte(haystack, needle[0])
		if h == -1 {
			break HAYSTACK
		}
		for n = 1; n < len(needle); n++ {
			if haystack[h+n] != needle[n] {
				h++
				continue HAYSTACK
			}
		}
		return h
	}
	return -1
}

/*

rz 上传
	init
	start
	end
	abort
*/
const (
	RZInit = iota
	RZStart
	RZEnd
	RZAbort
)
