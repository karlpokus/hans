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

	"github.com/fatih/color"
	"gopkg.in/yaml.v2"
)

var Version = "vX.Y.Z" // injected at build time

type Opt struct {
	Cwd string
	TTL string
}

type Hans struct {
	Stdout *log.Logger
	Stderr *log.Logger
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
	hans.Stdout.Printf("%s start", mod)
	ok := hans.interrupt() // SIGINT is trappable.
	if ok {
		hans.Stdout.Printf("%s pgid recieved SIGINT", mod)
		time.Sleep(2 * time.Second)
	}
	for _, app := range hans.Apps {
		if app.Running() {
			hans.kill(app) // SIGKILL is not trappable
			hans.Stdout.Printf("%s %s killed", mod, app.Name)
		}
		if app.Watcher.Running() {
			hans.kill(app.Watcher)
			hans.Stdout.Printf("%s %s watcher closed", mod, app.Name)
		}
	}
	hans.Stdout.Printf("%s done", mod)
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
			hans.Stderr.Printf("%s %s timeout", mod, c.GetName())
		case err := <-fail:
			if err != nil {
				hans.Stderr.Printf("%s %s failed start attempt: %v", mod, c.GetName(), err)
				break
			}
			c.RunningState(true)
			hans.Stdout.Printf("%s %s started", mod, c.GetName())
		}
	}
}

func (hans *Hans) manager(manc chan *App, runc chan Child) {
	mod := "[MANAGER]"
	for app := range manc {
		code := app.Cmd.ProcessState.ExitCode()
		switch {
		case code == -1:
			hans.Stdout.Printf("%s %s terminated", mod, app.Name)
		case code == 0:
			hans.Stdout.Printf("%s %s exited", mod, app.Name)
		case code > 0:
			hans.Stderr.Printf("%s %s exited %d", mod, app.Name, code)
			// bad exit dance
			app.BadExit.Init()
			app.BadExit.Inc()
			if app.BadExit.MaxReached() {
				if app.BadExit.WithinWindow() {
					hans.Stderr.Printf("%s maxBadExits reached. %s is dead", mod, app.Name)
					break
				}
				app.BadExit.Reset()
			}
			hans.Stdout.Printf("%s restarting %s", mod, app.Name)
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
		hans.Stdout.Printf("%s %s src change detected. Attempting build and restart", mod, app.Name)
		if app.Build == "" {
			hans.Stderr.Printf("%s %s build cmd missing. Build aborted. Attempting restart", mod, app.Name)
			hans.kill(app) // app is running during build
			app.SetCmd()
			runc <- app
			continue
		}
		res, err := app.build() // TODO: let app.build check Build string
		if err != nil {
			hans.Stderr.Printf("%s %s build failed: %v", mod, app.Name, err)
			hans.Stderr.Printf("%s %s", mod, res)
			hans.Stderr.Printf("%s restart attempt aborted", mod)
			continue
		}
		hans.Stdout.Printf("%s %s build successful. Attempting restart", mod, app.Name)
		hans.kill(app) // app is running during build
		app.SetCmd()
		runc <- app
	}
}

// setLogging sets logging level for hans based on verbosity flag
func (hans *Hans) setLogging(v bool) {
	if v {
		hans.Stdout = log.New(os.Stdout, formatName("hans", color.BlueString), log.Ldate|log.Ltime)
	} else {
		hans.Stdout = log.New(ioutil.Discard, "", 0)
	}
	hans.Stderr = log.New(os.Stderr, formatName("hans", color.RedString), log.Ldate|log.Ltime)
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
	const maxChars int = 9
	if len(name) >= maxChars {
		return colorfn(name[:9] + " ")
	}
	return colorfn(fmt.Sprintf("%-10v", name))
}

// splitBin formats a space-separated string command
func splitBin(s string) (string, []string) {
	args := strings.Split(s, " ")
	return args[0], args[1:]
}
