package net.consensys.zkevm.coordinator.blockcreation

import io.vertx.core.Vertx
import io.vertx.junit5.VertxExtension
import net.consensys.linea.async.get
import net.consensys.zkevm.ethereum.coordination.blockcreation.BlockCreated
import net.consensys.zkevm.ethereum.coordination.blockcreation.BlockCreationListener
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32
import org.apache.tuweni.units.bigints.UInt256
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.ArgumentMatchers.eq
import org.mockito.Mockito.inOrder
import org.mockito.kotlin.any
import org.mockito.kotlin.atLeast
import org.mockito.kotlin.atMost
import org.mockito.kotlin.doReturn
import org.mockito.kotlin.mock
import org.mockito.kotlin.never
import org.mockito.kotlin.times
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.bytes.Bytes20
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import java.math.BigInteger
import java.util.concurrent.Executors
import java.util.concurrent.TimeUnit
import kotlin.time.Duration
import kotlin.time.Duration.Companion.milliseconds

@ExtendWith(VertxExtension::class)
class BlockCreationMonitorTest {
  private val parentHash =
    Bytes32.fromHexString("0x1000000000000000000000000000000000000000000000000000000000000000")
  private val startingBlockNumberInclusive: Long = 100
  private lateinit var log: Logger
  private lateinit var web3jClient: ExtendedWeb3J
  private lateinit var blockCreationListener: BlockCreationListener
  private var config: BlockCreationMonitor.Config =
    BlockCreationMonitor.Config(
      pollingInterval = 100.milliseconds,
      blocksToFinalization = 2L
    )
  private val executor = Executors.newSingleThreadScheduledExecutor()
  private lateinit var monitor: BlockCreationMonitor

  @BeforeEach
  fun beforeEach(vertx: Vertx) {
    web3jClient = mock()
    blockCreationListener =
      mock { on { acceptBlock(any()) } doReturn SafeFuture.completedFuture(Unit) }

    log = mock()
    monitor =
      BlockCreationMonitor(
        vertx,
        web3jClient,
        startingBlockNumberInclusive,
        parentHash,
        blockCreationListener,
        config
      )
  }

  @AfterEach
  fun afterEach(vertx: Vertx) {
    monitor.stop()
    vertx.close().get()
  }

  // Tried to Test the "Vertx way". Could not make it work, to Will us classi thread sleep :(
  // @Test
  // fun `notifies listener after block is finalized`(testContext: VertxTestContext) {
  //   val payload = executionPayloadV1(blockNumber= startingBlockNumberInclusive)
  //   val headBlockNumber = startingBlockNumberInclusive + config.blocksToFinalization
  //   whenever(web3jClient.ethBlockNumber())
  //     .thenReturn(SafeFuture.completedFuture(BigInteger.valueOf(headBlockNumber)))
  //   whenever(web3jClient.ethGetExecutionPayloadByNumber(any()))
  //     .thenReturn(SafeFuture.completedFuture(payload))
  //   val listenerCalled = testContext.checkpoint(1)
  //   whenever(blockCreationListener.acceptBlock(any())).then {
  //     listenerCalled.flag()
  //     SafeFuture.completedFuture(Unit)
  //   }
  //
  //   monitor.start()
  //     .thenAccept {
  //       testContext.verify {
  //         verify(web3jClient).ethGetExecutionPayloadByNumber(eq(startingBlockNumberInclusive))
  //         verify(blockCreationListener, never()).acceptBlock(BlockCreated(payload))
  //         assertThat(monitor.nexBlockNumberToFetch()).isEqualTo(startingBlockNumberInclusive +
  // 1)
  //         // testContext.completeNow()
  //       }
  //     }
  //     .whenException(testContext::failNow)
  // }

  fun waitTicks(numberOfTicks: Int, function: () -> Unit) {
    Thread.sleep(config.pollingInterval.inWholeMilliseconds * numberOfTicks)
    function()
  }

  @Test
  fun `notifies listener after block is finalized sync`() {
    val payload =
      executionPayloadV1(blockNumber = startingBlockNumberInclusive, parentHash = parentHash)
    val payload2 =
      executionPayloadV1(
        blockNumber = startingBlockNumberInclusive + 1,
        parentHash = payload.blockHash
      )
    val headBlockNumber = startingBlockNumberInclusive + config.blocksToFinalization
    whenever(web3jClient.ethBlockNumber())
      .thenReturn(SafeFuture.completedFuture(BigInteger.valueOf(headBlockNumber)))
      .thenReturn(SafeFuture.completedFuture(BigInteger.valueOf(headBlockNumber + 1)))
    whenever(web3jClient.ethGetExecutionPayloadByNumber(any()))
      .thenReturn(SafeFuture.completedFuture(payload))
      .thenReturn(SafeFuture.completedFuture(payload2))
    whenever(blockCreationListener.acceptBlock(any())).thenReturn(SafeFuture.completedFuture(Unit))

    monitor.start().get()

    waitTicks(3) {
      verify(web3jClient).ethGetExecutionPayloadByNumber(eq(startingBlockNumberInclusive))
      verify(web3jClient).ethGetExecutionPayloadByNumber(eq(startingBlockNumberInclusive + 1))
      verify(blockCreationListener).acceptBlock(BlockCreated(payload))
      verify(blockCreationListener).acceptBlock(BlockCreated(payload2))
      assertThat(monitor.nexBlockNumberToFetch).isEqualTo(startingBlockNumberInclusive + 2)
    }
  }

