#!/bin/bash -x

docker build -f Dockerfile-win-dockcross --build-arg IMAGE=windows-shared-x64 -t native-windows-shared-x64-cross-compile .
docker build -f Dockerfile-win-dockcross --build-arg IMAGE=linux-x64 -t native-linux-x64-cross-compile .

mkdir -p compress/build/native/

docker run --rm native-windows-shared-x64-cross-compile > compress/build/native/native-windows-shared-x64-cross-compile
docker run --rm native-linux-x64-cross-compile > compress/build/native/native-linux-x64-cross-compile

compress/build/native/native-windows-shared-x64-cross-compile --image native-windows-shared-x64-cross-compile \
  bash -c "cd compress/compress-jni &&
    CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -buildmode=c-shared -o ../build/native/compress_jni.dll compress-jni.go"

compress/build/native/native-linux-x64-cross-compile --image native-linux-x64-cross-compile \
  bash -c "cd compress/compress-jni &&
    CGO_ENABLED=1 go build -buildmode=c-shared -o ../build/native/libcompress_jni.so compress-jni.go"
