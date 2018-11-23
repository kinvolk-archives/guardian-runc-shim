# RunC Wrapper

A wrapper around runc which allows altering [spec files][1].

The main use case for this wrapper is to be called by a [Concourse CI][2] worker (or - more
specifically - [Guardian][3]) instead of the actual `runc` executable. The wrapper can then perform
some desired customizations to the runc runtime spec file (namely `config.json`) and then call the
"real" `runc` binary to continue the normal flow.

The wrapper calls `runc` using the `execve(2)` syscall without forking, so that the actual runc
process replaces the wrapper and becomes a direct child of the process which invoked the wrapper.

## Requirements

- Linux
- Go 1.11.2 or above

## Downloading

A pre-compiled Linux binary may be downloaded from the [releases][5] page. Alternatively, you may
build the wrapper from source.

## Building from Source

```shell
git clone https://github.com/kinvolk/runc-wrapper.git
cd runc-wrapper
go build
```

Dependencies are handled automatically using [Go modules][4].

A binary named `runc-wrapper` will be created in the current directory.

## Running

Extract (or build) the binary and put it somewhere, for example `/usr/local/bin`.

Set the `RUNC_WRAPPER_BINARY` environment variable to point at the "real" `runc` binary which the
wrapper should execute after doing its thing. The wrapper executable can then be used as if it were
the real `runc` binary.

### Example

```shell
# Just for brevity
export BIN_PATH=/var/lib/concourse/<version>/assets/bin
```

1. Extract the wrapper to `/usr/local/bin/runc-wrapper`.
2. Rename the original `runc` binary at `$BIN_PATH/runc` to
`runc-real`.
3. Symlink the wrapper instead of the original binary:
`ln -s /usr/local/bin/runc-wrapper $BIN_PATH/runc`.
4. Set `RUNC_WRAPPER_BINARY` for the Concourse worker process to `$BIN_PATH/runc-real`.

## Logging

By default, the wrapper writes logs to `/var/log/runc-wrapper`. The log file path may be changed by
setting the `RUNC_WRAPPER_LOGFILE` environment variable.

The wrapper does **not** log anything to stdout/stderr on purpose. This is because Concourse relies
on the stdout/stderr of the actual `runc` binary, which should be passed unchanged to the parent
process. This behavior should not be modified.

[1]: https://github.com/opencontainers/runtime-spec/blob/master/config.md
[2]: https://concourse-ci.org/
[3]: https://github.com/cloudfoundry/guardian
[4]: https://github.com/golang/go/wiki/Modules
[5]: https://github.com/kinvolk/runc-wrapper/releases
