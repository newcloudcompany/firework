# List of variants
VARIANTS = minimal k8s tools

# List of target squashfs images
SQUASHFS_FILES = $(patsubst %,rootfs-%.squashfs,$(VARIANTS))

ROOTFS_ARCHIVE_URL="https://pub-1a5aeef625fc45b4a4bef89ee141047f.r2.dev/debian-bookworm-systemd-rootfs.tar.gz"
ROOTFS_ARCHIVE_PATH="debian-bookworm-systemd-rootfs.tar.gz"

all: $(SQUASHFS_FILES)

# Download debian-bookworm-systemd-rootfs.tar.gz if not exists
debian-bookworm-systemd-rootfs.tar.gz:
	curl -o $@ -L $(ROOTFS_ARCHIVE_URL)

# Define rules for squashfs files
rootfs-%.squashfs:
	mkdir -p /tmp/systemd-rootfs-$*-squashfs

	buildctl build --frontend=dockerfile.v0 \
        --local context=$* \
        --local dockerfile=. \
		--output type=tar \
        --opt "filename=$*/Dockerfile.$*" | tar -C /tmp/systemd-rootfs-$*-squashfs -xf -

	mksquashfs /tmp/systemd-rootfs-$*-squashfs $@ -noappend
	
	rm -f /tmp/systemd-rootfs-$*-squashfs.tar
	rm -rf /tmp/systemd-rootfs-$*-squashfs

cleanup:
	rm -f /tmp/systemd-rootfs-*-squashfs.tar
	rm -rf /tmp/systemd-rootfs-*-squashfs
	rm -f rootfs-*.squashfs

.PHONY: all cleanup

