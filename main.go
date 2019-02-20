package main

import (
	"github.com/karlpokus/hans/pkg/hans"
	"os"
	"fmt"
)

func main() {
	conf := os.Args[1]
	h, err := hans.New(conf)
	if err != nil {
		panic(err)
	}
	done, err := h.Start()
	if err != nil {
		panic(err)
	}
	<-done
}
