#!/usr/bin/env bash

set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $script_dir

variants=(minimal tools)

rootfs_archive_url="https://pub-1a5aeef625fc45b4a4bef89ee141047f.r2.dev/debian-bookworm-systemd-rootfs.tar.gz"
rootfs_archive_path="debian-bookworm-systemd-rootfs.tar.gz"

function ensure_rootfs_archive {
    if [[ ! -e "$rootfs_archive_path" ]]; then
        echo "Downloading debian bookworm rootfs ar archive..."
        curl -o "$rootfs_archive_path" -L "$rootfs_archive_url"
    fi
}

# ensure_rootfs_archive

for variant in "${variants[@]}"; do
    echo "Building $variant rootfs..."
    ./mkrootfs-$variant.sh "$rootfs_archive_path"
done