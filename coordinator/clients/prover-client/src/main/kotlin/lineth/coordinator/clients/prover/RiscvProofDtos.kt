package lineth.coordinator.clients.prover

import com.fasterxml.jackson.annotation.JsonInclude
import com.fasterxml.jackson.annotation.JsonProperty
import linea.clients.ConflationWitness
import linea.clients.ForcedTransaction
import linea.clients.L2ExecutionProofPublicInputs
import linea.clients.RollupAggregationProofPublicInputs
import linea.clients.RollupChunk
import linea.clients.RollupProofPublicInputs
import linea.domain.BlockInterval
import linea.domain.ExecutionPayload
import linea.ethapi.ExecutionWitness
import linea.forcedtx.ForcedTransactionInclusionResult
import linea.kotlin.decodeHex
import linea.kotlin.encodeHex
import linea.kotlin.toHexString
import kotlin.time.Instant

/**
 * Request/response DTOs for the RISC-V provers.
 *
 * These mirror the JSON schemas under `rollup_spec/src/rollup_spec/prover_io/schemas/`. Field names are kept
 * identical to the JSON keys so they serialize without custom naming.
 *
 * The DTO <-> domain mapping lives in `RiscvProofDtos.kt`.
 */

/** The 18-field PI tuple emitted by a l2-execution proof (rollup_spec §2.1). */
data class L2ExecutionProofPublicInputsDto(
  val parentBlockHash: String,
  val endBlockHash: String,
  val endBlockNumber: Long,
  val endBlockTimestamp: Long,
  val l2L1Messages: List<String>,
  val parentL1L2BridgeRollingHash: String,
  val parentL1L2BridgeRollingHashMessageNumber: Long,
  val endL1L2BridgeRollingHash: String,
  val endL1L2BridgeRollingHashMessageNumber: Long,
  val dynamicChainConfigHash: String,
  val parentFtxRollingHash: String,
  val parentFtxNumber: Long,
  val endFtxRollingHash: String,
  val endProcessedFtxNumber: Long,
  val filteredAddressesHash: String,
  val txFromsHash: String,
  val blockCount: Long,
  val l2MessagingBlocksOffsets: List<Long>,
)

/** The 22-field PI tuple emitted by a rollup proof (rollup_spec §2.2). */
data class RollupProofPublicInputsDto(
  val endBlockNumber: Long,
  val endBlockTimestamp: Long,
  val parentL1L2BridgeRollingHash: String,
  val parentL1L2BridgeRollingHashMessageNumber: Long,
  val endL1L2BridgeRollingHash: String,
  val endL1L2BridgeRollingHashMessageNumber: Long,
  val dynamicChainConfigHash: String,
  val parentFtxRollingHash: String,
  val parentFtxNumber: Long,
  val endFtxRollingHash: String,
  val endProcessedFtxNumber: Long,
  val parentDataRollingHash: String,
  val endDataRollingHash: String,
  val parentBlockHash: String,
  val endBlockHash: String,
  val parentDataTailTakeBytes: Int,
  val finalDataTailDiscardBytes: Int,
  val l2L1Messages: List<String>,
  val filteredAddresses: List<String>,
  val programVks: List<String>,
  val blockCount: Long,
  val l2MessagingBlocksOffsets: List<Long>,
)

/** The 22-field PI tuple emitted by a rollup-aggregation proof (rollup_spec §2.3). */
data class RollupAggregationProofPublicInputsDto(
  val endBlockNumber: Long,
  val endBlockTimestamp: Long,
  val parentL1L2BridgeRollingHash: String,
  val parentL1L2BridgeRollingHashMessageNumber: Long,
  val endL1L2BridgeRollingHash: String,
  val endL1L2BridgeRollingHashMessageNumber: Long,
  val dynamicChainConfigHash: String,
  val parentFtxRollingHash: String,
  val parentFtxNumber: Long,
  val endFtxRollingHash: String,
  val endProcessedFtxNumber: Long,
  val parentDataRollingHash: String,
  val endDataRollingHash: String,
  val parentBlockHash: String,
  val endBlockHash: String,
  val parentDataTailTakeBytes: Int,
  val finalDataTailDiscardBytes: Int,
  val l2L1Roots: List<String>,
  val l2L1TreeDepth: Long,
  val filteredAddresses: List<String>,
  val programIds: List<String>,
  val l2MessagingBlocksOffsets: List<Long>,
)

