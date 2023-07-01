#!/usr/bin/env bash

set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $script_dir

chown -R root:root initrd
chmod 500 initrd/init
cd initrd

find . -print0 | cpio --null --create --verbose --format=newc > ../initrd.cpio