package net.consensys.linea.web3j

import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameter
import org.web3j.protocol.core.Response
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger

/**
 * Web3J is quite complex to Mock and Unit test, this is just a tinny interface to make
 * BlockCreationMonitor more testable
 */
interface ExtendedWeb3J {
  val web3jClient: Web3j
  fun ethBlockNumber(): SafeFuture<BigInteger>
  fun ethGetExecutionPayloadByNumber(blockNumber: Long): SafeFuture<ExecutionPayloadV1>
  fun ethGetBlockTimestampByNumber(blockNumber: Long): SafeFuture<BigInteger>
}

class ExtendedWeb3JImpl(override val web3jClient: Web3j) : ExtendedWeb3J {

  private fun buildException(error: Response.Error): Exception =
    Exception("${error.code}: ${error.message}")

  override fun ethBlockNumber(): SafeFuture<BigInteger> {
    return SafeFuture.of(web3jClient.ethBlockNumber().sendAsync()).thenCompose { response ->
      if (response.hasError()) {
        SafeFuture.failedFuture(buildException(response.error))
      } else {
        SafeFuture.completedFuture(response.blockNumber)
      }
    }
  }

  override fun ethGetExecutionPayloadByNumber(blockNumber: Long): SafeFuture<ExecutionPayloadV1> {
    return SafeFuture.of(
      web3jClient
        .ethGetBlockByNumber(
          DefaultBlockParameter.valueOf(BigInteger.valueOf(blockNumber)),
          true
        )
        .sendAsync()
    )
      .thenCompose { response ->
        if (response.hasError()) {
          SafeFuture.failedFuture(buildException(response.error))
        } else {
          response.block?.let {
            SafeFuture.completedFuture(response.block.toExecutionPayloadV1())
          } ?: SafeFuture.failedFuture(Exception("Block $blockNumber not found!"))
        }
      }
  }

  override fun ethGetBlockTimestampByNumber(
    blockNumber: Long
  ): SafeFuture<BigInteger> {
    return SafeFuture.of(
      web3jClient
        .ethGetBlockByNumber(
          DefaultBlockParameter.valueOf(BigInteger.valueOf(blockNumber)),
          false
        )
        .sendAsync()
    )
      .thenCompose { response ->
        if (response.hasError()) {
          SafeFuture.failedFuture(buildException(response.error))
        } else {
          response.block?.let {
            SafeFuture.completedFuture(response.block.timestamp)
          } ?: SafeFuture.failedFuture(Exception("Block $blockNumber not found!"))
        }
      }
  }
}
