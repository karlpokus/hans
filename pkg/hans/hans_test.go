package hans

import (
	"testing"
	"fmt"
)

/*
	TODO:
	test formatName
	test splitBin
*/

var confPath = "/Users/pokus/golang/src/github.com/karlpokus/hans/test/conf.yaml"
var cwd = "/Users/pokus/golang/src/github.com/karlpokus/hans/test"

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
	hans, err := New(confPath, false)
	if err := shouldBeRunning(false, hans.Apps); err != nil {
		t.Error(err)
	}
	err = hans.Start()
	if err != nil {
		t.Errorf("Hans Start failed: %v", err)
		t.FailNow()
	}
	if err := shouldBeRunning(true, hans.Apps); err != nil {
		t.Error(err)
	}
	hans.cleanup()
	if err := shouldBeRunning(false, hans.Apps); err != nil {
		t.Error(err)
	}
}

func shouldBeRunning(b bool, apps []*App) error {
	for _, app := range apps {
		if app.Running != b {
			return fmt.Errorf("%s running state should be %v", app.Name, !b)
		}
	}
	return nil
}
