package net.consensys.linea

import net.consensys.linea.web3j.ExtendedWeb3JImpl
import org.web3j.protocol.Web3j
import org.web3j.protocol.http.HttpService
import org.web3j.utils.Async
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.infrastructure.async.SafeFuture

class BlockReader {
  private val web3jClient: Web3j = Web3j.build(
    HttpService("https://linea-sepolia.infura.io/v3/"),
    1000,
    Async.defaultExecutorService()
  )

  private val asyncWeb3J = ExtendedWeb3JImpl(web3jClient)

  fun getBlockPayload(blockNumber: Long): SafeFuture<ExecutionPayloadV1> {
    val encodedPayload = asyncWeb3J.ethGetExecutionPayloadByNumber(blockNumber)
    return encodedPayload
  }
}
