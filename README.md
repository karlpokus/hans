# hans
A process manager. A toy project.

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
$ go test -v ./pkg/hans/...
```

# todos
- [x] mvp
- [ ] hansd
- [ ] hansctl
- [x] config file
- [x] app src watcher option
- [x] app src rebuild on change
- [x] app restart on change
- [ ] status
- [ ] poll process cpu, mem
- [ ] colourized output
- [x] relative paths in config
- [ ] scale app
- [ ] os independent paths
- [x] write tests and remove app dependency on hans
- [ ] pass Env flags to children via `Cmd.Env`
- [ ] log levels
- [x] `Cmd.Dir` for global pwd
- [x] use conf structs
- [ ] split confs on main, app, watcher

# license
MIT
