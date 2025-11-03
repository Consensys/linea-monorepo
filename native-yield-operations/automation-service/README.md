# Linea Native Yield Automation Service

Trigger
- Event-based on Lido contract
- Fallback to cronjob

TODO

## Overview

TODO

## Configuration

### Environment Variables

See the [configuration schema file](native-yield-operations/automation-service/src/application/main/config/config.schema.ts)

## Development

### Running

#### Start the docker local stack

TODO

#### Run the automation service locally:

TODO

### Build

```bash
# Dependency on pnpm --filter @consensys/linea-shared-utils build
pnpm --filter @consensys/linea-native-yield-automation-service build
```

### Unit Test

```bash
pnpm --filter @consensys/linea-shared-utils test
```

## License

This package is licensed under the [Apache 2.0](../../LICENSE-APACHE) and the [MIT](../../LICENSE-MIT) licenses.
