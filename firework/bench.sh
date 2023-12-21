#!/bin/bash

set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $script_dir

nohup sudo ./firework start &

while ! sudo ./firework status | grep -q "Running"; do
    sleep 0.05
done