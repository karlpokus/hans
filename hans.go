package hans

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"io"

	"github.com/fatih/color"
	"gopkg.in/yaml.v2"
)

var Version = "vX.Y.Z" // injected at build time

type Opt struct {
	Cwd string
	TTL string
}

type Hans struct {
	//Stdout *log.Logger
	//Stderr *log.Logger
	StdoutWriter io.Writer
	StderrWriter io.Writer
	Apps   []*App
	Opts   Opt
	TTL    time.Duration
	Runc   chan Child
}

type Child interface {
	Run(chan error)
	Kill()
	RunningState(bool)
	GetName() string
}

var fd1 = log.New(os.Stdout, "", log.Ldate|log.Ltime)
var fd2 = log.New(os.Stderr, "", log.Ldate|log.Ltime)

type Logger interface {
	Setlog(bool) (string, io.Writer, func(string, ...interface{}) string)
}

// Stdout is a global logger and formats-, and outputs to whatever the passed Logger returns from Setlog
func Stdout(lgr Logger, format string, v ...interface{}) {
	name, w, colorFunc := lgr.Setlog(false)
	fd1.SetPrefix(formatName(name, colorFunc))
	fd1.SetOutput(w)
	fd1.Printf(format, v...)
}

// Stderr is a global logger and formats-, and outputs to whatever the passed Logger returns from Setlog
func Stderr(lgr Logger, format string, v ...interface{}) {
	name, w, colorFunc := lgr.Setlog(true)
	fd2.SetPrefix(formatName(name, colorFunc))
	fd2.SetOutput(w)
	fd2.Printf(format, v...)
}

func (hans Hans) interrupt() bool {
	anyRunning := func() (*App, bool) {
		for _, app := range hans.Apps {
			if app.Running() {
				return app, true
			}
		}
		return nil, false
	}
	app, ok := anyRunning()
	if ok {
		app.Cmd.Process.Signal(os.Interrupt)
	}
	return ok
}

// cleanup terminates running apps and associated watchers by first sending SIGINT to pgid,
// to allow for graceful exits, then SIGKILL to any still running apps after a short timeout.
// Logs will show `terminated` for apps that respond to SIGINT and `killed` for
// those that don't
func (hans *Hans) cleanup() {
	mod := "[CLEANUP]"
	Stdout(hans, "%s start", mod)
	ok := hans.interrupt() // SIGINT is trappable.
	if ok {
		Stdout(hans, "%s pgid recieved SIGINT", mod)
		time.Sleep(2 * time.Second)
	}
	for _, app := range hans.Apps {
		if app.Running() {
			hans.kill(app) // SIGKILL is not trappable
			Stdout(hans, "%s %s killed", mod, app.Name)
		}
		if app.Watcher.Running() {
			hans.kill(app.Watcher)
			Stdout(hans, "%s %s watcher closed", mod, app.Name)
		}
	}
	Stdout(hans, "%s done", mod)
}

// kill kills a child and toggles running state
func (hans *Hans) kill(c Child) {
	c.Kill()
	c.RunningState(false)
}

// run runs a child and toggles running state
func (hans *Hans) run() {
	mod := "[RUN]"
	for c := range hans.Runc {
		fail := make(chan error)
		go c.Run(fail)

		select {
		case <-time.After(hans.TTL):
			Stderr(hans, "%s %s timeout", mod, c.GetName())
		case err := <-fail:
			if err != nil {
				Stderr(hans, "%s %s failed start attempt: %v", mod, c.GetName(), err)
				break
			}
			c.RunningState(true)
			Stdout(hans, "%s %s started", mod, c.GetName())
		}
	}
}

func (hans *Hans) manager(manc chan *App, runc chan Child) {
	mod := "[MANAGER]"
	for app := range manc {
		code := app.Cmd.ProcessState.ExitCode()
		switch {
		case code == -1:
			Stdout(hans, "%s %s terminated", mod, app.Name)
		case code == 0:
			Stdout(hans, "%s %s exited", mod, app.Name)
		case code > 0:
			Stderr(hans, "%s %s exited %d", mod, app.Name, code)
			// bad exit dance
			app.BadExit.Init()
			app.BadExit.Inc()
			if app.BadExit.MaxReached() {
				if app.BadExit.WithinWindow() {
					Stderr(hans, "%s maxBadExits reached. %s is dead", mod, app.Name)
					break
				}
				app.BadExit.Reset()
			}
			Stdout(hans, "%s restarting %s", mod, app.Name)
			app.SetCmd()
			runc <- app
		}
	}
}