  @Test
  fun `does not notify listener when block is not safely finalized`() {
    val payload =
      executionPayloadV1(blockNumber = startingBlockNumberInclusive, parentHash = parentHash)
    val headBlockNumber = startingBlockNumberInclusive + config.blocksToFinalization - 1
    whenever(web3jClient.ethBlockNumber())
      .thenReturn(SafeFuture.completedFuture(BigInteger.valueOf(headBlockNumber)))
    whenever(web3jClient.ethGetExecutionPayloadByNumber(any()))
      .thenReturn(SafeFuture.completedFuture(payload))
    whenever(blockCreationListener.acceptBlock(any())).thenReturn(SafeFuture.completedFuture(Unit))

    monitor.start().get()

    waitTicks(2) {
      verify(web3jClient, never()).ethGetExecutionPayloadByNumber(any())
      verify(blockCreationListener, never()).acceptBlock(BlockCreated(payload))
      assertThat(monitor.nexBlockNumberToFetch).isEqualTo(startingBlockNumberInclusive)
    }
  }

  @Test
  fun `when listener throws retries on the next tick and moves on`() {
    val payload =
      executionPayloadV1(blockNumber = startingBlockNumberInclusive, parentHash = parentHash)
    val headBlockNumber = startingBlockNumberInclusive + config.blocksToFinalization + 1
    whenever(web3jClient.ethBlockNumber())
      .thenReturn(SafeFuture.completedFuture(BigInteger.valueOf(headBlockNumber)))
    whenever(web3jClient.ethGetExecutionPayloadByNumber(any()))
      .thenReturn(SafeFuture.completedFuture(payload))

    whenever(blockCreationListener.acceptBlock(any()))
      .thenReturn(SafeFuture.failedFuture<Unit>(Exception("Notification 1 Error")))
      .thenThrow(RuntimeException("Notification 2 Error"))
      .thenReturn(SafeFuture.failedFuture<Unit>(Exception("Notification 3 Error")))
      .thenReturn(SafeFuture.completedFuture(Unit))

    monitor.start().get()

    waitTicks(4) {
      verify(blockCreationListener, atLeast(4)).acceptBlock(BlockCreated(payload))
      assertThat(monitor.nexBlockNumberToFetch).isEqualTo(startingBlockNumberInclusive + 1)
    }
  }

  @Test
  fun `is resilient to connection failures`() {
    val payload =
      executionPayloadV1(blockNumber = startingBlockNumberInclusive, parentHash = parentHash)
    val headBlockNumber = startingBlockNumberInclusive + config.blocksToFinalization
    whenever(web3jClient.ethBlockNumber())
      .thenReturn(SafeFuture.failedFuture(Exception("ethBlockNumber Error 1")))
      .thenReturn(SafeFuture.completedFuture(BigInteger.valueOf(headBlockNumber)))
    whenever(web3jClient.ethGetExecutionPayloadByNumber(any()))
      .thenReturn(SafeFuture.failedFuture(Exception("ethGetExecutionPayloadByNumber Error 1")))
      .thenReturn(SafeFuture.completedFuture(payload))
    whenever(blockCreationListener.acceptBlock(any())).thenReturn(SafeFuture.completedFuture(Unit))

    monitor.start().get()

    waitTicks(6) {
      verify(blockCreationListener, times(1)).acceptBlock(BlockCreated(payload))
      assertThat(monitor.nexBlockNumberToFetch).isEqualTo(startingBlockNumberInclusive + 1)
    }
  }

  @Test
  fun `should stop when reorg is detected above blocksToFinalization limit - manual intervention necessary`() {
    val payload =
      executionPayloadV1(blockNumber = startingBlockNumberInclusive, parentHash = parentHash)
    val payload2 =
      executionPayloadV1(
        blockNumber = startingBlockNumberInclusive + 1,
        parentHash = Bytes32.random()
      )
    val headBlockNumber = startingBlockNumberInclusive + config.blocksToFinalization
    whenever(web3jClient.ethBlockNumber())
      .thenReturn(SafeFuture.completedFuture(BigInteger.valueOf(headBlockNumber)))
      .thenReturn(SafeFuture.completedFuture(BigInteger.valueOf(headBlockNumber + 1)))
    whenever(web3jClient.ethGetExecutionPayloadByNumber(any()))
      .thenReturn(SafeFuture.completedFuture(payload))
      .thenReturn(SafeFuture.completedFuture(payload2))
    whenever(blockCreationListener.acceptBlock(any())).thenReturn(SafeFuture.completedFuture(Unit))

    monitor.start().get()

    waitTicks(5) {
      verify(blockCreationListener, times(1)).acceptBlock(BlockCreated(payload))
      verify(blockCreationListener, never()).acceptBlock(BlockCreated(payload2))
      assertThat(monitor.nexBlockNumberToFetch).isEqualTo(startingBlockNumberInclusive + 1)
    }
  }

