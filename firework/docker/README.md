```
sudo nerdctl build -t fc .
sudo nerdctl run -it \
    -v $(pwd):/firework \
    -v /var/lib/firework:/var/lib/firework \
    --device /dev/net/tun:/dev/net/tun \
    --device /dev/kvm:/dev/kvm \
    --cap-add=NET_ADMIN \
    fc
```