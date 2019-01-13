package main

import (
	"log"
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
	hans.Stdout.Print("hans start")
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

func NewHans() *Hans {
	return &Hans{
		Stdout: log.New(os.Stdout, formatName("hans"), log.Ldate | log.Ltime),
		Stderr: log.New(os.Stderr, formatName("hans"), log.Ldate | log.Ltime),
	}
}
