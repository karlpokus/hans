package hans

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/time/rate"
)

type Watcher struct {
	*App
	Buildc      chan *App
	RootDir     string
	ExcludePath string
	State
	*fsnotify.Watcher
}

type WatcherConf struct {
	*App
	Buildc chan *App
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
		if w.ExcludePath != "" && strings.HasPrefix(path, w.ExcludePath) {
			return nil
		}
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
		w.Buildc <- w.App
	}
}

func (w *Watcher) Init(conf *WatcherConf) {
	w.Buildc = conf.Buildc
	w.App = conf.App
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err) // TODO: don't panic
	}
	w.Watcher = watcher
	w.RootDir = path.Join(conf.App.Cwd, conf.App.Watch)
	if conf.App.Watch_exclude != "" {
		w.ExcludePath = path.Join(w.RootDir, conf.App.Watch_exclude)
	}
}

func (w *Watcher) Kill() {
	w.Watcher.Close()
}

func (w *Watcher) GetName() string {
	return w.App.Name + " watcher"
}
