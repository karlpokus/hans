package main

import (
	"fmt"
	"flag"

	"github.com/karlpokus/hans"
)

var (
	v = flag.Bool("v", false, "toggle verbose logging")
	conf = flag.String("conf", "conf.yaml", "config file path")
	version = flag.Bool("version", false, "print version and exit")
)

func main() {
	flag.Parse()
	if *version {
    fmt.Println(hans.Version)
		return
  }
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
