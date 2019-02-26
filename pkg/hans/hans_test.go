package hans

import "testing"

/*
	TODO:
	test formatName
	test splitBin
*/

var confPath = "/Users/pokus/golang/src/github.com/karlpokus/hans/test/conf.yaml"
var cwd = "/Users/pokus/golang/src/github.com/karlpokus/hans/test/"

func TestHansNew(t *testing.T) {
	hans, err := New(confPath, true)
	if err != nil {
		t.Errorf("Hans New failed: %v", err)
		t.FailNow()
	}
	if hans.Opts.Cwd != cwd || hans.Opts.TTL != "5s" || hans.Verbose != true {
		t.Errorf("Wrong Hans fields set: %v, %v, %v", hans.Opts.Cwd, hans.Opts.TTL, hans.Verbose)
		t.FailNow()
	}
}

func TestHansStart(t *testing.T) {
	hans, err := New(confPath, false)
	_, err = hans.Start() // ignore done chan
	if err != nil {
		t.Errorf("Hans Start failed: %v", err)
		t.FailNow()
	}
	for _, app := range hans.Apps {
		if !app.Running {
			t.Errorf("%s is not running after Hans Start", app.Name)
		}
	}
	hans.cleanup()
	for _, app := range hans.Apps {
		if app.Running {
			t.Errorf("%s is running after Hans cleanup", app.Name)
		}
	}
}
