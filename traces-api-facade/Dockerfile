FROM openjdk:21-slim-bullseye

RUN apt-get update \
	&& apt-get install curl -y \
	&& apt-get clean \
	&& rm -rf /var/lib/apt/lists/*

WORKDIR /opt/consensys/linea/traces-api-facade

# copy application
COPY --from=jar ./app-all.jar /opt/consensys/linea/traces-api-facade/libs/

RUN mkdir -p /opt/consensys/linea/traces-api-facade/logs

# Build-time metadata as defined at http://label-schema.org
ARG BUILD_DATE
ARG VCS_REF
ARG VERSION
LABEL org.label-schema.build-date=$BUILD_DATE \
	org.label-schema.name="traces-api-facade" \
	org.label-schema.description="Linea Traces API" \
	org.label-schema.url="https://consensys.io/" \
	org.label-schema.vcs-ref=$VCS_REF \
	org.label-schema.vcs-url="https://github.com/ConsenSys/linea-monorepo" \
	org.label-schema.vendor="ConsenSys" \
	org.label-schema.version=$VERSION \
	org.label-schema.schema-version="1.0"

WORKDIR /opt/consensys/linea/traces-api-facade/


