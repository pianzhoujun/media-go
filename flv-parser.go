package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"media-go/h264"

	"os"
	"strconv"

	"github.com/torresjeff/rtmp/amf/amf0"
)

var (
	inputFile = flag.String("input_file", "", "input file")
)

const (
	PHEADER       = 0
	PBODY         = 1
	PTagHeader    = 2
	PTagBody      = 3
	HeaderSize    = 9
	TagHeaderSize = 11
)

const (
	Audio    = 0x08
	Video    = 0x09
	MetaData = 0x12
)

const (
	FLV_CODECID_PCM                   = 0
	FLV_CODECID_ADPCM                 = 1
	FLV_CODECID_MP3                   = 2
	FLV_CODECID_PCM_LE                = 3
	FLV_CODECID_NELLYMOSER_16KHZ_MONO = 4
	FLV_CODECID_NELLYMOSER_8KHZ_MONO  = 5
	FLV_CODECID_NELLYMOSER            = 6
	FLV_CODECID_PCM_ALAW              = 7
	FLV_CODECID_PCM_MULAW             = 8
	FLV_CODECID_AAC                   = 10
	FLV_CODECID_SPEEX                 = 11
)

var AudioCodec = map[int]string{
	FLV_CODECID_PCM:                   "pcm",
	FLV_CODECID_ADPCM:                 "adpcm",
	FLV_CODECID_MP3:                   "mp3",
	FLV_CODECID_PCM_LE:                "pcm_le",
	FLV_CODECID_NELLYMOSER_16KHZ_MONO: "nellymoser_16khz_mono",
	FLV_CODECID_NELLYMOSER_8KHZ_MONO:  "8khz_mono",
	FLV_CODECID_NELLYMOSER:            "nellymoser",
	FLV_CODECID_PCM_ALAW:              "pcm_alaw",
	FLV_CODECID_PCM_MULAW:             "pcm_mulaw",
	FLV_CODECID_AAC:                   "aac",
	FLV_CODECID_SPEEX:                 "speex",
}

var AudioSampleRate = map[int]string{
	0x00: "5.5kHz",
	0x01: "11kHz",
	0x02: "22kHz",
	0x03: "44kHz",
}

const (
	FLV_FRAME_KEY            = 1 //<< FLV_VIDEO_FRAMETYPE_OFFSET ///< key frame (for AVC, a seekable frame)
	FLV_FRAME_INTER          = 2 //<< FLV_VIDEO_FRAMETYPE_OFFSET ///< inter frame (for AVC, a non-seekable frame)
	FLV_FRAME_DISP_INTER     = 3 //<< FLV_VIDEO_FRAMETYPE_OFFSET ///< disposable inter frame (H.263 only)
	FLV_FRAME_GENERATED_KEY  = 4 //<< FLV_VIDEO_FRAMETYPE_OFFSET ///< generated key frame (reserved for server use only)
	FLV_FRAME_VIDEO_INFO_CMD = 5 //<< FLV_VIDEO_FRAMETYPE_OFFSET ///< video info/command frame
)

const (
	FLV_CODECID_H263     = 2
	FLV_CODECID_SCREEN   = 3
	FLV_CODECID_VP6      = 4
	FLV_CODECID_VP6A     = 5
	FLV_CODECID_SCREEN2  = 6
	FLV_CODECID_H264     = 7
	FLV_CODECID_REALH263 = 8
	FLV_CODECID_MPEG4    = 9
	FLV_CODECID_H265     = 12
)

var VideoFrameType = map[int]string{
	FLV_FRAME_KEY:            "key",
	FLV_FRAME_INTER:          "inter",
	FLV_FRAME_DISP_INTER:     "disp inter",
	FLV_FRAME_GENERATED_KEY:  "generated key",
	FLV_FRAME_VIDEO_INFO_CMD: "info cmd",
}

var VideoCodecMap = map[int]string{
	FLV_CODECID_H263:     "h263",
	FLV_CODECID_SCREEN:   "screen",
	FLV_CODECID_VP6:      "vp6",
	FLV_CODECID_VP6A:     "vp6a",
	FLV_CODECID_SCREEN2:  "screen2",
	FLV_CODECID_H264:     "h264",
	FLV_CODECID_REALH263: "realh263",
	FLV_CODECID_MPEG4:    "mpeg4",
	FLV_CODECID_H265:     "h265",
}

type FlvParser struct {
	InputFile string
	Status    int
	TagStatus int
	buffer    bytes.Buffer
	flv       *FLV
}

type FLVHeader struct {
	Magic      [3]byte
	Version    byte
	Flags      byte
	HeaderSize int32
}

type PacketType int

type Packet struct {
	Type PacketType
	Data interface{}
}

type PacketMetaData map[string]interface{}

type PacketVideo struct {
	Type  int
	Codec int
}

type PacketAudio struct {
}

func (h *FLVHeader) echo() {
	fmt.Println("Header:")
	fmt.Printf("\tmagic:		%v%s%s\n", string(h.Magic[0]), string(h.Magic[1]), string(h.Magic[2]))
	fmt.Printf("\tversion:  	%d\n", int(h.Version))
	fmt.Printf("\tflags:		%x\n", h.Flags)
	fmt.Printf("\theadersize:	%d\n", h.HeaderSize)
}

type TagHeader struct {
	Type        int8
	DataSize    int
	Timestamp   int
	TimestampEx int
	StreamId    int
}

