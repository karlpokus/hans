package hans

import (
	"log"
	"os"
	"os/exec"
	"time"
)

var execCommand = exec.Command

type App struct {
	Stdout  *log.Logger
	Stderr	*log.Logger
	Cmd     *exec.Cmd
	Running bool
	Name    string
	Bin     string
	Watch   string
	Build   string
	Watcher *Watcher
	Cwd     string
}

func (app *App) run(fail chan error) {
	if err := app.Cmd.Start(); err != nil {
		fail <- err
		return
	}
	fail <- nil
	app.Running = true
	app.Cmd.Wait() // blocks and closes the pipe on cmd exit
}

func (app *App) path(p string) string {
	if len(app.Cwd) > 0 {
		return app.Cwd + p
	}
	return p
}

func (app *App) init(cwd string) {
	app.Cwd = cwd
	app.Stdout = log.New(os.Stdout, formatName(app.Name), log.Ldate|log.Ltime)
	app.Stderr = log.New(os.Stderr, formatName(app.Name), log.Ldate|log.Ltime)
	cmd, args := splitBin(app.Bin)
	app.Cmd = execCommand(app.path(cmd), args...)
	app.Cmd.Stdout = app
	app.Watcher = &Watcher{}
	if len(app.Watch) > 0 {
		// "fswatch", "-r", "--exclude", ".*", "--include", "\.go$", app.Watch
		app.Watcher.Cmd = execCommand("fswatch", "-r", app.path(app.Watch))
		app.Watcher.AppName = app.Name
		app.Watcher.Cmd.Stdout = app.Watcher
	}
}

/*func (app *App) restart() {
	cmd, args := splitBin(app.Bin)
	app.Cmd = execCommand(app.path(cmd), args...)
	app.Cmd.Stdout = app
	go app.run()
}*/

func (app *App) kill() {
	app.Running = false
	app.Cmd.Process.Kill()
}

func (app *App) build(done chan<- bool) {
	cmd, args := splitBin(app.Build)
	out, err := execCommand(cmd, args...).CombinedOutput() // includes run
	if err != nil {
		app.Stderr.Print(err) // TODO: don't restart if build failed
	}
	if len(out) > 0 {
		app.Stderr.Print(string(out))
	} else {
		app.Stdout.Print("build done")
	}
	done <- true
}

func (app *App) Write(p []byte) (int, error) {
	app.Stdout.Print(string(p))
	return len(p), nil
}
