package h264

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"media-go/codec/golomb"
	"media-go/core"
	"os"
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
	Rbsp            []byte
}

func (n *Nalu) String() string {
	nt := nalTypeMap[n.NalUnitType]
	if len(nt) == 0 {
		nt = "unknown"
	}
	return fmt.Sprintf("|priority: %s|type: %s", nalPrioMap[n.NalReferenceIdc], nt)
}

type SPS struct {
	ProfileIdc                      int
	ConstraintSet0Flag              int
	ConstraintSet1Flag              int
	ConstraintSet2Flag              int
	ConstraintSet3Flag              int
	ConstraintSet4Flag              int
	ConstraintSet5Flag              int
	ReservedZero2Bits               int
	LevelIdc                        int
	SPSId                           int
	ChromaFormatIdc                 int
	ResidualColourTransformFlag     int
	BitDepthLumaMinus8              int
	BitDepthChromaMinus8            int
	QpprimeYZeroTransformBypassFlag int
	SeqScalingMatrixPresentFlag     int
	ScalingMatrix4                  [6][16]int
	ScalingMatrix8                  [6][64]int
	Log2MaxFrameNumMinus4           int
	PicOrderCntType                 int
	Log2MaxPicOrderCntLsbMinus4     int
	DeltaPicOrderAlwaysZeroFlag     int
	OffsetForNonRefPic              int
	OffsetForTopToBottomField       int
	NumRefFramesInPicOrderCntCycle  int
	OffsetForRefFrame               []int
	NumRefFrames                    int
	GapsInFrameNumValueAllowedFlag  int
	PicWidthInMbsMinus1             int
	PicHeightInMapUnitsMinus1       int
	FrmaeMbsOnlyFlag                int
	MbAdaptiveFrameFieldFlag        int
	Direct8x8InferenceFlag          int
	FrameCropingFlag                int
	FrameCropLeftOffset             int
	FrameCropRightOffset            int
	FrameCropTopOffset              int
	FrameCropButtomOffset           int
	VuiParametersPresentFlag        int
	//...
}

type PPS struct {
}

var gNaluSize int

func ParseCts(buffer *bytes.Buffer) int {
	var cts int
	for i := 0; i < 3; i++ {
		b, _ := buffer.ReadByte()
		cts = (cts << 8) | int(b)
	}
	return cts
}

