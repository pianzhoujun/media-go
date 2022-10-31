package core

import "io"

type BitStream struct {
	data      []byte
	offset    int
	offsetBit int
}

func NewBitStream(data []byte) *BitStream {
	return &BitStream{data: data, offset: len(data) - 1, offsetBit: 7}
}

func (bs *BitStream) Next() int {
	if bs.offset < 0 {
		return -1
	}

	rc := (int(bs.data[bs.offset]) >> bs.offsetBit) & 0x01
	if bs.offsetBit == 0 {
		bs.offsetBit = 7
		bs.offset--
	} else {
		bs.offsetBit--
	}

	return rc
}

func (bs *BitStream) ReadByte() (byte, error) {
	var rc byte

	for i := 0; i < 8; i++ {
		x := bs.Next()
		if x == -1 {
			return rc, io.EOF
		}

		rc = (rc << i) | byte(x)
	}

	return rc, nil
}
