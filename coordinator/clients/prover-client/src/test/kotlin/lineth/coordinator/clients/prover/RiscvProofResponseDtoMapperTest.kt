package lineth.coordinator.clients.prover

import linea.clients.L2ExecutionProofPublicInputs
import linea.clients.L2ExecutionProofResponseV1
import linea.clients.RollupAggregationProofPublicInputs
import linea.clients.RollupAggregationProofResponseV1
import linea.clients.RollupProofPublicInputs
import linea.clients.RollupProofResponseV1
import linea.kotlin.decodeHex
import lineth.coordinator.clients.prover.serialization.JsonSerialization
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import java.io.File
import kotlin.time.Instant

/**
 * Verifies that the RISC-V proof-response DTO -> domain mappers decode every field (DTO `String` (hex) -> domain
 * `ByteArray`, DTO `Long` -> domain `ULong`), and that the response fixtures shipped with the spec under
 * `rollup_spec/src/rollup_spec/prover_io/testdata/` deserialize into the response DTOs and map onto the domain types.
 */
class RiscvProofResponseDtoMapperTest {

  private val jsonMapper = JsonSerialization.proofResponseMapperV1

  private val executionPublicInputsDto = L2ExecutionProofPublicInputsDto(
    parentBlockHash = "0x0a",
    endBlockHash = "0x0b",
    endBlockNumber = 1000503,
    endBlockTimestamp = 1763000123,
    l2L1Messages = listOf("0x01", "0x1f"),
    parentL1L2BridgeRollingHash = "0x02",
    parentL1L2BridgeRollingHashMessageNumber = 3,
    endL1L2BridgeRollingHash = "0x04",
    endL1L2BridgeRollingHashMessageNumber = 5,
    dynamicChainConfigHash = "0xc0ffee",
    parentFtxRollingHash = "0x06",
    parentFtxNumber = 7,
    endFtxRollingHash = "0x07",
    endProcessedFtxNumber = 8,
    filteredAddressesHash = "0x09",
    txFromsHash = "0x0c",
    blockCount = 3,
    l2MessagingBlocksOffsets = listOf(1, 3),
  )

  private val expectedExecutionPublicInputs = L2ExecutionProofPublicInputs(
    parentBlockHash = "0x0a".decodeHex(),
    endBlockHash = "0x0b".decodeHex(),
    endBlockNumber = 1000503UL,
    endBlockTimestamp = Instant.fromEpochSeconds(1763000123L),
    l2L1Messages = listOf("0x01".decodeHex(), "0x1f".decodeHex()),
    parentL1L2BridgeRollingHash = "0x02".decodeHex(),
    parentL1L2BridgeRollingHashMessageNumber = 3UL,
    endL1L2BridgeRollingHash = "0x04".decodeHex(),
    endL1L2BridgeRollingHashMessageNumber = 5UL,
    dynamicChainConfigHash = "0xc0ffee".decodeHex(),
    parentFtxRollingHash = "0x06".decodeHex(),
    parentFtxNumber = 7UL,
    endFtxRollingHash = "0x07".decodeHex(),
    endFtxNumber = 8UL,
    filteredAddressesHash = "0x09".decodeHex(),
    txFromsHash = "0x0c".decodeHex(),
    blockCount = 3UL,
    l2MessagingBlocksOffsets = listOf(1UL, 3UL),
  )

  private val rollupPublicInputsDto = RollupProofPublicInputsDto(
    endBlockNumber = 1000520,
    endBlockTimestamp = 1763000457,
    parentL1L2BridgeRollingHash = "0x11",
    parentL1L2BridgeRollingHashMessageNumber = 12,
    endL1L2BridgeRollingHash = "0x13",
    endL1L2BridgeRollingHashMessageNumber = 14,
    dynamicChainConfigHash = "0xc0ffee",
    parentFtxRollingHash = "0x15",
    parentFtxNumber = 16,
    endFtxRollingHash = "0x16",
    endProcessedFtxNumber = 17,
    parentDataRollingHash = "0x19",
    endDataRollingHash = "0x1a",
    parentBlockHash = "0x1b",
    endBlockHash = "0x1c",
    parentDataTailTakeBytes = 4,
    finalDataTailDiscardBytes = 1024,
    l2L1Messages = listOf("0xaa"),
    filteredAddresses = listOf("0xbb"),
    programVks = listOf("0xcc"),
    blockCount = 20,
    l2MessagingBlocksOffsets = listOf(2, 7),
  )

