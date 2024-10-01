FROM eclipse-temurin:21-jre-alpine AS builder

RUN apt-get update \
	&& apt-get install curl -y \
  && apt-get install unzip -y \
	&& apt-get clean \
	&& rm -rf /var/lib/apt/lists/*

RUN mkdir -p src libs

COPY --from=zip ./coordinator.zip /src

RUN unzip /src/coordinator.zip -d libs/

FROM eclipse-temurin:21-jre-alpine

WORKDIR /opt/consensys/linea/coordinator

RUN apt-get update \
	&& apt-get install curl -y \
	&& apt-get clean \
	&& rm -rf /var/lib/apt/lists/* \
  && mkdir -p /opt/consensys/linea/coordinator/libs/ /opt/consensys/linea/coordinator/logs /opt/consensys/linea/coordinator/tmp/prover/request /opt/consensys/linea/coordinator/tmp/prover/response

COPY --from=builder libs/coordinator/lib/** libs/

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
