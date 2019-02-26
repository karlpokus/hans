package main

import (
	"github.com/karlpokus/hans/pkg/hans"
	"flag"
)

var v = flag.Bool("v", true, "toggle verbose logging")
var conf = flag.String("conf", "conf.yaml", "config file path")

func main() {
	flag.Parse()
	h, err := hans.New(*conf, *v)
	if err != nil {
		panic(err)
	}
	done, err := h.Start()
	if err != nil {
		panic(err)
	}
	<-done
}
