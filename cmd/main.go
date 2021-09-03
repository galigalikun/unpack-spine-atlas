package main

import (
	"flag"
	"log"

	unpackspineatlas "github.com/galigalikun/unpack-spine-atlas"
)

var (
	in  = flag.String("in", "skeleton.atlas.txt", "atlas file path")
	out = flag.String("out", "result", "output dir path")
)

func run() {
	parser, err := unpackspineatlas.NewParser(*in)
	if err != nil {
		panic(err)
	}
	defer parser.Close()
	atlas, err := parser.Parse()
	if err != nil {
		panic(err)
	}

	for _, a := range atlas {
		err := a.Unpack(*out)
		if err != nil {
			log.Printf("error:%s => %v", a.Image, err)
		}
	}
}

func init() {
	flag.Parse()
}

func main() {
	run()
}
