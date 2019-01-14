// usage: <cmd> <name> <interval>
package main

import (
	"fmt"
	"time"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "ERR: Args missing. Exiting.")
		os.Exit(1)
	}
	name := os.Args[1]
	interval, _ := strconv.Atoi(os.Args[2])
	for {
		fmt.Fprintf(os.Stdout, "This is %s logging every %d secs\n", name, interval)
		fmt.Fprintf(os.Stderr, "This is %s reporting an error\n", name)
		time.Sleep(time.Duration(interval) * time.Second)
	}
}
