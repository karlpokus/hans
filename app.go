package main

import (
	"log"
	"os/exec"
)

type App struct {
	Stdout *log.Logger
	Cmd *exec.Cmd
	Running bool
	Name string
	Bin string
	Args []string
}

func (app *App) Write(p []byte) (int, error) {
	app.Stdout.Print(string(p))
	return len(p), nil
}

func (app *App) Run(hans *Hans) {
	if err := app.Cmd.Start(); err != nil {
		hans.Stderr.Fatal(err)
	}
	app.Running = true

	if err := app.Cmd.Wait(); err != nil { // blocks and closes the pipe on cmd exit
		hans.Stderr.Printf("Wait done %s", err.Error())
	}
	hans.Stderr.Printf("ProcessState %v", app.Cmd.ProcessState)
}
