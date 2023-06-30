#!/usr/bin/env bash

set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $script_dir

mount_dir="$script_dir/rootfs"

output_dir="$script_dir"
output_ext4_path="$output_dir/rootfs.ext4"
output_gzip_path="$output_dir/rootfs.ext4.gz"

rm -rf "$mount_dir" || true

function cleanup {
    # Unmount the disk image and remove the temporary mount directory   
    {
        umount --lazy "$mount_dir" &> /dev/null || true
        rm -rf "$mount_dir"
    } || true
}

cleanup
trap cleanup EXIT

# Create an empty file
truncate -s 2G "$output_ext4_path" &> /dev/null
echo "Allocated an empty 2GB file..."

# Create an ext4 filesystem on the file
mkfs -t ext4 "$output_ext4_path" &> /dev/null
echo "Created an ext4 filesystem on the file..."

mkdir -p $mount_dir
mount -o loop "$output_ext4_path" $mount_dir

buildctl build --no-cache --frontend=dockerfile.v0 --local context=. --local dockerfile=. --output type=local,dest="$mount_dir"
echo "Pre-installed programs in the base rootfs image with buildctl..."

cp --remove-destination init "$mount_dir/sbin/init"

umount --lazy "$mount_dir"
rm -rf "$mount_dir"

gzip -c "$output_ext4_path" > "$output_gzip_path" &> /dev/null
echo "Gzipped the rootfs to $output_gzip_path..."