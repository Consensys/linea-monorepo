#!/bin/sh

# Initialize external vars - need this to get around unbound variable errors
SKIP_GRADLE="$SKIP_GRADLE"

# Exit script if you try to use an uninitialized variable.
set -o nounset

# Exit script if a statement returns a non-true return value.
set -o errexit

SCRIPTDIR=$(dirname -- "$( readlink -f -- "$0")")
OSTYPE=$(uname -o)

# delete old build dir, if exists
rm -rf "$SCRIPTDIR/compress/build/native" || true
mkdir -p "$SCRIPTDIR/compress/build/native"

if [ x"$OSTYPE" = x"msys" ]; then
  LIBRARY_EXTENSION=dll
elif [ x"$OSTYPE" = x"GNU/Linux" ]; then
  LIBRARY_EXTENSION=so
elif [ x"$OSTYPE" = x"Darwin" ]; then
  LIBRARY_EXTENSION=dylib
else
  echo "*** Unknown OS: $OSTYPE"
  exit 1
fi

cd "$SCRIPTDIR/compress/compress-jni"
echo "Building Go module libcompress_jni.$LIBRARY_EXTENSION for $OSTYPE"
CGO_ENABLED=1 go build -buildmode=c-shared -o libcompress_jni.$LIBRARY_EXTENSION compress-jni.go
mv libcompress_jni.* "$SCRIPTDIR/compress/build/native"
