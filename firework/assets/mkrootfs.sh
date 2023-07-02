#!/usr/bin/env bash

set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $script_dir

rootfs_base="debian-bullseye-rootfs"
squashfs_img="rootfs.squashfs"
packages="procps iproute2 ca-certificates curl dnsutils iptables iputils-ping cpu-checker git"

# mkdir -p "$rootfs_base"

function install_additional_tools {
    mkdir -p tmp

    # Install firecracker
    local fc_release="firecracker-v1.3.3-x86_64"
    curl -o "tmp/$fc_release.tgz" -L "https://github.com/firecracker-microvm/firecracker/releases/download/v1.3.3/$fc_release.tgz"
    tar -xvf "tmp/$fc_release.tgz" -C tmp
    cp "tmp/release-v1.3.3-x86_64/$fc_release" "$rootfs_base/usr/bin/firecracker"

    # Install golang
    local go_release="go1.20.5.linux-amd64"
    curl -o "tmp/$go_release.tar.gz" -L "https://go.dev/dl/$go_release.tar.gz"
    tar -C "$rootfs_base/usr/local" -xzf "tmp/$go_release.tar.gz"

    echo "export PATH=$PATH:/usr/local/go/bin" >> "$rootfs_base/etc/profile"
    rm -rf tmp
}

function debootstrap_rootfs {
    if [[ ! -e "$rootfs_base" ]]; then
        echo "Creating debian bullseye rootfs..."
        debootstrap \
            --arch=amd64 \
            --variant=minbase \
            --include=${packages// /,} \
            bullseye "$rootfs_base" \
            http://deb.debian.org/debian/    
    fi
    
    echo "Installing init in the rootfs..."
    mkdir -p "$rootfs_base/overlay" "$rootfs_base/mnt" "$rootfs_base/rom"
    cp init "$rootfs_base/sbin/init"
    cp overlay-init "$rootfs_base/sbin/overlay-init"
    cp .vimrc "$rootfs_base/root/.vimrc"
    # cp ../firework "$rootfs_base/usr/bin/firework"
    cp ../config_vm.json "$rootfs_base/config.json"
    

    echo "Installing additional tools in the rootfs..."
    install_additional_tools

    echo "Performing additional configuration..."
    chroot "$rootfs_base" /bin/bash -c "echo \"net.ipv4.conf.all.forwarding = 1\" >> /etc/sysctl.conf"
    chroot "$rootfs_base" /bin/bash -c "update-alternatives --set iptables /usr/sbin/iptables-legacy"
}

function mkroot_squashfs {
    if [[ ! -e "$squashfs_img" ]]; then
        echo "Creating squashfs image of the debian rootfs..."
        mksquashfs "$rootfs_base" "$squashfs_img" -noappend
    fi
}

time debootstrap_rootfs
time mkroot_squashfs

