package hans

import (
	"log"
	"os"
	"os/exec"
)

type App struct {
	Stdout  *log.Logger
	Cmd     *exec.Cmd
	Running bool
	Name    string
	Bin     string
	Watch   string
	Build   string
	Watcher *Watcher
	Cwd     string
}

func (app *App) run(hans *Hans) {
	if err := app.Cmd.Start(); err != nil {
		hans.Stderr.Fatal(err)
		return
	}
	app.Running = true
	hans.Stdout.Printf("%s started", app.Name)

	if err := app.Cmd.Wait(); err != nil { // blocks and closes the pipe on cmd exit
		hans.Stderr.Printf("%s wait err: %s", app.Name, err.Error())
	}
}

func (app *App) path(p string) string {
	if len(app.Cwd) > 0 {
		return app.Cwd + p
	}
	return p
}

func (app *App) start(hans *Hans) {
	app.Cwd = hans.Cwd
	app.Stdout = log.New(os.Stdout, formatName(app.Name), log.Ldate|log.Ltime)
	cmd, args := splitBin(app.Bin)
	app.Cmd = exec.Command(app.path(cmd), args...)
	app.Cmd.Stdout = app
	go app.run(hans)

	if len(app.Watch) > 0 {
		// "fswatch", "-r", "--exclude", ".*", "--include", "\.go$", app.Watch
		app.Watcher = &Watcher{
			Stdout:  log.New(os.Stdout, formatName("watcher"), log.Ldate|log.Ltime),
			Cmd:     exec.Command("fswatch", "-r", app.path(app.Watch)),
			AppName: app.Name,
		}
		app.Watcher.Cmd.Stdout = app.Watcher
		c := make(chan string, 1)
		app.Watcher.Ch = c
		go hans.restart(c)
		go app.Watcher.Watch(hans)
	} else {
		app.Watcher = &Watcher{}
	}
}

func (app *App) restart(hans *Hans) {
	cmd, args := splitBin(app.Bin)
	app.Cmd = exec.Command(app.path(cmd), args...)
	app.Cmd.Stdout = app
	go app.run(hans)
}

func (app *App) kill(hans *Hans) {
	hans.Stdout.Printf("killing %s", app.Name)
	app.Running = false
	if err := app.Cmd.Process.Kill(); err != nil {
		hans.Stderr.Printf("killing %s err: %s", app.Name, err.Error())
	}
}

func (app *App) build(hans *Hans, done chan<- bool) {
	cmd, args := splitBin(app.Build)
	out, err := exec.Command(cmd, args...).CombinedOutput() // includes run
	if err != nil {
		hans.Stderr.Print(err)
	}
	if len(out) > 0 {
		hans.Stderr.Print(out)
	} else {
		hans.Stdout.Printf("%s rebuilt", app.Name)
	}
	done <- true
}

func (app *App) Write(p []byte) (int, error) {
	app.Stdout.Print(string(p))
	return len(p), nil
}
