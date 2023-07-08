# List of variants
VARIANTS = minimal k8s tools

# List of target squashfs images
SQUASHFS_FILES = $(patsubst %,rootfs-%.squashfs,$(VARIANTS))

ROOTFS_ARCHIVE_URL="https://pub-1a5aeef625fc45b4a4bef89ee141047f.r2.dev/debian-bookworm-systemd-rootfs.tar.gz"

all: $(SQUASHFS_FILES)

rootfs-%.squashfs: artifacts/firework-agent
	mkdir -p /tmp/systemd-rootfs-$*-squashfs
	sudo buildctl build --frontend=dockerfile.v0 \
        --local context=. \
        --local dockerfile=. \
		--output type=tar \
        --opt "filename=firework-dev/Dockerfile.$*" | sudo tar -C /tmp/systemd-rootfs-$*-squashfs -xf -

	sudo mksquashfs /tmp/systemd-rootfs-$*-squashfs $@ -noappend

	rm -f /tmp/systemd-rootfs-$*-squashfs.tar
	rm -rf /tmp/systemd-rootfs-$*-squashfs

cleanup:
	rm -f /tmp/systemd-rootfs-*-squashfs.tar
	rm -rf /tmp/systemd-rootfs-*-squashfs
	rm -f rootfs-*.squashfs

install: firework-rootfs-dir
	sudo cp rootfs-*.squashfs /var/lib/firework/rootfs

firework-rootfs-dir:
	sudo mkdir -p /var/lib/firework/rootfs

artifacts/firework-agent: artifacts
	cargo build -p fwagent --release
	cp target/release/fwagent artifacts/firework-agent

artifacts:
	mkdir -p artifacts

.PHONY: all cleanup install firework-rootfs-dir

