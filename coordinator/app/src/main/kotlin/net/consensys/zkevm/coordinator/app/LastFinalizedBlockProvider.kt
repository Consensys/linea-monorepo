package net.consensys.zkevm.coordinator.app

import build.linea.contract.LineaRollupV5
import io.vertx.core.Vertx
import net.consensys.linea.async.AsyncRetryer
import net.consensys.toULong
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.web3j.protocol.core.DefaultBlockParameterName
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.util.concurrent.atomic.AtomicInteger
import java.util.concurrent.atomic.AtomicReference
import kotlin.time.Duration
import kotlin.time.Duration.Companion.seconds

interface LastFinalizedBlockProvider {
  fun getLastFinalizedBlock(): SafeFuture<ULong>
}

/**
 * This class infers when the last conflation happened based on
 * the last block L1 finalized on L1, and getting L1 block time stamp;
 *
 * It's not a very deterministic/accurate approach, but good enough and avoid managing state in a different database.
 */
class L1BasedLastFinalizedBlockProvider(
  private val vertx: Vertx,
  private val lineaRollupSmartContractWeb3jClient: LineaRollupV5,
  private val consistentNumberOfBlocksOnL1: UInt,
  private val numberOfRetries: UInt = Int.MAX_VALUE.toUInt(),
  private val pollingInterval: Duration = 2.seconds
) : LastFinalizedBlockProvider {
  private val log: Logger = LogManager.getLogger(this::class.java)

  override fun getLastFinalizedBlock(): SafeFuture<ULong> {
    lineaRollupSmartContractWeb3jClient.setDefaultBlockParameter(DefaultBlockParameterName.LATEST)

    return AsyncRetryer.retry(
      vertx,
      maxRetries = numberOfRetries.toInt(),
      backoffDelay = pollingInterval
    ) {
      SafeFuture.of(lineaRollupSmartContractWeb3jClient.currentL2BlockNumber().sendAsync())
    }
      .thenCompose { blockNumber ->
        log.info(
          "Rollup lastFinalizedBlockNumber={} waiting {} blocks for confirmation for no updates",
          blockNumber,
          consistentNumberOfBlocksOnL1
        )
        val lastObservedBlock = AtomicReference(blockNumber)
        val numberOfObservations = AtomicInteger(1)
        val isConsistentEnough = { lastPolledBlockNumber: BigInteger ->
          if (lastPolledBlockNumber == lastObservedBlock.get()) {
            numberOfObservations.incrementAndGet().toUInt() >= consistentNumberOfBlocksOnL1
          } else {
            log.info(
              "Rollup finalized block updated from {} to {}, waiting {} blocks for confirmation",
              blockNumber,
              lastPolledBlockNumber,
              consistentNumberOfBlocksOnL1
            )
            numberOfObservations.set(1)
            lastObservedBlock.set(lastPolledBlockNumber)
            false
          }
        }

        AsyncRetryer.retry(
          vertx,
          maxRetries = numberOfRetries.toInt(),
          backoffDelay = pollingInterval,
          stopRetriesPredicate = isConsistentEnough
        ) {
          SafeFuture.of(lineaRollupSmartContractWeb3jClient.currentL2BlockNumber().sendAsync())
        }
      }
      .thenApply { it.toULong() }
  }
}
