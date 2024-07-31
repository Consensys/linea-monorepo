package net.consensys.zkevm.ethereum.coordination.messageanchoring

import io.vertx.core.Vertx
import net.consensys.linea.async.toSafeFuture
import net.consensys.linea.contract.L2MessageService
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes32
import org.web3j.abi.EventEncoder
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameter
import org.web3j.protocol.core.methods.response.Log
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.util.concurrent.Callable

class L2QuerierImpl(
  private val l2Client: Web3j,
  private val messageService: L2MessageService,
  private val config: Config,
  private val vertx: Vertx
) : L2Querier {
  private val log: Logger = LogManager.getLogger(this::class.java)

  data class Config(
    val blocksToFinalizationL2: UInt,
    val lastHashSearchWindow: UInt,
    val contractAddressToListen: String
  )

  private fun finalizedBlockNumber(): SafeFuture<BigInteger> {
    return SafeFuture.of(
      l2Client.ethBlockNumber().sendAsync().thenApply {
        it.blockNumber.minus(BigInteger.valueOf(config.blocksToFinalizationL2.toLong()))
      }
    )
  }

  override fun findLastFinalizedAnchoredEvent(): SafeFuture<MessageHashAnchoredEvent?> {
    return vertx.executeBlocking(
      Callable {
        var finalizedBlockNumber = l2Client.ethBlockNumber().send().blockNumber
        if (finalizedBlockNumber > BigInteger.valueOf(config.blocksToFinalizationL2.toLong())) {
          finalizedBlockNumber = finalizedBlockNumber.minus(BigInteger.valueOf(config.blocksToFinalizationL2.toLong()))
        }

        val bigIntSearchWindow = BigInteger.valueOf(config.lastHashSearchWindow.toLong())
        var startingBlock: BigInteger = BigInteger.valueOf(0)
        if (finalizedBlockNumber > bigIntSearchWindow) {
          startingBlock = finalizedBlockNumber.minus(bigIntSearchWindow)
        }
        var endingBlock = finalizedBlockNumber

        log.debug(
          "Searching for event={} startingBlock={} endingBlock={}",
          L2MessageService.L1L2MESSAGEHASHESADDEDTOINBOX_EVENT.name,
          startingBlock,
          endingBlock
        )

        messageService.setDefaultBlockParameter(DefaultBlockParameter.valueOf(finalizedBlockNumber))
        var event: MessageHashAnchoredEvent? = null
        while (startingBlock >= BigInteger.ZERO) {
          val messageHashFilter =
            org.web3j.protocol.core.methods.request.EthFilter(
              DefaultBlockParameter.valueOf(startingBlock),
              DefaultBlockParameter.valueOf(endingBlock),
              messageService.contractAddress
            )

          messageHashFilter.addSingleTopic(EventEncoder.encode(L2MessageService.L1L2MESSAGEHASHESADDEDTOINBOX_EVENT))

          val logs = l2Client.ethGetLogs(messageHashFilter).send().logs
          if (logs.isNotEmpty()) {
            val lastLog = logs.last().get() as Log
            val messageHash =
              L2MessageService.getL1L2MessageHashesAddedToInboxEventFromLog(lastLog).messageHashes.last()
            log.debug("Returning found hash={}", Bytes32.wrap(messageHash))
            event = MessageHashAnchoredEvent(Bytes32.wrap(messageHash))
            break
          } else {
            endingBlock = startingBlock
            startingBlock = startingBlock.minus(bigIntSearchWindow)
          }
        }
        event
      },
      true
    )
      .toSafeFuture()
  }

  override fun getMessageHashStatus(messageHash: Bytes32): SafeFuture<BigInteger> {
    return SafeFuture.of(
      finalizedBlockNumber().thenApply {
        messageService.setDefaultBlockParameter(DefaultBlockParameter.valueOf(it))
      }.thenCompose {
        messageService.inboxL1L2MessageStatus(messageHash.toArray()).sendAsync()
      }
    )
  }
}
