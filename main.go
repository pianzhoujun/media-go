package main

import (
	"flag"
	"media-go/core"
	"media-go/flow"
)

var (
	source = flag.String("i", "", "input file")
	action = flag.String("filter", "", "video|audio|meta")
)

func main() {
	flag.Parse()
	if len(*source) == 0 {
		flag.Usage()
		return
	}

	ctx := core.NewContext()
	ctx.Source = *source
	switch *action {
	case "video":
		ctx.Filter = core.Video
	case "audio":
		ctx.Filter = core.Audio
	case "meta":
		ctx.Filter = core.MetaData
	default:
		ctx.Filter = core.MetaData
	}

	flow.Run(ctx)
}
