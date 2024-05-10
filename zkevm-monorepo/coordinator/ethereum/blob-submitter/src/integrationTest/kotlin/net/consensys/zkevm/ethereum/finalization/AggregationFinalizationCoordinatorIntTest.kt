package net.consensys.zkevm.ethereum.finalization

import io.vertx.core.Vertx
import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.decodeHex
import net.consensys.linea.contract.LineaRollupAsyncFriendly
import net.consensys.toBigInteger
import net.consensys.toULong
import net.consensys.zkevm.domain.ProofToFinalize
import net.consensys.zkevm.domain.createBlobRecord
import net.consensys.zkevm.domain.createProofToFinalizeFromBlobs
import net.consensys.zkevm.domain.defaultGasPriceCaps
import net.consensys.zkevm.ethereum.ContractsManager
import net.consensys.zkevm.ethereum.gaspricing.GasPriceCapProvider
import net.consensys.zkevm.ethereum.submission.BlobSubmitterAsCallData
import net.consensys.zkevm.persistence.aggregation.AggregationsRepository
import org.assertj.core.api.Assertions.assertThat
import org.awaitility.Awaitility.await
import org.awaitility.Awaitility.waitAtMost
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.kotlin.any
import org.mockito.kotlin.doReturn
import org.mockito.kotlin.mock
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.util.concurrent.TimeUnit
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.minutes
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration

@ExtendWith(VertxExtension::class)
class AggregationFinalizationCoordinatorIntTest {
  private val proofSubmissionDelay = 10.milliseconds
  private val maxAggregationsToFinalizePerIteration = 1U
  private val currentTimestamp = System.currentTimeMillis()
  private val fixedClock = mock<Clock> {
    on { now() } doReturn Instant.fromEpochMilliseconds(currentTimestamp)
  }
  private val expectedStartBlockTime = fixedClock.now()
  private var currentL2BlockNumber = 0UL
  private var startingParentStateRootHash: ByteArray =
    "0x113e9977cebf08f3b271d121342540fce95530c206c2a878ae957925e1d0fc02".decodeHex()
  private lateinit var lineaRollupContract: LineaRollupAsyncFriendly
  private val emptyHash: ByteArray = ByteArray(32)
  private var l1BlockNumber: ULong = 0UL

  private lateinit var blobSubmitter: BlobSubmitterAsCallData
  private val aggregationsRepository = mock<AggregationsRepository>()
  private lateinit var aggregationFinalization: AggregationFinalizationAsCallData
  private val pollingInterval = 2.seconds
  private lateinit var firstParentStateRootHash: ByteArray
  private val gasPriceCapProvider = mock<GasPriceCapProvider> {
    on { this.getGasPriceCaps(any()) } doReturn SafeFuture.completedFuture(defaultGasPriceCaps)
  }

  @BeforeEach
  fun beforeEach() {
    val rollupDeploymentResult = ContractsManager.get().deployLineaRollup().get()
    l1BlockNumber = rollupDeploymentResult.contractDeploymentBlockNumber
    lineaRollupContract = rollupDeploymentResult.rollupOperatorClient
    blobSubmitter = BlobSubmitterAsCallData(lineaRollupContract)
    aggregationFinalization = AggregationFinalizationAsCallData(lineaRollupContract, gasPriceCapProvider)

    currentL2BlockNumber = lineaRollupContract.currentL2BlockNumber().send().toULong()
    startingParentStateRootHash =
      lineaRollupContract.stateRootHashes(currentL2BlockNumber.toBigInteger()).send()
    firstParentStateRootHash =
      lineaRollupContract.stateRootHashes(BigInteger.valueOf(currentL2BlockNumber.toLong())).send()
  }

  @Test
  @Timeout(1, timeUnit = TimeUnit.MINUTES)
  fun `aggregation finalization coordinator submits multiple finalizations with increasing nonces`(
    vertx: Vertx,
    testContext: VertxTestContext
  ) {
    val blob1 = createBlobRecord(
      startBlockNumber = currentL2BlockNumber + 1UL,
      endBlockNumber = currentL2BlockNumber + 9UL,
      parentStateRootHash = firstParentStateRootHash,
      parentDataHash = emptyHash,
      startBlockTime = expectedStartBlockTime
    )
    val blob2 = createBlobRecord(
      startBlockNumber = currentL2BlockNumber + 10UL,
      endBlockNumber = currentL2BlockNumber + 19UL,
      startBlockTime = expectedStartBlockTime,
      parentBlobRecord = blob1
    )
    val blob3 = createBlobRecord(
      startBlockNumber = currentL2BlockNumber + 20UL,
      endBlockNumber = currentL2BlockNumber + 29UL,
      startBlockTime = expectedStartBlockTime,
      parentBlobRecord = blob2
    )
    val blob4 = createBlobRecord(
      startBlockNumber = currentL2BlockNumber + 30UL,
      endBlockNumber = currentL2BlockNumber + 39UL,
      startBlockTime = expectedStartBlockTime,
      parentBlobRecord = blob3
    )
    val blob5 = createBlobRecord(
      startBlockNumber = currentL2BlockNumber + 40UL,
      endBlockNumber = currentL2BlockNumber + 49UL,
      startBlockTime = expectedStartBlockTime,
      parentBlobRecord = blob4
    )
    val blob6 = createBlobRecord(
      startBlockNumber = currentL2BlockNumber + 50UL,
      endBlockNumber = currentL2BlockNumber + 59UL,
      startBlockTime = expectedStartBlockTime,
      parentBlobRecord = blob5
    )

    val blobs = listOf(blob1, blob2, blob3, blob4, blob5, blob6)
    var lastFinalizedBlockTime = Instant.fromEpochSeconds(lineaRollupContract.currentTimestamp().send().toLong())
    val aggregation1 = createProofToFinalizeFromBlobs(
      blobRecords = blobs.slice(0..2),
      lastFinalizedBlockTime = lastFinalizedBlockTime,
      parentStateRootHash = startingParentStateRootHash,
      dataParentHash = emptyHash
    )

    lastFinalizedBlockTime = lastFinalizedBlockTime.plus(
      (blobs[2].endBlockNumber.toLong() * 1).seconds
    )
    val aggregation2 = createProofToFinalizeFromBlobs(
      blobRecords = blobs.slice(3..4),
      lastFinalizedBlockTime = lastFinalizedBlockTime,
      parentRecord = blobs[2],
      aggregation1
    )
    val aggregation3 = createProofToFinalizeFromBlobs(
      blobRecords = blobs.slice(5..5),
      lastFinalizedBlockTime = lastFinalizedBlockTime,
      parentRecord = blobs[4],
      aggregation2
    )
    val aggregations = listOf(aggregation1, aggregation2, aggregation3)
    mockRepositories(aggregations)

    blobs.forEach { blob -> blobSubmitter.submitBlob(blob).get() }

    lineaRollupContract.resetNonce().get()
    val aggregationFinalizationCoordinator = setupAggregationFinalizationCoordinator(vertx)

    aggregationFinalizationCoordinator.start()
      .thenApply {
        waitAtMost(2.minutes.toJavaDuration())
          .pollInterval(1.seconds.toJavaDuration())
          .untilAsserted {
            val finalizedBlockNumber = lineaRollupContract.currentL2BlockNumber().send()
            assertThat(finalizedBlockNumber.toLong()).isEqualTo(aggregations.last().finalBlockNumber)
          }
        testContext.completeNow()
      }.whenException(testContext::failNow)
  }

