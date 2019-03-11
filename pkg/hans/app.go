package hans

import (
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/fatih/color"
)

var execCommand = exec.Command

type App struct {
	Stdout  *log.Logger
	Stderr  *log.Logger
	Cmd     *exec.Cmd
	Running bool
	Name    string
	Bin     string
	Watch   string
	Build   string
	*Watcher
	Cwd string
}

type AppConf struct {
	StdoutWriter io.Writer
	StderrWriter io.Writer
	Cwd          string
}

func (app *App) Run(fail chan error) {
	err := app.Cmd.Start()
	fail <- err
	if err != nil {
		return
	}
	app.Running = true
	app.Cmd.Wait() // blocks and closes the pipe on cmd exit
}

// setLogging sets the logging for the app
func (app *App) setLogging(conf *AppConf) {
	if conf.StdoutWriter != nil {
		app.Stdout = log.New(conf.StdoutWriter, "", 0)
	} else {
		app.Stdout = log.New(os.Stdout, formatName(app.Name, color.GreenString), log.Ldate|log.Ltime)
	}
	if conf.StderrWriter != nil {
		app.Stdout = log.New(conf.StderrWriter, "", 0)
	} else {
		app.Stderr = log.New(os.Stderr, formatName(app.Name, color.RedString), log.Ldate|log.Ltime)
	}
}

// init prepares an app to be run later
func (app *App) Init(conf *AppConf) {
	app.Cwd = conf.Cwd
	app.setLogging(conf)
	app.setCmd()
	app.Watcher = &Watcher{}
}

func (app *App) setCmd() {
	cmd, args := splitBin(app.Bin)
	app.Cmd = execCommand(cmd, args...)
	app.Cmd.Stdout = app
	app.Cmd.Dir = app.Cwd
}

func (app *App) Kill() {
	app.Running = false
	app.Cmd.Process.Kill()
}

func (app *App) build() ([]byte, error) {
	cmd, args := splitBin(app.Build)
	Cmd := execCommand(cmd, args...)
	Cmd.Dir = app.Cwd
	return Cmd.CombinedOutput() // includes run
}

func (app *App) Write(p []byte) (int, error) {
	app.Stdout.Print(string(p))
	return len(p), nil
}
