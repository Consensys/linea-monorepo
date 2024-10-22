package net.consensys.zkevm.ethereum.coordination.conflation

import com.github.michaelbull.result.Ok
import io.vertx.core.Vertx
import io.vertx.junit5.VertxExtension
import net.consensys.linea.BlockNumberAndHash
import net.consensys.linea.traces.TracesCountersV1
import net.consensys.zkevm.coordinator.clients.GetTracesCountersResponse
import net.consensys.zkevm.coordinator.clients.TracesCountersClientV1
import net.consensys.zkevm.ethereum.coordination.blockcreation.BlockCreated
import net.consensys.zkevm.toULong
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.assertj.core.api.Assertions
import org.awaitility.Awaitility
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.Mockito.mock
import org.mockito.kotlin.any
import org.mockito.kotlin.argumentCaptor
import org.mockito.kotlin.eq
import org.mockito.kotlin.times
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import tech.pegasys.teku.ethereum.executionclient.schema.randomExecutionPayload
import tech.pegasys.teku.infrastructure.async.SafeFuture
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.toJavaDuration

@ExtendWith(VertxExtension::class)
class BlockToBatchSubmissionCoordinatorTest {
  companion object {
    private val defaultConflationService = ConflationServiceImpl(mock(), mock())
    private const val ARBITRARY_BLOCK_NUMBER = 100L
    private val randomExecutionPayload = randomExecutionPayload(blockNumber = ARBITRARY_BLOCK_NUMBER)
    private val baseBlock = BlockCreated(randomExecutionPayload)
    private val blockRlpEncoded = ByteArray(0)
    private val tracesCounters = TracesCountersV1.EMPTY_TRACES_COUNT
  }

  private fun createBlockToBatchSubmissionCoordinator(
    vertx: Vertx,
    conflationService: ConflationService = defaultConflationService,
    log: Logger = LogManager.getLogger(this::class.java)
  ): BlockToBatchSubmissionCoordinator {
    val tracesCountersClient =
      mock<TracesCountersClientV1>().also {
        whenever(
          it.rollupGetTracesCounters(
            BlockNumberAndHash(
              randomExecutionPayload.blockNumber.toULong(),
              randomExecutionPayload.blockHash.toArray()
            )
          )
        ).thenReturn(
          SafeFuture.completedFuture(Ok(GetTracesCountersResponse(tracesCounters, "")))
        )
      }
    return BlockToBatchSubmissionCoordinator(
      conflationService = conflationService,
      tracesCountersClient = tracesCountersClient,
      vertx = vertx,
      payloadEncoder = { blockRlpEncoded },
      log = log
    )
  }

  @Test
  fun `if conflation service fails, error is logged`(vertx: Vertx) {
    val failingConflationService: ConflationService = mock()
    val expectedException = RuntimeException("Conflation service failed!")
    whenever(failingConflationService.newBlock(any(), any())).thenThrow(expectedException)
    val testLogger: Logger = mock()
    val blockToBatchSubmissionCoordinator = createBlockToBatchSubmissionCoordinator(
      vertx = vertx,
      conflationService = failingConflationService,
      log = testLogger
    )

    val captor = argumentCaptor<Throwable>()
    Assertions.assertThat(blockToBatchSubmissionCoordinator.acceptBlock(baseBlock)).isCompleted
    Awaitility.await().atMost(200.milliseconds.toJavaDuration())
      .untilAsserted {
        verify(testLogger, times(1)).error(
          eq("Failed to conflate block={} errorMessage={}"),
          any(),
          any(),
          captor.capture()
        )
      }

    Assertions.assertThat(captor.allValues).hasSize(1)
    Assertions.assertThat(captor.allValues.first().cause).isEqualTo(expectedException)
  }
}
