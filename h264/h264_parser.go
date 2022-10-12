package h264

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	AVC_PKT_SEQ_HEADER = 0
	AVC_PKT_NALU       = 1
	AVC_PKT_END_SEQ    = 2
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

// func isStartCode(data []byte, pos int) (int, bool) {
// 	if data[pos] == 0 && data[pos+1] == 0 && data[pos+2] == 1 {
// 		return 3, true
// 	}

// 	if data[pos] == 0 && data[pos+1] == 0 && data[pos+2] == 0 && data[pos+3] == 1 {
// 		return 4, true
// 	}

// 	return 0, false
// }

type Nalu struct {
	Len             int
	ForbiddenBit    int
	NalReferenceIdc int
	NalUnitType     int
}

func (n *Nalu) String() string {
	nt := nalTypeMap[n.NalUnitType]
	if len(nt) == 0 {
		nt = "unknown"
	}
	return fmt.Sprintf("|priority: %s|type: %s", nalPrioMap[n.NalReferenceIdc], nt)
}

// func hexBytes(data []byte) {
// 	fmt.Println()
// 	for _, b := range data {
// 		fmt.Printf("%02X ", b)
// 	}
// 	fmt.Println()
// }

var gNaluSize int

func ParseNalu(buffer *bytes.Buffer) {

	var cts int
	for i := 0; i < 3; i++ {
		b, _ := buffer.ReadByte()
		cts = (cts << 8) | int(b)
	}

	fmt.Printf("|cts=%d", cts)

	if gNaluSize == 0 {
		panic("nalu size is 0")
	}

	data := buffer.Next(gNaluSize)
	length := binary.BigEndian.Uint32(data)

	// fmt.Printf("\nnalu size: %d nalu length: %d\n", gNaluSize, length)

	b := int(buffer.Bytes()[0])
	n := Nalu{ForbiddenBit: b >> 7, NalReferenceIdc: (b >> 5) & 0x03, NalUnitType: b & 0x1f, Len: int(length)}
	fmt.Print(n.String())
}

// type SeqHeader struct {
// 	version       int //8
// 	provile       int //8
// 	compatibility int //8
// 	level         int //8
// 	reserved   	  int //6
// 	naluLenSize   int //2
// 	//...
// }

func ParseSeq(buffer *bytes.Buffer) {
	var cts int
	for i := 0; i < 3; i++ {
		b, _ := buffer.ReadByte()
		cts = (cts << 8) | int(b)
	}

	fmt.Printf("|type=seq header|cts=%d", cts)

	buffer.ReadByte()
	buffer.ReadByte()
	buffer.ReadByte()
	buffer.ReadByte()
	b, _ := buffer.ReadByte()
	gNaluSize = (int(b) & 0x03) + 1
}

func Parse(buffer *bytes.Buffer) {
	fmt.Printf("|type=%d", int(buffer.Bytes()[0]))
	switch buffer.Bytes()[0] {
	case 0:
		ParseSeq(buffer)
	case 1:
		ParseNalu(buffer)
	case 2:
		break
	}
}