data class MetaDataDto(
  val startBlockNumber: Long,
  val endBlockNumber: Long,
  val startBlockTimestamp: Long,
  @get:JsonInclude(JsonInclude.Include.NON_NULL)
  val endBlockTimestamp: Long? = null,
  @get:JsonInclude(JsonInclude.Include.NON_NULL)
  val transactionsCount: Long? = null,
  @get:JsonInclude(JsonInclude.Include.NON_NULL)
  val totalGasUsed: Long? = null,
)

// ---------------------------------------------------------------------------------------------------------------------
// getZkL2ExecutionProof.request.json (§2.1)
// ---------------------------------------------------------------------------------------------------------------------

/** One of the five `ForcedTransactionAcceptance` variants (rollup_spec §6.5). */
enum class ForcedTransactionAcceptance {
  INCLUDED,
  BAD_NONCE,
  BAD_BALANCE,
  FILTERED_ADDRESS_FROM,
  FILTERED_ADDRESS_TO,
}

/** Per-FTX metadata (rollup_spec §6.5). */
data class ForcedTransactionDto(
  val number: Long,
  val deadline: Long,
  val signedTxRlp: String,
  val acceptance: ForcedTransactionAcceptance,
)

data class WithdrawalDto(
  val index: Long,
  val validatorIndex: Long,
  val address: String,
  val amount: Long,
)

// Execution payload including Amsterdam blockAccessList and slotNumber.
data class ExecutionPayloadDto(
  val parentHash: String,
  val feeRecipient: String,
  val stateRoot: String,
  val receiptsRoot: String,
  val logsBloom: String,
  val prevRandao: String,
  val blockNumber: Long,
  val gasLimit: Long,
  val gasUsed: Long,
  val timestamp: Long,
  val extraData: String,
  val baseFeePerGas: String,
  val blockHash: String,
  val transactions: List<String>,
  val withdrawals: List<WithdrawalDto>,
  val blobGasUsed: Long,
  val excessBlobGas: Long,
  val blockAccessList: String,
  @get:JsonInclude(JsonInclude.Include.NON_NULL)
  val slotNumber: Long? = null,
)

data class NewPayloadRequestDto(
  val executionPayload: ExecutionPayloadDto,
  val versionedHashes: List<String>,
  val parentBeaconBlockRoot: String,
  val executionRequests: List<String>,
)

/** Static chain configuration (preimage inputs of `dynamicChainConfigHash`). */
data class ChainConfigDto(
  val l2MessageServiceAddress: String,
  val coinbase: String,
  val chainId: Long,
  val forkName: String,
)

/** A single `debug_executionWitness` entry, one per block in canonical order. */
data class ExecutionWitnessDto(
  val state: List<String>,
  val codes: List<String>,
  val headers: List<String>,
)

data class RollupExtensionDto(
  val forcedTransactions: List<ForcedTransactionDto>,
)

data class StatelessInputDto(
  val newPayloadRequest: NewPayloadRequestDto,
  val executionWitness: ExecutionWitnessDto,
)

data class PayloadInputDto(
  val statelessInput: StatelessInputDto,
  val rollupExtension: RollupExtensionDto,
)

data class L2ExecutionProofRequestParamsDto(
  val parentFtxRollingHash: String,
  val parentFtxNumber: Long,
  val payloads: List<PayloadInputDto>,
  val chainConfig: ChainConfigDto,
)

data class L2ExecutionProofRequestDto(
  val guestProgramId: String,
  val provingSystem: String,
  val proofRequest: L2ExecutionProofRequestParamsDto,
  val metadata: MetaDataDto,
)

data class L2ExecutionProofResponseDto(
  val startBlockNumber: Long,
  val proof: String,
  val publicInputs: L2ExecutionProofPublicInputsDto,
  val filteredAddresses: List<String>,
  val programVk: String,
)

// ---------------------------------------------------------------------------------------------------------------------
// getZkRollupProof.request.json (§2.2)
// ---------------------------------------------------------------------------------------------------------------------

data class ConflationWitnessDto(
  val blockRlps: List<String>,
)

data class BlockIntervalDto(
  val startBlockNumber: Long,
  val endBlockNumber: Long,
) {
  companion object {
    fun BlockInterval.toDto() = BlockIntervalDto(
      startBlockNumber = this.startBlockNumber.toLong(),
      endBlockNumber = this.endBlockNumber.toLong(),
    )
  }
}

