package hans

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type Opt struct {
	Cwd string
	TTL string
}

type Hans struct {
	Stdout    *log.Logger
	Stderr    *log.Logger
	Apps      []*App
	Opts      Opt
	TTL       time.Duration
}

type Child interface {
	Run(chan error)
	Kill()
}

// cleanup kills running apps and associated watchers
func (hans *Hans) cleanup() {
	if len(hans.Apps) > 0 {
		for _, app := range hans.Apps {
			if app.Running {
				hans.Stdout.Printf("killing %s", app.Name)
				app.Kill()
			}
			if app.Watcher.Running {
				hans.Stdout.Printf("killing %s watcher", app.Name)
				app.Watcher.Kill()
			}
		}
	}
}

// cleanupOnExit runs cleanup on hans exit
// then writes to the done channel
func (hans *Hans) cleanupOnExit(done chan<- bool) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	hans.cleanup()
	done <- true
}

// run runs a child
func (hans *Hans) run(c Child) error {
	fail := make(chan error)
	go c.Run(fail)

	select {
	case <-time.After(hans.TTL):
		return fmt.Errorf("timeout")
	case err := <-fail:
		return err
	}
}

// shouldStartWatcher determines if we should run hans.restart
func (hans *Hans) shouldStartWatcher() bool {
	for _, app := range hans.Apps {
		if app.Watch != "" {
			return true
		}
	}
	return false
}

// Wait prepares resource cleanup on hans exit
func (hans *Hans) Wait() <-chan bool {
	done := make(chan bool, 1)
	go hans.cleanupOnExit(done)
	return done
}

// Start inits and starts all apps and associated watchers
func (hans *Hans) Start() error {
	if len(hans.Apps) == 0 { // TODO: move to New
		return errors.New("no apps to run")
	}
	var restart chan string
	if hans.shouldStartWatcher() {
		restart = make(chan string)
		go hans.restart(restart)
	}
	for _, app := range hans.Apps {
		app.Init(&AppConf{ // TODO: move init to New
			Cwd: hans.Opts.Cwd,
		})
		err := hans.run(app)
		if err != nil {
			hans.Stderr.Printf("%s did not start: %s", app.Name, err)
		}
		hans.Stdout.Printf("%s started", app.Name)

		if app.Watch != "" {
			app.Watcher.Init(&WatcherConf{
				Restart: restart,
				App: app,
			})
			err := hans.run(app.Watcher)
			if err != nil {
				hans.Stderr.Printf("%s watcher did not start: %s", app.Name, err)
			}
			hans.Stdout.Printf("%s watcher started", app.Name)
		}
	}
	return nil
}

// appFromName returns an App type from an app name
func (hans *Hans) appFromName(appName string) *App {
	for _, app := range hans.Apps {
		if app.Name == appName {
			return app
		}
	}
	return &App{}
}

func (hans *Hans) build(app *App) error {
	// TODO remove chatter
	hans.Stdout.Printf("detected change on %s src", app.Name)
	hans.Stdout.Printf("attempting %s restart", app.Name)
	if app.Build != "" {
		hans.Stdout.Printf("rebuilding %s first", app.Name)
		res, err := app.build()
		if err != nil {
			hans.Stderr.Printf("%s build err: %v", app.Name, err)
			hans.Stderr.Printf("%s", res)
			hans.Stderr.Printf("%s restart aborted", app.Name)
			return err
		}
		hans.Stdout.Printf("%s build succesful", app.Name)
	}
	return nil
}

// restart restarts an app when signaled from a watcher
// also runs build before restarting if the build field is set in the apps config
func (hans *Hans) restart(c chan string) {
	for {
		// TODO: wait for build and restart if multiple watchers share the chan
		app := hans.appFromName(<-c)
		err := hans.build(app)
		if err != nil {
			continue // don't restart
		}
		if app.Running {
			hans.Stdout.Printf("restarting %s", app.Name)
			app.Kill()
			app.setCmd()
			err := hans.run(app)
			if err != nil {
				hans.Stderr.Printf("%s did not restart: %s", app.Name, err)
			}
			hans.Stdout.Printf("%s restarted", app.Name)
		}
	}
}

// setLogging sets logging level for hans based on verbosity flag
func (hans *Hans) setLogging(v bool) {
	if v {
		hans.Stdout = log.New(os.Stdout, formatName("hans"), log.Ldate|log.Ltime)
	} else {
		hans.Stdout = log.New(ioutil.Discard, "", 0)
	}
	hans.Stderr = log.New(os.Stderr, formatName("hans"), log.Ldate|log.Ltime)
}

// New takes a path to a config file, a verbose flag and returns a complete Hans type
func New(path string, v bool) (*Hans, error) {
	hans := &Hans{}
	hans.setLogging(v)
	err := readConf(hans, path)
	if err != nil {
		return hans, err
	}
	// set defaults
	if hans.Opts.TTL == "" {
		hans.Opts.TTL = "1s"
	}
	hans.TTL, err = time.ParseDuration(hans.Opts.TTL)
	return hans, err
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
func formatName(name string) string {
	const maxChars int = 9
	if len(name) >= maxChars {
		return name[:9] + " "
	}
	return fmt.Sprintf("%-10v", name)
}

// splitBin formats a space-separated string command
func splitBin(s string) (string, []string) {
	args := strings.Split(s, " ")
	return args[0], args[1:]
}
