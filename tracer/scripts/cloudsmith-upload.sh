#!/bin/bash
set -euo pipefail

VERSION=${1:?Must specify Linea Arithmetization version}
BUILD_DIR=${2:?Must specify Linea Arithmetization module build directory}

ENV_DIR=./build/tmp/cloudsmith-env
if [[ -d ${ENV_DIR} ]] ; then
    source ${ENV_DIR}/bin/activate
else
    python3 -m venv ${ENV_DIR}
    source ${ENV_DIR}/bin/activate
fi

python3 -m pip install --upgrade cloudsmith-cli

echo ">>>>>>>>>>>>>> Uploading Maven Artifact for linea-arithmetization-${VERSION} to Cloudsmith ..."
cloudsmith push maven consensys/linea-arithmetization ${BUILD_DIR}/libs/linea-arithmetization-${VERSION}.jar --pom-file ${BUILD_DIR}/publications/mavenJava/pom-default.xml