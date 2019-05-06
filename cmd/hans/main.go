package main

import (
	"github.com/karlpokus/hans"
	"flag"
)

var v = flag.Bool("v", false, "toggle verbose logging")
var conf = flag.String("conf", "conf.yaml", "config file path")

func main() {
	flag.Parse()
	h, err := hans.New(*conf, *v)
	if err != nil {
		panic(err)
	}
	err = h.Start()
	if err != nil {
		panic(err)
	}
	h.Wait()
}
