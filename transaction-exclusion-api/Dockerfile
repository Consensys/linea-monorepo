FROM openjdk:22-ea-17-slim-bookworm

RUN apt-get update \
	&& apt-get install unzip -y \
	&& apt-get clean \
	&& rm -rf /var/lib/apt/lists/*

WORKDIR /opt/consensys/linea/transaction-exclusion-api

# copy application
COPY --from=zip ./transaction-exclusion-api.zip /opt/consensys/linea/transaction-exclusion-api/libs/
RUN unzip libs/transaction-exclusion-api.zip -d libs/ \
	&& mv libs/transaction-exclusion-api/lib/** libs/ \
	&& rm -R libs/transaction-exclusion-api/

RUN mkdir -p /opt/consensys/linea/transaction-exclusion-api/logs

# Build-time metadata as defined at http://label-schema.org
ARG BUILD_DATE
ARG VCS_REF
ARG VERSION
LABEL org.label-schema.build-date=$BUILD_DATE \
	org.label-schema.name="transaction-exclusion-api" \
	org.label-schema.description="Linea Transaction Exclusion API" \
	org.label-schema.url="https://consensys.io/" \
	org.label-schema.vcs-ref=$VCS_REF \
	org.label-schema.vcs-url="https://github.com/ConsenSys/linea-monorepo" \
	org.label-schema.vendor="ConsenSys" \
	org.label-schema.version=$VERSION \
	org.label-schema.schema-version="1.0"

WORKDIR /opt/consensys/linea/transaction-exclusion-api/