func ParseNalu(vf *VideoFrameInfo, buffer *bytes.Buffer) {
	cts := ParseCts(buffer)

	if gNaluSize == 0 {
		panic("nalu size is 0")
	}

	data := buffer.Next(gNaluSize)
	length := binary.BigEndian.Uint32(data)

	b := int(buffer.Bytes()[0])
	vf.Cts = cts
	vf.NaluInfo = &Nalu{ForbiddenBit: b >> 7, NalReferenceIdc: (b >> 5) & 0x03, NalUnitType: b & 0x1f, Len: int(length)}
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

// parse sps nalu and rbsp
func ParseSPS(data []byte) {
	buffer := bytes.NewBuffer(data)
	nalu := Nalu{}
	b, _ := buffer.ReadByte()
	n := int(b)

	nalu.ForbiddenBit = n >> 7
	nalu.NalReferenceIdc = (n >> 5) & 0x03
	nalu.NalUnitType = n & 0x1f

	if nalu.NalUnitType != 7 {
		panic(fmt.Sprintf("nalu type not match %d", nalu.NalUnitType))
	}

	nalu.Rbsp = make([]byte, len(data))
	n = 0

	for buffer.Len() > 0 {
		nalu.Rbsp[n], _ = buffer.ReadByte()
		if n > 2 && (nalu.Rbsp[n-2] == 0 && nalu.Rbsp[n-1] == 0 && nalu.Rbsp[n] == 3) {
			if buffer.Len() == 0 {
				break
			}

			nalu.Rbsp[n], _ = buffer.ReadByte()
		}

		n++
	}

	// TODO: parse sps rbsp
	buffer = bytes.NewBuffer(nalu.Rbsp[:n])
	var sps SPS
	b, _ = buffer.ReadByte()
	sps.ProfileIdc = int(b)

	// fmt.Printf("idc: %v\n", sps.ProfileIdc)

	b, _ = buffer.ReadByte()
	x := int(b)
	sps.ConstraintSet0Flag = (x >> 7) & 0x01
	sps.ConstraintSet1Flag = (x >> 6) & 0x01
	sps.ConstraintSet2Flag = (x >> 5) & 0x01
	sps.ConstraintSet3Flag = (x >> 4) & 0x01
	sps.ConstraintSet4Flag = (x >> 3) & 0x01
	sps.ConstraintSet5Flag = (x >> 2) & 0x01

	b, _ = buffer.ReadByte()
	sps.LevelIdc = int(b)
	// fmt.Printf("level idc: %v\n", sps.LevelIdc)
	bs := core.NewBitStream(buffer.Bytes())
	sps.SPSId = golomb.ReadUEV(bs)
	// fmt.Printf("sps id: %v\n", sps.SPSId)

	if sps.ProfileIdc == 100 || sps.ProfileIdc == 110 || sps.ProfileIdc == 122 ||
		sps.ProfileIdc == 244 || sps.ProfileIdc == 44 || sps.ProfileIdc == 83 ||
		sps.ProfileIdc == 86 || sps.ProfileIdc == 118 || sps.ProfileIdc == 128 {
		sps.ChromaFormatIdc = golomb.ReadUEV(bs)

		if sps.ChromaFormatIdc == 3 {
			sps.ResidualColourTransformFlag = bs.Next()
		}

		sps.BitDepthLumaMinus8 = golomb.ReadUEV(bs)
		sps.BitDepthChromaMinus8 = golomb.ReadUEV(bs)

		sps.QpprimeYZeroTransformBypassFlag = bs.Next()
		sps.SeqScalingMatrixPresentFlag = bs.Next()

		if sps.SeqScalingMatrixPresentFlag == 1 {
			scmpfsNum := 8
			if sps.ChromaFormatIdc == 3 {
				scmpfsNum = 12
			}

			for i := 0; i < scmpfsNum; i++ {
				// flagI := bs.Next()
			}
		}

		sps.Log2MaxFrameNumMinus4 = golomb.ReadUEV(bs)
		sps.PicOrderCntType = golomb.ReadUEV(bs)
		if sps.PicOrderCntType == 0 {
			sps.Log2MaxPicOrderCntLsbMinus4 = golomb.ReadUEV(bs)
		} else if sps.PicOrderCntType == 1 {
			sps.DeltaPicOrderAlwaysZeroFlag = bs.Next()
			sps.OffsetForNonRefPic = golomb.ReadUEV(bs)
			sps.OffsetForTopToBottomField = golomb.ReadUEV(bs)
			sps.NumRefFramesInPicOrderCntCycle = golomb.ReadUEV(bs)

			sps.OffsetForRefFrame = make([]int, 0)
			fmt.Printf("NumRefFramesInPicOrderCntCycle: %d\n", sps.NumRefFramesInPicOrderCntCycle)
			for i := 0; i < sps.NumRefFramesInPicOrderCntCycle; i++ {
				sps.OffsetForRefFrame = append(sps.OffsetForRefFrame, golomb.ReadUEV(bs))
			}
		}

		sps.NumRefFrames = golomb.ReadUEV(bs)
		sps.GapsInFrameNumValueAllowedFlag = bs.Next()
		sps.PicWidthInMbsMinus1 = golomb.ReadUEV(bs)
		sps.PicHeightInMapUnitsMinus1 = golomb.ReadUEV(bs)

		sps.FrmaeMbsOnlyFlag = bs.Next()
		if sps.FrmaeMbsOnlyFlag == 0 {
			sps.MbAdaptiveFrameFieldFlag = bs.Next()
		}

		sps.Direct8x8InferenceFlag = bs.Next()
		sps.FrameCropingFlag = bs.Next()
		if sps.FrameCropingFlag == 1 {
			sps.FrameCropLeftOffset = golomb.ReadUEV(bs)
			sps.FrameCropRightOffset = golomb.ReadUEV(bs)
			sps.FrameCropTopOffset = golomb.ReadUEV(bs)
			sps.FrameCropButtomOffset = golomb.ReadUEV(bs)
		}

		sps.VuiParametersPresentFlag = bs.Next()
		// vui paramters
	}

	os.Exit(0)
}

func ParsePPS(data []byte) {}

func ParseSeq(vf *VideoFrameInfo, buffer *bytes.Buffer) {
	cts := ParseCts(buffer)

	buffer.ReadByte()         // Configuration VerSion
	buffer.ReadByte()         // AVC Profile SPS[1]
	buffer.ReadByte()         // profile_compatibility SPS[2]
	buffer.ReadByte()         // AVC Level SPS[3]
	b, _ := buffer.ReadByte() // lengthSizeMinusOne
	gNaluSize = (int(b) & 0x03) + 1

	vf.Type = "seq header"
	vf.Cts = cts
	vf.NaluSize = gNaluSize

	buffer.ReadByte()
	// spsNum := int(b) & 0x1f
	// sps number shoud always be 1

	spsSize := int(binary.BigEndian.Uint16(buffer.Next(2)))
	data := buffer.Next(spsSize)
	ParseSPS(data)

	b, _ = buffer.ReadByte()
	ppsNum := int(b)
	for i := 0; i < ppsNum; i++ {
		spsSize := int(binary.BigEndian.Uint16(buffer.Next(2)))
		data := buffer.Next(spsSize)
		ParsePPS(data)
	}
}

type VideoFrameInfo struct {
	CodecType string
	Type      string
	Cts       int
	NaluSize  int
	NaluInfo  *Nalu
	Frame     interface{} // TODO
}

func (f *VideoFrameInfo) String() string {
	if f.NaluInfo == nil {
		return fmt.Sprintf("code: %s\ttype: %s\tcts:%d", f.CodecType, f.Type, f.Cts)
	}

	return fmt.Sprintf("code: %s\ttype: %s\tcts:%d\tnalu: %s", f.CodecType, f.Type, f.Cts, f.NaluInfo.String())
}

func Parse(buffer *bytes.Buffer) *VideoFrameInfo {
	vf := &VideoFrameInfo{}
	b, _ := buffer.ReadByte()
	switch int(b) {
	case 0:
		vf.CodecType = "seq"
		ParseSeq(vf, buffer)
	case 1:
		vf.CodecType = "nalu"
		ParseNalu(vf, buffer)
	case 2:
		vf.CodecType = "seq end"
	}

	fmt.Printf("\t%s\n", vf.String())

	return vf
}
