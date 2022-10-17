package flow

import (
	"media-go/core"
	"media-go/muxer/flv"
	"media-go/reader"
)

func Run(ctx *core.Context) {
	reader := reader.FileReader{Source: ctx.Source}
	if err := reader.Open(); err != nil {
		panic(err)
	}

	flv := flv.NewFLV(ctx)

	for !reader.Done() {
		reader.Read()
		flv.Decode(reader.Buffer.Bytes())
		reader.Buffer.Reset()
	}
}
