#!/usr/bin/env bash
# Copyright (C) 2024, Nuklai. All rights reserved.
# See the file LICENSE for licensing terms.

set -o errexit
set -o nounset
set -o pipefail

realpath() {
    [[ $1 = /* ]] && echo "$1" || echo "$PWD/${1#./}"
}

build_project() {
    local project_path
    project_path=$(realpath "$1")
    local project_name=$2

    local binary_path
    if [[ $# -eq 3 ]]; then
        local binary_dir
        local binary_name
        # Ensure binary_dir is an absolute path
        binary_dir=$(realpath "$project_path/build/$(dirname "$3")")
        binary_name=$(basename "$3")
        binary_path="$binary_dir/$binary_name"
    else
        # Set default binary directory location
        binary_path=$project_path/build/$project_name
    fi

    cd "$project_path"

    echo "Building "${project_name}vm" in $binary_path"
    mkdir -p "$(dirname "$binary_path")"
    go build -o "$binary_path" ./cmd/"${project_name}vm"

    cli_path=$project_path/build/nuklai-cli
    echo "Building "${project_name}-cli" in $cli_path"
    mkdir -p "$(dirname "$cli_path")"
    go build -o "$cli_path" ./cmd/nuklai-cli
}
