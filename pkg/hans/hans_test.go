package hans

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/karlpokus/bufw"
)

var (
	cwd      = "/Users/pokus/dev/hans/testdata"
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
	if err := shouldBeRunning(true, hans.Apps); err != nil {
		t.Error(err)
	}

	// wait for app to write to stdout
	if err := w.Wait(); err != nil {
		t.Errorf("wait err %s", err)
	}
	stdout := w.String()
	if stdout != old {
		t.Errorf("app stdout want: %s got: %s", old, stdout)
	}

	// update file, wait for watcher to trigger hans to build and restart app
	if err := replaceLineInFile(old, new); err != nil {
		t.Errorf("replaceLineInFile failed: %v", err)
	}
	if err := w.Wait(); err != nil {
		t.Errorf("wait err %s", err)
	}
	stdout = w.String()
	if stdout != new {
		t.Errorf("app stdout want: %s got: %s", new, stdout)
	}

	// cleanup
	hans.cleanup()
	if err := shouldBeRunning(false, hans.Apps); err != nil {
		t.Error(err)
	}

	// reset test state
	if err := replaceLineInFile(new, old); err != nil {
		t.Errorf("replaceLineInFile failed: %v", err)
	}
	hans.build(hans.Apps[0])
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

func replaceLineInFile(old, new string) error {
	filepath := cwd + "/src/hello/hello.go"
	f, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(f), "\n")
	for i, line := range lines {
		if strings.Contains(line, old) {
			lines[i] = strings.Replace(line, old, new, 1)
		}
	}
	return ioutil.WriteFile(filepath, []byte(strings.Join(lines, "\n")), 0644)
}
