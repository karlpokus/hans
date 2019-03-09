package hans

import (
	"os/exec"
)

type Watcher struct {
	Cmd     *exec.Cmd
	Running bool
	AppName string
	Restart chan string
}

type WatcherConf struct {
	Restart chan string
	App     *App
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

func (w *Watcher) Init(conf *WatcherConf) {
	// "fswatch", "-r", "--exclude", ".*", "--include", "\.go$", app.Watch
	w.Cmd = execCommand("fswatch", "-r", conf.App.Watch)
	w.Cmd.Dir = conf.App.Cwd
	w.Cmd.Stdout = w
	w.Restart = conf.Restart
	w.AppName = conf.App.Name
}

func (w Watcher) Write(p []byte) (int, error) {
	w.Restart <- w.AppName // check chan != nil?
	return len(p), nil
}

func (w *Watcher) Kill() {
	w.Running = false
	w.Cmd.Process.Kill()
}
