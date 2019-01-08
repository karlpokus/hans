package main

import (
	"log"
	"fmt"
	"os"
	"os/signal"
	"os/exec"
	"syscall"
)

type Hans struct {
	Stdout *log.Logger
	Stderr *log.Logger
	Apps []*App
}

func (hans *Hans) killAppsOnSignal(done chan<- bool) {
	sigs := make(chan os.Signal, 1) // signals are strings
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs // block

	// kill all running apps
	if len(hans.Apps) > 0 {
		for _, app := range hans.Apps {
			if app.Running {
				hans.Stdout.Printf("Killing %s", app.Name)
				if err := app.Cmd.Process.Kill(); err != nil {
					hans.Stderr.Printf("err from cmd.Process.Kill %s", err.Error())
				}
			}
			// TODO: remove app struct from arr
		}
	}
	done <- true
}

func getPath(p string) string {
	pwd, _ := os.Getwd()
	return pwd + p
}

func createApp(name, interval string) *App {
	return &App{
		Stdout: log.New(os.Stdout, fmt.Sprintf("%-7v", name), log.Ldate | log.Ltime),
		Cmd: exec.Command(getPath("/apps/dummy"), name, interval),
		Name: name,
		Running: false,
	}
}

func main() {
	hans := &Hans{
		Stdout: log.New(os.Stdout, fmt.Sprintf("%-7v", "hans"), log.Ldate | log.Ltime),
		Stderr: log.New(os.Stderr, fmt.Sprintf("%-7v", "hans"), log.Ldate | log.Ltime),
	}
	bixa := createApp("bixa", "5")
	bixa.Cmd.Stdout = bixa
	rex := createApp("rex", "3")
	rex.Cmd.Stdout = rex
	
	hans.Apps = []*App{bixa, rex}
	hans.Stdout.Print("hans start")
	for _, app := range hans.Apps {
		go app.Run(hans)
	}

	done := make(chan bool, 1)
	go hans.killAppsOnSignal(done)
	<-done
	hans.Stdout.Println("hans end")
}
