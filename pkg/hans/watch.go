package hans

import (
	"log"
	"os/exec"
)

type Watcher struct {
	Cmd     *exec.Cmd
	Running bool
	AppName string
	Ch      chan string
}

func (w *Watcher) Watch(c chan string, stdout, stderr *log.Logger) {
	w.Ch = c
	if err := w.Cmd.Start(); err != nil {
		stderr.Print(err)
		return
	}
	w.Running = true
	stdout.Print("watcher running")
	w.Cmd.Wait()
}

func (w Watcher) Write(p []byte) (int, error) {
	w.Ch <- w.AppName
	return len(p), nil
}

func (w *Watcher) kill(stdout *log.Logger) {
	w.Running = false
	w.Cmd.Process.Kill()
	stdout.Print("watcher terminated")
}
