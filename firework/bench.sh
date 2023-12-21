#!/bin/bash

set -euo pipefail

script_dir=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd $script_dir

nohup sudo ./firework start &

while ! nc -z 172.18.0.242 3000; do
    sleep 0.05
done