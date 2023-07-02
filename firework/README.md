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