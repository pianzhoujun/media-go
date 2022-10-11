package main

import "flag"

func main() {
	flag.Parse()
	if len(*inputFile) == 0 {
		flag.Usage()
		return
	}

	flv := NewFLV(*inputFile)
	flv.Parser.DoParse()
}
