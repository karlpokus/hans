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
	Stdout *log.Logger
	Stderr *log.Logger
	Apps   []*App
	Opts   Opt
	TTL time.Duration // TODO: make a conf struct instead
}

// cleanup kills running apps and associated watchers on os.signals
// when done it writes to the passed in done channel
func (hans *Hans) cleanup(done chan<- bool) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	if len(hans.Apps) > 0 {
		for _, app := range hans.Apps {
			if app.Running {
				hans.Stdout.Printf("killing %s", app.Name)
				app.kill()
			}
			if app.Watcher.Running {
				hans.Stdout.Printf("killing %s watcher", app.Name)
				app.Watcher.kill()
			}
		}
	}
	done <- true
}

// Start starts all apps and associated watchers
// it also prepares cleanup on main exit
func (hans *Hans) Start() (<-chan bool, error) {
	if len(hans.Apps) == 0 {
		return nil, errors.New("no apps to run")
	}
	for _, app := range hans.Apps {
		hans.Stdout.Printf("%s starting", app.Name)
		app.init(hans.Opts.Cwd)
		fail := make(chan error)
		go app.run(fail)

		select {
		case <-time.After(hans.TTL):
			hans.Stderr.Printf("%s timed out", app.Name)
			continue
		case err := <-fail:
			if err != nil {
				hans.Stderr.Printf("%s did not start %s", app.Name, err)
				continue
			}
			hans.Stdout.Printf("%s started", app.Name)
		}

		if len(app.Watch) > 0 {
			hans.Stdout.Printf("%s watcher starting", app.Name)
			restart := make(chan string)
			go app.Watcher.Watch(fail, restart)

			select {
			case <-time.After(hans.TTL):
				hans.Stderr.Printf("%s watcher timed out", app.Name)
				continue
			case err := <-fail:
				if err != nil {
					hans.Stderr.Printf("%s watcher did not start %s", app.Name, err)
					continue
				}
				hans.Stdout.Printf("%s watcher started", app.Name)
				go hans.restart(restart)
			}
		}
	}
	done := make(chan bool, 1)
	go hans.cleanup(done)
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
		// TODO: only unblock when build and restart are done: semaphore?
		app := hans.appFromName(<-c)
		if len(app.Build) > 0 {
			done := make(chan bool)
			go app.build(done)
			<-done
		}
		if app.Running {
			app.kill()
			//app.restart()
		}
	}
}

// New takes a path to a config file and returns a complete Hans type
func New(path string) (*Hans, error) {
	hans := &Hans{
		Stdout: log.New(os.Stdout, formatName("hans"), log.Ldate|log.Ltime),
		Stderr: log.New(os.Stderr, formatName("hans"), log.Ldate|log.Ltime),
	}
	err := hans.conf(path)
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
func (hans *Hans) conf(path string) error {
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
