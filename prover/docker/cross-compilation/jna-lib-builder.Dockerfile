FROM golang:alpine

RUN apk add --no-cache musl-dev libgcc g++
RUN mkdir -p /build

WORKDIR  /usr/src/
COPY . .

ENV LINUX_AMD64_FLAGS   CGO_ENABLED=1 GOOS="linux" GOARCH="amd64" CC="x86_64-linux-musl-gcc" CXX="x86_64-linux-musl-g++"
ENV DARWIN_ARM64_FLAGS  CGO_ENABLED=1 GOOS="darwin" GOARCH="arm64"

ENV SHNARF_CALCULATOR_SRC_DIR       lib/shnarf_calculator/shnarf_calculator.go
ENV SHNARF_CALCULATOR_TARGET_DIR    /build/shnarf_calculator/libshnarf_calculator
ENV BLOB_COMPRESSOR_SRC_DIR         lib/compressor/blob_compressor.go
ENV BLOB_COMPRESSOR_TARGET_DIR      /build/compressor/libblob_compressor

# Build of the Shnarf calculator
RUN $(LINUX_AMD64_FLAGS) 	go build -buildmode=c-shared -o ${SHNARF_CALCULATOR_TARGET_DIR}_linux_amd64_jna.so ${SHNARF_CALCULATOR_SRC_DIR}
RUN $(DARWIN_ARM64_FLAGS) 	go build -buildmode=c-shared -o ${SHNARF_CALCULATOR_TARGET_DIR}_darwin_arm64_jna.so ${SHNARF_CALCULATOR_SRC_DIR}

# Build of the compressor
RUN $(LINUX_AMD64_FLAGS) 	go build -buildmode=c-shared -o ${BLOB_COMPRESSOR_TARGET_DIR}_linux_amd64_jna.so ${BLOB_COMPRESSOR_SRC_DIR}
RUN $(DARWIN_ARM64_FLAGS) 	go build -buildmode=c-shared -o ${BLOB_COMPRESSOR_TARGET_DIR}_darwin_arm64_jna.so ${BLOB_COMPRESSOR_SRC_DIR}

