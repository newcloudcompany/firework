## Development requirements

### Common
1. Linux Kernel version 5 or higher.
2. `sudo` privileges.

### Building and running the binary
1. [Firecracker](https://github.com/firecracker-microvm/firecracker) v1.3.3 or higher installed at `/bin/firecracker`.
2. CPU that supports virtualization. To check (on Debian-based OS):

```
$ apt install -y cpu-checker
$ kvm-ok
```

3. `gcc` available (probably better install `build-essential` package) in `$PATH`. This is used for `CGO_ENABLED` build because the `firework` binary depends on `sqlite`.
4. `go` binary available in `$PATH` (TODO (jlkiri): containerize builds).
5. `16GB` of RAM would be good for running a multi-VM cluster.

### Building the base rootfs and customized squashfs images
1. `containerd` installed and daemon running.
2. [`buildctl`](https://github.com/moby/buildkit) installed and `buildkitd` daemon running.
3. Rust toolchain installed (`rustc`, `cargo`). This is to build the `firework` agent and the packages it depends on (`ptyca`).
4. `mksquashfs` available in `$PATH`.