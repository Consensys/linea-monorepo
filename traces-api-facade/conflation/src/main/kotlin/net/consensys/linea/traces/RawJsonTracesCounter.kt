package net.consensys.linea.traces

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import com.github.michaelbull.result.map
import io.vertx.core.json.JsonObject
import net.consensys.linea.BlockCounters
import net.consensys.linea.ErrorType
import net.consensys.linea.TracesCounter
import net.consensys.linea.TracesError
import net.consensys.linea.VersionedResult
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger

class RawJsonTracesCounter(private val tracesVersion: String) : TracesCounter {
  private val log: Logger = LogManager.getLogger(this::class.java)

  private val emptyCounters = TracingModule.values().associateWith { 0.toUInt() }

  fun concreteCountTraces(trace: JsonObject): Result<BlockCounters, TracesError> {
    val counters = emptyCounters.toMutableMap()
    // Module limits
    MODULES.forEach { moduleInfo ->

      val (jsonPath, klass) = moduleInfo
      trace.getTrace(jsonPath)?.let { traceJsonObject ->
        if (!traceJsonObject.isEmpty) {
          traceJsonObject.mapTo(klass)?.also {
            when (it) {
              is Add -> counters[TracingModule.ADD] = it.ACC_1.size.toUInt() + 2U
              is Bin -> counters[TracingModule.BIN] = it.ACC_1.size.toUInt() + 16U
              is Ext -> counters[TracingModule.EXT] = it.ACC_A_0.size.toUInt() + 8U
              is EcData -> counters[TracingModule.EC_DATA] = it.ACC_DELTA.size.toUInt() + 12U
              is HashData -> counters[TracingModule.PUB_HASH] = it.INDEX.size.toUInt()
              is HashInfo -> counters[TracingModule.PUB_HASH_INFO] = it.HASH_HI.size.toUInt()
              is Hub -> counters[TracingModule.HUB] = it.ALPHA.size.toUInt() + 2U
              is LogData -> counters[TracingModule.PUB_LOG] = it.INDEX.size.toUInt()
              is LogInfo -> counters[TracingModule.PUB_LOG_INFO] = it.ADDR_HI.size.toUInt()
              is Mmio -> counters[TracingModule.MMIO] = it.ACC_1.size.toUInt() // TODO: get spilling
              is Mmu -> counters[TracingModule.MMU] = it.ACC_1.size.toUInt() // TODO: get spilling
              is Mod -> counters[TracingModule.MOD] = it.ACC_1_2.size.toUInt() + 8U
              is Mul -> counters[TracingModule.MUL] = it.ACC_A_0.size.toUInt() + 9U
              is Mxp -> counters[TracingModule.MXP] = it.ACC_1.size.toUInt() + 4U
              is PhoneyRlp -> counters[TracingModule.PHONEY_RLP] = it.INDEX.size.toUInt()
              is Rlp -> counters[TracingModule.RLP] = it.ADDR_HI.size.toUInt() + 8U
              is Rom -> counters[TracingModule.ROM] = it.PC.size.toUInt() + 2U
              is Shf -> counters[TracingModule.SHF] = it.ACC_1.size.toUInt() + 16U
              is TxRlp -> counters[TracingModule.TX_RLP] = it.ABS_TX_NUM.size.toUInt()
              is Wcp -> counters[TracingModule.WCP] = it.ACC_1.size.toUInt() + 16U
              //
              // These modules are constant-sized, so they do not matter to the blocks
              // conflation counters
              //
              is MmuId -> counters[TracingModule.MMU_ID] = 0.toUInt()
              is ShfRt -> counters[TracingModule.SHF_RT] = 0.toUInt()
              is InstructionDecoder -> counters[TracingModule.INSTRUCTION_DECODER] = 0.toUInt()
              is BinRt -> counters[TracingModule.BIN_RT] = 0.toUInt()
              else -> log.warn("Unrecognized evm module {}", it::class)
            }
          }
        }
      }
        ?: run {
          log.warn("Traces do not contain object with path: '{}'", jsonPath.joinToString("."))
        }
    }

    // Block limits
    counters[TracingModule.BLOCK_TX] = trace.getLong("TxCount").toUInt()
    counters[TracingModule.BLOCK_L2L1LOGS] = trace.getLong("L2L1logsCount").toUInt()
    counters[TracingModule.BLOCK_KECCAK] = trace.getLong("KeccakCount").toUInt()

    // Precompile limits
    val precompiles = trace.getJsonObject("PrecompileCalls")
    counters[TracingModule.PRECOMPILE_ECRECOVER] = precompiles.getLong("EcRecover").toUInt()
    counters[TracingModule.PRECOMPILE_SHA2] = precompiles.getLong("Sha2").toUInt()
    counters[TracingModule.PRECOMPILE_RIPEMD] = precompiles.getLong("RipeMD").toUInt()
    counters[TracingModule.PRECOMPILE_IDENTITY] = precompiles.getLong("Identity").toUInt()
    counters[TracingModule.PRECOMPILE_MODEXP] = precompiles.getLong("ModExp").toUInt()
    counters[TracingModule.PRECOMPILE_ECADD] = precompiles.getLong("EcAdd").toUInt()
    counters[TracingModule.PRECOMPILE_ECMUL] = precompiles.getLong("EcMul").toUInt()
    counters[TracingModule.PRECOMPILE_ECPAIRING] = precompiles.getLong("EcPairing").toUInt()
    counters[TracingModule.PRECOMPILE_BLAKE2F] = precompiles.getLong("Blake2f").toUInt()

    val blockL1Size = trace.getLong("blockL1Size")
    if (blockL1Size == null || blockL1Size < 0) {
      Err(TracesError(ErrorType.WRONG_JSON_CONTENT, "Invalid blockL1Size $blockL1Size"))
    }

    return Ok(BlockCounters(counters, blockL1Size.toUInt()))
  }

  override fun countTraces(
    traces: JsonObject
  ): Result<VersionedResult<BlockCounters>, TracesError> {
    return concreteCountTraces(traces).map { VersionedResult(tracesVersion, it) }
  }
}
