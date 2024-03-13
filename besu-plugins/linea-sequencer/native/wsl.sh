#!/bin/bash

docker build -f Dockerfile-win-dockcross -t native-windows-cross-compile .

docker run --rm native-windows-cross-compile > compress/build/native/native-windows-cross-compile

compress/build/native/native-windows-cross-compile \
  bash -c "cd compress/compress-jni &&
    CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -buildmode=c-shared -o ../build/native/compress_jni.dll compress-jni.go"

