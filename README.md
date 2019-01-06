# hans
A process manager for go apps. A toy project.

# requirements
- apps should not spawn children of their own
- apps should not deamonize
- apps should log to stdout/stderr not files

# usage
```bash
$ go run hans.go
```

# todos
- [x] mvp
- [ ] hansd
- [ ] hansctl
- [ ] config file
- [ ] file watcher
- [ ] status
- [ ] poll process cpu, mem
- [ ] colourized output

# license
MIT
