# EIP-7825 Transaction Gas Limit Design

## Context

Issue 1798 asks the Linea sequencer TxGasLimit plugin to align with EIP-7825, which caps each
transaction at `16_777_216` gas.

Current behavior:

- `LineaTransactionPoolValidatorCliOptions.DEFAULT_MAX_TRANSACTION_GAS_LIMIT` is `30_000_000`.
- `GasLimitValidator` rejects only when `transaction.getGasLimit() > maxTxGasLimit`.
- `GasLimitValidator` is used for txpool admission through `LineaTransactionPoolValidatorFactory`
  and for bundle validation through `LineaBundleEndpointsPlugin`.
- `LineaEstimateGas` rejects a caller-provided `gas` value above the configured
  `txValidatorConf.maxTxGasLimit()`.
- `LineaEstimateGas` does not check the decoded `eth_estimateGas` result against the configured
  transaction gas limit before returning it.
- `linea-besu-package/linea-besu/profiles/advanced-mainnet.toml` and
  `linea-besu-package/linea-besu/profiles/advanced-sepolia.toml` currently configure
  `plugin-linea-max-tx-gas-limit=24000000`.

## Decisions

- Enforce EIP-7825 unconditionally in the Linea plugin layer. Do not add fork-aware branching.
- Preserve existing packaged profile files. Do not edit `linea-besu-package/linea-besu/profiles/*.toml`.
- Keep `--plugin-linea-max-tx-gas-limit` as an operator-configured local maximum, but cap the
  runtime effective limit at the EIP-7825 value.
- Do not add a new Linea block-import validator. Block-level protocol validation remains outside
  this change.
- Document rejection metrics and alerts as a follow-up. Do not implement new metrics in this change.

## Effective Limit

Define one authoritative EIP-7825 cap for the sequencer plugin:

```text
EIP_7825_MAX_TRANSACTION_GAS_LIMIT = 16_777_216
```

All runtime validation uses:

```text
effectiveMaxTxGasLimit = min(configuredMaxTxGasLimit, EIP_7825_MAX_TRANSACTION_GAS_LIMIT)
```

This preserves startup compatibility with existing production profiles that configure
`24_000_000`, while ensuring those profiles cannot raise the effective runtime limit above
`16_777_216`.

Operators may still configure a lower value. For example, `9_000_000` remains a stricter local cap.

## Components

Add a small shared constant/helper in the sequencer code near the existing transaction pool
validator configuration or validator package.

The helper should provide:

- The EIP-7825 max transaction gas limit constant.
- A function that computes the effective limit from the configured limit.
- A shared rejection message formatter.

Update `LineaTransactionPoolValidatorCliOptions` so the default max transaction gas limit is
`16_777_216`.

Update `GasLimitValidator` so it stores and enforces the effective limit. Keep the inclusive
boundary:

```text
accept gasLimit <= effectiveMaxTxGasLimit
reject gasLimit > effectiveMaxTxGasLimit
```

`LineaTransactionPoolValidatorFactory` and `LineaBundleEndpointsPlugin` should continue passing
`transactionPoolValidatorConfiguration().maxTxGasLimit()` into `GasLimitValidator`. The validator
itself should apply the EIP-7825 cap.

Update `LineaEstimateGas` so it uses the same effective limit for both:

- Caller-provided `gas` in request parameters.
- The decoded `eth_estimateGas` result before line-count simulation and before fee-estimation
  transaction construction.

## Error Handling

Use clear EIP-7825 wording in gas-limit rejections. The message should include the effective limit.

`GasLimitValidator` and bundle validation should return the same style of string rejection used
today, but with EIP-7825 wording and the effective cap.

`LineaEstimateGas` should preserve existing JSON-RPC error categories:

- Caller-provided `gas` above the cap returns an invalid-params error.
- Computed estimate above the cap returns an estimate error.

If configured max transaction gas is above `16_777_216`, do not fail startup. A one-time warning is
acceptable if it is attached to configuration or validator construction. Do not log a warning for
each rejected transaction.

## Documentation

Update sequencer plugin docs to say:

- The default `--plugin-linea-max-tx-gas-limit` is `16_777_216`.
- Since EIP-7825, the effective max transaction gas limit is capped at `16_777_216`.
- A configured value above `16_777_216` is tolerated for compatibility with existing profiles, but
  does not raise the effective limit.
- A configured value below `16_777_216` is enforced as a stricter local cap.

Documentation files to review:

- `besu-plugins/linea-sequencer/docs/plugins.md`
- `docs/tech/components/besu-plugins.md`
- `tracer/PLUGINS.md`
- `tracer/docs/plugins.md`

Do not update `linea-besu-package/linea-besu/profiles/*.toml`.

## Testing

Unit tests:

- `GasLimitValidatorTest` accepts a transaction at exactly `16_777_216`.
- `GasLimitValidatorTest` rejects a transaction at `16_777_217`.
- `GasLimitValidatorTest` verifies configured values above `16_777_216` are clamped.
- `GasLimitValidatorTest` verifies configured values below `16_777_216` remain stricter local caps.
- Add focused tests for the helper if it contains branching beyond the effective-limit calculation
  and message formatting.

Acceptance tests:

- `TransactionGasLimitTest` should use an explicit EIP-7825 cap constant rather than
  `DefaultGasProvider.GAS_LIMIT`.
- Bundle validation should assert the same cap behavior through `linea_sendBundle`.
- `EstimateGasTest` should cover explicit request `gas > 16_777_216`.
- Cover computed `eth_estimateGas` results above the cap with a unit test or a stable acceptance
  test. Prefer a unit test if an acceptance scenario would require heavy or fragile execution.

Avoid changing unrelated tests that use `DefaultGasProvider.GAS_LIMIT` for normal contract calls
unless those tests actually depend on the maximum transaction gas limit.

Verification commands:

```bash
./gradlew :besu-plugins:linea-sequencer:test
./gradlew :besu-plugins:linea-sequencer:acceptance-tests:acceptanceTests --tests '*TransactionGasLimitTest'
./gradlew :besu-plugins:linea-sequencer:acceptance-tests:acceptanceTests --tests '*SendBundleTest'
./gradlew :besu-plugins:linea-sequencer:acceptance-tests:acceptanceTests --tests '*EstimateGasTest'
```

Run the full sequencer acceptance suite if the environment and time allow:

```bash
./gradlew :besu-plugins:linea-sequencer:acceptance-tests:acceptanceTests
```

## Non-Goals

- No edits to `linea-besu-package/linea-besu/profiles/*.toml`.
- No new Linea block-import validator.
- No metrics or alert implementation.
- No fork-aware behavior.
- No broad refactor of transaction validation or estimate-gas flow.

## Follow-Up

Create a follow-up task for operational visibility:

- Track txpool/API rejection count by reason.
- Alert on post-upgrade spikes in EIP-7825 gas-limit rejections.
- Monitor transaction gas-limit distribution around `16_777_216`.
