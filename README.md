# hans
A process manager for go apps. A toy project.

# requirements
- apps should not spawn children of their own
- apps should not deamonize
- apps should log to stdout/stderr not files
- app src watcher option requires [fswatch](https://github.com/emcrisostomo/fswatch)

# usage
```bash
$ go run *.go
```

# todos
- [x] mvp
- [ ] hansd
- [ ] hansctl
- [x] config file
- [x] app src watcher option
- [ ] app src rebuild on change
- [ ] app restart on change
- [ ] status
- [ ] poll process cpu, mem
- [ ] colourized output

# license
MIT
