package net.consensys.linea.traces

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import com.github.michaelbull.result.map
import net.consensys.linea.BlockCounters
import net.consensys.linea.ErrorType
import net.consensys.linea.TracesCounter
import net.consensys.linea.TracesError
import net.consensys.linea.VersionedResult
import net.consensys.linea.traces.JsonParserHelper.Companion.countRow
import net.consensys.linea.traces.JsonParserHelper.Companion.findKey
import net.consensys.linea.traces.JsonParserHelper.Companion.getPrecomputedLimit
import net.consensys.linea.traces.JsonParserHelper.Companion.getTracesPosition
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger

class RawJsonTracesCounter(private val tracesVersion: String) :
  TracesCounter {
  private val log: Logger = LogManager.getLogger(this::class.java)

  private val emptyCounters = TracingModuleV1.values().associateWith { 0.toUInt() }

  fun concreteCountTraces(trace: String): Result<BlockCounters, TracesError> {
    val counters = emptyCounters.toMutableMap()

    val traceInfoPosition = getTracesPosition(trace)
    for (pos: Int in traceInfoPosition) {
      val key = findKey(trace, pos)
      if (key == null) {
        log.info("Key of Trace at $pos not found.")
        continue
      }
      val count = countRow(trace, pos)
      if (count < 0) {
        return Err(TracesError(ErrorType.WRONG_JSON_CONTENT, "Invalid count at $pos."))
      }
      val module = JsonParserHelper.from(key)
      if (module == null) {
        log.warn("Unrecognized module {}", key)
      } else {
        counters[module] = count.toUInt() + getOffset(module)
      }
    }
    // some constant limits
    counters[TracingModuleV1.MMU_ID] = 0.toUInt()
    counters[TracingModuleV1.SHF_RT] = 0.toUInt()
    counters[TracingModuleV1.INSTRUCTION_DECODER] = 0.toUInt()
    counters[TracingModuleV1.BIN_RT] = 0.toUInt()

    // Block limits
    val blockTx = getPrecomputedLimit(trace, "TxCount")
    if (blockTx < 0) {
      return Err(TracesError(ErrorType.WRONG_JSON_CONTENT, "Error while looking for blockTx: TxCount."))
    } else {
      counters[TracingModuleV1.BLOCK_TX] = blockTx.toUInt()
    }

    val l2L1logsCount = getPrecomputedLimit(trace, "L2L1logsCount")
    if (l2L1logsCount < 0) {
      return Err(TracesError(ErrorType.WRONG_JSON_CONTENT, "Error while looking for L2L1logsCount."))
    } else {
      counters[TracingModuleV1.BLOCK_L2L1LOGS] = l2L1logsCount.toUInt()
    }

    val keccakCount = getPrecomputedLimit(trace, "KeccakCount")
    if (keccakCount < 0) {
      return Err(TracesError(ErrorType.WRONG_JSON_CONTENT, "Error while looking for KeccakCount."))
    } else {
      counters[TracingModuleV1.BLOCK_KECCAK] = keccakCount.toUInt()
    }

    // Precompile limits
    val precompiles = JsonParserHelper.getPrecompiles(trace)
    if (precompiles == null) {
      return Err(TracesError(ErrorType.WRONG_JSON_CONTENT, "Missing precompiles."))
    } else {
      counters[TracingModuleV1.PRECOMPILE_ECRECOVER] = ((precompiles.get("EcRecover") ?: 0) as Int).toUInt()
      counters[TracingModuleV1.PRECOMPILE_SHA2] = ((precompiles.get("Sha2") ?: 0) as Int).toUInt()
      counters[TracingModuleV1.PRECOMPILE_RIPEMD] = ((precompiles.get("RipeMD") ?: 0) as Int).toUInt()
      counters[TracingModuleV1.PRECOMPILE_IDENTITY] = ((precompiles.get("Identity") ?: 0) as Int).toUInt()
      counters[TracingModuleV1.PRECOMPILE_MODEXP] = ((precompiles.get("ModExp") ?: 0) as Int).toUInt()
      counters[TracingModuleV1.PRECOMPILE_ECADD] = ((precompiles.get("EcAdd") ?: 0) as Int).toUInt()
      counters[TracingModuleV1.PRECOMPILE_ECMUL] = ((precompiles.get("EcMul") ?: 0) as Int).toUInt()
      counters[TracingModuleV1.PRECOMPILE_ECPAIRING] = ((precompiles.get("EcPairing") ?: 0) as Int).toUInt()
      counters[TracingModuleV1.PRECOMPILE_BLAKE2F] = ((precompiles.get("Blake2f") ?: 0) as Int).toUInt()
    }
    val blockL1Size = getPrecomputedLimit(trace, "blockL1Size")
    if (blockL1Size < 0) {
      return Err(TracesError(ErrorType.WRONG_JSON_CONTENT, "Invalid blockL1Size $blockL1Size"))
    }

    return Ok(BlockCounters(TracesCountersV1(counters), blockL1Size.toUInt()))
  }

  private fun getOffset(module: TracingModuleV1): UInt {
    return when (module) {
      TracingModuleV1.ADD -> 2U
      TracingModuleV1.BIN -> 16U
      TracingModuleV1.BIN_RT -> 0U
      TracingModuleV1.EC_DATA -> 12U
      TracingModuleV1.EXT -> 8U
      TracingModuleV1.HUB -> 2U
      TracingModuleV1.INSTRUCTION_DECODER -> 0U
      TracingModuleV1.MMIO -> 0U
      TracingModuleV1.MMU -> 0U
      TracingModuleV1.MMU_ID -> 0U
      TracingModuleV1.MOD -> 8U
      TracingModuleV1.MUL -> 9U
      TracingModuleV1.MXP -> 4U
      TracingModuleV1.PHONEY_RLP -> 0U
      TracingModuleV1.PUB_HASH -> 0U
      TracingModuleV1.PUB_HASH_INFO -> 0U
      TracingModuleV1.PUB_LOG -> 0U
      TracingModuleV1.PUB_LOG_INFO -> 0U
      TracingModuleV1.RLP -> 8U
      TracingModuleV1.ROM -> 2U
      TracingModuleV1.SHF -> 16U
      TracingModuleV1.SHF_RT -> 0U
      TracingModuleV1.TX_RLP -> 0U
      TracingModuleV1.WCP -> 16U
      TracingModuleV1.BLOCK_TX -> 0U
      TracingModuleV1.BLOCK_L2L1LOGS -> 0U
      TracingModuleV1.BLOCK_KECCAK -> 0U
      TracingModuleV1.PRECOMPILE_ECRECOVER -> 0U
      TracingModuleV1.PRECOMPILE_SHA2 -> 0U
      TracingModuleV1.PRECOMPILE_RIPEMD -> 0U
      TracingModuleV1.PRECOMPILE_IDENTITY -> 0U
      TracingModuleV1.PRECOMPILE_MODEXP -> 0U
      TracingModuleV1.PRECOMPILE_ECADD -> 0U
      TracingModuleV1.PRECOMPILE_ECMUL -> 0U
      TracingModuleV1.PRECOMPILE_ECPAIRING -> 0U
      TracingModuleV1.PRECOMPILE_BLAKE2F -> 0U
    }
  }

  override fun countTraces(traces: String): Result<VersionedResult<BlockCounters>, TracesError> {
    return concreteCountTraces(traces).map { VersionedResult(tracesVersion, it) }
  }
}
