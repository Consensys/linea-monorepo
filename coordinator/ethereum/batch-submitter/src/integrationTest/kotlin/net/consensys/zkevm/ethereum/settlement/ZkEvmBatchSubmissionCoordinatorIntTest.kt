package net.consensys.zkevm.ethereum.settlement

import io.vertx.core.Vertx
import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import kotlinx.datetime.Clock
import kotlinx.datetime.toKotlinInstant
import net.consensys.linea.contract.AsyncFriendlyTransactionManager
import net.consensys.linea.contract.ZkEvmV2AsyncFriendly
import net.consensys.zkevm.coordinator.clients.GetProofResponse
import net.consensys.zkevm.ethereum.coordination.conflation.Batch
import net.consensys.zkevm.ethereum.settlement.persistence.BatchesRepository
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.TestInstance
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.kotlin.any
import org.mockito.kotlin.doReturn
import org.mockito.kotlin.mock
import org.mockito.kotlin.whenever
import org.web3j.crypto.Credentials
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameter
import org.web3j.protocol.http.HttpService
import org.web3j.tx.gas.StaticEIP1559GasProvider
import org.web3j.tx.response.PollingTransactionReceiptProcessor
import org.web3j.utils.Async
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import java.math.BigInteger
import java.time.Instant
import java.util.*
import java.util.concurrent.TimeUnit
import kotlin.time.Duration.Companion.seconds

