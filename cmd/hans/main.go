package main

import (
	"fmt"
	"flag"

	"hans"
)

var (
	v = flag.Bool("v", false, "toggle verbose logging")
	conf = flag.String("conf", "conf.yaml", "config file path")
	version = flag.Bool("version", false, "print version and exit")
	mem = flag.Bool("mem", false, "print mem usage periodically for all running apps")
)

func main() {
	flag.Parse()
	if *version {
    fmt.Println(hans.Version)
		return
  }
	h, err := hans.New(*conf, *v, *mem)
	if err != nil {
		panic(err)
	}
	err = h.Start()
	if err != nil {
		panic(err)
	}
	h.Wait()
}
