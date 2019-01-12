package main

import (
	"log"
	"os/exec"
)

type Watcher struct {
	Cmd *exec.Cmd
	Running bool
}

func (w *Watcher) Watch(hans *Hans, name string) {
	if err := w.Cmd.Start(); err != nil {
		hans.Stderr.Fatal(err)
		return
	}
	w.Running = true
	hans.Stdout.Printf("%s watcher started", name)

	if err := w.Cmd.Wait(); err != nil { // blocks and closes the pipe on cmd exit
		hans.Stderr.Printf("watcher watch wait err: %s", err.Error())
	}
}

type App struct {
	Stdout *log.Logger
	Cmd *exec.Cmd
	Running bool
	Name string
	Bin string
	Args []string
	Watch string
	Watcher Watcher
}

func (app *App) Write(p []byte) (int, error) {
	app.Stdout.Print(string(p))
	return len(p), nil
}

func (app *App) Run(hans *Hans) {
	if err := app.Cmd.Start(); err != nil {
		hans.Stderr.Fatal(err)
		return
	}
	app.Running = true
	hans.Stdout.Printf("%s started", app.Name)

	if err := app.Cmd.Wait(); err != nil { // blocks and closes the pipe on cmd exit
		hans.Stderr.Printf("%s wait err: %s", app.Name, err.Error())
	}
}
