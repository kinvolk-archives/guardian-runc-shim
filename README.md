# RunC Wrapper

A wrapper around runc which allows altering [spec files][1].

The main use case for this wrapper is to be called by a [Concourse CI][2] worker (or - more
specifically - [Guardian][3]) instead of the actual `runc` executable. The wrapper can then perform
some desired customizations to the runc runtime spec file (namely `config.json`) and then call the
"real" `runc` binary to continue the normal flow.

[1]: https://github.com/opencontainers/runtime-spec/blob/master/config.md
[2]: https://concourse-ci.org/
[3]: https://github.com/cloudfoundry/guardian
