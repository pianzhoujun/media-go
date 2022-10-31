package golomb

import (
	"media-go/core"
)

//Exp-Golomb
func ReadUEV(bs *core.BitStream) int {
	var leadingZerosBit int

	for {
		x := bs.Next()
		if x == 1 {
			break
		}

		leadingZerosBit++
	}

	if leadingZerosBit >= 31 {
		panic("invalid UEV")
	}

	v := (1 << leadingZerosBit) - 1
	for i := 0; i < leadingZerosBit; i++ {
		b := bs.Next()
		v += b << (leadingZerosBit - 1 - i)
	}

	return v
}
