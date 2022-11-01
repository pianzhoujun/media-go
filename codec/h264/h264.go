package h264

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"media-go/codec/golomb"
	"media-go/core"
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
	RbspSize        int
}

func (n *Nalu) String() string {
	nt := nalTypeMap[n.NalUnitType]
	if len(nt) == 0 {
		nt = "unknown"
	}
	return fmt.Sprintf("|priority: %s|type: %s", nalPrioMap[n.NalReferenceIdc], nt)
}

type SPS struct {
	ProfileIdc                      int        `json:"profile_idc"`
	ConstraintSet0Flag              int        `json:"constraint_set0_flag"`
	ConstraintSet1Flag              int        `json:"constraint_set1_flag"`
	ConstraintSet2Flag              int        `json:"constraint_set2_flag"`
	ConstraintSet3Flag              int        `json:"constraint_set3_flag"`
	ConstraintSet4Flag              int        `json:"constraint_set4_flag"`
	ConstraintSet5Flag              int        `json:"constraint_set5_flag"`
	ReservedZero2Bits               int        `json:"-"`
	LevelIdc                        int        `json:"level_idc"`
	SPSId                           int        `json:"sps_id"`
	ChromaFormatIdc                 int        `json:"chroma_format_idc"`
	ResidualColourTransformFlag     int        `json:"residual_colour_transform_flag"`
	BitDepthLumaMinus8              int        `json:"bit_depth_luma_minus8"`
	BitDepthChromaMinus8            int        `json:"bit_deptn_chroma_minus8"`
	QpprimeYZeroTransformBypassFlag int        `json:"qpprime_y_zero_transform_bypass_flag"`
	SeqScalingMatrixPresentFlag     int        `json:"seq_scaling_matrix_present_flag"`
	ScalingMatrix4                  [6][16]int `json:"-"`
	ScalingMatrix8                  [6][64]int `json:"-"`
	Log2MaxFrameNumMinus4           int        `json:"log2_max_frame_num_minus4"`
	PicOrderCntType                 int        `json:"pic_order_cnt_type"`
	Log2MaxPicOrderCntLsbMinus4     int        `json:"log2_max_pic_order_cnt_lsb_minus4"`
	DeltaPicOrderAlwaysZeroFlag     int        `json:"delta_pic_order_always_zero_flag"`
	OffsetForNonRefPic              int        `json:"offset_for_non_ref_pic"`
	OffsetForTopToBottomField       int        `json:"offset_fot_top_to_bottom_field"`
	NumRefFramesInPicOrderCntCycle  int        `json:"num_ref_frames_in_pic_order_cnt_cycle"`
	OffsetForRefFrame               []int      `json:"offset_for_ref_frame"`
	NumRefFrames                    int        `json:"num_ref_frames"`
	GapsInFrameNumValueAllowedFlag  int        `json:"gaps_in_frame_num_value_allowed_flag"`
	PicWidthInMbsMinus1             int        `json:"pic_width_in_mbs_minus1"`
	PicHeightInMapUnitsMinus1       int        `json:"pic_height_in_map_units_minus1"`
	FrameMbsOnlyFlag                int        `json:"frame_mbs_only_flag"`
	MbAdaptiveFrameFieldFlag        int        `json:"mb_adaptive_frame_field_flag"`
	Direct8x8InferenceFlag          int        `json:"direct_8x8_in_ference_flag"`
	FrameCropingFlag                int        `json:"frame_croping_flag"`
	FrameCropLeftOffset             int        `json:"frame_crop_left_offset"`
	FrameCropRightOffset            int        `json:"frame_crop_right_offset"`
	FrameCropTopOffset              int        `json:"frame_crop_top_offset"`
	FrameCropButtomOffset           int        `json:"frame_crop_buttom_offset"`
	VuiParametersPresentFlag        int        `json:"vui_parameters_present_flag"`
	//...
}

type PPS struct {
	PPSId                                 int `json:"pps_id"`
	SPSId                                 int `json:"sps_id"`
	EntropyCodingModeFlag                 int `json:"entropy_coding_mode_flag"`
	BottomFieldPicOrderInFramePresentFlag int `json:"bottom_field_pic_order_in_frame_present_flag"`
	NumSliceGroupsMinus1                  int `json:"num_slice_groups_minus1"`
	// ……H.264-AVC-ISO_IEC_14496-10-2012.pdf P-65
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
	nalu := decodeNalu(data)
	if nalu.NalUnitType != 7 {
		panic(fmt.Sprintf("nalu type not match %d", nalu.NalUnitType))
	}

	bs := core.NewBitStream(nalu.Rbsp)
	var sps SPS
	b, _ := bs.ReadByte()
	sps.ProfileIdc = int(b)

	b, _ = bs.ReadByte()
	x := int(b)
	sps.ConstraintSet0Flag = (x >> 7) & 0x01
	sps.ConstraintSet1Flag = (x >> 6) & 0x01
	sps.ConstraintSet2Flag = (x >> 5) & 0x01
	sps.ConstraintSet3Flag = (x >> 4) & 0x01
	sps.ConstraintSet4Flag = (x >> 3) & 0x01
	sps.ConstraintSet5Flag = (x >> 2) & 0x01

	b, _ = bs.ReadByte()
	sps.LevelIdc = int(b)
	sps.SPSId = golomb.ReadUEV(bs)

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
				bs.Next()
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

		sps.FrameMbsOnlyFlag = bs.Next()
		if sps.FrameMbsOnlyFlag == 0 {
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

	content, _ := json.MarshalIndent(sps, "", "\t")
	fmt.Printf("\tsps: \n%v\n", string(content))
}

func decodeNalu(data []byte) *Nalu {
	buffer := bytes.NewBuffer(data)
	nalu := &Nalu{}
	b, _ := buffer.ReadByte()
	n := int(b)

	nalu.ForbiddenBit = n >> 7
	nalu.NalReferenceIdc = (n >> 5) & 0x03
	nalu.NalUnitType = n & 0x1f

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

	nalu.RbspSize = n
	return nalu
}

func ParsePPS(data []byte) {
	nalu := decodeNalu(data)
	pps := PPS{}
	bs := core.NewBitStream(nalu.Rbsp)
	pps.PPSId = golomb.ReadUEV(bs)
	pps.SPSId = golomb.ReadUEV(bs)
	pps.EntropyCodingModeFlag = bs.Next()
	pps.BottomFieldPicOrderInFramePresentFlag = bs.Next()
	pps.NumSliceGroupsMinus1 = golomb.ReadUEV(bs)

	content, _ := json.MarshalIndent(pps, "", "\t")
	fmt.Printf("\tsps: \n%v\n", string(content))
}

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
