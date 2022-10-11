package h264

import (
	"bytes"
	"fmt"
)

const (
	NAL_SLICE           = 1
	NAL_DPA             = 2
	NAL_DPB             = 3
	NAL_DPC             = 4
	NAL_IDR_SLICE       = 5
	NAL_SEI             = 6
	NAL_SPS             = 7
	NAL_PPS             = 8
	NAL_AUD             = 9
	NAL_END_SEQUENCE    = 10
	NAL_END_STREAM      = 11
	NAL_FILLER_DATA     = 12
	NAL_SPS_EXT         = 13
	NAL_AUXILIARY_SLICE = 19
	NAL_KWAI_PRIV       = 31
	NAL_FF_IGNORE       = 0xff0f001
)

const (
	NALU_PRIORITY_DISPOSABLE = 0
	NALU_PRIRITY_LOW         = 1
	NALU_PRIORITY_HIGH       = 2
	NALU_PRIORITY_HIGHEST    = 3
)

var nalPrioMap = map[int]string{
	NALU_PRIORITY_DISPOSABLE: "disposable",
	NALU_PRIRITY_LOW:         "low",
	NALU_PRIORITY_HIGH:       "high",
	NALU_PRIORITY_HIGHEST:    "highest",
}

var nalTypeMap = map[int]string{
	NAL_SLICE:           "slice",
	NAL_DPA:             "dpa",
	NAL_DPB:             "dpb",
	NAL_DPC:             "dpc",
	NAL_IDR_SLICE:       "idr_slice",
	NAL_SEI:             "sei",
	NAL_SPS:             "sps",
	NAL_PPS:             "pps",
	NAL_AUD:             "aud",
	NAL_END_SEQUENCE:    "end_sequence",
	NAL_END_STREAM:      "end_stream",
	NAL_FILLER_DATA:     "filler_data",
	NAL_SPS_EXT:         "sps_ext",
	NAL_AUXILIARY_SLICE: "auxiliary_slice",
	NAL_KWAI_PRIV:       "kwai_priv",
	NAL_FF_IGNORE:       "ff_ignore",
}

func isStartCode(data []byte, pos int) (int, bool) {
	if data[pos] == 0 && data[pos+1] == 0 && data[pos+2] == 1 {
		return 3, true
	}

	if data[pos] == 0 && data[pos+1] == 0 && data[pos+2] == 0 && data[pos+3] == 1 {
		return 4, true
	}

	return 0, false
}

type Nalu struct {
	Len             int
	ForbiddenBit    int
	NalReferenceIdc int
	NalUnitType     int
}

func (n *Nalu) String() string {
	return fmt.Sprintf("|priority: %s|type: %s", nalPrioMap[n.NalReferenceIdc], nalTypeMap[n.NalUnitType])
}

func hexBytes(data []byte) {
	fmt.Println()
	for _, b := range data {
		fmt.Printf("%02X ", b)
	}
	fmt.Println()
}

func Parse(buffer *bytes.Buffer) {
	data, p := buffer.Bytes(), 0

	for p < len(data) {
		size, flag := isStartCode(data, p)
		if !flag {
			p++
			continue
		}

		p += size
		if p >= len(data) {
			break
		}

		b := int(data[p])
		n := Nalu{ForbiddenBit: b >> 7, NalReferenceIdc: (b >> 5) & 0x03, NalUnitType: b & 0x1f, Len: len(data) - p}
		fmt.Print(n.String())
		break
	}
}
