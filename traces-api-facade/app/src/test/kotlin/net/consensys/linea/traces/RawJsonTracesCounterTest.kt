package net.consensys.linea.traces

import net.consensys.linea.async.get
import net.consensys.linea.traces.repository.FilesystemHelper
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.mockito.Mockito
import java.nio.file.Path

class RawJsonTracesCounterTest {
  private val log: Logger = LogManager.getLogger(this::class.java)
  var counter = RawJsonTracesCounter("test")

  @Test
  fun parse_file() {
    val fsHelper = FilesystemHelper(Mockito.mock(), log = log)

    val parsed = fsHelper.readGzippedJsonFileAsString(Path.of("../../testdata/traces/raw/small.json.gz"))

    val counted = counter.concreteCountTraces(parsed.get()).component1()
      ?.tracesCounters
      ?.entries()
      ?.associate { e -> e.first.name to e.second }

    assertThat(counted).isEqualTo(
      mapOf(
        "ADD" to 514U,
        "BIN" to 256U,
        "BIN_RT" to 0U,
        "EC_DATA" to 12U,
        "EXT" to 8U,
        "HUB" to 1361U,
        "INSTRUCTION_DECODER" to 0U,
        "MMIO" to 209U,
        "MMU" to 188U,
        "MMU_ID" to 0U,
        "MOD" to 16U,
        "MUL" to 14U,
        "MXP" to 128U,
        "PHONEY_RLP" to 0U,
        "PUB_HASH" to 8U,
        "PUB_HASH_INFO" to 2U,
        "PUB_LOG" to 14U,
        "PUB_LOG_INFO" to 1U,
        "RLP" to 8U,
        "ROM" to 11266U,
        "SHF" to 48U,
        "SHF_RT" to 0U,
        "TX_RLP" to 365U,
        "WCP" to 226U,
        "BLOCK_TX" to 1U,
        "BLOCK_L2L1LOGS" to 0U,
        "BLOCK_KECCAK" to 11U,
        "PRECOMPILE_ECRECOVER" to 0U,
        "PRECOMPILE_SHA2" to 0U,
        "PRECOMPILE_RIPEMD" to 0U,
        "PRECOMPILE_IDENTITY" to 0U,
        "PRECOMPILE_MODEXP" to 0U,
        "PRECOMPILE_ECADD" to 0U,
        "PRECOMPILE_ECMUL" to 0U,
        "PRECOMPILE_ECPAIRING" to 0U,
        "PRECOMPILE_BLAKE2F" to 0U
      )
    )
  }

  @Test
  fun parse_string() {
    val parsed = """
        {"blockL1Size":1000,"KeccakCount":11,"L2L1logsCount":0,"TxCount":1,"PrecompileCalls":{"EcRecover":1,"Sha2":0,"RipeMD":0,"Identity":0,"ModExp":0,
        "EcAdd":0,"EcMul":0,"EcPairing":0,"Blake2f":0}}
    """.trimIndent()

    val counted = counter.concreteCountTraces(parsed)
    assertThat(counted.component1()?.blockL1Size).isEqualTo(1000U)
    assertThat(counted.component1()?.tracesCounters?.entries()?.associate { e -> e.first.name to e.second }).isEqualTo(
      mapOf(
        "ADD" to 0U,
        "BIN" to 0U,
        "BIN_RT" to 0U,
        "BLOCK_KECCAK" to 11U,
        "BLOCK_L2L1LOGS" to 0U,
        "BLOCK_TX" to 1U,
        "EC_DATA" to 0U,
        "EXT" to 0U,
        "HUB" to 0U,
        "INSTRUCTION_DECODER" to 0U,
        "MMIO" to 0U,
        "MMU" to 0U,
        "MMU_ID" to 0U,
        "MOD" to 0U,
        "MUL" to 0U,
        "MXP" to 0U,
        "PHONEY_RLP" to 0U,
        "PRECOMPILE_BLAKE2F" to 0U,
        "PRECOMPILE_ECADD" to 0U,
        "PRECOMPILE_ECMUL" to 0U,
        "PRECOMPILE_ECPAIRING" to 0U,
        "PRECOMPILE_ECRECOVER" to 1U,
        "PRECOMPILE_IDENTITY" to 0U,
        "PRECOMPILE_MODEXP" to 0U,
        "PRECOMPILE_RIPEMD" to 0U,
        "PRECOMPILE_SHA2" to 0U,
        "PUB_HASH" to 0U,
        "PUB_HASH_INFO" to 0U,
        "PUB_LOG" to 0U,
        "PUB_LOG_INFO" to 0U,
        "RLP" to 0U,
        "ROM" to 0U,
        "SHF" to 0U,
        "SHF_RT" to 0U,
        "TX_RLP" to 0U,
        "WCP" to 0U
      )
    )
  }
}
