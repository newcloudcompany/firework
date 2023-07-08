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