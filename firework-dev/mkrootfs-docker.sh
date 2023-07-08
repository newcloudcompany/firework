#!/usr/bin/env bash

set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $script_dir

rootfs_archive_path="debian-bookworm-rootfs.tar.gz"

buildctl build --frontend=dockerfile.v0 \
    --local context=docker \
    --local dockerfile=docker \
    --output type=tar,dest=rootfs.tar \
    --opt build-arg:ROOTFS_ARCHIVE_PATH="$rootfs_archive_path"