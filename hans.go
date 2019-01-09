package main

import (
	"log"
	"fmt"
	"os"
	"os/signal"
	"os/exec"
	"syscall"
	"gopkg.in/yaml.v2"
	"io/ioutil"
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

func (hans *Hans) getConf(path string) error {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(f, hans); err != nil {
		return err
	}
	return nil
}

func (hans *Hans) createApps() error {
	if err := hans.getConf("conf.yaml"); err != nil {
		return err
	}

	for _, app := range hans.Apps {
		app.Stdout = log.New(os.Stdout, fmt.Sprintf("%-7v", app.Name), log.Ldate | log.Ltime)
		app.Cmd = exec.Command(getPath(app.Bin), app.Args...)
		app.Cmd.Stdout = app
		go app.Run(hans)
	}
	return nil
}

func getPath(p string) string {
	pwd, _ := os.Getwd()
	return pwd + p
}

func main() {
	hans := &Hans{
		Stdout: log.New(os.Stdout, fmt.Sprintf("%-7v", "hans"), log.Ldate | log.Ltime),
		Stderr: log.New(os.Stderr, fmt.Sprintf("%-7v", "hans"), log.Ldate | log.Ltime),
	}
	if err := hans.createApps(); err != nil {
		hans.Stderr.Fatal()
	}
	hans.Stdout.Print("hans start")

	done := make(chan bool, 1)
	go hans.killAppsOnSignal(done)
	<-done
	hans.Stdout.Println("hans end")
}
