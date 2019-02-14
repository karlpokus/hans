package hans

import "testing"

func TestPath(t *testing.T) {
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