func (h *TagHeader) String() string {
	types := ""
	switch h.Type {
	case 0x08:
		types = "audio"
	case 0x09:
		types = "video"
	case 0x12:
		types = "metadata"
	default:
		types = strconv.Itoa(int(h.Type))
	}

	str := fmt.Sprintf("%10s(0x%02x)|data size: %10d|dts: %10d",
		types, h.Type, h.DataSize, h.Timestamp|(h.TimestampEx<<24))
	return str
}

type FLVTag struct {
	Header TagHeader
	Data   []byte
}

type FLVBody struct {
	PreviousTagSize uint32
	Tag             FLVTag
}

type FLV struct {
	Header *FLVHeader
	Body   *FLVBody

	Parser FlvParser
}

func NewFLV(input string) *FLV {
	flv := &FLV{Header: &FLVHeader{},
		Body:   nil,
		Parser: FlvParser{InputFile: input}}
	flv.Parser.flv = flv
	return flv
}

func (fp *FlvParser) DoParse() {
	fd, err := os.Open(fp.InputFile)
	if err != nil {
		panic(err)
	}

	defer fd.Close()

	buffer := make([]byte, 1024)

	for {
		size, err := fd.Read(buffer)
		if err == io.EOF || size < 0 {
			fmt.Println("EOF")
			break
		}

		fp.buffer.Write(buffer[:size])
		fp.parseNext()
	}
}

func (fp *FlvParser) parseNext() {
	if fp.Status == PHEADER {
		fp.parseHeader()
	}

	if fp.Status == PBODY {
		fp.parseBody()
	}
}

func (fp *FlvParser) parseHeader() {
	if fp.buffer.Len() < HeaderSize {
		return
	}

	header := fp.buffer.Next(HeaderSize)
	var i int
	for ; i < 3; i++ {
		fp.flv.Header.Magic[i] = header[i]
	}

	fp.flv.Header.Version = header[i]
	i++

	fp.flv.Header.Flags = header[i]
	i++

	fp.flv.Header.HeaderSize = int32(binary.BigEndian.Uint32(header[i:]))

	fp.flv.Header.echo()
	fp.Status = PBODY
	fp.TagStatus = PTagHeader
}

func (fp *FlvParser) parseBody() {
	if fp.TagStatus == PTagHeader {
		fp.parseTagHeader()
	}

	if fp.TagStatus == PTagBody {
		fp.parseTagBody()
	}
}

func bytesToInt64(bs []byte) uint64 {
	var x uint64
	for _, b := range bs {
		x = (x << 8) | uint64(b)
	}

	return x
}

func (fp *FlvParser) parseTagHeader() {
	if fp.buffer.Len() < TagHeaderSize+4 {
		return
	}

	body := &FLVBody{}

	body.PreviousTagSize = uint32(bytesToInt64(fp.buffer.Next(4)))

	body.Tag.Header.Type = int8(bytesToInt64(fp.buffer.Next(1)))
	body.Tag.Header.DataSize = int(bytesToInt64(fp.buffer.Next(3)))
	body.Tag.Header.Timestamp = int(bytesToInt64(fp.buffer.Next(3)))
	body.Tag.Header.TimestampEx = int(bytesToInt64(fp.buffer.Next(1)))
	body.Tag.Header.StreamId = int(bytesToInt64(fp.buffer.Next(3)))

	fp.flv.Body = body
	fp.TagStatus = PTagBody

	fmt.Print(fp.flv.Body.Tag.Header.String())
}

func (fp *FlvParser) parseTagBody() {
	if fp.buffer.Len() < fp.flv.Body.Tag.Header.DataSize {
		return
	}

	fp.flv.Body.Tag.Data = fp.buffer.Next(fp.flv.Body.Tag.Header.DataSize)
	fp.TagStatus = PTagHeader

	fp.readPacket()
}

func (fp *FlvParser) readPacket() {
	if fp.flv.Body == nil {
		return
	}

	switch fp.flv.Body.Tag.Header.Type {
	case MetaData:
		fp.readMetadata()
	case Video:
		fp.readVideo()
	case Audio:
		fp.readAudio()
	}

	fmt.Println()
}

func (fp *FlvParser) readMetadata() {
	packet := fp.flv.Body.Tag.Data
	pos := 0

	for {
		content, err := amf0.Decode(packet[pos:])
		if err != nil {
			panic(err)
		}

		pos += int(amf0.Size(content))
		switch content.(type) {
		case amf0.ECMAArray:
			for key, value := range content.(amf0.ECMAArray) {
				fmt.Printf("%s : %v\n", key, value)
			}
		default:
			fmt.Printf("%v\n", content)
		}

		if pos >= len(packet) {
			break
		}
	}
}

func (fp *FlvParser) readVideo() {
	packet := fp.flv.Body.Tag.Data
	flag := int(packet[0])
	ft := VideoFrameType[flag>>4]
	codec := VideoCodecMap[flag&0x0f]
	fmt.Printf("|%s|%s", ft, codec)

	if flag&0x0f == FLV_CODECID_H264 {
		buffer := bytes.NewBuffer(packet[1:])
		h264.Parse(buffer)
	}
}

func (fp *FlvParser) readAudio() {
	packet := fp.flv.Body.Tag.Data
	flag := int(packet[0])

	codec := AudioCodec[flag>>4]
	sampleRate := AudioSampleRate[(flag&0x0f)>>2]

	var accuracy string
	var audioType string

	if flag&0x02 == 1 {
		accuracy = "8bits"
	} else {
		accuracy = "16bits"
	}

	if flag&0x01 == 1 {
		audioType = "sndMono"
	} else {
		audioType = "sndStereo"
	}

	fmt.Printf("|%s-%s-%s-%s", codec, sampleRate, accuracy, audioType)
}
