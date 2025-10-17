package linea.ethapi

import linea.domain.Block
import linea.domain.BlockParameter
import linea.domain.BlockWithTxHashes
import linea.domain.FeeHistory
import linea.domain.Transaction
import linea.domain.TransactionForEthCall
import linea.domain.TransactionReceipt
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger

/**
 * Failed requests with JSON-RPC error responses will be rejected with JsonRpcErrorResponseException
 */
interface EthApiClient : EthLogsClient {
  // ==============================================================================
  // Legacy methods (keeping existing)
  // to avoid compilation errors until we aggree on final interface and refactor usage
  fun getChainId(): SafeFuture<ULong> = ethChainId()
  fun blockNumber(): SafeFuture<ULong> = ethBlockNumber()
  fun findBlockByNumber(blockParameter: BlockParameter): SafeFuture<Block?> = ethGetBlockByNumberFullTxs(blockParameter)
  fun getBlockByNumber(blockParameter: BlockParameter): SafeFuture<Block> {
    return findBlockByNumber(blockParameter).thenApply { block ->
      block ?: throw IllegalArgumentException("block=$blockParameter not found!")
    }
  }
  fun getBlockByNumberWithoutTransactionsData(blockParameter: BlockParameter): SafeFuture<BlockWithTxHashes> {
    return ethGetBlockByNumberTxHashes(blockParameter).thenApply { block ->
      block ?: throw IllegalArgumentException("block=$blockParameter not found!")
    }
  }
  // ==============================================================================
  // ==============================================================================

  // Ethereum JSON-RPC API methods with eth prefix
  // Protocol version and chain information
  fun ethChainId(): SafeFuture<ULong>
  fun ethProtocolVersion(): SafeFuture<Int>

  fun ethCoinbase(): SafeFuture<ByteArray>
  fun ethMining(): SafeFuture<Boolean>
  fun ethGasPrice(): SafeFuture<BigInteger>
  fun ethMaxPriorityFeePerGas(): SafeFuture<BigInteger>
  fun ethFeeHistory(
    blockCount: Int,
    newestBlock: BlockParameter,
    rewardPercentiles: List<Double>,
  ): SafeFuture<FeeHistory>

  // Count methods
  fun ethBlockNumber(): SafeFuture<ULong>
  fun ethGetBalance(address: ByteArray, blockParameter: BlockParameter): SafeFuture<BigInteger>
  fun ethGetTransactionCount(address: ByteArray, blockParameter: BlockParameter): SafeFuture<ULong>
  // fun ethGetCode(address: ByteArray, blockParameter: BlockParameter): SafeFuture<ByteArray>
  // fun ethGetStorageAt(address: ByteArray, position: BigInteger, blockParameter: BlockParameter): SafeFuture<ByteArray>
  // fun ethGetBlockTransactionCountByHash(blockHash: ByteArray): SafeFuture<ULong?>
  // fun ethGetBlockTransactionCountByNumber(blockParameter: BlockParameter): SafeFuture<ULong?>

  // Block retrieval methods
  fun ethGetBlockByNumberFullTxs(blockParameter: BlockParameter): SafeFuture<Block?>
  fun ethGetBlockByNumberTxHashes(blockParameter: BlockParameter): SafeFuture<BlockWithTxHashes?>
  // fun ethGetBlockByHashFullTxs(blockHash: ByteArray): SafeFuture<Block?>
  // fun ethGetBlockByHashTxHashes(blockHash: ByteArray): SafeFuture<BlockWithTxHashes?>

  // Transaction methods
  fun ethGetTransactionByHash(transactionHash: ByteArray): SafeFuture<Transaction?>
  fun ethGetTransactionReceipt(transactionHash: ByteArray): SafeFuture<TransactionReceipt?>
  // fun ethGetTransactionByBlockHashAndIndex(blockHash: ByteArray, transactionIndex: ULong): SafeFuture<Transaction?>
  // fun ethGetTransactionByBlockNumberAndIndex(blockParameter: BlockParameter, transactionIndex: ULong): SafeFuture<Transaction?>

  // Transaction submission and estimation
  fun ethSendRawTransaction(signedTransactionData: ByteArray): SafeFuture<ByteArray>
  fun ethCall(
    transaction: TransactionForEthCall,
    blockParameter: BlockParameter = BlockParameter.Tag.LATEST,
    stateOverride: StateOverride? = null,
  ): SafeFuture<ByteArray>

  fun ethEstimateGas(transaction: TransactionForEthCall): SafeFuture<BigInteger>
}

data class StateOverride(
  val balance: BigInteger? = null, // Temporary account balance for the call execution.
  val nonce: ULong? = null, // Temporary nonce value for the call execution.
  val code: ByteArray? = null, // Bytecode to inject into the account.
  // Data, 20 bytes	Address to which the precompile address should be moved.
  val movePrecompileToAddress: ByteArray? = null,
  // key:value pairs (ByteArray hexadecimal encoded) to override all slots in the account storage.
  // You cannot set both the state and stateDiff options simultaneously.
  val state: Map<String, String>? = null,
  // key:value pairs (ByteArray hexadecimal encoded) to override individual slots in the account storage.
  // You cannot set both the state and stateDiff options simultaneously.
  val stateDiff: Map<String, String>? = null,
) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as StateOverride

    if (balance != other.balance) return false
    if (nonce != other.nonce) return false
    if (!code.contentEquals(other.code)) return false
    if (!movePrecompileToAddress.contentEquals(other.movePrecompileToAddress)) return false
    if (state != other.state) return false
    if (stateDiff != other.stateDiff) return false

    return true
  }

  override fun hashCode(): Int {
    var result = balance.hashCode()
    result = 31 * result + nonce.hashCode()
    result = 31 * result + code.contentHashCode()
    result = 31 * result + movePrecompileToAddress.contentHashCode()
    result = 31 * result + state.hashCode()
    result = 31 * result + stateDiff.hashCode()
    return result
  }
}
