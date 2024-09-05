FROM openjdk:22-ea-17-slim-bookworm

RUN apt-get update \
	&& apt-get install curl -y \
  && apt-get install unzip -y \
	&& apt-get clean \
	&& rm -rf /var/lib/apt/lists/*

WORKDIR /opt/consensys/linea/coordinator

# copy application
COPY --from=zip ./coordinator.zip /opt/consensys/linea/coordinator/libs/
RUN unzip libs/coordinator.zip -d libs/ && mv libs/coordinator/lib/** libs/ && rm -R libs/coordinator/

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
	org.label-schema.url="https://consensys.io/" \
	org.label-schema.vcs-ref=$VCS_REF \
	org.label-schema.vcs-url="https://github.com/ConsenSys/zkevm-monorepo" \
	org.label-schema.vendor="ConsenSys" \
	org.label-schema.version=$VERSION \
	org.label-schema.schema-version="1.0"

WORKDIR /opt/consensys/linea/coordinator/
