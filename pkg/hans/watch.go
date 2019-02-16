package hans

import (
	"os/exec"
)

type Watcher struct {
	Cmd     *exec.Cmd
	Running bool
	AppName string
	Ch      chan string
}

func (w *Watcher) Watch(fail chan error, restart chan string) {
	w.Ch = restart
	if err := w.Cmd.Start(); err != nil {
		fail <- err
		return
	}
	fail <- nil
	w.Running = true
	w.Cmd.Wait()
}

func (w Watcher) Write(p []byte) (int, error) {
	w.Ch <- w.AppName
	return len(p), nil
}

func (w *Watcher) kill() {
	w.Running = false
	w.Cmd.Process.Kill()
}
