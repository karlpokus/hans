package main

import "github.com/karlpokus/hans/pkg/hans"

func main() {
	h, err := hans.New("conf.yaml")
	if err != nil {
		panic(err)
	}
	done, err := h.Start()
	if err != nil {
		panic(err)
	}
	<-done
}
