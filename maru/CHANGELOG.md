# Changelog

### Upcoming Breaking Changes

### Breaking Changes

### Additions and Improvements

Added a validator for the execution layer (EL) payload block number to ensure block numbers increment sequentially between blocks, improving consistency checks between consensus and execution layers.

Validators no longer wait for EL sync completion before starting QBFT consensus — they catch up via block building instead. Added `roundExpiryCoefficient` config field (default `2.0`) for tunable QBFT round timeout scaling. (#499)

Added `ConsensusMetrics` with Micrometer histograms for each QBFT phase latency, broken down by proposer vs. non-proposer role. (#500)

### Bug Fixes

Fixed a ~75% CPU overhead on the event loop caused by repeated `BouncyCastleProvider` initialization on every signature operation. Fixed a race condition where `engine_getPayload` could be called before the preceding FCU completed. (#498)
