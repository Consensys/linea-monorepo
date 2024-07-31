#!/usr/bin/env bash
set -Eeu

if [ $# -lt 1 ]; then
    echo "Description: copy all the relevant files under the contracts folder to the local linea-contracts repo root folder"
    echo "Usage: $0 destRepoRootDir"
    echo Example: $0 ../../linea-contracts
    exit
fi

export destRepoRootDir=$1
echo destRepoRootDir = $destRepoRootDir

echo "Copying all files under contracts with exclude list to $destRepoRootDir ..."
rsync -av --exclude-from='.exclude_copy_list' ./ $destRepoRootDir
echo "Done!!!"

echo "Renaming package.json and package-lock.json to "linea-contracts" in $destRepoRootDir ..."
cd $destRepoRootDir
node -e "let pkg=require('./package.json'); pkg.name='linea-contracts'; require('fs').writeFileSync('package.json', JSON.stringify(pkg, null, 2));"
node -e "let pkg=require('./package-lock.json'); pkg.name='linea-contracts'; pkg.packages[''].name='linea-contracts'; require('fs').writeFileSync('package-lock.json', JSON.stringify(pkg, null, 2));"
cd -
echo "All Done!!!!"
