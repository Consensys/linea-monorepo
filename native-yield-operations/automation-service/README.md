# Linea Native Yield Automation Service

## Overview

The Linea Native Yield Automation Service automates native yield yield management operations by continuously monitoring the YieldManager contract's ossification state and executing mode-specific processors. 

The service is triggered through an event-driven mechanism that watches for `VaultsReportDataUpdated` events from the LazyOracle contract. The `waitForVaultsReportDataUpdatedEvent()` function implements a race condition between event detection and a configurable timeout — whichever occurs first triggers the operation cycle, ensuring timely execution even when events are delayed.

Operation modes are selected dynamically based on the yield provider's state:
- `YIELD_REPORTING_MODE` for normal operations
- `OSSIFICATION_PENDING_MODE` during transition
- `OSSIFICATION_COMPLETE_MODE` once ossified.

## Configuration

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
