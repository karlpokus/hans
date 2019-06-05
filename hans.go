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
}

type Child interface {
	Run(chan error)
	Kill()
	RunningState(bool)
}

// cleanup kills running apps and associated watchers
func (hans *Hans) cleanup() {
	for _, app := range hans.Apps {
		if app.Running() {
			hans.kill(app)
			hans.Stdout.Printf("%s killed", app.Name)
		}
		if app.Watcher.Running() {
			hans.kill(app.Watcher)
			hans.Stdout.Printf("%s watcher killed", app.Name)
		}
	}
}

// kill kills a child and toggles running state
func (hans *Hans) kill(c Child) {
	c.Kill()
	c.RunningState(false)
}

// run runs a child and toggles running state
func (hans *Hans) run(c Child) error {
	fail := make(chan error)
	go c.Run(fail)

	select {
	case <-time.After(hans.TTL):
		return fmt.Errorf("timeout")
	case err := <-fail:
		if err == nil {
			c.RunningState(true)
		}
		return err
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
		err := hans.run(app)
		if err != nil {
			hans.Stderr.Printf("%s did not start: %s", app.Name, err)
			continue
		}
		hans.Stdout.Printf("%s started", app.Name)
		if app.Watch != "" {
			err := hans.run(app.Watcher)
			if err != nil {
				hans.Stderr.Printf("%s watcher did not start: %s", app.Name, err)
				continue
			}
			hans.Stdout.Printf("%s watcher started", app.Name)
		}
	}
	return nil
}

// build builds an app and sends it off for restarting if successful
func (hans *Hans) build(buildChan, restartChan chan *App) {
	for app := range buildChan {
		hans.Stdout.Printf("%s src change detected, attempting build and restart", app.Name)
		if app.Build == "" {
			hans.Stderr.Printf("%s is missing a build cmd, build and restart aborted", app.Name)
			continue
		}
		res, err := app.build()
		if err != nil {
			hans.Stderr.Printf("%s build failed: %v", app.Name, err)
			hans.Stderr.Printf("%s", res)
			hans.Stderr.Println("restart aborted")
			continue
		}
		hans.Stdout.Printf("%s build successful", app.Name)
		restartChan <- app
	}
}

// restart restarts an app
func (hans *Hans) restart(c chan *App) {
	for app := range c {
		if app.BadExit.Ko {
			hans.Stderr.Printf("maxBadExits reached for %s. No more restarts", app.Name)
			continue
		}
		hans.Stdout.Printf("restarting %s", app.Name)
		if app.Running() { // app is still running on src change
			hans.kill(app)
		}
		app.setCmd()
		err := hans.run(app)
		if err != nil {
			hans.Stderr.Printf("%s did not restart: %s", app.Name, err)
		}
		hans.Stdout.Printf("%s restarted", app.Name)
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
	restart := make(chan *App)
	build := make(chan *App)
	go hans.restart(restart)
	go hans.build(build, restart)
	// init apps and watchers
	for _, app := range hans.Apps {
		app.Init(&AppConf{
			Restart: restart,
			Cwd:     hans.Opts.Cwd,
		})
		if app.Watch != "" {
			app.Watcher.Init(&WatcherConf{
				Build: build,
				App:   app,
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
		return name[:9] + " "
	}
	return colorfn(fmt.Sprintf("%-10v", name))
}

// splitBin formats a space-separated string command
func splitBin(s string) (string, []string) {
	args := strings.Split(s, " ")
	return args[0], args[1:]
}
