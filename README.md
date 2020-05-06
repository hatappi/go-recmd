# recmd
`recmd` is simple live reloading tool that doesn't require **a configuration file**.
This tools supports MacOS, Linux.

## Installation

```sh
$ go get github.com/hatappi/go-recmd/cmd/recmd
```

or 

https://github.com/hatappi/go-recmd/releases

## Usage

```sh
$ recmd watch -h
watch path and execute command

Usage:
  recmd watch [flags] [your command]

Aliases:
  watch, w

Examples:

$ recmd watch go run main.go
$ recmd watch -p "./main.go" go run main.go
$ recmd watch --exclude testA --exclude testB go run main.go


Flags:
  -e, --exclude strings   exclude path. you can specify multiple it
  -h, --help              help for watch
  -p, --path string       watch path (default "**/*")

Global Flags:
  -v, --verbose   enable verbose
```

### Example

The simplest way is to specify `recmd watch` followed by your command.
In this case, it watches change in the files under the current directory.

```sh
$ recmd watch go run cmd/http-server/main.go
```

If you want to specify a file to watch change, you specify `--path`.

```sh
$ recmd watch \
	--path "**/*.go"
	go run cmd/http-server/main.go
```

If you want to exclude to watch file, you specify `--exclude`.
This option can be specified multiple.

```sh
$ recmd watch \
	--path "**/*.go"
	--exclude "vendor/**/*.go"
	--exclude "tmp"
	go run cmd/http-server/main.go
```