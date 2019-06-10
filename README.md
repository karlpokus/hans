# hans

A process manager can be two things 1: a development tool that restarts processes on file change and provide a shared log stream, and 2: a production runtime daemon to manage process life cycle, scaling and resource consumption. I originally wanted both but I'll settle for 1 for now.

# requirements

- apps should not spawn children of their own
- apps should not deamonize
- apps should log to stdout/stderr - not files

# usage

config
```yaml
apps:
  - name: the name of the app # required
    bin: path to binary and space separated args to run # required
    watch: path to a src dir or file to watch for changes. Src change will trigger a restart of bin # optional
    watch_exclude: path relative to watch to exclude from watching # optional
    build: path to build command to run on src changes before restart # optional
    env: array of key=value pairs for the app environment # optional
    cwd: local base path for relative app paths. Overrides the global one # optional
opts:
  cwd: global base path for relative app paths # optional
  ttl: global timeout for app and watcher startups # optional, defaults to 1s
```

start
```bash
$ go run ./cmd/hans <flags>
```

flags
- `-conf` path/to/config. Defaults to conf.yaml
- `-v` verbose logging. Defaults to false
- `-version` print version and exit

# test

```bash
$ go test -v -race -cover
```

# build

```bash
# creates a new version of the hans cmd by
# updating the version file,
# building amd64 binaries for linux and darwin,
# adds-, and pushes new commit,
# adds-, and pushes new git tag
$ ./release.sh vX.Y.Z
```

# todos

- [x] mvp
- [x] hansd - won't do
- [x] hansctl - won't do
- [x] poll process cpu, mem - won't do
- [x] scale app - won't do
- [x] config file
- [x] app src watcher option
- [x] app src rebuild on change
- [x] app restart on change
- [x] colourize output
- [x] relative paths in config
- [ ] os independent paths
- [x] write tests and remove app dependency on hans
- [x] pass Env flags to children via `Cmd.Env`
- [ ] log levels
- [x] `Cmd.Dir` for global pwd
- [x] use conf structs
- [ ] check `Cmd.ProcessState` if proc exits on its own
- [ ] notification channel on app restart
- [x] go mod
- [x] replace fsnotify with lib
- [ ] support globbing build cmd
- [x] restart child on bad exit
- [x] maxBadExits
- [x] version flag
- [x] local cwd
- [ ] maybe stop hans when last app exited
- [x] allow graceful exits
- [x] log any exit
- [x] log by module
- [x] watch exclude rules

# known bugs
- passing the env opt requires an explicit interpreter (like node) regardless of she-bang
- `Process.Signal(os.Kill)` will not allow for capturing anything after `Cmd.Wait()`

# license
MIT