// Wait blocks on the done chan until an os.Signal is triggered
// runs cleanup before returning
func (hans *Hans) Wait() {
	done := make(chan bool, 1)
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
		done <- true
	}()
	<-done
	hans.cleanup()
}

// Start starts all apps and associated watchers
func (hans *Hans) Start() error {
	for _, app := range hans.Apps {
		hans.Runc <- app
		if app.Watch != "" {
			hans.Runc <- app.Watcher
		}
	}
	return nil
}

// build builds an app and sends it off for restarting if successful
func (hans *Hans) build(buildc chan *App, runc chan Child) {
	mod := "[BUILD]"
	for app := range buildc {
		Stdout(hans, "%s %s src change detected. Attempting build and restart", mod, app.Name)
		if app.Build == "" {
			Stderr(hans, "%s %s build cmd missing. Build aborted. Attempting restart", mod, app.Name)
			hans.kill(app) // app is running during build
			app.SetCmd()
			runc <- app
			continue
		}
		res, err := app.build() // TODO: let app.build check Build string
		if err != nil {
			Stderr(hans, "%s %s build failed: %v", mod, app.Name, err)
			Stderr(hans, "%s %s", mod, res)
			Stderr(hans, "%s restart attempt aborted", mod)
			continue
		}
		Stdout(hans, "%s %s build successful. Attempting restart", mod, app.Name)
		hans.kill(app) // app is running during build
		app.SetCmd()
		runc <- app
	}
}

func (hans *Hans) Setlog(isErr bool) (string, io.Writer, func(string, ...interface{}) string) {
	if isErr {
		return "hans", hans.StderrWriter, color.RedString
	}
	return "hans", hans.StdoutWriter, color.BlueString
}

// setLogging sets logging level for hans based on verbosity flag
func (hans *Hans) setLogging(v bool) {
	if v {
		hans.StdoutWriter = os.Stdout
		//hans.Stdout = log.New(os.Stdout, formatName("hans", color.BlueString), log.Ldate|log.Ltime)
	} else {
		//hans.Stdout = log.New(ioutil.Discard, "", 0)
		hans.StdoutWriter = ioutil.Discard
	}
	//hans.Stderr = log.New(os.Stderr, formatName("hans", color.RedString), log.Ldate|log.Ltime)
	hans.StderrWriter = os.Stderr
}

// New inits apps and watchers and returns a complete Hans type
func New(path string, v bool) (*Hans, error) {
	hans := &Hans{}
	hans.setLogging(v)
	err := readConf(hans, path)
	if err != nil {
		return hans, err
	}
	if len(hans.Apps) == 0 {
		return hans, fmt.Errorf("no apps to run")
	}
	// set defaults
	if hans.Opts.TTL == "" {
		hans.Opts.TTL = "1s"
	}
	hans.TTL, err = time.ParseDuration(hans.Opts.TTL)
	if err != nil {
		return hans, err
	}
	// run services
	runc := make(chan Child)
	hans.Runc = runc
	go hans.run()
	manc := make(chan *App)
	go hans.manager(manc, runc)
	buildc := make(chan *App)
	go hans.build(buildc, runc)
	// init apps and watchers
	for _, app := range hans.Apps {
		app.Init(&AppConf{
			Manc: manc,
			Cwd:  hans.Opts.Cwd,
		})
		if app.Watch != "" {
			app.Watcher.Init(&WatcherConf{
				Buildc: buildc,
				App:    app,
				Verbose: v,
			})
		}
	}
	return hans, nil
}

// conf reads a config file and populates the Hans type
func readConf(hans *Hans, path string) error {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(f, hans); err != nil {
		return err
	}
	// TODO: validate fields
	return nil
}

// formatName limits app name length for logging purposes
func formatName(name string, colorfn func(string, ...interface{}) string) string {
	const maxChars int = 15 //9
	if len(name) >= maxChars {
		return colorfn(name[:14] + " ")
	}
	return colorfn(fmt.Sprintf("%-15v", name))
}

// splitBin formats a space-separated string command
func splitBin(s string) (string, []string) {
	args := strings.Split(s, " ")
	return args[0], args[1:]
}
