package hans

import (
	"fmt"
	"testing"
	"bytes"
	"time"
	"io/ioutil"
	"strings"
)

/*
	TODO:
	test formatName
	test splitBin
*/

var (
	confPath = "/Users/pokus/golang/src/github.com/karlpokus/hans/test/conf.yaml"
	cwd = "/Users/pokus/golang/src/github.com/karlpokus/hans/test"
	old = "hello"
	new = "bye"
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

	// capture app io
	var stdoutBuf bytes.Buffer
	hans.Apps[0].setLogging(&AppConf{
		StdoutWriter: &stdoutBuf,
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
	stdout := trimBuf(&stdoutBuf)
	if stdout != old {
		t.Errorf("app stdout want: %s got: %s", old, stdout)
	}
	stdoutBuf.Reset()

	// update file, wait for hans to build and restart app and check stdout
	if err := replaceLineInFile(old, new); err != nil {
		t.Errorf("replaceLineInFile failed: %v", err)
	}
	time.Sleep(2 * time.Second)
	stdout = trimBuf(&stdoutBuf)
	if stdout != new {
		t.Errorf("app stdout want: %s got: %s", new, stdout)
	}
	stdoutBuf.Reset()

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

func trimBuf(b *bytes.Buffer) string {
	return strings.TrimSpace(b.String())
}

func shouldBeRunning(b bool, apps []*App) error {
	for _, app := range apps {
		if app.Running != b {
			return fmt.Errorf("%s running state is %v", app.Name, !b)
		}
		if app.Watch != "" && app.Watcher.Running != b {
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
