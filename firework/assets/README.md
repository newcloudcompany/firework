## Misc

### Build base Debian rootfs
```
debootstrap --arch=amd64 --variant=minbase bullseye ./debian-bullseye http://deb.debian.org/debian/
```