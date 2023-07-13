## Misc

### Build base Debian rootfs
```
debootstrap --arch=amd64 --variant=minbase bullseye ./debian-bullseye http://deb.debian.org/debian/
```

TODO: Use SDK-provided vsock client

```
// g := new(errgroup.Group)

	// machines := []*firecracker.Machine{}
	// for _, node := range config.Nodes {
	// 	name := node.Name
	// 	g.Go(func() error {
	// 		m, err := runVmm(ctx, name, bridge, ipamDb)
	// 		if err != nil {
	// 			return fmt.Errorf("failed to run VMM: %w", err)
	// 		}
	// 		defer func() {
	// 			if err := m.StopVMM(); err != nil {
	// 				log.Println("An error occurred while stopping Firecracker VMM: ", err)
	// 			}
	// 		}()

	// 		machines = append(machines, m)

	// 		// wait for the VMM to exit
	// 		if err := m.Wait(ctx); err != nil {
	// 			log.Println("An error occurred while waiting for the Firecracker VMM to exit: ", err)
	// 		}

	// 		return nil
	// 	})
	// }
```


```
update-alternatives --set iptables /usr/sbin/iptables-legacy
```

## k8s memos
* Make sure SystemdCgroup in runtimes.runc.options is set to true
* Set snapshotter in  plugins."io.containerd.grpc.v1.cri".containerd to "native", because default is overlayfs and nested overlayfs does not work
* When running nerdctl make sure the --snapshotter is explicitly set to native

## TODO:

* Remove udev from pre-installed pkgs

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