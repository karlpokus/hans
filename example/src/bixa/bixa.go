// usage: <cmd> <name> <interval>
package main

import (
	"fmt"
	"time"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Arg missing. Exiting.")
		os.Exit(1)
	}
	interval, _ := strconv.Atoi(os.Args[1])
	for {
		fmt.Fprintf(os.Stdout, "reporting every %d secs", interval)
		fmt.Fprintf(os.Stderr, "reporting an error")
		time.Sleep(time.Duration(interval) * time.Second)
	}
}
