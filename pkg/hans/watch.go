package hans

import (
	"log"
	"os/exec"
)

type Watcher struct {
	Stdout  *log.Logger
	Cmd     *exec.Cmd
	Running bool
	AppName string
	Ch      chan string
}

func (w *Watcher) Watch(hans *Hans) {
	if err := w.Cmd.Start(); err != nil {
		hans.Stderr.Fatal(err)
		return
	}
	w.Running = true
	hans.Stdout.Printf("%s watcher started", w.AppName)

	if err := w.Cmd.Wait(); err != nil { // blocks and closes the pipe on cmd exit
		hans.Stderr.Printf("watcher watch wait err: %s", err.Error())
	}
}

func (w Watcher) Write(p []byte) (int, error) {
	w.Ch <- w.AppName
	return len(p), nil
}

func (w *Watcher) kill(hans *Hans, appName string) {
	hans.Stdout.Printf("killing %s watcher", appName)
	w.Running = false
	if err := w.Cmd.Process.Kill(); err != nil {
		hans.Stderr.Printf("killing %s err: %s", appName, err.Error())
	}
}
