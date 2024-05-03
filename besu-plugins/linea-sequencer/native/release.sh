#!/bin/bash -x

docker build -f Dockerfile-dockcross --build-arg IMAGE=windows-shared-x64 -t native-windows-shared-x64-cross-compile .
docker build -f Dockerfile-dockcross --build-arg IMAGE=linux-x64 -t native-linux-x64-cross-compile .
docker build -f Dockerfile-dockcross --build-arg IMAGE=linux-arm64 -t native-linux-arm64-cross-compile .

mkdir -p compress/build/native/linux-x86_64
mkdir -p compress/build/native/linux-aarch64

docker run --rm native-windows-shared-x64-cross-compile > compress/build/native/native-windows-shared-x64-cross-compile
docker run --rm native-linux-x64-cross-compile > compress/build/native/native-linux-x64-cross-compile
docker run --rm native-linux-arm64-cross-compile > compress/build/native/native-linux-arm64-cross-compile

chmod +x \
compress/build/native/native-windows-shared-x64-cross-compile \
compress/build/native/native-linux-x64-cross-compile \
compress/build/native/native-linux-arm64-cross-compile

OCI_EXE=docker \
compress/build/native/native-windows-shared-x64-cross-compile --image native-windows-shared-x64-cross-compile \
  bash -c "cd compress/compress-jni &&
    CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -buildmode=c-shared -o ../build/native/compress_jni.dll compress-jni.go"

OCI_EXE=docker \
compress/build/native/native-linux-x64-cross-compile --image native-linux-x64-cross-compile \
  bash -c "cd compress/compress-jni &&
    CGO_ENABLED=1 go build -buildmode=c-shared -o ../build/native/linux-x86_64/libcompress_jni.so compress-jni.go"

OCI_EXE=docker \
compress/build/native/native-linux-arm64-cross-compile --image native-linux-arm64-cross-compile \
  bash -c "cd compress/compress-jni &&
    CGO_ENABLED=1 GOARCH=arm64 go build -buildmode=c-shared -o ../build/native/linux-aarch64/libcompress_jni.so compress-jni.go"
