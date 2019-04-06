# hans
A process manager. A toy project.

A process manager can be two things 1: a development tool that restarts processes on file change and provide a shared log stream, and 2: a production runtime daemon to manage process life cycle, scaling and resource consumption. I originally wanted both but I'll settle for 1 for now.

# requirements
- apps should not spawn children of their own
- apps should not deamonize
- apps should log to stdout/stderr - not files
- app src watch option requires [fswatch](https://github.com/emcrisostomo/fswatch)

# usage
config
```yaml
apps:
- name: the name of the app # required
  bin: path to bin and space separated args to run # required
  watch: path to a src dir or file to watch for changes. Will trigger a restart of bin # optional
  build: absolute path to build command to run on src changes before restart # optional
opts:
  cwd: global base path for relative app paths # optional
  ttl: global timeout for app and watcher startups # optional, defaults to 1s
```
start
```bash
$ go run main.go
```
options
- `-conf` path/to/config. Defaults to config.yaml
- `-v` verbose logging. Defaults to true

run test
```bash
$ go test -v ./pkg/hans/... -race -cover
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
- [ ] pass Env flags to children via `Cmd.Env`
- [ ] log levels
- [x] `Cmd.Dir` for global pwd
- [x] use conf structs
- [ ] split confs on main, app, watcher
- [ ] check `Cmd.ProcessState` if proc exits on its own
- [ ] notification channel on app restart

# license
MIT
