package net.consensys.zkevm.ethereum.settlement

import net.consensys.linea.contract.ZkEvmV2
import net.consensys.linea.contract.ZkEvmV2AsyncFriendly
import net.consensys.zkevm.coordinator.clients.GetProofResponse
import net.consensys.zkevm.ethereum.coordination.conflation.Batch
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes
import org.web3j.abi.FunctionEncoder
import org.web3j.abi.TypeReference
import org.web3j.abi.datatypes.DynamicArray
import org.web3j.abi.datatypes.DynamicBytes
import org.web3j.abi.datatypes.Function
import org.web3j.abi.datatypes.Type
import org.web3j.abi.datatypes.generated.Bytes32
import org.web3j.abi.datatypes.generated.Uint256
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.util.concurrent.CompletionStage

class ZkEvmBatchSubmitter(
  private val contract: ZkEvmV2AsyncFriendly
) :
  BatchSubmitter {
  private val log: Logger = LogManager.getLogger(this::class.java)

  private fun fromHexStringsToByteArrays(strings: List<String>): List<ByteArray> {
    return strings.map { Bytes.fromHexString(it).toArray() }
  }

  override fun submitBatchCall(batch: Batch): SafeFuture<String?> {
    val function = Function(
      ZkEvmV2.FUNC_FINALIZEBLOCKS,
      listOf<Type<*>>(
        DynamicArray(ZkEvmV2.BlockData::class.java, convertBlocksData(batch.proverResponse.blocksData)),
        DynamicBytes(batch.proverResponse.proof.toArray()),
        Uint256(batch.proverResponse.verifierIndex),
        Bytes32(batch.proverResponse.zkParentStateRootHash.toArray())
      ),
      emptyList<TypeReference<*>>()
    )
    val calldata = FunctionEncoder.encode(function)
    return contract.executeEthCall(calldata)
  }

  @Synchronized
  override fun submitBatch(batch: Batch): SafeFuture<*> {
    log.debug(
      "Submitting batch={} parent_root_hash={} new_root_hash={}",
      batch.intervalString(),
      batch.proverResponse.zkParentStateRootHash.toHexString(),
      batch.proverResponse.blocksData.last().zkRootHash.toHexString()
    )

    val finalizationFuture = contract
      .finalizeBlocks(
        convertBlocksData(batch.proverResponse.blocksData),
        batch.proverResponse.proof.toArray(),
        BigInteger.valueOf(batch.proverResponse.verifierIndex.toLong()),
        batch.proverResponse.zkParentStateRootHash.toArray()
      )
      .sendAsync()
    log.debug("Batch has been submitted to L1: batch={}", batch.intervalString())
    return SafeFuture.of(finalizationFuture)
      .thenApply { transactionReceipt ->
        log.info(
          "Batch has been accepted: batch={} tx_hash={}",
          batch.intervalString(),
          transactionReceipt.transactionHash
        )
        transactionReceipt
      }.whenException { error ->
        handleError(batch, error)
      }
  }

  private fun handleError(batch: Batch, error: Throwable): CompletionStage<Unit> {
    // This should fail with an exception and clearer revert reason. If it doesn't fail, throw what we have
    return submitBatchCall(batch)
      .whenException { errorWithRevertReason ->
        val txHash = extractTransactionHashFromErrorMessage(error.message!!)
        log.error(
          "Batch submission failed: batch={} tx hash: {} error: {}",
          batch.intervalString(),
          txHash,
          errorWithRevertReason.message,
          errorWithRevertReason
        )
      }
      .thenApply<Unit> {
        log.error(
          "Batch submission failed: batch={} error: {}",
          batch.intervalString(),
          error.message,
          error
        )
        throw error
      }.minimalCompletionStage()
  }

  private fun extractTransactionHashFromErrorMessage(message: String): String {
    val regex = Regex("Transaction (.+?) ")
    return regex.findAll(message).first().groupValues[1]
  }

  private fun convertBlocksData(blocksData: List<GetProofResponse.BlockData>): List<ZkEvmV2.BlockData> =
    blocksData.map { blockData: GetProofResponse.BlockData ->
      ZkEvmV2.BlockData(
        blockData.zkRootHash.toArray(),
        BigInteger.valueOf(blockData.timestamp.epochSecond),
        fromHexStringsToByteArrays(blockData.rlpEncodedTransactions),
        fromHexStringsToByteArrays(blockData.l2ToL1MsgHashes),
        blockData.fromAddresses.toArray(),
        blockData.batchReceptionIndices.map { BigInteger.valueOf(it.toLong()) }
      )
    }
}
