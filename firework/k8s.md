# Building squashfs images with kubeadm pre-installed

- Make sure `SystemdCgroup` in `runtimes.runc.options` section of `containerd` config is set to `true`.
- Set `snapshotter` in `plugins."io.containerd.grpc.v1.cri".containerd` section of `containerd` config to `native`, because default is `overlayfs` and nested overlayfs does not work (squashfs images are mounted as overlayfs upper directories).

# Running VMs with kubeadm pre-installed

## Initialize cluster

```sh
# Ignore kernel module check because the required modules are guaranteed to be built in the kernel. (TODO (jlkiri): Check if more granular ignoring is possible)
kubeadm init --ignore-preflight-errors SystemVerification
```

## Join

```sh
# Ignore kernel module check because the required modules are guaranteed to be built in the kernel. (TODO (jlkiri): Check if more granular ignoring is possible)
kubeadm join --ignore-preflight-errors SystemVerification <control plane's IP> --token <token> --discovery-token-ca-cert-hash <hash>
```

To join a node later, the join command can be issued with:
```sh
kubeadm token create --print-join-command
```

## Memo
- When running `nerdctl` inside the VM make sure the `--snapshotter` flag is explicitly set to `native`.

## Configure kubectl
```sh
mkdir -p $HOME/.kube
cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
chown $(id -u):$(id -g) $HOME/.kube/config
```