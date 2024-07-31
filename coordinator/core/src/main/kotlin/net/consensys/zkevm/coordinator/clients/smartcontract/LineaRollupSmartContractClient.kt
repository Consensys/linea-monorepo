package net.consensys.zkevm.coordinator.clients.smartcontract

import net.consensys.zkevm.domain.BlobRecord
import net.consensys.zkevm.domain.ProofToFinalize
import net.consensys.zkevm.ethereum.gaspricing.GasPriceCaps
import tech.pegasys.teku.infrastructure.async.SafeFuture

enum class LineaContractVersion : Comparable<LineaContractVersion> {
  V5 // "EIP4844 multiple blobs per tx support - version in all networks",
}

interface LineaRollupSmartContractClientReadOnly : ContractVersionProvider<LineaContractVersion> {

  fun getAddress(): String

  /**
   * Get the current L2 block number
   */
  fun finalizedL2BlockNumber(blockParameter: BlockParameter = BlockParameter.Tag.LATEST): SafeFuture<ULong>

  /**
   * Get the current L2 block timestamp
   */
  fun finalizedL2BlockTimestamp(blockParameter: BlockParameter = BlockParameter.Tag.LATEST): SafeFuture<ULong>

  fun getMessageRollingHash(
    blockParameter: BlockParameter = BlockParameter.Tag.LATEST,
    messageNumber: Long
  ): SafeFuture<ByteArray>

  /**
   * Get the final block number of a shnarf
   */
  fun findBlobFinalBlockNumberByShnarf(
    blockParameter: BlockParameter = BlockParameter.Tag.LATEST,
    shnarf: ByteArray
  ): SafeFuture<ULong?>

  /**
   * Gets Type 2 StateRootHash for Linea Block
   */
  fun blockStateRootHash(
    blockParameter: BlockParameter,
    lineaL2BlockNumber: ULong
  ): SafeFuture<ByteArray>
}

data class BlockAndNonce(
  val blockNumber: ULong,
  val nonce: ULong
)

interface LineaRollupSmartContractClient : LineaRollupSmartContractClientReadOnly {

  fun currentNonce(): ULong

  /**
   * Fetches LATEST block from L1, correspondent nonce at that block
   * and sets internal state to those
   */
  fun updateNonceAndReferenceBlockToLastL1Block(): SafeFuture<BlockAndNonce>

  /**
   *  Simulates the sending of a list of blobs to the smart contract, with EIP4844 transaction.
   */
  fun submitBlobsEthCall(
    blobs: List<BlobRecord>,
    gasPriceCaps: GasPriceCaps?
  ): SafeFuture<String?>

  fun finalizeBlocksEthCall(
    aggregation: ProofToFinalize,
    aggregationLastBlob: BlobRecord,
    parentShnarf: ByteArray,
    parentL1RollingHash: ByteArray,
    parentL1RollingHashMessageNumber: Long
  ): SafeFuture<String?>

  /**
   * Submit a list of blobs to the smart contract, with EIP4844 transaction
   */
  fun submitBlobs(
    blobs: List<BlobRecord>,
    gasPriceCaps: GasPriceCaps?
  ): SafeFuture<String>

  fun finalizeBlocks(
    aggregation: ProofToFinalize,
    aggregationLastBlob: BlobRecord,
    parentShnarf: ByteArray,
    parentL1RollingHash: ByteArray,
    parentL1RollingHashMessageNumber: Long,
    gasPriceCaps: GasPriceCaps?
  ): SafeFuture<String>
}

interface LineaGenesisStateProvider {
  val stateRootHash: ByteArray
  val shnarf: ByteArray
}
