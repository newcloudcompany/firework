# Firework

Firework is a tool for launching networked clusters of Firecracker microVMs locally.

## Development requirements

1. Linux Kernel version 5 or higher
2. [Firecracker](https://github.com/firecracker-microvm/firecracker) v1.3.3 or higher installed at `/bin/firecracker`.
3. CPU that supports virtualization. To check (on Debian-based OS):

```
$ apt install -y cpu-checker
$ kvm-ok
```

## Usage

```
A tool for launching local Firecracker VM clusters

Usage:
  firework [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  connect     Connect to a VM
  help        Help about any command
  logs        View VMM logs or logs of a running VM
  start       Start a VM cluster from config
  status      View status of running VMs
  stop        Stop a VM cluster from config

Flags:
  -h, --help   help for firework

Use "firework [command] --help" for more information about a command.
```

### firework start

Use to launch a cluster of Firecracker microVMs. The configuration is read from a `config.json` file that must exist in the working directory. Here's the example configuration:

```json
{
    "subnet_cidr": "172.18.0.240/28",
    "gateway": "172.18.0.241/28",
    "nodes": [
        {
            "name": "ctrl",
            "vcpu": 2,
            "memory": 8192,
            "rootfs_path": "/var/lib/firework/rootfs/rootfs-k8s.squashfs",
            "disk": "32G"
        },
        {
            "name": "worker-1",
            "vcpu": 2,
            "memory": 4096,
            "rootfs_path": "/var/lib/firework/rootfs/rootfs-k8s.squashfs",
            "disk": "32G"
        }
    ]
}
```

`firework` first creates a bridge network and an IP address database with addresses in the `subnet_cidr`. Then it checks whether a Linux kernel is available locally, and if not,`firework` downloads it  to `/var/lib/firework`, which also stores runtime VM files and logs.

Every time a cluster is created with `start`:
- a TAP network interface is created for each VM
- a free IP address is allocated to each VM from the database
- appropriate `iptables` rules are inserted to enable traffic between the VMs and from the VMs to the Internet
- a sparse file with capacity in `disk` is created to be attached as non-root block device for each VM

Every VM node configuration must include a number of `vcpu`s, memory in megabytes, `disk` capacity and an absolute path to `squashfs` image of rootfs. The image must have an init system installed. init can be anything but `systemd` is a good choice. For quick start, here is an image with `systemd` as init as kubeadm pre-installed: https://pub-1a5aeef625fc45b4a4bef89ee141047f.r2.dev/rootfs-k8s.squashfs

### firework stop

Undoes what `firework start` does. Cleans up created resources, and network configuration (`iptables`).

### firework status

Prints a table of VMs with their unique IDs, IPv4 address and status which can be `Running` or `Not Running`.

### firework logs

Prints aggregated logs of a Virtual Machine Monitor of each VM. These are messages about VM's status, some other metadata.

### firework logs \<VMID\>

Prints logs of an individual VM with VMID (which can be looked up with `firework status`).

### firework connect

Creates a session with remote shell using VSOCK connection to an `firework` agent running inside a VM. This requires `firework` agent to be pre-installed in the `squashfs` rootfs.




## Misc

### Build base Debian rootfs
```
debootstrap --arch=amd64 --variant=minbase bullseye ./debian-bullseye http://deb.debian.org/debian/
```


## k8s memos
* Make sure SystemdCgroup in runtimes.runc.options is set to true
* Set snapshotter in  plugins."io.containerd.grpc.v1.cri".containerd to "native", because default is overlayfs and nested overlayfs does not work
* When running nerdctl make sure the --snapshotter is explicitly set to native


## Cmd

```
# Ignore kernel module check
kubeadm init --ignore-preflight-errors SystemVerification
```

```
kubeadm join --ignore-preflight-errors SystemVerification 10.0.0.242:6443 --token kjordx.wajkd8rn2cb5zgs4 \
        --discovery-token-ca-cert-hash sha256:840e4779c5c215d1e78b05883634386104649ceb12dd36483a5b46f1126b94c4
```

```
mkdir -p $HOME/.kube
cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
chown $(id -u):$(id -g) $HOME/.kube/config
```