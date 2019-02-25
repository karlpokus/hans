package hans

import (
	"bytes"
	"log"
	"os"
	"os/exec"
)

var execCommand = exec.Command

type App struct {
	Stdout    *log.Logger
	Stderr    *log.Logger
	Cmd       *exec.Cmd
	Running   bool
	Name      string
	Bin       string
	Watch     string
	Build     string
	Watcher   *Watcher
	Cwd       string
	StdoutBuf bytes.Buffer
	StderrBuf bytes.Buffer
}

func (app *App) run(fail chan error) {
	err := app.Cmd.Start()
	fail <- err
	if err != nil {
		return
	}
	app.Running = true
	app.Cmd.Wait() // blocks and closes the pipe on cmd exit
}

func (app *App) path(p string) string {
	if len(app.Cwd) > 0 {
		return app.Cwd + p
	}
	return p
}

// setLogging sets the logging for hans based on RUNTIME env
func (app *App) setLogging() {
	if RUNTIME == "TEST" {
		app.Stdout = log.New(&app.StdoutBuf, "", 0)
		app.Stderr = log.New(&app.StderrBuf, "", 0)
	} else {
		app.Stdout = log.New(os.Stdout, formatName(app.Name), log.Ldate|log.Ltime)
		app.Stderr = log.New(os.Stderr, formatName(app.Name), log.Ldate|log.Ltime)
	}
}

func (app *App) init(cwd string) {
	app.Cwd = cwd
	app.setLogging()
	app.setCmd()
	app.Watcher = &Watcher{}
	if len(app.Watch) > 0 {
		// "fswatch", "-r", "--exclude", ".*", "--include", "\.go$", app.Watch
		app.Watcher.Cmd = execCommand("fswatch", "-r", app.path(app.Watch))
		app.Watcher.AppName = app.Name
		app.Watcher.Cmd.Stdout = app.Watcher
	}
}

func (app *App) setCmd() {
	cmd, args := splitBin(app.Bin)
	app.Cmd = execCommand(app.path(cmd), args...)
	app.Cmd.Stdout = app
}

func (app *App) restart(fail chan error) {
	app.setCmd()
	go app.run(fail)
}

func (app *App) kill() { // TODO: check return value of Kill()
	app.Running = false
	app.Cmd.Process.Kill()
}

func (app *App) build() ([]byte, error) {
	cmd, args := splitBin(app.Build)
	out, err := execCommand(cmd, args...).CombinedOutput() // includes run
	if err != nil {
		return out, err
	}
	return out, nil
}

func (app *App) Write(p []byte) (int, error) {
	app.Stdout.Print(string(p))
	return len(p), nil
}
