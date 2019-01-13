package main

import (
	"os"
	"fmt"
)

func absPath(p string) string {
	pwd, _ := os.Getwd()
	return pwd + p
}

func formatName(name string) string {
	const maxChars int = 9
	if len(name) >= maxChars {
		return name[:9] + " "
	}
	return fmt.Sprintf("%-10v", name)
}

func main() {
	hans := NewHans()
	err := hans.getConf("conf.yaml")
	if err != nil {
		hans.Stderr.Fatal(err)
		return
	}
	err = hans.createApps()
	if err != nil {
		hans.Stderr.Fatal(err)
		return
	}
	done := make(chan bool, 1)
	go hans.killAppsOnSignal(done)
	<-done
	hans.Stdout.Println("hans end")
}
