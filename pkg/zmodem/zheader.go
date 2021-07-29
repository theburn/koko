package zmodem

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
)

var HEX_OCTET_VALUE = getTableMap()

func getTableMap() map[byte]byte {
	HEX_DIGITS := [16]byte{48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 97, 98, 99, 100, 101, 102}
	ret := map[byte]byte{}
	for i := 0; i < len(HEX_DIGITS); i++ {
		ret[HEX_DIGITS[i]] = byte(i)
	}
	return ret
}

const (
	// ZPAD *
	ZPAD = 0x2a /* pad character; begins frames */

	// ZBIN A
	ZBIN = 0x41 /* binary frame indicator (CRC16) */

	// ZHEX B
	ZHEX = 0x42 /* hex frame indicator */

	// ZBIN32 C
	ZBIN32 = 0x43 /* binary frame indicator (CRC32) */

)

var (
	HEX_HEADER_CRLF    = []byte{CR, LF}
	HEX_HEADERCRLF_XON = []byte{0x0a, XON}

	HEX_HEADER_PREFIX = []byte{ZPAD, ZPAD, ZDLE, ZHEX}

	INARY16_HEADER_PREFIX = []byte{ZPAD, ZDLE, ZBIN}

	BINARY32_HEADER_PREFIX = []byte{ZPAD, ZDLE, ZBIN32}
)

func TrimLeadingGarbage(ibuffer []byte) []byte {
	var garbage = make([]byte, 0, len(ibuffer))

	var (
		discardAll, parser bool
	)

TrimLoop:
	for len(ibuffer) > 0 && !parser {
		firstZpad := bytes.IndexByte(ibuffer, ZPAD)
		if firstZpad == -1 {
			discardAll = true
			break TrimLoop
		} else {

			garbage = append(garbage, ZPAD)
			if len(ibuffer) < 2 {
				break TrimLoop
			} else if ibuffer[1] == ZPAD {
				//Two leading ZPADs should be a hex header.
			}

		}

	}
	if discardAll {
		garbage = append(garbage, ibuffer[0])
		ibuffer = ibuffer[1:]
	}
	return garbage
}

func ParseZFileHeader(p []byte) {

}

func parseHex(bytes_arr []byte) error {
	lf_pos := bytes.IndexByte(bytes_arr, 0x8a)
	if lf_pos == -1 {
		lf_pos = bytes.IndexByte(bytes_arr, LF)
	}
	if lf_pos == -1 {
		lf_pos = bytes.IndexByte(bytes_arr, CR)
	}
	if lf_pos == -1 {
		if len(bytes_arr) > 11 {
			return fmt.Errorf("invalid hex header - no LF detected within 12 bytes! %d", len(bytes_arr))
		}
	}

	hex_bytes := bytes_arr[:lf_pos]
	fmt.Println(hex_bytes, len(hex_bytes))

	if len(hex_bytes) == 19 {
		preceding := hex_bytes[len(hex_bytes)-1]
		hex_bytes = hex_bytes[:len(hex_bytes)-1]
		fmt.Println(hex_bytes, len(hex_bytes))
		if preceding != CR {
			return errors.New("invalid hex header: (cr/) lf doesn’t have cr")
		}
	}
	if len(hex_bytes) != 18 {
		return errors.New("invalid hex header: invalid number of bytes before LF",
		)
	}
	hex_bytes = hex_bytes[4:]
	fmt.Println(hex.Dump(hex_bytes))
	//Should be 7 bytes ultimately:
	//  1 for typenum
	//  4 for header data
	//  2 for CRC
	octets := parse_hex_octets(hex_bytes)
	fmt.Println(octets)
	
	return nil
}

func parse_hex_octets(p []byte) []byte {
	octets := make([]byte, len(p)/2)
	for i := 0; i < len(octets); i++ {
		octets[i] = (HEX_OCTET_VALUE[p[2*i]] << 4) + (HEX_OCTET_VALUE[p[1+2*i]])
	}
	binary.BigEndian.GoString()
	return octets
}

