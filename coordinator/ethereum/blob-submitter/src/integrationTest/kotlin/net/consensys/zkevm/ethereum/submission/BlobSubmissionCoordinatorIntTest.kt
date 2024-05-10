package net.consensys.zkevm.ethereum.submission

import io.vertx.core.Vertx
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import kotlinx.datetime.Clock
import kotlinx.datetime.toKotlinInstant
import net.consensys.linea.contract.LineaRollupAsyncFriendly
import net.consensys.toULong
import net.consensys.zkevm.domain.BlobRecord
import net.consensys.zkevm.domain.createBlobRecord
import net.consensys.zkevm.domain.defaultGasPriceCaps
import net.consensys.zkevm.ethereum.AccountTransactionManager
import net.consensys.zkevm.ethereum.ContractsManager
import net.consensys.zkevm.ethereum.gaspricing.GasPriceCapProvider
import net.consensys.zkevm.ethereum.settlement.BlobSubmitter
import net.consensys.zkevm.persistence.blob.BlobsRepository
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.InOrder
import org.mockito.Mockito
import org.mockito.Mockito.inOrder
import org.mockito.kotlin.any
import org.mockito.kotlin.doReturn
import org.mockito.kotlin.mock
import org.mockito.kotlin.never
import org.mockito.kotlin.spy
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import org.web3j.protocol.core.DefaultBlockParameterName
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.time.Instant
import kotlin.time.Duration.Companion.seconds

@ExtendWith(VertxExtension::class)
class BlobSubmissionCoordinatorIntTest {
  private val proofSubmissionDelay = 1.seconds
  private val fixedClock =
    mock<Clock> { on { now() } doReturn Instant.ofEpochMilli(System.currentTimeMillis()).toKotlinInstant() }
  private lateinit var rollupOperatorAccount: AccountTransactionManager
  private lateinit var blobsRepositoryMock: BlobsRepository
  private val expectedStartBlockTime = fixedClock.now()
  private val pollingInterval = 1.seconds
  private lateinit var lineaRollupContract: LineaRollupAsyncFriendly
  private lateinit var blobSubmitterAsCallData: BlobSubmitter
  private lateinit var blobSubmitterAsEIP4844: BlobSubmitter
  private lateinit var blobSubmissionCoordinator: BlobSubmissionCoordinatorImpl
  private val gasPriceCapProvider = mock<GasPriceCapProvider> {
    on { this.getGasPriceCaps(any()) } doReturn SafeFuture.completedFuture(defaultGasPriceCaps)
  }
  private val zeroHash: ByteArray = Bytes32.ZERO.toArray()
  private lateinit var inOrder: InOrder
  private val log: Logger = Mockito.spy(LogManager.getLogger(BlobSubmissionCoordinatorImpl::class.java))

  @BeforeEach
  fun beforeEach(vertx: Vertx) {
    val deploymentResult = ContractsManager.get().deployLineaRollup().get()
    lineaRollupContract = spy(deploymentResult.rollupOperatorClient)

    rollupOperatorAccount = deploymentResult.rollupOperator
    blobsRepositoryMock = mock()
    blobSubmitterAsCallData = spy(BlobSubmitterAsCallData(lineaRollupContract))
    blobSubmitterAsEIP4844 = spy(BlobSubmitterAsEIP4844(lineaRollupContract, gasPriceCapProvider))
    blobSubmissionCoordinator = blobSubmissionCoordinator(
      vertx,
      lineaRollupContract,
      Eip4844SwitchAwareBlobSubmitter(blobSubmitterAsCallData, blobSubmitterAsEIP4844)
    )
    inOrder = inOrder(lineaRollupContract, blobSubmitterAsCallData)
  }

  private fun blobSubmissionCoordinator(
    vertx: Vertx,
    lineaRollupContractOverride: LineaRollupAsyncFriendly,
    blobSubmitter: BlobSubmitter
  ): BlobSubmissionCoordinatorImpl {
    return BlobSubmissionCoordinatorImpl(
      BlobSubmissionCoordinatorImpl.Config(
        pollingInterval,
        proofSubmissionDelay,
        maxBlobsToSubmitPerTick = 20U
      ),
      blobSubmitter,
      blobsRepositoryMock,
      lineaRollupContractOverride,
      vertx,
      fixedClock,
      log
    )
  }

