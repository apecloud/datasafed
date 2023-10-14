#!/bin/sh

if [ $# -ne 1 ]; then
    echo "Usage: $0 <target_dir>"
    exit 1
fi

target_dir="$1"
if [ ! -d "$target_dir" ]; then
    echo "Error: Target directory '${target_dir}' does not exist"
    exit 1
fi

cp /datasafed "$target_dir"
cp -r /etc/ssl/certs "$target_dir/certs"