/** One touched DA chunk with its anchored hash and exact physical bytes (rollup_spec §3.1). */
data class RollupChunkDto(
  val chunkHash: String,
  @get:JsonProperty("isCalldata")
  val isCalldata: Boolean,
  val blobBytes: String,
  val calldataBytes: String,
)

data class RollupProofRequestParamsDto(
  val chainId: Long,
  val conflations: List<ConflationWitnessDto>,
  @get:JsonInclude(JsonInclude.Include.NON_NULL)
  val l2ExecutionProofs: List<L2ExecutionProofResponseDto>? = null,
  @get:JsonInclude(JsonInclude.Include.NON_NULL)
  val l2ExecutionProofIndexes: List<BlockIntervalDto>? = null,
  val chunks: List<RollupChunkDto>,
  val parentDataRollingHash: String,
  val parentDataTailTakeBytes: Int,
  @get:JsonInclude(JsonInclude.Include.NON_NULL)
  val boundaryPrevDataRollingHash: String? = null,
)

data class FileBasedRollupProofRequestDto(
  val guestProgramId: String,
  val provingSystem: String,
  val proofRequest: RollupProofRequestParamsDto,
  val metadata: MetaDataDto,
)

data class RollupProofResponseDto(
  // val proverVersion: String,
  val startBlockNumber: Long,
  val proof: String,
  val publicInputs: RollupProofPublicInputsDto,
  val programVk: String,
)

// ---------------------------------------------------------------------------------------------------------------------
// getZkRollupAggregationProof.request.json (§2.3)
// ---------------------------------------------------------------------------------------------------------------------

data class FileBasedRollupAggregationProofRequestParamsDto(
  @get:JsonInclude(JsonInclude.Include.NON_NULL)
  val rollupProofs: List<RollupProofResponseDto>? = null,
  @get:JsonInclude(JsonInclude.Include.NON_NULL)
  val rollupProofsIndexes: List<BlockIntervalDto>? = null,
)

data class FileBasedRollupAggregationProofRequestDto(
  val guestProgramId: String,
  val provingSystem: String,
  val proofRequest: FileBasedRollupAggregationProofRequestParamsDto,
  val metadata: MetaDataDto,
)

/** Response of a rollup-aggregation proof: the aggregated proof bytes plus the 22-field PI tuple (§2.3). */
data class RollupAggregationProofResponseDto(
  val proverVersion: String,
  val startBlockNumber: Long,
  val proof: String,
  val publicInputs: RollupAggregationProofPublicInputsDto,
)

// ---------------------------------------------------------------------------------------------------------------------
// to/fromDomainObject helper functions between the RISC-V proof DTOs (RiscvProofDtos.kt) and their domain twins.
// ---------------------------------------------------------------------------------------------------------------------

internal fun L2ExecutionProofPublicInputsDto.toDomainObject(): L2ExecutionProofPublicInputs {
  return L2ExecutionProofPublicInputs(
    parentBlockHash = parentBlockHash.decodeHex(),
    endBlockHash = endBlockHash.decodeHex(),
    endBlockNumber = endBlockNumber.toULong(),
    endBlockTimestamp = Instant.fromEpochSeconds(endBlockTimestamp),
    l2L1Messages = l2L1Messages.map { it.decodeHex() },
    parentL1L2BridgeRollingHash = parentL1L2BridgeRollingHash.decodeHex(),
    parentL1L2BridgeRollingHashMessageNumber = parentL1L2BridgeRollingHashMessageNumber.toULong(),
    endL1L2BridgeRollingHash = endL1L2BridgeRollingHash.decodeHex(),
    endL1L2BridgeRollingHashMessageNumber = endL1L2BridgeRollingHashMessageNumber.toULong(),
    dynamicChainConfigHash = dynamicChainConfigHash.decodeHex(),
    parentFtxRollingHash = parentFtxRollingHash.decodeHex(),
    parentFtxNumber = parentFtxNumber.toULong(),
    endFtxRollingHash = endFtxRollingHash.decodeHex(),
    endFtxNumber = endProcessedFtxNumber.toULong(),
    filteredAddressesHash = filteredAddressesHash.decodeHex(),
    txFromsHash = txFromsHash.decodeHex(),
    blockCount = blockCount.toULong(),
    l2MessagingBlocksOffsets = l2MessagingBlocksOffsets.map { it.toULong() },
  )
}

