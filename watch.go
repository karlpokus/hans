package hans

import (
	"os"
	"path"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/time/rate"
)

type Watcher struct {
	*App
	Build   chan *App
	RootDir string
	State
	*fsnotify.Watcher
}

type WatcherConf struct {
	*App
	Build chan *App
}

// Watch starts watching files and dirs recursively
func (w *Watcher) Watch() error {
	fi, err := os.Stat(w.RootDir)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		err := w.Watcher.Add(w.RootDir) // only watching a file
		if err != nil {
			return err
		}
		return nil
	}
	walk := func(path string, fi os.FileInfo, err error) error {
		if fi.IsDir() {
			return w.Watcher.Add(path)
		}
		return nil
	}
	err = filepath.Walk(w.RootDir, walk) // recursively add more dirs
	if err != nil {
		return err
	}
	return nil
}

func debounce(a chan fsnotify.Event) chan struct{} {
	l := rate.NewLimiter(0.2, 1) // once per 5 secs
	b := make(chan struct{})
	go func() {
		for _ = range a {
			if l.Allow() {
				b <- struct{}{}
			}
		}
		close(b)
	}()
	return b
}

func (w *Watcher) Run(fail chan error) {
	err := w.Watch()
	fail <- err
	if err != nil {
		w.Watcher.Close()
		return
	}
	// TODO: handle w.Watcher.Errors
	for _ = range debounce(w.Watcher.Events) {
		w.Build <- w.App
	}
}

func (w *Watcher) Init(conf *WatcherConf) {
	w.Build = conf.Build
	w.App = conf.App
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err) // TODO: don't panic
	}
	w.Watcher = watcher
	w.RootDir = path.Join(conf.App.Cwd, conf.App.Watch)
}

func (w *Watcher) Kill() {
	w.Watcher.Close()
}
