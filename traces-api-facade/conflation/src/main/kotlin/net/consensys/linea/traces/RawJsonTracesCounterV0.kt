package net.consensys.linea.traces

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import com.github.michaelbull.result.map
import io.vertx.core.json.JsonObject
import net.consensys.linea.BlockCounters
import net.consensys.linea.ErrorType
import net.consensys.linea.TracesCounterV0
import net.consensys.linea.TracesError
import net.consensys.linea.VersionedResult
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger

class RawJsonTracesCounterV0(private val tracesVersion: String) : TracesCounterV0 {
  private val log: Logger = LogManager.getLogger(this::class.java)

  private val emptyCounters = TracingModuleV1.values().associateWith { 0.toUInt() }

  fun concreteCountTraces(trace: JsonObject): Result<BlockCounters, TracesError> {
    val counters = emptyCounters.toMutableMap()
    // Module limits
    MODULES.forEach { moduleInfo ->

      val (jsonPath, klass) = moduleInfo
      trace.getTrace(jsonPath)?.let { traceJsonObject ->
        if (!traceJsonObject.isEmpty) {
          traceJsonObject.mapTo(klass)?.also {
            when (it) {
              is Add -> counters[TracingModuleV1.ADD] = it.ACC_1.size.toUInt() + 2U
              is Bin -> counters[TracingModuleV1.BIN] = it.ACC_1.size.toUInt() + 16U
              is Ext -> counters[TracingModuleV1.EXT] = it.ACC_A_0.size.toUInt() + 8U
              is EcData -> counters[TracingModuleV1.EC_DATA] = it.ACC_DELTA.size.toUInt() + 12U
              is HashData -> counters[TracingModuleV1.PUB_HASH] = it.INDEX.size.toUInt()
              is HashInfo -> counters[TracingModuleV1.PUB_HASH_INFO] = it.HASH_HI.size.toUInt()
              is Hub -> counters[TracingModuleV1.HUB] = it.ALPHA.size.toUInt() + 2U
              is LogData -> counters[TracingModuleV1.PUB_LOG] = it.INDEX.size.toUInt()
              is LogInfo -> counters[TracingModuleV1.PUB_LOG_INFO] = it.ADDR_HI.size.toUInt()
              is Mmio -> counters[TracingModuleV1.MMIO] = it.ACC_1.size.toUInt() // TODO: get spilling
              is Mmu -> counters[TracingModuleV1.MMU] = it.ACC_1.size.toUInt() // TODO: get spilling
              is Mod -> counters[TracingModuleV1.MOD] = it.ACC_1_2.size.toUInt() + 8U
              is Mul -> counters[TracingModuleV1.MUL] = it.ACC_A_0.size.toUInt() + 9U
              is Mxp -> counters[TracingModuleV1.MXP] = it.ACC_1.size.toUInt() + 4U
              is PhoneyRlp -> counters[TracingModuleV1.PHONEY_RLP] = it.INDEX.size.toUInt()
              is Rlp -> counters[TracingModuleV1.RLP] = it.ADDR_HI.size.toUInt() + 8U
              is Rom -> counters[TracingModuleV1.ROM] = it.PC.size.toUInt() + 2U
              is Shf -> counters[TracingModuleV1.SHF] = it.ACC_1.size.toUInt() + 16U
              is TxRlp -> counters[TracingModuleV1.TX_RLP] = it.ABS_TX_NUM.size.toUInt()
              is Wcp -> counters[TracingModuleV1.WCP] = it.ACC_1.size.toUInt() + 16U
              //
              // These modules are constant-sized, so they do not matter to the blocks
              // conflation counters
              //
              is MmuId -> counters[TracingModuleV1.MMU_ID] = 0.toUInt()
              is ShfRt -> counters[TracingModuleV1.SHF_RT] = 0.toUInt()
              is InstructionDecoder -> counters[TracingModuleV1.INSTRUCTION_DECODER] = 0.toUInt()
              is BinRt -> counters[TracingModuleV1.BIN_RT] = 0.toUInt()
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
    counters[TracingModuleV1.BLOCK_TX] = trace.getLong("TxCount").toUInt()
    counters[TracingModuleV1.BLOCK_L2L1LOGS] = trace.getLong("L2L1logsCount").toUInt()
    counters[TracingModuleV1.BLOCK_KECCAK] = trace.getLong("KeccakCount").toUInt()

    // Precompile limits
    val precompiles = trace.getJsonObject("PrecompileCalls")
    counters[TracingModuleV1.PRECOMPILE_ECRECOVER] = precompiles.getLong("EcRecover").toUInt()
    counters[TracingModuleV1.PRECOMPILE_SHA2] = precompiles.getLong("Sha2").toUInt()
    counters[TracingModuleV1.PRECOMPILE_RIPEMD] = precompiles.getLong("RipeMD").toUInt()
    counters[TracingModuleV1.PRECOMPILE_IDENTITY] = precompiles.getLong("Identity").toUInt()
    counters[TracingModuleV1.PRECOMPILE_MODEXP] = precompiles.getLong("ModExp").toUInt()
    counters[TracingModuleV1.PRECOMPILE_ECADD] = precompiles.getLong("EcAdd").toUInt()
    counters[TracingModuleV1.PRECOMPILE_ECMUL] = precompiles.getLong("EcMul").toUInt()
    counters[TracingModuleV1.PRECOMPILE_ECPAIRING] = precompiles.getLong("EcPairing").toUInt()
    counters[TracingModuleV1.PRECOMPILE_BLAKE2F] = precompiles.getLong("Blake2f").toUInt()

    val blockL1Size = trace.getLong("blockL1Size")
    if (blockL1Size == null || blockL1Size < 0) {
      Err(TracesError(ErrorType.WRONG_JSON_CONTENT, "Invalid blockL1Size $blockL1Size"))
    }

    return Ok(BlockCounters(TracesCountersV1(counters), blockL1Size.toUInt()))
  }

  override fun countTraces(
    traces: JsonObject
  ): Result<VersionedResult<BlockCounters>, TracesError> {
    return concreteCountTraces(traces).map { VersionedResult(tracesVersion, it) }
  }
}
