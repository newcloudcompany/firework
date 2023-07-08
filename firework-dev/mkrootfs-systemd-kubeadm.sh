#!/usr/bin/env bash

set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $script_dir

rootfs_base="debian-bookworm-rootfs-systemd"
squashfs_img="rootfs.squashfs"
packages="procps iproute2 ca-certificates curl dnsutils iptables iputils-ping cpu-checker git systemd systemd-sysv udev gnupg"

# mkdir -p "$rootfs_base"

function install_vm_tools {
    mkdir -p tmp

    # Install firecracker
    local fc_release="firecracker-v1.3.3-x86_64"
    curl -o "tmp/$fc_release.tgz" -L "https://github.com/firecracker-microvm/firecracker/releases/download/v1.3.3/$fc_release.tgz"
    tar -xvf "tmp/$fc_release.tgz" -C tmp
    cp "tmp/release-v1.3.3-x86_64/$fc_release" "$rootfs_base/usr/bin/firecracker"

    rm -rf tmp
}

function debootstrap_rootfs {
    if [[ ! -e "$rootfs_base" ]]; then
        echo "Creating debian bookworm rootfs..."
        debootstrap \
            --arch=amd64 \
            --variant=minbase \
            --include=${packages// /,} \
            bookworm "$rootfs_base" \
            http://deb.debian.org/debian/    
    fi
    
    echo "Installing overlay-init-systemd and fwagent in the rootfs..."
    mkdir -p "$rootfs_base/overlay" "$rootfs_base/mnt" "$rootfs_base/rom"
    cp overlay-init-systemd "$rootfs_base/sbin/overlay-init-systemd"

    cp fwagent.service "$rootfs_base/etc/systemd/system/fwagent.service"
    ln -sf "$rootfs_base/etc/systemd/system/fwagent.service" "$rootfs_base/etc/systemd/system/multi-user.target.wants/fwagent.service"
    cp fwagent "$rootfs_base/usr/bin/fwagent"

    rm -rf "$rootfs_base/etc/systemd/system/getty.target.wants/"

    echo "Installing vm tools in the rootfs..."
    install_vm_tools

    echo "Performing additional configuration..."
    chroot "$rootfs_base" /bin/bash -c "update-alternatives --set iptables /usr/sbin/iptables-legacy"
    chroot "$rootfs_base" /bin/bash -c "update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy"

    chmod +x "$script_dir/install_kubeadm_prerequisites.sh"
    cp "$script_dir/install_kubeadm_prerequisites.sh" "$rootfs_base/install_kubeadm_prerequisites.sh"
    cp "$script_dir/containerd.service" "$rootfs_base/etc/systemd/system/containerd.service"
    mkdir -p "$rootfs_base/etc/containerd"
    cp "$script_dir/config.toml" "$rootfs_base/etc/containerd/config.toml"
    chroot "$rootfs_base" /install_kubeadm_prerequisites.sh
}

function mkroot_squashfs {
    if [[ ! -e "$squashfs_img" ]]; then
        echo "Creating squashfs image of the debian rootfs..."
        mksquashfs "$rootfs_base" "$squashfs_img" -noappend
    fi
}

time debootstrap_rootfs
time mkroot_squashfs

