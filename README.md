# hans
A process manager. A toy project.

# requirements
- apps should not spawn children of their own
- apps should not deamonize
- apps should log to stdout/stderr not files
- app src watch option requires [fswatch](https://github.com/emcrisostomo/fswatch)

# usage
config
```yaml
cwd: global base path # optional
apps:
- name: the name of the app # required
  bin: absolute|relative path to bin and space separated args to run # required
  watch: absolute|relative path to a src dir or file to watch for changes. Will trigger a restart of bin # optional
  build: absolute path to build command to run on src changes before restart # optional
```
start
```bash
$ go run main.go
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

# license
MIT
