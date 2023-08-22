package net.consensys.zkevm.coordinator.app

import kotlinx.datetime.Instant
import net.consensys.linea.contract.ZkEvmV2
import net.consensys.linea.toBigInteger
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.units.bigints.UInt256
import org.web3j.abi.EventEncoder
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameter
import org.web3j.protocol.core.DefaultBlockParameterName
import org.web3j.protocol.core.methods.request.EthFilter
import org.web3j.protocol.core.methods.response.EthLog
import org.web3j.protocol.core.methods.response.Log
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.util.concurrent.CompletableFuture

data class LastConflatedBlock(
  val l2blockNumber: ULong,
  val conflationTimestamp: Instant
)

interface LastConflationProvider {
  fun getLastConflation(): SafeFuture<LastConflatedBlock>
}

/**
 * This class infers when the last conflation happened based on
 * the last finalized block on L1, and getting L1 block time stamp;
 *
 * It's not a very deterministic/accurate approach, but good enough and avoid managing state in a different database.
 */
class L1BasedLastConflationProvider(
  private val lastFinalizedBlockProvider: LastFinalizedBlockProvider,
  private val l1Web3j: Web3j,
  private val l1EarliestBlockNumberToSearchEvents: ULong,
  private val zkEvmSmartContractWeb3jClient: ZkEvmV2,
  private val l2Web3j: Web3j
) : LastConflationProvider {
  private val log: Logger = LogManager.getLogger(this::class.java)
  override fun getLastConflation(): SafeFuture<LastConflatedBlock> {
    return lastFinalizedBlockProvider.getLastFinalizedBlock()
      .thenCompose { lastProvenBlockNumber ->
        if (lastProvenBlockNumber == 0UL) {
          // beginning of time. Use L1 Genesis block timestamp
          l2Web3j.ethGetBlockByNumber(DefaultBlockParameter.valueOf(BigInteger.ZERO), false)
            .sendAsync()
            .thenApply { block ->
              Instant.fromEpochSeconds(block.block.timestamp.toLong())
            }
            .thenApply { blockTimestamp ->
              LastConflatedBlock(
                lastProvenBlockNumber,
                blockTimestamp
              )
            }
        } else {
          getLatestFinalizedBlockEmittedEvent(lastProvenBlockNumber)
            .thenCompose { log ->
              l1Web3j.ethGetBlockByNumber(DefaultBlockParameter.valueOf(log.blockNumber), false)
                .sendAsync()
                .thenApply { block ->
                  Instant.fromEpochSeconds(block.block.timestamp.toLong())
                }
                .thenApply { blockTimestamp ->
                  LastConflatedBlock(
                    lastProvenBlockNumber,
                    blockTimestamp
                  )
                }
            }
        }
      }
  }

  private fun getLatestFinalizedBlockEmittedEvent(
    targetFinalizedL2BlockNumber: ULong
  ): CompletableFuture<Log> {
    val filter = EthFilter(
      DefaultBlockParameter.valueOf(l1EarliestBlockNumberToSearchEvents.toBigInteger()),
      DefaultBlockParameterName.LATEST,
      zkEvmSmartContractWeb3jClient.contractAddress
    )
    filter.addSingleTopic(EventEncoder.encode(ZkEvmV2.BLOCKFINALIZED_EVENT))
    filter.addSingleTopic(UInt256.valueOf(targetFinalizedL2BlockNumber.toLong()).toString())
    return l1Web3j.ethGetLogs(filter)
      .sendAsync()
      .thenCompose { ethLogs: EthLog ->
        if (ethLogs.logs.isEmpty()) {
          val errorMessage = "BlockFinalized event not found for block $targetFinalizedL2BlockNumber, " +
            "between l1 blocks: $l1EarliestBlockNumberToSearchEvents..LATEST"
          log.error(errorMessage)
          SafeFuture.failedFuture(Exception(errorMessage))
        } else {
          val log: Log = (ethLogs.logs.last() as EthLog.LogObject).get()
          SafeFuture.completedFuture(log)
        }
      }
  }
}
