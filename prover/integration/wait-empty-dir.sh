#!/bin/sh
DIR=$1

sleep 10

while [ "$(ls -A "$DIR")" ]
do
  echo "waiting"
  sleep 2
done
