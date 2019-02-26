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

func TestAppPath(t *testing.T) {
	// empty init
	app := &App{}
	got := app.path("/bar")
	want := "/bar"
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	// Cwd set
	app.Cwd = "foo"
	got = app.path("/bar")
	want = "foo/bar"
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestAppRun(t *testing.T) {
	cwd := "/Users/pokus/golang/src/github.com/karlpokus/hans/test/"
	app := &App{
		Bin: "apps/hello",
	}
	app.init(cwd)
	fail := make(chan error)
	go app.run(fail)

	select {
	case <-time.After(1 * time.Second):
		t.Errorf("%s timed out", app.Name)
		app.kill()
		t.FailNow()
	case err := <-fail:
		if err != nil {
			t.Errorf("%s did not start %s", app.Name, err)
			app.kill()
			t.FailNow()
		}
	}
	time.Sleep(1 * time.Second) // wait for app to start

	if trimBuf(&app.StdoutBuf) != "hello" {
		t.Errorf("app.StdoutBuf fail: %s", trimBuf(&app.StdoutBuf))
	}
	if app.Running == false {
		t.Error("app should be running")
	}
	app.kill()
	if app.Running == true {
		t.Error("app should not be running")
	}
}
