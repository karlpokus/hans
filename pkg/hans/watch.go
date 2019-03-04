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

func (w *Watcher) Run(fail chan error) {
	err := w.Cmd.Start()
	fail <- err
	//close(fail)
	if err != nil {
		return
	}
	w.Running = true
	w.Cmd.Wait()
}

func (w *Watcher) Init(restart chan string, app *App) {
	// "fswatch", "-r", "--exclude", ".*", "--include", "\.go$", app.Watch
	w.Cmd = execCommand("fswatch", "-r", app.path(app.Watch))
	w.AppName = app.Name
	w.Cmd.Stdout = w
	w.Ch = restart
}

func (w Watcher) Write(p []byte) (int, error) {
	w.Ch <- w.AppName
	return len(p), nil
}

func (w *Watcher) Kill() {
	w.Running = false
	w.Cmd.Process.Kill()
}