  private fun <V> delay(delay: Duration, action: () -> SafeFuture<V>): SafeFuture<V> {
    val future = SafeFuture<V>()
    executor.schedule(
      { action().propagateTo(future) },
      delay.inWholeMilliseconds,
      TimeUnit.MILLISECONDS
    )
    return future
  }

  @Test
  fun `should poll in order when response takes longer that polling interval`() {
    val payload =
      executionPayloadV1(blockNumber = startingBlockNumberInclusive, parentHash = parentHash)
    val payload2 =
      executionPayloadV1(
        blockNumber = startingBlockNumberInclusive + 1,
        parentHash = payload.blockHash
      )
    val headBlockNumber = startingBlockNumberInclusive + config.blocksToFinalization

    whenever(web3jClient.ethBlockNumber())
      .then {
        delay(config.pollingInterval.times(2)) {
          SafeFuture.completedFuture(BigInteger.valueOf(headBlockNumber))
        }
      }
      .thenReturn(SafeFuture.completedFuture(BigInteger.valueOf(headBlockNumber + 1)))
    whenever(web3jClient.ethGetExecutionPayloadByNumber(any()))
      .then { delay(config.pollingInterval.times(2)) { SafeFuture.completedFuture(payload) } }
      .thenReturn(SafeFuture.completedFuture(payload2))
    whenever(blockCreationListener.acceptBlock(any())).thenReturn(SafeFuture.completedFuture(Unit))

    val inOrder = inOrder(blockCreationListener)

    monitor.start().get()

    waitTicks(20) {
      inOrder.verify(blockCreationListener).acceptBlock(BlockCreated(payload))
      inOrder.verify(blockCreationListener).acceptBlock(BlockCreated(payload2))
      assertThat(monitor.nexBlockNumberToFetch).isEqualTo(startingBlockNumberInclusive + 2)
    }
  }

  @Test
  fun `start allow 2nd call when already started`() {
    monitor.start().get()
    monitor.start().get()
  }

  @Test
  fun `stop should be idempotent`() {
    val payload =
      executionPayloadV1(blockNumber = startingBlockNumberInclusive, parentHash = parentHash)
    val payload2 =
      executionPayloadV1(
        blockNumber = startingBlockNumberInclusive + 1,
        parentHash = payload.blockHash
      )
    val headBlockNumber = startingBlockNumberInclusive + config.blocksToFinalization
    whenever(web3jClient.ethBlockNumber())
      .thenReturn(SafeFuture.completedFuture(BigInteger.valueOf(headBlockNumber)))
      .then {
        delay(config.pollingInterval.times(30)) {
          SafeFuture.completedFuture(BigInteger.valueOf(headBlockNumber + 1))
        }
      }
    whenever(web3jClient.ethGetExecutionPayloadByNumber(any()))
      .thenReturn(SafeFuture.completedFuture(payload))
      .then { delay(config.pollingInterval.times(30)) { SafeFuture.completedFuture(payload2) } }
    whenever(blockCreationListener.acceptBlock(any())).thenReturn(SafeFuture.completedFuture(Unit))

    monitor.start().get()
    waitTicks(1) { verify(blockCreationListener, times(1)).acceptBlock(any()) }
    monitor.stop().get()
    monitor.stop().get()

    waitTicks(2) { verify(blockCreationListener, atMost(1)).acceptBlock(any()) }
  }

  private fun executionPayloadV1(
    blockNumber: Long = 0,
    parentHash: Bytes32 = Bytes32.random(),
    feeRecipient: Bytes20 = Bytes20(Bytes.random(20)),
    stateRoot: Bytes32 = Bytes32.random(),
    receiptsRoot: Bytes32 = Bytes32.random(),
    logsBloom: Bytes = Bytes32.random(),
    prevRandao: Bytes32 = Bytes32.random(),
    gasLimit: UInt64 = UInt64.valueOf(0),
    gasUsed: UInt64 = UInt64.valueOf(0),
    timestamp: UInt64 = UInt64.valueOf(0),
    extraData: Bytes = Bytes32.random(),
    baseFeePerGas: UInt256 = UInt256.valueOf(256),
    blockHash: Bytes32 = Bytes32.random(),
    transactions: List<Bytes> = emptyList()
  ): ExecutionPayloadV1 {
    return ExecutionPayloadV1(
      parentHash,
      feeRecipient,
      stateRoot,
      receiptsRoot,
      logsBloom,
      prevRandao,
      UInt64.valueOf(blockNumber),
      gasLimit,
      gasUsed,
      timestamp,
      extraData,
      baseFeePerGas,
      blockHash,
      transactions
    )
  }
}
