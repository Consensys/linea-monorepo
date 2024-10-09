FROM openjdk:21-jdk-slim AS builder

RUN apt-get update && apt-get install unzip -y

RUN mkdir -p /src /libs

COPY --from=zip ./coordinator.zip /src

RUN unzip /src/coordinator.zip -d /libs

RUN mkdir -p /libs/coordinator/lib/unpacked-blob-compressor /libs/coordinator/lib/unpacked-blob-shnarf-calculator && \
  cd /libs/coordinator/lib/unpacked-blob-compressor/ && jar -xf ../blob-compressor-0.0.4.jar && \
  cd /libs/coordinator/lib/unpacked-blob-shnarf-calculator/ && jar -xf /libs/coordinator/lib/blob-shnarf-calculator-0.0.4.jar

RUN ARCH=`uname -m` && \
    if [ "$ARCH" == "x86_64" ]; then \
        rm -rf /libs/coordinator/lib/unpacked-blob-compressor/darwin-**; \
        rm -rf /libs/coordinator/lib/unpacked-blob-compressor/linux-aarch64/; \
        rm -rf /libs/coordinator/lib/unpacked-blob-shnarf-calculator/darwin-**; \
        rm -rf /libs/coordinator/lib/unpacked-blob-shnarf-calculator/linux-aarch64/; \
    elif [ "$ARCH" == "aarch64" ]; then \
        rm -rf /libs/coordinator/lib/unpacked-blob-compressor/darwin-**; \
        rm -rf /libs/coordinator/lib/unpacked-blob-compressor/linux-x86-64/; \
        rm -rf /libs/coordinator/lib/unpacked-blob-shnarf-calculator/darwin-**; \
        rm -rf /libs/coordinator/lib/unpacked-blob-shnarf-calculator/linux-x86-64/; \
    fi

RUN jar -cf /libs/coordinator/lib/blob-compressor-0.0.4.jar -C /libs/coordinator/lib/unpacked-blob-compressor . && \
    jar -cf /libs/coordinator/lib/blob-shnarf-calculator-0.0.4.jar -C /libs/coordinator/lib/unpacked-blob-shnarf-calculator . && \
    rm -rf /libs/coordinator/lib/unpacked-blob-compressor && rm -rf /libs/coordinator/lib/unpacked-blob-shnarf-calculator

FROM openjdk:21-jdk-slim

WORKDIR /opt/consensys/linea/coordinator

RUN apt-get update \
    && apt-get install curl -y \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /libs/ /logs /tmp/prover/request /tmp/prover/response

COPY --from=builder /libs/coordinator/lib/** libs/

# Build-time metadata as defined at http://label-schema.org
ARG BUILD_DATE
ARG VCS_REF
ARG VERSION
LABEL org.label-schema.build-date=$BUILD_DATE \
	org.label-schema.name="coordinator" \
	org.label-schema.description="Coordinator for Linea" \
	org.label-schema.url="https://consensys.io/" \
	org.label-schema.vcs-ref=$VCS_REF \
	org.label-schema.vcs-url="https://github.com/ConsenSys/linea-monorepo" \
	org.label-schema.vendor="ConsenSys" \
	org.label-schema.version=$VERSION \
	org.label-schema.schema-version="1.0"
