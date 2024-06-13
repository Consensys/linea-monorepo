package net.consensys.linea.traces

import net.consensys.KMath

/** Module's traces counters are not expected to go above 2^31 */
typealias TracesCounters = Map<TracingModule, UInt>
fun TracesCounters.toLogString(): String {
  return this.entries.joinToString(prefix = "[", postfix = "]", separator = " ") { (module, count) ->
    "$module=$count"
  }
}

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

fun emptyTracesCounts(): TracesCounters {
  return TracingModule.values().associateWith { 0u }
}

fun sumTracesCounters(
  vararg tracesCounters: TracesCounters
): TracesCounters {
  val result = emptyTracesCounts().toMutableMap()

  TracingModule.values().forEach { module ->
    tracesCounters.forEach { counters ->
      result[module] = KMath.addExact(result[module]!!, (counters[module] ?: 0u))
    }
  }

  return result
}

fun allTracesWithinLimits(tracesCounters: TracesCounters, tracesCountersLimits: TracesCounters): Boolean {
  return tracesCounters.entries.all { (moduleName, moduleCount) ->
    val moduleCap = tracesCountersLimits[moduleName]!!
    moduleCount <= moduleCap
  }
}

fun allTracesEmpty(tracesCounters: TracesCounters): Boolean {
  return tracesCounters.values.all { it == 0u }
}

fun allModulesAreDefined(tracesCounters: TracesCounters): Boolean {
  return tracesCounters.keys.containsAll(TracingModule.allModules)
}
