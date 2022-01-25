Simple backup of IMAP attachements to local folder

### Cross compile

```
env GOOS=<OS> GOARCH=<architecture> go build -o <output file> main.go
```

Examples

```
env GOOS=linux GOARCH=amd64 go build -o attache-1.0_linux_amd64 main.go
env GOOS=darwin GOARCH=amd64 go build -o attache-1.0_darwin_amd64 main.go
env GOOS=darwin GOARCH=arm64 go build -o attache-1.0_darwin_arm64 main.go
env GOOS=windows GOARCH=amd64 go build -o attache-1.0_windows_amd64.exe main.go
````
