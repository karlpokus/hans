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
				hans.Stdout.Printf("killing %s", app.Name)
				if err := app.Cmd.Process.Kill(); err != nil {
					hans.Stderr.Printf("killing %s err: %s", app.Name, err.Error())
				}
			}
			// TODO:
			// - remove app struct from arr
			// x end watcher
			if app.Watcher.Running {
				hans.Stdout.Printf("killing %s watcher", app.Name)
				if err := app.Watcher.Cmd.Process.Kill(); err != nil {
					hans.Stderr.Printf("killing %s err: %s", app.Name, err.Error())
				}
			}
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
	// TODO: validate fields
	return nil
}

func (hans *Hans) createApps() error {
	if err := hans.getConf("conf.yaml"); err != nil {
		return err
	}

	for _, app := range hans.Apps {
		app.Stdout = log.New(os.Stdout, formatName(app.Name), log.Ldate | log.Ltime)
		app.Cmd = exec.Command(absPath(app.Bin), app.Args...)
		app.Cmd.Stdout = app
		go app.Run(hans)

		if len(app.Watch) > 0 {
			// "fswatch", "-r", "--exclude", ".*", "--include", "\.go$", app.Watch
			app.Watcher.Cmd = exec.Command("fswatch", "-r", absPath(app.Watch))
			app.Watcher.Cmd.Stdout = app
				go app.Watcher.Watch(hans, app.Name)
		}
	}
	return nil
}

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
	hans := &Hans{
		Stdout: log.New(os.Stdout, formatName("hans"), log.Ldate | log.Ltime),
		Stderr: log.New(os.Stderr, formatName("hans"), log.Ldate | log.Ltime),
	}
	hans.Stdout.Print("hans start")

	if err := hans.createApps(); err != nil {
		hans.Stderr.Fatal()
	}
	done := make(chan bool, 1)
	go hans.killAppsOnSignal(done)
	<-done
	hans.Stdout.Println("hans end")
}
