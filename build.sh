#!/usr/bin/bash

directory=$( cd "$(dirname "${BASH_SOURCE[0]}")" ; pwd -P )
install_directory="/usr/bin"
wgmesh_bin="$directory/cmd/wg-mesh/wg-mesh"
wgmeshd_bin="$directory/cmd/wgmeshd/wgmeshd"

echo "BUILDING to $install_directory"

cd "$(dirname $wgmesh_bin)"
go build 

cd "$(dirname $wgmeshd_bin)"
go build

cd $directory

mv $wgmesh_bin $install_directory 
mv $wgmeshd_bin $install_directory 

echo "BUILT to $install_directory"