  private val expectedRollupPublicInputs = RollupProofPublicInputs(
    endBlockNumber = 1000520UL,
    endBlockTimestamp = Instant.fromEpochSeconds(1763000457L),
    parentL1L2BridgeMessageRollingHash = "0x11".decodeHex(),
    parentL1L2BridgeMessageNumber = 12UL,
    endL1L2BridgeMessageRollingHash = "0x13".decodeHex(),
    endL1L2BridgeMessageNumber = 14UL,
    dynamicChainConfigHash = "0xc0ffee".decodeHex(),
    parentFtxRollingHash = "0x15".decodeHex(),
    parentFtxNumber = 16UL,
    endFtxRollingHash = "0x16".decodeHex(),
    endFtxNumber = 17UL,
    parentDataRollingHash = "0x19".decodeHex(),
    endDataRollingHash = "0x1a".decodeHex(),
    parentBlockHash = "0x1b".decodeHex(),
    endBlockHash = "0x1c".decodeHex(),
    parentDataTailTakeBytes = 4,
    finalDataTailDiscardBytes = 1024,
    l2L1Messages = listOf("0xaa".decodeHex()),
    filteredAddresses = listOf("0xbb".decodeHex()),
    programVks = listOf("0xcc".decodeHex()),
    blockCount = 20UL,
    l2MessagingBlocksOffsets = listOf(2UL, 7UL),
  )

  private val rollupAggregationPublicInputsDto = RollupAggregationProofPublicInputsDto(
    endBlockNumber = 1000567,
    endBlockTimestamp = 1763002301,
    parentL1L2BridgeRollingHash = "0x11",
    parentL1L2BridgeRollingHashMessageNumber = 12,
    endL1L2BridgeRollingHash = "0x13",
    endL1L2BridgeRollingHashMessageNumber = 14,
    dynamicChainConfigHash = "0xc0ffee",
    parentFtxRollingHash = "0x15",
    parentFtxNumber = 16,
    endFtxRollingHash = "0x16",
    endProcessedFtxNumber = 17,
    parentDataRollingHash = "0x19",
    endDataRollingHash = "0x1a",
    parentBlockHash = "0x1b",
    endBlockHash = "0x1c",
    parentDataTailTakeBytes = 0,
    finalDataTailDiscardBytes = 1024,
    l2L1Roots = listOf("0xaa", "0xcc"),
    l2L1TreeDepth = 5,
    filteredAddresses = listOf("0xbb", "0xdd"),
    programIds = listOf("0xee"),
    l2MessagingBlocksOffsets = listOf(1, 20, 100),
  )

  private val expectedRollupAggregationPublicInputs = RollupAggregationProofPublicInputs(
    endBlockNumber = 1000567UL,
    endBlockTimestamp = Instant.fromEpochSeconds(1763002301L),
    parentL1L2BridgeMessageRollingHash = "0x11".decodeHex(),
    parentL1L2BridgeMessageNumber = 12UL,
    endL1L2BridgeMessageRollingHash = "0x13".decodeHex(),
    endL1L2BridgeMessageNumber = 14UL,
    dynamicChainConfigHash = "0xc0ffee".decodeHex(),
    parentFtxRollingHash = "0x15".decodeHex(),
    parentFtxNumber = 16UL,
    endFtxRollingHash = "0x16".decodeHex(),
    endFtxNumber = 17UL,
    parentDataRollingHash = "0x19".decodeHex(),
    endDataRollingHash = "0x1a".decodeHex(),
    parentBlockHash = "0x1b".decodeHex(),
    endBlockHash = "0x1c".decodeHex(),
    parentDataTailTakeBytes = 0,
    finalDataTailDiscardBytes = 1024,
    l2L1Roots = listOf("0xaa".decodeHex(), "0xcc".decodeHex()),
    l2L1TreeDepth = 5UL,
    filteredAddresses = listOf("0xbb".decodeHex(), "0xdd".decodeHex()),
    programIds = listOf("0xee".decodeHex()),
    l2MessagingBlocksOffsets = listOf(1UL, 20UL, 100UL),
  )

  @Test
  fun `L2ExecutionProofResponseDtoMapper decodes every field`() {
    val programVkHex = "0x" + "dd".repeat(32)
    val dto = L2ExecutionProofResponseDto(
      startBlockNumber = 1000500L,
      proof = "0xabcd",
      publicInputs = executionPublicInputsDto,
      filteredAddresses = listOf("0xcc"),
      programVk = programVkHex,
    )

    assertThat(
      L2ExecutionProofResponseDtoMapper(dto),
    ).isEqualTo(
      L2ExecutionProofResponseV1(
        startBlockNumber = 1000500UL,
        endBlockNumber = 1000503UL,
        proof = "0xabcd".decodeHex(),
        publicInputs = expectedExecutionPublicInputs,
        filteredAddresses = listOf("0xcc".decodeHex()),
        programVk = programVkHex.decodeHex(),
      ),
    )
  }

  @Test
  fun `L2ExecutionProofPublicInputs round-trips through its DTO`() {
    assertThat(expectedExecutionPublicInputs.toDto()).isEqualTo(executionPublicInputsDto)
  }

