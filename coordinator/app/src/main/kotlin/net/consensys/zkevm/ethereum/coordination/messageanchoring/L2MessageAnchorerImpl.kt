package net.consensys.zkevm.ethereum.coordination.messageanchoring

import io.vertx.core.Vertx
import net.consensys.linea.async.retryWithInterval
import net.consensys.linea.contract.L2MessageService
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes32
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.methods.response.EthBlock
import org.web3j.protocol.core.methods.response.TransactionReceipt
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import kotlin.time.Duration

class L2MessageAnchorerImpl(
  private val vertx: Vertx,
  private val l2Web3j: Web3j,
  private val l2Client: L2MessageService,
  private val config: Config
) : L2MessageAnchorer {

  class Config(
    val receiptPollingInterval: Duration,
    val maxReceiptRetries: UInt,
    val blocksToFinalisation: Long
  )

  private val log: Logger = LogManager.getLogger(this::class.java)

  override fun anchorMessages(messageHashes: List<Bytes32>): SafeFuture<TransactionReceipt> {
    log.debug("Anchoring {}, hashes={}", messageHashes.count(), messageHashes.map { it.toHexString() })

    return SafeFuture.of(
      l2Client.addL1L2MessageHashes(messageHashes.map { arr -> arr.toArray() })
        .sendAsync()
        .thenCompose { txReceipt ->
          val safeBlock = txReceipt.blockNumber.add(
            BigInteger.valueOf(config.blocksToFinalisation)
          )
          retryWithInterval(
            config.maxReceiptRetries.toInt(),
            config.receiptPollingInterval,
            vertx,
            transactionIsSafe(
              safeBlock
            )
          ) {
            SafeFuture.of(l2Web3j.ethGetBlockByNumber({ "latest" }, false).sendAsync())
          }.whenException {
            log.warn(it.message, messageHashes)
          }.thenApply {
            txReceipt
          }
        }
    )
  }

  private fun transactionIsSafe(safeBlockNumber: BigInteger) =
    { result: EthBlock -> result.block.number >= safeBlockNumber }
}
