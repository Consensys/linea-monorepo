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

  private val emptyCounters = TracingModule.values().associateWith { 0.toUInt() }

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
    counters[TracingModule.MMU_ID] = 0.toUInt()
    counters[TracingModule.SHF_RT] = 0.toUInt()
    counters[TracingModule.INSTRUCTION_DECODER] = 0.toUInt()
    counters[TracingModule.BIN_RT] = 0.toUInt()

    // Block limits
    val blockTx = getPrecomputedLimit(trace, "TxCount")
    if (blockTx < 0) {
      return Err(TracesError(ErrorType.WRONG_JSON_CONTENT, "Error while looking for blockTx: TxCount."))
    } else {
      counters[TracingModule.BLOCK_TX] = blockTx.toUInt()
    }

    val l2L1logsCount = getPrecomputedLimit(trace, "L2L1logsCount")
    if (l2L1logsCount < 0) {
      return Err(TracesError(ErrorType.WRONG_JSON_CONTENT, "Error while looking for L2L1logsCount."))
    } else {
      counters[TracingModule.BLOCK_L2L1LOGS] = l2L1logsCount.toUInt()
    }

    val keccakCount = getPrecomputedLimit(trace, "KeccakCount")
    if (keccakCount < 0) {
      return Err(TracesError(ErrorType.WRONG_JSON_CONTENT, "Error while looking for KeccakCount."))
    } else {
      counters[TracingModule.BLOCK_KECCAK] = keccakCount.toUInt()
    }

    // Precompile limits
    val precompiles = JsonParserHelper.getPrecompiles(trace)
    if (precompiles == null) {
      return Err(TracesError(ErrorType.WRONG_JSON_CONTENT, "Missing precompiles."))
    } else {
      counters[TracingModule.PRECOMPILE_ECRECOVER] = ((precompiles.get("EcRecover") ?: 0) as Int).toUInt()
      counters[TracingModule.PRECOMPILE_SHA2] = ((precompiles.get("Sha2") ?: 0) as Int).toUInt()
      counters[TracingModule.PRECOMPILE_RIPEMD] = ((precompiles.get("RipeMD") ?: 0) as Int).toUInt()
      counters[TracingModule.PRECOMPILE_IDENTITY] = ((precompiles.get("Identity") ?: 0) as Int).toUInt()
      counters[TracingModule.PRECOMPILE_MODEXP] = ((precompiles.get("ModExp") ?: 0) as Int).toUInt()
      counters[TracingModule.PRECOMPILE_ECADD] = ((precompiles.get("EcAdd") ?: 0) as Int).toUInt()
      counters[TracingModule.PRECOMPILE_ECMUL] = ((precompiles.get("EcMul") ?: 0) as Int).toUInt()
      counters[TracingModule.PRECOMPILE_ECPAIRING] = ((precompiles.get("EcPairing") ?: 0) as Int).toUInt()
      counters[TracingModule.PRECOMPILE_BLAKE2F] = ((precompiles.get("Blake2f") ?: 0) as Int).toUInt()
    }
    val blockL1Size = getPrecomputedLimit(trace, "blockL1Size")
    if (blockL1Size < 0) {
      return Err(TracesError(ErrorType.WRONG_JSON_CONTENT, "Invalid blockL1Size $blockL1Size"))
    }

    return Ok(BlockCounters(counters, blockL1Size.toUInt()))
  }

  private fun getOffset(module: TracingModule): UInt {
    return when (module) {
      TracingModule.ADD -> 2U
      TracingModule.BIN -> 16U
      TracingModule.BIN_RT -> 0U
      TracingModule.EC_DATA -> 12U
      TracingModule.EXT -> 8U
      TracingModule.HUB -> 2U
      TracingModule.INSTRUCTION_DECODER -> 0U
      TracingModule.MMIO -> 0U
      TracingModule.MMU -> 0U
      TracingModule.MMU_ID -> 0U
      TracingModule.MOD -> 8U
      TracingModule.MUL -> 9U
      TracingModule.MXP -> 4U
      TracingModule.PHONEY_RLP -> 0U
      TracingModule.PUB_HASH -> 0U
      TracingModule.PUB_HASH_INFO -> 0U
      TracingModule.PUB_LOG -> 0U
      TracingModule.PUB_LOG_INFO -> 0U
      TracingModule.RLP -> 8U
      TracingModule.ROM -> 2U
      TracingModule.SHF -> 16U
      TracingModule.SHF_RT -> 0U
      TracingModule.TX_RLP -> 0U
      TracingModule.WCP -> 16U
      TracingModule.BLOCK_TX -> 0U
      TracingModule.BLOCK_L2L1LOGS -> 0U
      TracingModule.BLOCK_KECCAK -> 0U
      TracingModule.PRECOMPILE_ECRECOVER -> 0U
      TracingModule.PRECOMPILE_SHA2 -> 0U
      TracingModule.PRECOMPILE_RIPEMD -> 0U
      TracingModule.PRECOMPILE_IDENTITY -> 0U
      TracingModule.PRECOMPILE_MODEXP -> 0U
      TracingModule.PRECOMPILE_ECADD -> 0U
      TracingModule.PRECOMPILE_ECMUL -> 0U
      TracingModule.PRECOMPILE_ECPAIRING -> 0U
      TracingModule.PRECOMPILE_BLAKE2F -> 0U
    }
  }

  override fun countTraces(traces: String): Result<VersionedResult<BlockCounters>, TracesError> {
    return concreteCountTraces(traces).map { VersionedResult(tracesVersion, it) }
  }
}
