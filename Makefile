all: rootfs.squashfs

rootfs.alp.squashfs: artifacts/firework-agent artifacts/init
	sudo mkdir -p /tmp/rootfs-squashfs
	# sudo cp -r firework-dev/alpine-rootfs/* /tmp/rootfs-squashfs
	sudo firework-dev/alpine-make-rootfs.sh --branch 3.18 /tmp/rootfs-squashfs

	sudo cp artifacts/firework-agent /tmp/rootfs-squashfs/firework-agent
	sudo cp artifacts/init /tmp/rootfs-squashfs/init
	sudo cp firework-dev/overlay-init /tmp/rootfs-squashfs/sbin/overlay-init

	echo "nameserver 8.8.8.8" | sudo tee /tmp/rootfs-squashfs/etc/resolv.conf
	
	sudo mkdir /tmp/rootfs-squashfs/mnt
	sudo mkdir /tmp/rootfs-squashfs/rom
	sudo mkdir /tmp/rootfs-squashfs/overlay

	sudo mksquashfs /tmp/rootfs-squashfs $@ -noappend

	sudo rm -rf /tmp/rootfs-squashfs

rootfs.deb.squashfs: artifacts/firework-agent
	sudo mkdir -p /tmp/rootfs-squashfs
	sudo cp -r firework-dev/debian-bookworm-rootfs/* /tmp/rootfs-squashfs

	sudo cp artifacts/firework-agent /tmp/rootfs-squashfs/firework-agent
	sudo cp firework-dev/overlay-init /tmp/rootfs-squashfs/sbin/overlay-init
	sudo cp firework-dev/init /tmp/rootfs-squashfs/sbin/init

	# sudo mkdir /tmp/rootfs-squashfs/mnt
	sudo mkdir /tmp/rootfs-squashfs/rom
	sudo mkdir /tmp/rootfs-squashfs/overlay

	sudo mksquashfs /tmp/rootfs-squashfs $@ -noappend

	sudo rm -rf /tmp/rootfs-squashfs

cleanup:
	sudo rm -f rootfs*.squashfs
	sudo rm -rf /tmp/rootfs-squashfs

install: firework-rootfs-dir rootfs.alp.squashfs
	# sudo cp rootfs.deb.squashfs /var/lib/firework/rootfs
	sudo cp rootfs.alp.squashfs /var/lib/firework/rootfs

firework-rootfs-dir:
	sudo mkdir -p /var/lib/firework/rootfs

artifacts/firework-agent: artifacts
	cargo build --target x86_64-unknown-linux-musl -p fwagent --release
	cp target/x86_64-unknown-linux-musl/release/fwagent artifacts/firework-agent

artifacts/init: artifacts
	cargo build --target x86_64-unknown-linux-musl -p init --release
	cp target/x86_64-unknown-linux-musl/release/init artifacts/init

artifacts:
	mkdir -p artifacts

.PHONY: all cleanup install firework-rootfs-dir

