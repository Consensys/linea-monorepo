#!/bin/sh

FROM=$1
TO=$2
IMAGE_NAME=consensys/jna-lib-builder

echo "Copying from $FROM to $TO in docker image $IMAGE_NAME"
id=$(docker create $IMAGE_NAME)

# Copy the folder from the container in a placeholder. That way, the content of
# the folder is copied into the target directory instead of the folder itself.
docker cp $id:$FROM $TO
docker rm $id