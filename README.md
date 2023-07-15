# Firework

A repository for `firework` - a collection of resources, tools and libraries for launching networked clusters of Firecracker microVMs.

## Development requirements

### Common
1. Linux Kernel version 5 or higher.
2. `sudo` privileges.

### Building and running the binary
1. [Firecracker](https://github.com/firecracker-microvm/firecracker) `v1.3.3` or higher installed at `/bin/firecracker`.
2. CPU that supports virtualization. To check (on Debian-based OS):

```
$ apt install -y cpu-checker
$ kvm-ok
```

3. `gcc` available in `$PATH`. Install `build-essential` package. This is used for `CGO_ENABLED=1` build because the `firework` CLI binary depends on `sqlite`.
4. `go` binary available in `$PATH` (TODO (jlkiri): containerize CLI builds).
5. `16GB` of RAM would be good for running a multi-VM cluster.

### Building the base rootfs and customized squashfs images
1. `containerd` runtime installed and daemon running.
2. [`buildctl`](https://github.com/moby/buildkit) installed and `buildkitd` daemon running.
3. Rust toolchain installed (`rustc`, `cargo`). This is to build the `firework` agent (`fwagent`) and the packages it depends on (`ptyca`).
4. `mksquashfs` tool available in `$PATH`.

## Build base rootfs
Use Debian `bookworm` minimal base as base rootfs which serves as the base for custom squashfs images. To create a directory that contains the Debian rootfs, run [`firework-dev/debootstrap.sh`](firework-dev/debootstrap.sh) script. This takes about 7 minutes.

Because the squashfs image Dockerfiles download the rootfs from R2, it needs to be uploaded again in case you want to update something. The best way to do it is to use `aws-cli` with configured credentials for R2.

Example:

```sh
aws s3api put-object --endpoint-url https://7d9bec7ddc058a107bbd85fd4f8cc6d6.r2.cloudflarestorage.com --bucket firework --key debian-bookworm-systemd-rootfs.tar.gz --body docker/debian-bookworm-rootfs.tar.gz
```

TODO (jlkiri): currently this points to `jlkiri`'s individual Cloudflare account so other people or CI cannot have an access key.

## Kernel

Firecracker VM require a modern Linux kernel and the Firecracker project provides minimal recommended kernel configs and versions. For this project we use the longterm 5.x.x release of the kernel which is `5.15.120` at the time of writing.

To build kernel:

1. Install required dependencies
```
sudo apt-get install git fakeroot build-essential ncurses-dev xz-utils libssl-dev bc flex libelf-dev bison
```
2. Clone the linux repository and checkout the release.

```sh
git clone --depth 1 --branch v5.15 git://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git linux-5.15.120
```

3. Copy the config from [firework-dev/kernel/.config](./firework-dev/kernel/.config) to the working directory (root of the Linux repo)
4. Run `make menuconfig` and add new parametes if needed. Save.
5. Run `make -j<cpu>` where `<cpu>` is the number of cores on your machine.
6. Upload the `vmlinux` file to the shared object storage.

```sh
aws s3api put-object --endpoint-url https://7d9bec7ddc058a107bbd85fd4f8cc6d6.r2.cloudflarestorage.com --bucket firework --key vmlinux --body vmlinux
```

TODO (jlkiri): currently this points to `jlkiri`'s individual Cloudflare account so other people or CI cannot have an access key.

### Image specific kernel config

Some squashfs images that we build like the ones the come with `kubeadm` pre-installed require more parameters to be set than in the recommended official config (TODO (jlkiri): distribute kernels in pairs with the squashfs images that require it?). For example, every tool in the k8s has its own requirements and while the tool specific requirements are lost in the history of experiments, here's the diff between the end result for `kubeadm` & `cilium` and the minimal officially recommended config:

