// usage: go run apps/dummy <name> <interval>
package main

import (
	"fmt"
	"time"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("ERR. Args missing. Exiting.")
		os.Exit(1)
	}
	t, _ := strconv.Atoi(os.Args[2])
	for {
		fmt.Printf("I am %s and I log every %d secs\n", os.Args[1], t)
		time.Sleep(time.Duration(t) * time.Second)
	}
}
