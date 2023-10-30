#!/usr/bin/bash
directory=$( cd "$(dirname "${BASH_SOURCE[0]}")" ; pwd -P )
wgmesh_bin="$directory/cmd/wg-mesh/wg-mesh"
wgmeshd_bin="$directory/cmd/wgmeshd/wgmeshd"

echo "BUILDING to $install_directory"

cd "$(dirname $wgmesh_bin)"
go mod tidy
go build 

cd "$(dirname $wgmeshd_bin)"
go mod tidy
go build

cd $directory

echo "BUILT"
