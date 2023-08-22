package net.consensys.zkevm.ethereum.coordination.blockcreation

import com.github.michaelbull.result.mapBoth
import kotlinx.datetime.Instant
import kotlinx.datetime.toJavaInstant
import net.consensys.zkevm.coordinator.clients.BlocksStore
import net.consensys.zkevm.ethereum.executionclient.ExecutionEngineClient
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes32
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.ethereum.executionclient.schema.ForkChoiceStateV1
import tech.pegasys.teku.ethereum.executionclient.schema.ForkChoiceUpdatedResult
import tech.pegasys.teku.ethereum.executionclient.schema.PayloadAttributesV1
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.bytes.Bytes20
import tech.pegasys.teku.infrastructure.bytes.Bytes8
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import java.util.concurrent.atomic.AtomicReference
import kotlin.concurrent.timer
import kotlin.time.Duration
import kotlin.time.Duration.Companion.seconds

data class BlockHeightAndHash(val height: UInt64, val hash: Bytes32)

class ExecutionClientBlockCreationCoordinator(
  private val config: Config,
  private val executionEngineClient: ExecutionEngineClient,
  private val blocksStore: BlocksStore<ExecutionPayloadV1>
) : BlockCreationCoordinator {
  data class Config(
    val timeout: Duration = 4.seconds,
    val suggestedFeeRecipient: Bytes20 = Bytes20.ZERO,
    val initialFinalizedBlock: BlockHeightAndHash,
    val initialHeadBlock: BlockHeightAndHash
  )

  private val log: Logger = LogManager.getLogger(this::class.java)

  private val payloadInProcess: AtomicReference<Bytes8> = AtomicReference(null)

  private val finalizedBlock: AtomicReference<BlockHeightAndHash> =
    AtomicReference(config.initialFinalizedBlock)
  private val headBlock: AtomicReference<BlockHeightAndHash> =
    AtomicReference(config.initialHeadBlock)

  private fun triggerBlockCreation(timestamp: Instant, slotNumber: ULong): SafeFuture<Bytes8> {
    val forkChoiceStateV1 =
      ForkChoiceStateV1(headBlock.get().hash, headBlock.get().hash, finalizedBlock.get().hash)
    val payloadAttributes =
      PayloadAttributesV1(
        UInt64.valueOf(timestamp.toJavaInstant().epochSecond),
        Bytes32.ZERO,
        config.suggestedFeeRecipient
      )

    return executionEngineClient
      .forkChoiceUpdatedV1(forkChoiceStateV1, payloadAttributes)
      .thenCompose { result ->
        result.mapBoth(
          { fcuResult: ForkChoiceUpdatedResult ->
            val fcuResult2 = fcuResult.asInternalExecutionPayload()
            if (fcuResult2.payloadId.isPresent) {
              SafeFuture.completedFuture(fcuResult2.payloadId.get())
            } else {
              log.error("Forkchoice update to create payload failed. Result: {}", fcuResult2)
              SafeFuture.failedFuture(
                Exception("Forkchoice update to create payload failed: $fcuResult2")
              )
            }
          },
          { SafeFuture.failedFuture(it.asException()) }
        )
      }
  }

  private fun fetchPayloadAfterBlockTime(payloadId: Bytes8): SafeFuture<ExecutionPayloadV1> {
    val resultFuture = SafeFuture<ExecutionPayloadV1>()
    timer("payload-fetcher", true, config.timeout.inWholeMilliseconds, 100) {
      this.cancel()
      executionEngineClient
        .getPayloadV1(payloadId)
        .thenCompose { result ->
          result.mapBoth(
            { executionPayload -> SafeFuture.completedFuture(executionPayload) },
            {
              log.error("Error fetching payload: {}", it)
              SafeFuture.failedFuture(it.asException("Sequencer error"))
            }
          )
        }
        .propagateTo(resultFuture)
    }

    return resultFuture
  }

  private fun saveBlock(executionPayload: ExecutionPayloadV1): SafeFuture<ExecutionPayloadV1> {
    return blocksStore.saveBlock(executionPayload).thenApply { executionPayload }
  }

  override fun createBlock(timestamp: Instant, slotNumber: ULong): SafeFuture<BlockCreated> {
    log.trace("Triggering block creation: slot={}, timestamp={}", slotNumber, timestamp)
    return triggerBlockCreation(timestamp, slotNumber)
      .thenCompose { payloadId ->
        payloadInProcess.set(payloadId)
        fetchPayloadAfterBlockTime(payloadId)
      }
      .thenCompose(this::saveBlock)
      .thenCompose { executionPayload ->
        executionEngineClient
          .newPayloadV1(executionPayload)
          .thenCompose {
            executionEngineClient.forkChoiceUpdatedV1(
              ForkChoiceStateV1(
                executionPayload.blockHash,
                executionPayload.blockHash,
                finalizedBlock.get().hash
              )
            )
          }
          .thenApply {
            headBlock.set(
              BlockHeightAndHash(executionPayload.blockNumber, executionPayload.blockHash)
            )
            BlockCreated(executionPayload)
          }
      }
  }

  override fun finalizeBlock(block: BlockHeightAndHash): SafeFuture<Unit> {
    return SafeFuture.completedFuture(finalizedBlock.set(block))
  }
}
