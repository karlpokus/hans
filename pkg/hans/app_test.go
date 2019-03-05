package hans

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func trimBuf(b *bytes.Buffer) string {
	return strings.TrimSpace(b.String())
}

func TestAppRun(t *testing.T) {
	app := &App{ // don't read conf file
		Bin: "apps/hello",
	}
	var stdoutBuf bytes.Buffer
	app.Init(&AppConf{
		Cwd: "/Users/pokus/golang/src/github.com/karlpokus/hans/test/",
		StdoutWriter: &stdoutBuf,
	})
	fail := make(chan error)
	go app.Run(fail)

	select {
	case <-time.After(1 * time.Second):
		t.Errorf("%s timed out", app.Name)
		app.Kill()
		t.FailNow()
	case err := <-fail:
		if err != nil {
			t.Errorf("%s did not start %s", app.Name, err)
			app.Kill()
			t.FailNow()
		}
	}
	time.Sleep(1 * time.Second) // wait for app to start
	stdout := trimBuf(&stdoutBuf)
	if stdout != "hello" {
		t.Errorf("app.StdoutBuf fail: %s", stdout)
	}
	if app.Running == false {
		t.Error("app should be running")
	}
	app.Kill()
	if app.Running == true {
		t.Error("app should not be running")
	}
}
