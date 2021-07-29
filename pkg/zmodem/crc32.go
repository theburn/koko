package zmodem

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
)

type CRC32 struct {
}

func (c CRC32) Checksum(data []byte) []byte {
	var out bytes.Buffer
	value := crc32.ChecksumIEEE(data)
	err := binary.Write(&out, binary.BigEndian, value)
	if err != nil {
		panic(err)
	}
	return out.Bytes()

}

func (c CRC32) CRC16Validate(data []byte, expected []byte) error {

	actual := c.Checksum(data)

	// validate the provided checksum against the calculated
	if !bytes.Equal(actual, expected) {
		return ErrInvalidChecksum
	}

	return nil
}
