#!/usr/bin/env bash

set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $script_dir

source vars.sh

mount_dir="$script_dir/rootfs"
rootfs="$rootfs_dir/rootfs.ext4"
base_rootfs="$downloads_dir/rootfs.tar.xz"
tmp_rootfs="$script_dir/tmp_rootfs"

rm -rf "$rootfs" || true

function cleanup {
    # Unmount the disk image and remove the temporary mount directory   
    {
        umount --lazy "$mount_dir" &> /dev/null || true
        rm -rf "$mount_dir"
        rm -rf "$tmp_rootfs"
    } || true
}

cleanup
trap cleanup EXIT

mkdir -p "$rootfs_dir"

# Create an empty file
echo "Allocating an empty 512MB file..."
truncate -s 512MB "$rootfs"

# Create an ext4 filesystem on the file
echo "Creating an ext4 filesystem on the file..."
mkfs -t ext4 "$rootfs"

mkdir -p $mount_dir
mount -o loop "$rootfs" $mount_dir

echo "Unpacking the base rootfs image to the mount dir..."
tar -xf "$base_rootfs" -C "$mount_dir"

echo "Pre-installing programs in the base rootfs image with Docker..."

# img_id=$(docker build . -t rootfs)
# container_id=$(docker run --rm --tty --detach rootfs /bin/bash)

# echo "Copying the contents of the container back to the rootfs..."
# docker cp $container_id:/ $mount_dir

# ssh-keygen -t ed25519 -C "firecracker" -f "$tmp_ssh_key_dir/$vm_id" -q -N "" <<<y &>/dev/null
# cp "$tmp_ssh_key_dir/$vm_id.pub" .

mkdir -p "$tmp_rootfs"
buildctl build --frontend=dockerfile.v0 --local context=. --local dockerfile=./vm --output type=local,dest="$tmp_rootfs"

cp --remove-destination -r "$tmp_rootfs"/* "$mount_dir"

echo "Done."
