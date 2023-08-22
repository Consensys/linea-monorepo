package net.consensys.linea.traces

/** Module's traces counters are not expected to go above 2^31 */
typealias TracesCounters = Map<TracingModule, UInt>

/** More info: https://github.com/ConsenSys/zkevm-monorepo/issues/525 */
enum class TracingModule {
  // EMV Module limits
  ADD,
  BIN,
  BIN_RT,
  EC_DATA,
  EXT,
  HUB,
  INSTRUCTION_DECODER,
  MMIO,
  MMU,
  MMU_ID,
  MOD,
  MUL,
  MXP,
  PHONEY_RLP,
  PUB_HASH,
  PUB_HASH_INFO,
  PUB_LOG,
  PUB_LOG_INFO,
  RLP,
  ROM,
  SHF,
  SHF_RT,
  TX_RLP,
  WCP,

  // Block-specific limits
  BLOCK_TX,
  BLOCK_L2L1LOGS,
  BLOCK_KECCAK,

  // Precompiles call limits
  PRECOMPILE_ECRECOVER,
  PRECOMPILE_SHA2,
  PRECOMPILE_RIPEMD,
  PRECOMPILE_IDENTITY,
  PRECOMPILE_MODEXP,
  PRECOMPILE_ECADD,
  PRECOMPILE_ECMUL,
  PRECOMPILE_ECPAIRING,
  PRECOMPILE_BLAKE2F;

  companion object {
    val evmModules: Set<TracingModule> = setOf(
      ADD,
      BIN,
      BIN_RT,
      EC_DATA,
      EXT,
      HUB,
      INSTRUCTION_DECODER,
      MMIO,
      MMU,
      MMU_ID,
      MOD,
      MUL,
      MXP,
      PHONEY_RLP,
      PUB_HASH,
      PUB_HASH_INFO,
      PUB_LOG,
      PUB_LOG_INFO,
      RLP,
      ROM,
      SHF,
      SHF_RT,
      WCP
    )
    val allModules: Set<TracingModule> = TracingModule.values().toSet()
  }
}
