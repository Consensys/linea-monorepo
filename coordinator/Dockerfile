FROM openjdk:17-slim-bullseye

RUN apt-get update \
 && apt-get install curl -y \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /opt/consensys/linea/coordinator

# copy application
COPY --from=jar ./app-all.jar /opt/consensys/linea/coordinator/libs/

RUN mkdir -p /opt/consensys/linea/coordinator/logs
RUN mkdir -p /opt/consensys/linea/coordinator/tmp/prover/request
RUN mkdir -p /opt/consensys/linea/coordinator/tmp/prover/response

# Build-time metadata as defined at http://label-schema.org
ARG BUILD_DATE
ARG VCS_REF
ARG VERSION
LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.name="coordinator" \
      org.label-schema.description="Coordinator for Linea" \
      org.label-schema.url="https://consensys.net/" \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url="https://github.com/ConsenSys/zkevm-monorepo" \
      org.label-schema.vendor="ConsenSys" \
      org.label-schema.version=$VERSION \
      org.label-schema.schema-version="1.0"

WORKDIR /opt/consensys/linea/coordinator/


