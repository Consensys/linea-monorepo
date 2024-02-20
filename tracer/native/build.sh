#!/usr/bin/env bash

# Initialize external vars - need this to get around unbound variable errors
SKIP_GRADLE="$SKIP_GRADLE"

# Exit script if you try to use an uninitialized variable.
set -o nounset

# Exit script if a statement returns a non-true return value.
set -o errexit

# Use the error status of the first failure, rather than that of the last item in a pipeline.
set -o pipefail

# Resolve the directory that contains this script. We have to jump through a few
# hoops for this because the usual one-liners for this don't work if the script
# is a symlink
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ]; do # resolve $SOURCE until the file is no longer a symlink
  DIR="$( cd -P "$( dirname "$SOURCE" )" >/dev/null 2>&1 && pwd )"
  SOURCE="$(readlink "$SOURCE")"
  [[ $SOURCE != /* ]] && SOURCE="$DIR/$SOURCE" # if $SOURCE was a relative symlink, we need to resolve it relative to the path where the symlink file was located
done
SCRIPTDIR="$( cd -P "$( dirname "$SOURCE" )" >/dev/null 2>&1 && pwd )"

# Determine core count for parallel make
if [[ "$OSTYPE" == "linux-gnu" ]];  then
  CORE_COUNT=$(nproc)
  OSARCH=${OSTYPE%%[0-9.]*}-`arch`
fi

if [[ "$OSTYPE" == "darwin"* ]];  then
  CORE_COUNT=$(sysctl -n hw.ncpu)
  if [[ "`machine`" == "arm"* ]]; then
    arch_name="aarch64"
    export CFLAGS="-arch arm64"
  else
    arch_name="x86-64"
    export CFLAGS="-arch x86_64"
  fi
  OSARCH="darwin-$arch_name"
fi

# add to path cargo
[ -f $HOME/.cargo/env ] && . $HOME/.cargo/env

# add to path brew
[ -f $HOME/.zprofile ] && . $HOME/.zprofile

build_compress() {
  cat <<EOF
  ############################
  ####### build compress #######
  ############################
EOF

  cd "$SCRIPTDIR/compress/compress-jni"

  # delete old build dir, if exists
  rm -rf "$SCRIPTDIR/compress/build" || true
  mkdir -p "$SCRIPTDIR/compress/build/lib"

  if [[ "$OSTYPE" == "msys" ]]; then
    	LIBRARY_EXTENSION=dll
  elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    LIBRARY_EXTENSION=so
  elif [[ "$OSTYPE" == "darwin"* ]]; then
    LIBRARY_EXTENSION=dylib
  fi

  go build -buildmode=c-shared -o libcompress_jni.$LIBRARY_EXTENSION compress-jni.go

  mkdir -p "$SCRIPTDIR/compress/build/${OSARCH}/lib"
  cp libcompress_jni.* "$SCRIPTDIR/compress/build/${OSARCH}/lib"
}

build_compress

exit