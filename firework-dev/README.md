## Misc

### Build base Debian rootfs
```
debootstrap --arch=amd64 --variant=minbase bullseye ./debian-bullseye http://deb.debian.org/debian/
```

```
buildctl build --frontend=dockerfile.v0 \
    --local context=. \
    --local dockerfile=. \
    --output type \
    --opt build-arg:foo=bar
```

```
aws s3api list-buckets --endpoint-url https://7d9bec7ddc058a107bbd85fd4f8cc6d6.r2.cloudflarestorage.com
```

```
aws s3api put-object --endpoint-url https://7d9bec7ddc058a107bbd85fd4f8cc6d6.r2.cloudflarestorage.com --bucket firework --key debian-bookworm-systemd-rootfs.tar.gz --body docker/debian-bookworm-rootfs.tar.gz
```

## k8s specific kernel config

Cilium, and likely Weave-Net

```
CONFIG_NETFILTER_XT_SET=m
CONFIG_IP_SET=m
CONFIG_IP_SET_HASH_IP=m
```

Probably `y` should work too.