  @Test
  @Timeout(15, timeUnit = TimeUnit.SECONDS)
  fun `aggregation finalization transaction isn't sent until all submission data is accounted for`(
    vertx: Vertx,
    testContext: VertxTestContext
  ) {
    val blob1 = createBlobRecord(
      startBlockNumber = currentL2BlockNumber + 1UL,
      endBlockNumber = currentL2BlockNumber + 10UL,
      parentStateRootHash = firstParentStateRootHash,
      parentDataHash = emptyHash,
      startBlockTime = expectedStartBlockTime
    )
    val blob2A = createBlobRecord(
      startBlockNumber = currentL2BlockNumber + 11UL,
      endBlockNumber = currentL2BlockNumber + 20UL,
      parentStateRootHash = blob1.blobCompressionProof!!.finalStateRootHash,
      parentDataHash = blob1.blobCompressionProof!!.dataHash,
      startBlockTime = expectedStartBlockTime
    )
    val blob2B = createBlobRecord(
      startBlockNumber = currentL2BlockNumber + 11UL,
      endBlockNumber = currentL2BlockNumber + 20UL,
      conflationCalculationVersion = "0.1.1",
      parentStateRootHash = blob1.blobCompressionProof!!.finalStateRootHash,
      parentDataHash = blob1.blobCompressionProof!!.dataHash,
      startBlockTime = expectedStartBlockTime
    )
    val lastFinalizedBlockTime = Instant.fromEpochSeconds(lineaRollupContract.currentTimestamp().send().toLong())

    val unsubmittedAggregation = createProofToFinalizeFromBlobs(
      blobRecords = listOf(blob1, blob2A),
      lastFinalizedBlockTime = lastFinalizedBlockTime,
      parentStateRootHash = startingParentStateRootHash,
      dataParentHash = emptyHash
    )

    val fullySubmittedAggregation = createProofToFinalizeFromBlobs(
      blobRecords = listOf(blob1, blob2B),
      lastFinalizedBlockTime = lastFinalizedBlockTime,
      parentStateRootHash = startingParentStateRootHash,
      dataParentHash = emptyHash
    )
    val blobsToSubmit = listOf(blob1, blob2B)

    lineaRollupContract.resetNonce().get()
    blobsToSubmit.forEach { blob -> blobSubmitter.submitBlob(blob).get() }

    mockRepositories(listOf(fullySubmittedAggregation, unsubmittedAggregation))
    val aggregationFinalizationCoordinator = setupAggregationFinalizationCoordinator(vertx)

    aggregationFinalizationCoordinator.start()
      .thenApply {
        await()
          .untilAsserted {
            val finalizedBlockNumber = lineaRollupContract.currentL2BlockNumber().send()
            assertThat(finalizedBlockNumber.toLong()).isEqualTo(fullySubmittedAggregation.finalBlockNumber)
          }
        testContext.completeNow()
      }
  }

  private fun setupAggregationFinalizationCoordinator(vertx: Vertx): AggregationFinalizationCoordinator {
    val aggregationFinalizationCoordinator = AggregationFinalizationCoordinator(
      config = AggregationFinalizationCoordinator.Config(
        pollingInterval,
        proofSubmissionDelay,
        maxAggregationsToFinalizePerIteration
      ),
      aggregationFinalization = aggregationFinalization,
      aggregationsRepository = aggregationsRepository,
      lineaRollup = lineaRollupContract,
      vertx = vertx,
      clock = fixedClock
    )
    return aggregationFinalizationCoordinator
  }

  private fun mockRepositories(proofsToFinalize: List<ProofToFinalize>) {
    whenever(aggregationsRepository.getProofsToFinalize(any(), any(), any(), any()))
      .thenAnswer { invocation ->
        val firstBlockNumber = invocation.arguments[0] as Long
        val proofsToReturn = proofsToFinalize.filter { it.finalBlockNumber >= firstBlockNumber }
        SafeFuture.completedFuture(proofsToReturn)
      }
  }
}