internal fun L2ExecutionProofPublicInputs.toDto(): L2ExecutionProofPublicInputsDto {
  return L2ExecutionProofPublicInputsDto(
    parentBlockHash = parentBlockHash.encodeHex(),
    endBlockHash = endBlockHash.encodeHex(),
    endBlockNumber = endBlockNumber.toLong(),
    endBlockTimestamp = endBlockTimestamp.epochSeconds,
    l2L1Messages = l2L1Messages.map { it.encodeHex() },
    parentL1L2BridgeRollingHash = parentL1L2BridgeRollingHash.encodeHex(),
    parentL1L2BridgeRollingHashMessageNumber = parentL1L2BridgeRollingHashMessageNumber.toLong(),
    endL1L2BridgeRollingHash = endL1L2BridgeRollingHash.encodeHex(),
    endL1L2BridgeRollingHashMessageNumber = endL1L2BridgeRollingHashMessageNumber.toLong(),
    dynamicChainConfigHash = dynamicChainConfigHash.encodeHex(),
    parentFtxRollingHash = parentFtxRollingHash.encodeHex(),
    parentFtxNumber = parentFtxNumber.toLong(),
    endFtxRollingHash = endFtxRollingHash.encodeHex(),
    endProcessedFtxNumber = endFtxNumber.toLong(),
    filteredAddressesHash = filteredAddressesHash.encodeHex(),
    txFromsHash = txFromsHash.encodeHex(),
    blockCount = blockCount.toLong(),
    l2MessagingBlocksOffsets = l2MessagingBlocksOffsets.map { it.toLong() },
  )
}

internal fun ExecutionPayload.toDto(): ExecutionPayloadDto {
  return ExecutionPayloadDto(
    parentHash = parentHash.encodeHex(),
    feeRecipient = feeRecipient.encodeHex(),
    stateRoot = stateRoot.encodeHex(),
    receiptsRoot = receiptsRoot.encodeHex(),
    logsBloom = logsBloom.encodeHex(),
    prevRandao = prevRandao.encodeHex(),
    blockNumber = blockNumber.toLong(),
    gasLimit = gasLimit.toLong(),
    gasUsed = gasUsed.toLong(),
    timestamp = timestamp.toLong(),
    extraData = extraData.encodeHex(),
    baseFeePerGas = baseFeePerGas.toHexString(),
    blockHash = blockHash.encodeHex(),
    transactions = transactions.map { it.encodeHex() },
    withdrawals = withdrawals.map {
      WithdrawalDto(
        index = it.index.toLong(),
        validatorIndex = it.validatorIndex.toLong(),
        address = it.address.encodeHex(),
        amount = it.amount.toLong(),
      )
    },
    blobGasUsed = blobGasUsed.toLong(),
    excessBlobGas = excessBlobGas.toLong(),
    blockAccessList = blockAccessList.encodeHex(),
    slotNumber = slotNumber?.toLong(),
  )
}

internal fun ExecutionWitness.toDto(): ExecutionWitnessDto {
  return ExecutionWitnessDto(
    state = state.map { it.encodeHex() },
    codes = codes.map { it.encodeHex() },
    headers = headers.map { it.encodeHex() },
  )
}

private fun mapFtxInclusionResultToAcceptance(
  inclusionResult: ForcedTransactionInclusionResult,
): ForcedTransactionAcceptance {
  return when (inclusionResult) {
    ForcedTransactionInclusionResult.Included -> ForcedTransactionAcceptance.INCLUDED
    ForcedTransactionInclusionResult.BadNonce -> ForcedTransactionAcceptance.BAD_NONCE
    ForcedTransactionInclusionResult.BadBalance -> ForcedTransactionAcceptance.BAD_BALANCE
    ForcedTransactionInclusionResult.FilteredAddressFrom -> ForcedTransactionAcceptance.FILTERED_ADDRESS_FROM
    ForcedTransactionInclusionResult.FilteredAddressTo -> ForcedTransactionAcceptance.FILTERED_ADDRESS_TO
    else -> throw IllegalArgumentException("Unsupported FTX inclusion result: $inclusionResult")
  }
}

fun ForcedTransaction.toDto(): ForcedTransactionDto {
  return ForcedTransactionDto(
    number = ftxNumber.toLong(),
    deadline = deadlineBlockNumber.toLong(),
    signedTxRlp = signedTxRlp.encodeHex(),
    acceptance = mapFtxInclusionResultToAcceptance(acceptance),
  )
}

