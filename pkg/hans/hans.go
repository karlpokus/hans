package hans

import (
	"bytes"
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

var RUNTIME = os.Getenv("RUNTIME")

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
	StdoutBuf bytes.Buffer
	StderrBuf bytes.Buffer
	Verbose   bool
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
				if hans.Verbose {
					hans.Stdout.Printf("killing %s", app.Name)
				}
				app.Kill()
			}
			if app.Watcher.Running {
				if hans.Verbose {
					hans.Stdout.Printf("killing %s watcher", app.Name)
				}
				app.Watcher.Kill()
			}
		}
	}
}

// cleanupOnExit kills running apps and associated watchers on exit
// when done it writes to the passed in done channel
func (hans *Hans) cleanupOnExit(done chan<- bool) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	hans.cleanup()
	done <- true
}

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

// Start starts all apps and associated watchers
// it also prepares cleanup on exit
func (hans *Hans) Start() (<-chan bool, error) {
	if len(hans.Apps) == 0 {
		return nil, errors.New("no apps to run")
	}
	for _, app := range hans.Apps {
		if hans.Verbose {
			hans.Stdout.Printf("%s starting", app.Name)
		}
		app.Init(hans.Opts.Cwd)

		if err := hans.run(app); err != nil {
			hans.Stderr.Printf("%s did not start: %s", app.Name, err)
		}
		if hans.Verbose {
			hans.Stdout.Printf("%s started", app.Name)
		}
		if len(app.Watch) > 0 {
			restart := make(chan string) // share this among all watchers?
			app.Watcher.Init(restart, app)

			if hans.Verbose {
				hans.Stdout.Printf("%s watcher starting", app.Name)
			}
			if err := hans.run(app.Watcher); err != nil {
				hans.Stderr.Printf("%s watcher did not start: %s", app.Name, err)
			}
			if hans.Verbose {
				hans.Stdout.Printf("%s watcher started", app.Name)
			}
			go hans.restart(restart) // TODO: only start one of these for all watchers?
		}
	}
	done := make(chan bool, 1)
	go hans.cleanupOnExit(done)
	return done, nil
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

// restart restarts an app when signaled from a watcher
// also runs build before restarting if the build field is set in the apps config
func (hans *Hans) restart(c chan string) {
	for {
		// TODO: wait for build and restart if multiple watchers share the chan
		app := hans.appFromName(<-c)
		hans.Stdout.Printf("detected change on %s src", app.Name)
		hans.Stdout.Printf("attempting %s restart", app.Name)
		if len(app.Build) > 0 {
			hans.Stdout.Printf("rebuilding %s first", app.Name)
			res, err := app.build()
			if err != nil {
				hans.Stderr.Printf("%s build err: %v", app.Name, err)
				hans.Stderr.Printf("%s", res)
				hans.Stderr.Printf("%s restart aborted", app.Name)
				continue // don't restart
			}
			hans.Stdout.Printf("%s build succesful", app.Name)
		}
		if app.Running {
			app.Kill()
			hans.Stdout.Printf("restarting %s", app.Name)
			fail := make(chan error)
			app.restart(fail)

			select {
			case <-time.After(hans.TTL):
				hans.Stderr.Printf("%s timed out", app.Name)
				continue
			case err := <-fail:
				if err != nil {
					hans.Stderr.Printf("%s did not restart %s", app.Name, err)
					continue
				}
				hans.Stdout.Printf("%s restarted", app.Name)
			}
		}
	}
}

// setLogging sets the logging for hans based on RUNTIME env
// it also sets verbosity
func (hans *Hans) setLogging(v bool) {
	hans.Verbose = v
	if RUNTIME == "TEST" {
		hans.Stdout = log.New(&hans.StdoutBuf, "", 0)
		hans.Stderr = log.New(&hans.StderrBuf, "", 0)
	} else {
		hans.Stdout = log.New(os.Stdout, formatName("hans"), log.Ldate|log.Ltime)
		hans.Stderr = log.New(os.Stderr, formatName("hans"), log.Ldate|log.Ltime)
	}
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

/*func absPath(p string) string {
	pwd, _ := os.Getwd()
	return pwd + p
}*/

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
