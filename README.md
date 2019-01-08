# Guardian runC Shim

A tiny shim for [Guardian][3] which allows altering [runC][7] [spec files][1].

## UPDATE: You Might Not Need This!

There are ways to accomplish the same thing the code in this repository does without interfering
with the way Concourse operates. For example, the following Concourse pipeline configuration can be
used to modify a task's environment *from within a privileged container*, in this case in order to
run a KVM virtual machine inside a Concourse task:

```bash
mkdir /tmp/devices-cgroup
mount -t cgroup -o devices none /tmp/devices-cgroup
echo 'c 10:232 rwm' > /tmp/devices-cgroup/system.slice/${concourse_systemd_unit}/$(hostname)/devices.allow
mknod -m 0660 /dev/kvm c 10 232
chown root:kvm /dev/kvm

mkdir /dev/net
mknod -m 0600 /dev/net/tun c 10 200

mount -o rw,remount /sys

service libvirtd start
service virtlogd start
```

It is more stable, secure and convenient to keep the required customizations within the pipeline
configuration itself. If, however, the customization you require can't be accomplished by altering
the pipeline you may benefit from using this code (with relevant modifications) to build a shim
which can perform your desired changes whenever a Concourse task runs.

## Why?

[Concourse CI][2] runs tasks inside containers. It uses an API called [Garden][6] to accomplish
this, with a runC-based implementation called [Guardian][3] that ships with Concourse by default.

Concourse spawns containers with a specific [OCI spec][8] by default. This spec is quite rigid and
could prevent certain operations from running in a Concourse task. One example is running
**virtual machines** as part of a CI pipeline. Setting a task step as [privileged][9] might not be
enough: certain tasks might require special Linux devices to be created for the task's container,
or fine-grained capabilities to be set.

At the time of writing, Concourse and Garden don't support fine-grained OCI spec customizations. As
a result, certain tasks cannot be accomplished using Concourse. Examples:

- https://github.com/concourse/concourse/issues/1905
- https://github.com/concourse/concourse/issues/2784

Additionally, some use cases might require hardening Concourse task containers beyond the default
spec for security reasons.

This project was created to allow performing customizations to the OCI spec used by Concourse with
the hopes that such customizations would be supported directly in Concourse in the future.

## How?

This project is a tiny wrapper around the `runc` executable. Guardian normally interacts with
`runc` directly to manage containers. By replacing the `runc` binary with Guardian runC Shim it is
possible to intercept calls to `runc`, perform arbitrary operations and finally call the actual
`runc` binary.

The shim calls `runc` using the `execve(2)` system call **without forking**, so that the actual
runc process *replaces* the shim and becomes a direct child of the process which invoked the shim
in the first place (i.e the Concourse worker).

## Requirements

- Linux
- Go 1.11.2 or above

## Downloading

A pre-compiled Linux binary may be downloaded from the [releases][5] page. Alternatively, you may
build the project from source.

## Building from Source

```shell
git clone https://github.com/kinvolk/guardian-runc-shim.git
cd guardian-runc-shim
go build
```

Dependencies are handled automatically using [Go modules][4].

A binary named `guardian-runc-shim` will be created in the current directory.

## Running

Extract (or build) the binary and put it somewhere, for example `/usr/local/bin`.

Set the `GUARDIAN_RUNC_SHIM_BINARY` environment variable to point at the "real" `runc` binary which
the shim should execute after doing its thing. The shim executable can then be used as if it were
the real `runc` binary.

### Example

```shell
# Just for brevity
export BIN_PATH=/var/lib/concourse/<version>/assets/bin
```

1. Extract the shim to `/usr/local/bin/guardian-runc-shim`.
2. Rename the original `runc` binary at `$BIN_PATH/runc` to `runc-real`.
3. Symlink the shim instead of the original binary:
`ln -s /usr/local/bin/guardian-runc-shim $BIN_PATH/runc`.
4. Set `GUARDIAN_RUNC_SHIM_BINARY` for the Concourse worker process to `$BIN_PATH/runc-real`.

## Logging

By default, the shim writes logs to `/var/log/guardian-runc-shim.log`. The log file path may be
changed by setting the `GUARDIAN_RUNC_SHIM_LOGFILE` environment variable.

The shim does **not** log anything to stdout/stderr on purpose. This is because Concourse relies
on the stdout/stderr of the actual `runc` binary, which should be passed unchanged to the parent
process. This behavior should not be modified.

## Known Issues

### Can't run non-privileged workloads

It seems to be impossible to run non-privileged containers when using the shim. When Concourse
attempts to start the container, the following error is returned from runc:

```
runc run: exit status 1: container_linux.go:348: starting container process caused "process_linux.go:402: container init caused \"process_linux.go:367: setting cgroup config for procHooks process caused \\\"failed to write a *:* rwm to devices.allow: write /sys/fs/cgroup/devices/system.slice/concourse-worker.service/garden/200e7e87-3af9-4a79-4ea5-5ca0364c6cc8/devices.allow: operation not permitted\\\"\""
```

[1]: https://github.com/opencontainers/runtime-spec/blob/master/config.md
[2]: https://concourse-ci.org/
[3]: https://github.com/cloudfoundry/guardian
[4]: https://github.com/golang/go/wiki/Modules
[5]: https://github.com/kinvolk/guardian-runc-shim/releases
[6]: https://github.com/cloudfoundry/garden
[7]: https://github.com/opencontainers/runc
[8]: https://github.com/opencontainers/runtime-spec
[9]: https://concourse-ci.org/task-step.html#task-step-privileged