  @Test
  fun `when first blob is valid, all blobs are submitted with success`(
    vertx: Vertx,
    testContext: VertxTestContext
  ) {
    val currentL2BlockNumber = lineaRollupContract.currentL2BlockNumber().send().toULong()

    val firstParentStateRootHash =
      lineaRollupContract.stateRootHashes(BigInteger.valueOf(currentL2BlockNumber.toLong()))
        .send()

    val blobsToSubmit = createBlobsToSubmit(
      startingBlockNumber = currentL2BlockNumber + 1UL,
      firstParentStateRootHash = firstParentStateRootHash,
      parentDataHash = zeroHash,
      blobsCount = 20
    )

    whenever(blobsRepositoryMock.getConsecutiveBlobsFromBlockNumber(any(), any())).thenAnswer {
      SafeFuture.completedFuture(blobsToSubmit)
    }.thenAnswer {
      SafeFuture.completedFuture(emptyList<BlobRecord>())
    }

    blobSubmissionCoordinator.start()
      .thenApply {
        vertx.setTimer(pollingInterval.inWholeMilliseconds * 5) {
          testContext.verify {
            val finalCompressedData = blobsToSubmit.last().blobCompressionProof
            lineaRollupContract.setDefaultBlockParameter(DefaultBlockParameterName.LATEST)
            val actualStateRootHash = lineaRollupContract.dataFinalStateRootHashes(
              finalCompressedData!!.dataHash
            ).send()
            assertThat(actualStateRootHash).isEqualTo(finalCompressedData.finalStateRootHash)
            inOrder.verify(lineaRollupContract).resetNonce(any())
            inOrder.verify(blobSubmitterAsCallData).submitBlobCall(blobsToSubmit[0])
            inOrder.verify(blobSubmitterAsCallData).submitBlob(blobsToSubmit[0])
            inOrder.verify(blobSubmitterAsCallData).submitBlob(blobsToSubmit[1])
          }.completeNow()
        }
      }.whenException(testContext::failNow)
  }

  @Test
  fun `when first blob is invalid, no blobs are submitted`(
    vertx: Vertx,
    testContext: VertxTestContext
  ) {
    val currentL2BlockNumber = lineaRollupContract.currentL2BlockNumber().send().toULong()
    val invalidRootHash = Bytes32.random()
    val blobsToSubmit = createBlobsToSubmit(
      startingBlockNumber = currentL2BlockNumber + 1UL,
      firstParentStateRootHash = invalidRootHash.toArray(),
      parentDataHash = zeroHash,
      blobsCount = 2
    )

    whenever(blobsRepositoryMock.getConsecutiveBlobsFromBlockNumber(any(), any())).thenAnswer {
      SafeFuture.completedFuture(blobsToSubmit)
    }.thenAnswer {
      SafeFuture.completedFuture(emptyList<BlobRecord>())
    }

    blobSubmissionCoordinator.start()
      .thenApply {
        vertx.setTimer(pollingInterval.inWholeMilliseconds * 3) {
          testContext.verify {
            verify(blobSubmitterAsCallData).submitBlobCall(blobsToSubmit[0])
            verify(blobSubmitterAsCallData, never()).submitBlob(any())
          }.completeNow()
        }
      }.whenException(testContext::failNow)
  }

  @Test
  fun `when blob submission in the middle fails, prevents subsequent submissions`(
    vertx: Vertx,
    testContext: VertxTestContext
  ) {
    val currentL2BlockNumber = lineaRollupContract.currentL2BlockNumber().send().toULong()
    val firstParentStateRootHash =
      lineaRollupContract.stateRootHashes(BigInteger.valueOf(currentL2BlockNumber.toLong()))
        .send()
    val blobsToSubmit = createBlobsToSubmit(
      startingBlockNumber = currentL2BlockNumber + 1UL,
      firstParentStateRootHash = firstParentStateRootHash,
      parentDataHash = zeroHash,
      blobsCount = 3
    )

    class FakeBlobSubmitter(val delegate: BlobSubmitter) : BlobSubmitter {
      override fun submitBlob(blobRecord: BlobRecord): SafeFuture<String> {
        return if (blobRecord.startBlockNumber == blobsToSubmit[1].startBlockNumber) {
          SafeFuture.failedFuture(RuntimeException("Blob Submission Failed"))
        } else {
          delegate.submitBlob(blobRecord)
        }
      }

      override fun submitBlobCall(blobRecord: BlobRecord): SafeFuture<*> = delegate.submitBlobCall(blobRecord)
    }

    whenever(blobsRepositoryMock.getConsecutiveBlobsFromBlockNumber(any(), any())).thenAnswer {
      SafeFuture.completedFuture(blobsToSubmit)
    }.thenAnswer {
      SafeFuture.completedFuture(emptyList<BlobRecord>())
    }

    whenever(blobsRepositoryMock.getConsecutiveBlobsFromBlockNumber(any(), any())).thenAnswer {
      SafeFuture.completedFuture(blobsToSubmit)
    }.thenAnswer {
      SafeFuture.completedFuture(emptyList<BlobRecord>())
    }

    val blobSubmissionCoordinator = blobSubmissionCoordinator(
      vertx,
      lineaRollupContract,
      Eip4844SwitchAwareBlobSubmitter(
        FakeBlobSubmitter(blobSubmitterAsCallData),
        FakeBlobSubmitter(blobSubmitterAsEIP4844)
      )
    )

    blobSubmissionCoordinator.start()
      .thenApply {
        vertx.setTimer(pollingInterval.inWholeMilliseconds * 3) {
          testContext.verify {
            verify(blobSubmitterAsCallData).submitBlob(blobsToSubmit[0])
            verify(blobSubmitterAsCallData, never()).submitBlob(blobsToSubmit[2])
          }.completeNow()
        }
      }.whenException(testContext::failNow)
  }