  @Test
  fun `RollupProofResponseDtoMapper decodes every field`() {
    val programVkHex = "0x" + "bb".repeat(32)
    val dto = RollupProofResponseDto(
      startBlockNumber = 1000500L,
      proof = "0xabcd",
      publicInputs = rollupPublicInputsDto,
      programVk = programVkHex,
    )

    assertThat(
      RollupProofResponseDtoMapper(dto),
    ).isEqualTo(
      RollupProofResponseV1(
        startBlockNumber = 1000500UL,
        endBlockNumber = 1000520UL,
        proof = "0xabcd".decodeHex(),
        publicInputs = expectedRollupPublicInputs,
        programVk = programVkHex.decodeHex(),
      ),
    )
  }

  @Test
  fun `RollupAggregationProofResponseDtoMapper decodes every field`() {
    val dto = RollupAggregationProofResponseDto(
      proverVersion = "4.0.0-riscv",
      proof = "0xabcd",
      startBlockNumber = 1000500L,
      publicInputs = rollupAggregationPublicInputsDto,
    )

    assertThat(
      RollupAggregationProofResponseDtoMapper(dto),
    ).isEqualTo(
      RollupAggregationProofResponseV1(
        startBlockNumber = 1000500UL,
        endBlockNumber = 1000567UL,
        proof = "0xabcd".decodeHex(),
        publicInputs = expectedRollupAggregationPublicInputs,
      ),
    )
  }

  @Test
  fun `spec l2-execution proof response fixture parses into the DTO and maps to the domain response`() {
    val dto = jsonMapper.readValue(
      specFixture("10-11-getZkL2ExecutionProofV1.response.json"),
      L2ExecutionProofResponseDto::class.java,
    )
    val response = L2ExecutionProofResponseDtoMapper(dto)

    assertThat(response.startBlockNumber).isEqualTo(1000501UL)
    assertThat(response.endBlockNumber).isEqualTo(1000503UL)
    assertThat(response.publicInputs.l2L1Messages).hasSize(1)
    assertThat(response.publicInputs.l2L1Messages.first()).isEqualTo(ByteArray(32) { 0x08 })
    assertThat(response.publicInputs.txFromsHash).isEqualTo(ByteArray(32) { 0x07 })
    assertThat(response.publicInputs.blockCount).isEqualTo(3UL)
    assertThat(response.publicInputs.l2MessagingBlocksOffsets).isEmpty()
    assertThat(response.filteredAddresses).hasSize(1)
    assertThat(response.programVk).isEqualTo(ByteArray(32) { 0xaa.toByte() })
  }

  @Test
  fun `spec rollup proof response fixture parses into the DTO and maps to the domain response`() {
    val dto = jsonMapper.readValue(
      specFixture("10-14-getZkRollupProofV1.response.json"),
      RollupProofResponseDto::class.java,
    )
    val response = RollupProofResponseDtoMapper(dto)

    assertThat(response.startBlockNumber).isEqualTo(1000501UL)
    assertThat(response.endBlockNumber).isEqualTo(1000520UL)
    assertThat(response.publicInputs.parentDataTailTakeBytes).isEqualTo(4)
    assertThat(response.publicInputs.finalDataTailDiscardBytes).isEqualTo(0)
    assertThat(response.publicInputs.l2L1Messages).hasSize(2)
    assertThat(response.publicInputs.filteredAddresses).hasSize(2)
    assertThat(response.publicInputs.programVks).hasSize(1)
    assertThat(response.publicInputs.blockCount).isEqualTo(20UL)
    assertThat(response.programVk).isEqualTo(ByteArray(32) { 0xbb.toByte() })
  }

  @Test
  fun `spec rollup-aggregation proof response fixture parses into the DTO and maps to the domain response`() {
    val dto = jsonMapper.readValue(
      specFixture("10-18-getZkRollupAggregationProofV1.response.json"),
      RollupAggregationProofResponseDto::class.java,
    )
    val response = RollupAggregationProofResponseDtoMapper(dto)

    assertThat(response.startBlockNumber).isEqualTo(1000501UL)
    assertThat(response.endBlockNumber).isEqualTo(1000520UL)
    assertThat(response.publicInputs.l2L1TreeDepth).isEqualTo(5UL)
    assertThat(response.publicInputs.l2L1Roots).hasSize(1)
    assertThat(response.publicInputs.filteredAddresses).hasSize(2)
    assertThat(response.publicInputs.programIds)
      .containsExactly(ByteArray(32) { 0x11 }, ByteArray(32) { 0x22 })
    assertThat(response.publicInputs.l2MessagingBlocksOffsets).isEmpty()
  }

  private fun specFixture(fileName: String): File =
    File("../../../rollup_spec/src/rollup_spec/prover_io/testdata/$fileName")
}
