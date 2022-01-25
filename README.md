Simple backup of IMAP attachements to local folder

### Cross compile

```
env GOOS=<OS> GOARCH=<architecture> go build -o <output file>
```

Examples

```
Linux x64
env GOOS=linux GOARCH=amd64 go build -o attache-1.0_linux_amd64

MacOS x64
env GOOS=darwin GOARCH=amd64 go build -o attache-1.0_darwin_amd64

MacOS Apple Silicon
env GOOS=darwin GOARCH=arm64 go build -o attache-1.0_darwin_arm64

Windows x64
env GOOS=windows GOARCH=amd64 go build -o attache-1.0_windows_amd64.exe

Linux Raspberry PI
env GOOS=linux GOARCH=arm GOARM=7 go build -o attache-1.0_linux_armv7
````
