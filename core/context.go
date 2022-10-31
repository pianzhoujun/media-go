package core

type PktType int

var (
	Video    PktType = 0
	Audio    PktType = 1
	MetaData PktType = 2
	All      PktType = 2
)

type VideoFrame struct {
}

type Packet struct {
	Type PktType
	Data interface{}
}

type PktCallback func(ctx *Context, pkt *Packet) interface{}
type FrameCallback func(ctx *Context, frame interface{}) interface{}
type StreamReader func(ctx *Context) ([]byte, error)

type Context struct {
	Source  string
	Filter  PktType
	PktCb   PktCallback
	FrameCb FrameCallback
	SR      StreamReader
	Done    bool
}

func NewContext() *Context {
	return &Context{}
}

func (ctx *Context) SetFrameCallback(cb FrameCallback)   { ctx.FrameCb = cb }
func (ctx *Context) SetPktCallback(cb PktCallback)       { ctx.PktCb = cb }
func (ctx *Context) SetStreamReader(reader StreamReader) { ctx.SR = reader }
