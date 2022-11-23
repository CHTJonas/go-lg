# Go-Looking Glass

go-lg is a web-based network looking glass that's written in Go. Like other looking glasses it allows you to ping remote hosts and perform traceroutes. Unlike other looking glasses it also produces reports with a unique persistent URL that can be shared with your colleagues or friends.

go-lg is designed to be a 'simple' looking glass application and is intended to be run on an end-host machine. It's not designed to be a fully-fledged network looking glass with BGP functionality that you might run on a route server in an internet exchange point, for example.

## Usage

go-lg is compiled into a single static binary and there is minimal configuration to worry about. All reports are stored in an embedded [BadgerDB](https://github.com/dgraph-io/badger) database and web assets are compiled into the binary itself. It's expected that go-lg will be run behind a reverse proxy such as nginx which will take care of things like TLS.

```
Usage:
  go-lg server [flags]

Flags:
  -b, --bind string     address and port to bind to (default "localhost:8080")
  -h, --help            help for server
  -p, --path string     path to database storage directory (default "/var/lib/go-lg")
```

## Compiling

It should be relatively simple to checkout and build the code, assuming you have a suitable [Go toolchain installed](https://golang.org/doc/install). Running the following commands in a terminal will compile binaries for various operating systems and processor architectures and place them in `./bin`:

```bash
git clone https://github.com/CHTJonas/go-lg.git
cd go-lg
make clean && make all
```

## Copyright

go-lg is licensed under the [BSD 2-Clause License](https://opensource.org/licenses/BSD-2-Clause).

Copyright (c) 2020–2022 Charlie Jonas.