/** Maps the RISC-V rollup PI tuple DTO onto its domain twin (rollup_spec §2.2). */
fun RollupProofPublicInputsDto.toDomainObject(): RollupProofPublicInputs {
  return RollupProofPublicInputs(
    endBlockNumber = endBlockNumber.toULong(),
    endBlockTimestamp = Instant.fromEpochSeconds(endBlockTimestamp),
    parentL1L2BridgeMessageRollingHash = parentL1L2BridgeRollingHash.decodeHex(),
    parentL1L2BridgeMessageNumber = parentL1L2BridgeRollingHashMessageNumber.toULong(),
    endL1L2BridgeMessageRollingHash = endL1L2BridgeRollingHash.decodeHex(),
    endL1L2BridgeMessageNumber = endL1L2BridgeRollingHashMessageNumber.toULong(),
    dynamicChainConfigHash = dynamicChainConfigHash.decodeHex(),
    parentFtxRollingHash = parentFtxRollingHash.decodeHex(),
    parentFtxNumber = parentFtxNumber.toULong(),
    endFtxRollingHash = endFtxRollingHash.decodeHex(),
    endFtxNumber = endProcessedFtxNumber.toULong(),
    parentDataRollingHash = parentDataRollingHash.decodeHex(),
    endDataRollingHash = endDataRollingHash.decodeHex(),
    parentBlockHash = parentBlockHash.decodeHex(),
    endBlockHash = endBlockHash.decodeHex(),
    parentDataTailTakeBytes = parentDataTailTakeBytes,
    finalDataTailDiscardBytes = finalDataTailDiscardBytes,
    l2L1Messages = l2L1Messages.map { it.decodeHex() },
    filteredAddresses = filteredAddresses.map { it.decodeHex() },
    programVks = programVks.map { it.decodeHex() },
    blockCount = blockCount.toULong(),
    l2MessagingBlocksOffsets = l2MessagingBlocksOffsets.map { it.toULong() },
  )
}

/** Maps the RISC-V rollup-aggregation PI tuple DTO onto its domain twin (rollup_spec §2.3). */
fun RollupAggregationProofPublicInputsDto.toDomainObject(): RollupAggregationProofPublicInputs {
  return RollupAggregationProofPublicInputs(
    endBlockNumber = endBlockNumber.toULong(),
    endBlockTimestamp = Instant.fromEpochSeconds(endBlockTimestamp),
    parentL1L2BridgeMessageRollingHash = parentL1L2BridgeRollingHash.decodeHex(),
    parentL1L2BridgeMessageNumber = parentL1L2BridgeRollingHashMessageNumber.toULong(),
    endL1L2BridgeMessageRollingHash = endL1L2BridgeRollingHash.decodeHex(),
    endL1L2BridgeMessageNumber = endL1L2BridgeRollingHashMessageNumber.toULong(),
    dynamicChainConfigHash = dynamicChainConfigHash.decodeHex(),
    parentFtxRollingHash = parentFtxRollingHash.decodeHex(),
    parentFtxNumber = parentFtxNumber.toULong(),
    endFtxRollingHash = endFtxRollingHash.decodeHex(),
    endFtxNumber = endProcessedFtxNumber.toULong(),
    parentDataRollingHash = parentDataRollingHash.decodeHex(),
    endDataRollingHash = endDataRollingHash.decodeHex(),
    parentBlockHash = parentBlockHash.decodeHex(),
    endBlockHash = endBlockHash.decodeHex(),
    parentDataTailTakeBytes = parentDataTailTakeBytes,
    finalDataTailDiscardBytes = finalDataTailDiscardBytes,
    l2L1Roots = l2L1Roots.map { it.decodeHex() },
    l2L1TreeDepth = l2L1TreeDepth.toULong(),
    filteredAddresses = filteredAddresses.map { it.decodeHex() },
    programIds = programIds.map { it.decodeHex() },
    l2MessagingBlocksOffsets = l2MessagingBlocksOffsets.map { it.toULong() },
  )
}

fun RollupChunk.toDto(): RollupChunkDto {
  return RollupChunkDto(
    chunkHash = chunkHash.encodeHex(),
    isCalldata = isCalldata,
    blobBytes = blobBytes.encodeHex(),
    calldataBytes = calldataBytes.encodeHex(),
  )
}

fun ConflationWitness.toDto(): ConflationWitnessDto {
  return ConflationWitnessDto(
    blockRlps = blockRlps.map { it.encodeHex() },
  )
}
