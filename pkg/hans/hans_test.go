package hans

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"testing"
	"time"
)

var (
	cwd      = "/Users/pokus/golang/src/github.com/karlpokus/hans/testdata"
	confPath = cwd + "/conf.yaml"
	old      = "hello"
	new      = "bye"
)

type mockWriter struct {
	buf bytes.Buffer
	mu  sync.Mutex
}

func (mw *mockWriter) Write(b []byte) (int, error) {
	mw.mu.Lock()
	defer mw.mu.Unlock()
	mw.buf.Write(b)
	return len(b), nil
}

func (mw *mockWriter) Read() string {
	mw.mu.Lock()
	defer mw.mu.Unlock()
	return strings.TrimSpace(mw.buf.String())
}

func (mw *mockWriter) Reset() {
	mw.buf.Reset()
}

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

	// capture app io
	mw := &mockWriter{}
	hans.Apps[0].setLogging(&AppConf{
		StdoutWriter: mw,
	})

	// start
	if err := hans.Start(); err != nil {
		t.Errorf("Hans Start failed: %v", err)
		t.FailNow()
	}
	if err := shouldBeRunning(true, hans.Apps); err != nil {
		t.Error(err)
	}

	// wait for app to start and check stdout
	time.Sleep(1 * time.Second)
	stdout := mw.Read()
	if stdout != old {
		t.Errorf("app stdout want: %s got: %s", old, stdout)
	}
	mw.Reset()

	// update file, wait for hans to build and restart app and check stdout
	if err := replaceLineInFile(old, new); err != nil {
		t.Errorf("replaceLineInFile failed: %v", err)
	}
	time.Sleep(2 * time.Second)
	stdout = mw.Read()
	if stdout != new {
		t.Errorf("app stdout want: %s got: %s", new, stdout)
	}
	mw.Reset()

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