@ExtendWith(VertxExtension::class)
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class ZkEvmBatchSubmissionCoordinatorIntTest {
  private val proofSubmissionDelay = 1.seconds
  private val defaultGasLimit = BigInteger.valueOf(10000000)
  private val gwei = BigInteger.valueOf(1000000000L)
  private val maxFeePerGas = gwei.multiply(BigInteger.valueOf(5L))
  private val currentTimestamp = System.currentTimeMillis()
  private val fixedClock =
    mock<Clock> { on { now() } doReturn Instant.ofEpochMilli(currentTimestamp).toKotlinInstant() }
  private val zkProofVerifierVersion = 1

  // WARNING: FOR LOCAL DEV ONLY - DO NOT REUSE THESE KEYS ELSEWHERE
  private val privateKey = "202454d1b4e72c41ebf58150030f649648d3cf5590297fb6718e27039ed9c86d"
  private val credentiials = Credentials.create(privateKey)
  private val l1Client =
    Web3j.build(HttpService("http://localhost:8445"), 1000, Async.defaultExecutorService())
  private val pollingTransactionReceiptProcessor = PollingTransactionReceiptProcessor(l1Client, 1000, 40)
  private val asyncFriendlyTransactionManager = AsyncFriendlyTransactionManager(
    l1Client,
    credentiials,
    pollingTransactionReceiptProcessor
  )
  private val contractAddress = System.getProperty("ContractAddress")
  private lateinit var zkEvmV2Contract: ZkEvmV2AsyncFriendly
  private lateinit var firstRootHash: Bytes32
  private val validTransactionRlp =
    this::class.java.getResource("/valid-transaction.rlp")!!.readText().trim()

  @BeforeAll
  fun beforeAll() {
    zkEvmV2Contract = connectToZkevmContract(defaultGasLimit)
  }

  @BeforeEach
  fun beforeEach() {
    firstRootHash =
      Bytes32.wrap(zkEvmV2Contract.stateRootHashes(zkEvmV2Contract.currentL2BlockNumber().send()).send())
  }

  @Test
  @Timeout(2, timeUnit = TimeUnit.MINUTES)
  fun `batch submission coordinator submits big number of batches with increasing nonces without reverts`(
    vertx: Vertx,
    testContext: VertxTestContext
  ) {
    val verifierIndex = 0L
    val batchesRepositoryMock = mock<BatchesRepository>()
    val batchSubmitter = ZkEvmBatchSubmitter(zkEvmV2Contract)
    val pollingInterval = 24.seconds

    val batchSubmissionCoordinator = ZkEvmBatchSubmissionCoordinator(
      ZkEvmBatchSubmissionCoordinator.Config(pollingInterval, proofSubmissionDelay),
      batchSubmitter,
      batchesRepositoryMock,
      zkEvmV2Contract,
      vertx,
      fixedClock
    )

    val firstBatchToSubmit = createBatchToSubmit(verifierIndex, firstRootHash, UInt64.ONE, UInt64.ONE)
    val batchesToSubmit = (2..200).runningFold(firstBatchToSubmit) { previousBatch, index ->
      createBatchToSubmit(
        verifierIndex,
        previousBatch.proverResponse.blocksData.last().zkRootHash,
        UInt64.valueOf(index.toLong()),
        UInt64.valueOf(index.toLong())
      )
    }

    whenever(batchesRepositoryMock.getConsecutiveBatchesFromBlockNumber(any(), any())).thenAnswer {
      SafeFuture.completedFuture(
        batchesToSubmit
      )
    }.thenAnswer {
      SafeFuture.completedFuture(
        emptyList<Batch>()
      )
    }

    val transactionCountBeforeSubmissions = getCurrentTransactionCount()

    batchSubmissionCoordinator.start()
      .thenApply {
        vertx.setTimer(pollingInterval.inWholeMilliseconds) {
          testContext.verify {
            val currentL2BlockNumber = zkEvmV2Contract.currentL2BlockNumber().send()
            val actualStateRootHash = Bytes32.wrap(zkEvmV2Contract.stateRootHashes(currentL2BlockNumber).send())
            assertThat(actualStateRootHash).isEqualTo(
              batchesToSubmit.last().proverResponse.blocksData.last().zkRootHash
            )
            val transactionsSinceStart = getCurrentTransactionCount() - transactionCountBeforeSubmissions
            assertThat(transactionsSinceStart).isEqualTo(batchesToSubmit.size)
          }.completeNow()
        }
      }.whenException(testContext::failNow)
  }

  @Test
  @Timeout(30, timeUnit = TimeUnit.SECONDS)
  fun `first invalid batch prevents whole series submission`(
    vertx: Vertx,
    testContext: VertxTestContext
  ) {
    val verifierIndex = 0L
    val batchesRepositoryMock = mock<BatchesRepository>()
    val batchSubmitter = ZkEvmBatchSubmitter(zkEvmV2Contract)
    val pollingInterval = 4.seconds

    val batchSubmissionCoordinator = ZkEvmBatchSubmissionCoordinator(
      ZkEvmBatchSubmissionCoordinator.Config(pollingInterval, proofSubmissionDelay),
      batchSubmitter,
      batchesRepositoryMock,
      zkEvmV2Contract,
      vertx,
      fixedClock
    )

    val invalidRootHash = Bytes32.random()
    val firstBatchToSubmit = createBatchToSubmit(verifierIndex, invalidRootHash, UInt64.ONE, UInt64.ONE)
    val batchesToSubmit = (2..3).runningFold(firstBatchToSubmit) { previousBatch, index ->
      createBatchToSubmit(
        verifierIndex,
        previousBatch.proverResponse.blocksData.last().zkRootHash,
        UInt64.valueOf(index.toLong()),
        UInt64.valueOf(index.toLong())
      )
    }

    whenever(batchesRepositoryMock.getConsecutiveBatchesFromBlockNumber(any(), any())).thenAnswer {
      SafeFuture.completedFuture(
        batchesToSubmit
      )
    }.thenAnswer {
      SafeFuture.completedFuture(
        emptyList<Batch>()
      )
    }

    val transactionCountBeforeSubmissions = getCurrentTransactionCount()

    batchSubmissionCoordinator.start()
      .thenApply {
        vertx.setTimer(pollingInterval.inWholeMilliseconds * 3) {
          testContext.verify {
            val currentL2BlockNumber = zkEvmV2Contract.currentL2BlockNumber().send()
            val actualStateRootHash = Bytes32.wrap(zkEvmV2Contract.stateRootHashes(currentL2BlockNumber).send())
            assertThat(actualStateRootHash).isEqualTo(
              firstRootHash
            )
            assertThat(getCurrentTransactionCount()).isEqualTo(transactionCountBeforeSubmissions)
          }.completeNow()
        }
      }.whenException(testContext::failNow)
  }

  @Test
  @Timeout(30, timeUnit = TimeUnit.SECONDS)
  fun `if first transaction fails because of gas limit, none of the transactions should go out`(
    vertx: Vertx,
    testContext: VertxTestContext
  ) {
    val zkEvmV2Contract = connectToZkevmContract(BigInteger.valueOf(30000))
    val verifierIndex = 0L
    val batchesRepositoryMock = mock<BatchesRepository>()
    val batchSubmitter = ZkEvmBatchSubmitter(zkEvmV2Contract)
    val pollingInterval = 4.seconds

    val batchSubmissionCoordinator = ZkEvmBatchSubmissionCoordinator(
      ZkEvmBatchSubmissionCoordinator.Config(pollingInterval, proofSubmissionDelay),
      batchSubmitter,
      batchesRepositoryMock,
      zkEvmV2Contract,
      vertx,
      fixedClock
    )

    val firstBatchToSubmit = createBatchToSubmit(verifierIndex, firstRootHash, UInt64.ONE, UInt64.ONE)
    val batchesToSubmit = (2..3).runningFold(firstBatchToSubmit) { previousBatch, index ->
      createBatchToSubmit(
        verifierIndex,
        previousBatch.proverResponse.blocksData.last().zkRootHash,
        UInt64.valueOf(index.toLong()),
        UInt64.valueOf(index.toLong())
      )
    }

    whenever(batchesRepositoryMock.getConsecutiveBatchesFromBlockNumber(any(), any())).thenAnswer {
      SafeFuture.completedFuture(
        batchesToSubmit
      )
    }.thenAnswer {
      SafeFuture.completedFuture(
        emptyList<Batch>()
      )
    }

    val transactionCountBeforeSubmissions = getCurrentTransactionCount()

    batchSubmissionCoordinator.start()
      .thenApply {
        vertx.setTimer(pollingInterval.inWholeMilliseconds * 3) {
          testContext.verify {
            val currentL2BlockNumber = zkEvmV2Contract.currentL2BlockNumber().send()
            val actualStateRootHash = Bytes32.wrap(zkEvmV2Contract.stateRootHashes(currentL2BlockNumber).send())
            assertThat(actualStateRootHash).isEqualTo(
              firstRootHash
            )
            assertThat(getCurrentTransactionCount()).isEqualTo(transactionCountBeforeSubmissions)
          }.completeNow()
        }
      }.whenException(testContext::failNow)
  }

  private fun getCurrentTransactionCount(): BigInteger {
    return l1Client.ethGetTransactionCount(credentiials.address, DefaultBlockParameter.valueOf("latest"))
      .send().transactionCount
  }

  private fun connectToZkevmContract(gasLimit: BigInteger): ZkEvmV2AsyncFriendly {
    val gasProvider = StaticEIP1559GasProvider(
      l1Client.ethChainId().send().chainId.toLong(),
      maxFeePerGas,
      maxFeePerGas.minus(BigInteger.valueOf(100)),
      gasLimit
    )
    return ZkEvmV2AsyncFriendly.load(contractAddress, l1Client, asyncFriendlyTransactionManager, gasProvider)
  }

  private fun createBatchToSubmit(
    verifierIndex: Long,
    parentRootHash: Bytes32,
    startNumber: UInt64,
    endNumber: UInt64
  ): Batch {
    val proof = Bytes.random(100)
    val zkStateRootHash = Bytes32.random()
    val batchReceptionIndices = emptyList<UShort>()
    val l2ToL1MsgHashes = emptyList<String>()
    val fromAddresses = Bytes.EMPTY
    val blocks =
      listOf(
        GetProofResponse.BlockData(
          zkStateRootHash,
          Instant.ofEpochMilli(currentTimestamp),
          listOf(validTransactionRlp, validTransactionRlp),
          batchReceptionIndices,
          l2ToL1MsgHashes,
          fromAddresses
        )
      )
    val proverResponse =
      GetProofResponse(
        proof,
        verifierIndex,
        parentRootHash,
        blocks,
        zkProofVerifierVersion.toString()
      )
    return Batch(startNumber, endNumber, proverResponse)
  }
}
