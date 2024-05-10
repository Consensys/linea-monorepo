#!/usr/bin/env bash
set -Eeu

function prepareDatadir {
  datadir=$1
  genesisfile=$2
  if [ ! -d $datadir/geth ]; then
    echo -e "\n\n----------> A new data directory '$datadir' will be created!"
    geth --datadir $datadir init $genesisfile
    echo -e "----------> A new data directory '$datadir' created!\n\n"
  else
    echo -e "\n\n----------> Data directory '$datadir' already exists! Contents:"
    find $datadir
    echo -e "\n\n"
  fi
}
