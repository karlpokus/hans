package hans

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/karlpokus/bufw"
)

var (
	cwd      = "testdata"
	confPath = cwd + "/conf.yaml"
	old      = "hello"
	new      = "bye"
)

func TestHansNew(t *testing.T) {
	hans, err := New(confPath, true)
	if err != nil {
		t.Errorf("Hans New failed: %v", err)
		t.FailNow()
	}
	if hans.Opts.Cwd != cwd || hans.Opts.TTL != "5s" {
		t.Errorf("Wrong Hans fields set: %v, %v", hans.Opts.Cwd, hans.Opts.TTL)
		t.FailNow()
	}
}

func TestHansStart(t *testing.T) {
	hans, _ := New(confPath, false)
	if err := shouldBeRunning(false, hans.Apps); err != nil {
		t.Error(err)
	}
	// prepare to capture app io
	w := bufw.New(true)
	hans.Apps[0].setLogging(&AppConf{
		StdoutWriter: w,
	})
	// start
	if err := hans.Start(); err != nil {
		t.Errorf("Hans Start failed: %v", err)
		t.FailNow()
	}
	if err := w.Wait(); err != nil {
		t.Errorf("%s", err)
	}
	if err := shouldBeRunning(true, hans.Apps); err != nil {
		t.Error(err)
	}
	stdout := w.String()
	if stdout != old {
		t.Errorf("app stdout want: %s got: %s", old, stdout)
	}
	// update file, wait for watcher to trigger hans to build and restart app
	go updateSrc(old, new)
	if err := w.Wait(); err != nil {
		t.Errorf("%s", err)
	}
	stdout = w.String()
	if stdout != new {
		t.Errorf("app stdout want: %s got: %s", new, stdout)
	}
	// reset test state and cleanup
	hans.cleanup()
	updateSrc(new, old)
	hans.Apps[0].build()
	if err := shouldBeRunning(false, hans.Apps); err != nil {
		t.Error(err)
	}
}

func shouldBeRunning(b bool, apps []*App) error {
	for _, app := range apps {
		if app.Running() != b {
			return fmt.Errorf("%s running state is %v", app.Name, !b)
		}
		if app.Watch != "" && app.Watcher.Running() != b {
			return fmt.Errorf("%s watcher running state is %v", app.Name, !b)
		}
	}
	return nil
}

func updateSrc(old, new string) {
	filepath := cwd + "/src/hello/hello.go"
	f, _ := ioutil.ReadFile(filepath)
	lines := strings.Split(string(f), "\n")
	for i, line := range lines {
		if strings.Contains(line, old) {
			lines[i] = strings.Replace(line, old, new, 1)
		}
	}
	ioutil.WriteFile(filepath, []byte(strings.Join(lines, "\n")), 0644)
}
