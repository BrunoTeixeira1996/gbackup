# TODO

- `postgresql.go`
  - Make comments
  - Fix instance const in (with toml file)
  - Fix paths

- `utils.go`
  - Clean `wrapCmd` function

- Create new struct

```
.
├── main.go
├── config.toml
├── internal
│ ├── ssh.go
│ └── wrapprom.go
└── targets
    ├── leaks.go
    ├── postgresql.go
    └── syncthing.go
```
  
- Create other rsyncs and cronjobs
  - Create toml for hosts, keypath, instance and other consts stuff
  - Fix leak_backup cronjob architecture
- Document (properly) what every rsync and cronjobs do