  @Test
  fun `when first transaction fails because of gas limit, none of the transactions should go out`(
    vertx: Vertx,
    testContext: VertxTestContext
  ) {
    lineaRollupContract = ContractsManager.get().connectToLineaRollupContract(
      contractAddress = lineaRollupContract.contractAddress,
      transactionManager = rollupOperatorAccount.txManager,
      gasLimit = 30000UL
    )
    blobSubmitterAsCallData = spy(BlobSubmitterAsCallData(lineaRollupContract))
    blobSubmitterAsEIP4844 = spy(BlobSubmitterAsEIP4844(lineaRollupContract, gasPriceCapProvider))

    val currentL2BlockNumber = lineaRollupContract.currentL2BlockNumber().send().toULong()
    val firstParentStateRootHash = lineaRollupContract
      .stateRootHashes(BigInteger.valueOf(currentL2BlockNumber.toLong())).send()
    val blobsToSubmit = createBlobsToSubmit(
      startingBlockNumber = currentL2BlockNumber + 1UL,
      firstParentStateRootHash = firstParentStateRootHash,
      parentDataHash = zeroHash,
      blobsCount = 3
    )

    whenever(blobsRepositoryMock.getConsecutiveBlobsFromBlockNumber(any(), any())).thenAnswer {
      SafeFuture.completedFuture(blobsToSubmit)
    }.thenAnswer {
      SafeFuture.completedFuture(emptyList<BlobRecord>())
    }

    val blobSubmissionCoordinator = blobSubmissionCoordinator(
      vertx,
      lineaRollupContract,
      Eip4844SwitchAwareBlobSubmitter(blobSubmitterAsCallData, blobSubmitterAsEIP4844)
    )

    blobSubmissionCoordinator.start()
      .thenApply {
        vertx.setTimer(pollingInterval.inWholeMilliseconds * 3) {
          testContext.verify {
            verify(blobSubmitterAsCallData).submitBlobCall(blobsToSubmit[0])
            verify(blobSubmitterAsCallData, never()).submitBlob(blobsToSubmit[0])
            verify(blobSubmitterAsCallData, never()).submitBlob(blobsToSubmit[1])
            verify(blobSubmitterAsCallData, never()).submitBlob(blobsToSubmit[2])
          }.completeNow()
        }
      }.whenException(testContext::failNow)
  }

  private fun createBlobsToSubmit(
    startingBlockNumber: ULong,
    firstParentStateRootHash: ByteArray,
    parentDataHash: ByteArray,
    blobsCount: Int,
    eip4844EnabledBlobs: Set<Int> = setOf() // Blob numbers from 1 to blobsCount with eip4844 enabled
  ): List<BlobRecord> {
    require(eip4844EnabledBlobs.isEmpty() || eip4844EnabledBlobs.max() <= blobsCount)
    var blobNumber = 1
    val firstBlobToSubmit = createBlobRecord(
      startBlockNumber = startingBlockNumber,
      endBlockNumber = startingBlockNumber,
      parentStateRootHash = firstParentStateRootHash,
      parentDataHash = parentDataHash,
      eip4844Enabled = blobNumber in eip4844EnabledBlobs,
      startBlockTime = expectedStartBlockTime
    )
    return (1 until blobsCount).runningFold(firstBlobToSubmit) { previousBlob, index ->
      blobNumber += 1

      createBlobRecord(
        startBlockNumber = startingBlockNumber + index.toULong(),
        endBlockNumber = startingBlockNumber + index.toULong(),
        parentBlobRecord = previousBlob,
        eip4844Enabled = blobNumber in eip4844EnabledBlobs,
        startBlockTime = expectedStartBlockTime
      )
    }
  }
}
