package hans

import (
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/fatih/color"
)

var execCommand = exec.Command

type LogWriter struct {
	Out *log.Logger
}

func (w *LogWriter) Write(b []byte) (int, error) {
	w.Out.Print(string(b))
	return len(b), nil
}

type App struct {
	Stdout  *LogWriter
	Stderr  *LogWriter
	Cmd     *exec.Cmd
	Running bool
	Name    string
	Bin     string
	Watch   string
	Build   string
	Cwd     string
	*Watcher
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
	app.Cmd.Wait() // blocks and closes the io pipe on cmd exit
}

// setLogging sets prefix, flags and io.Writer for the app loggers
func (app *App) setLogging(conf *AppConf) {
	if conf.StdoutWriter != nil {
		app.Stdout.Out = log.New(conf.StdoutWriter, "", 0)
	} else {
		app.Stdout.Out = log.New(os.Stdout, formatName(app.Name, color.GreenString), log.Ldate|log.Ltime)
	}
	if conf.StderrWriter != nil {
		app.Stderr.Out = log.New(conf.StderrWriter, "", 0)
	} else {
		app.Stderr.Out = log.New(os.Stderr, formatName(app.Name, color.RedString), log.Ldate|log.Ltime)
	}
}

// init prepares an app to be run later
func (app *App) Init(conf *AppConf) {
	app.Cwd = conf.Cwd
	app.Stdout = &LogWriter{}
	app.Stderr = &LogWriter{}
	app.setLogging(conf)
	app.setCmd()
	app.Watcher = &Watcher{}
}

func (app *App) setCmd() {
	cmd, args := splitBin(app.Bin)
	app.Cmd = execCommand(cmd, args...)
	app.Cmd.Stdout = app.Stdout
	app.Cmd.Stderr = app.Stderr
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
