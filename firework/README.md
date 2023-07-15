# Firework

Firework is a tool for launching networked clusters of Firecracker microVMs locally.

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
- appropriate `iptables` rules are inserted to enable traffic between the VMs and from the VMs to the Internet and back
- a sparse file with capacity in `disk` is created with `truncate` to be attached as non-root block device for each VM

Every VM node configuration must include a number of `vcpu`s, memory in megabytes, `disk` capacity in units acceptable by `truncate` and an absolute path to `squashfs` image of rootfs. The image must have an init system installed. init can be anything but `systemd` is a good choice. For quick start, here is an image with `systemd` as init as kubeadm pre-installed: https://pub-1a5aeef625fc45b4a4bef89ee141047f.r2.dev/rootfs-k8s.squashfs

### firework stop

Gracefully stops all VMs in the cluster and undoes what `firework start` does. Cleans up created resources, and network configuration (`iptables`).

### firework status

Prints a table of VM statuses. Each entry has a unique VMID, IP address and status which can be `Running` or `Not Running`.

### firework logs

Prints aggregated logs of a Virtual Machine Monitor (VMM) of each VM. These are messages about VM's status, some other metadata.

TODO: In a client CLI implementation just use `isatty` and poll the endpoint for logs to keep the worker agent stateless.

### firework logs \<VMID\>

Prints logs of an individual VM with VMID (which can be looked up with `firework status`). This is where the `stdout` and `stderr` of the init goes.

### firework connect

Allocates a TTY and creates a session with remote shell through VSOCK connection to an `firework` agent running inside a VM. This requires `firework` agent to be pre-installed in the `squashfs` rootfs image.
