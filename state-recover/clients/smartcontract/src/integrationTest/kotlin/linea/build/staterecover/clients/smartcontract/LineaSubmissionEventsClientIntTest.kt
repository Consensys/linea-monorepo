package linea.build.staterecover.clients.smartcontract

import build.linea.staterecover.clients.DataFinalizedV3
import build.linea.staterecover.clients.DataSubmittedV3
import build.linea.staterecover.clients.LineaRollupSubmissionEventsClient
import io.vertx.core.Vertx
import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import net.consensys.linea.contract.Web3JLogsClient
import net.consensys.linea.testing.submission.loadBlobsAndAggregations
import net.consensys.tuweni.bytes.toBytes32
import net.consensys.zkevm.coordinator.clients.smartcontract.LineaContractVersion
import net.consensys.zkevm.coordinator.clients.smartcontract.LineaRollupSmartContractClient
import net.consensys.zkevm.domain.Aggregation
import net.consensys.zkevm.domain.BlobRecord
import net.consensys.zkevm.ethereum.ContractsManager
import net.consensys.zkevm.ethereum.Web3jClientManager
import org.apache.logging.log4j.Level
import org.apache.logging.log4j.LogManager
import org.assertj.core.api.Assertions.assertThat
import org.awaitility.Awaitility.await
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.TimeUnit
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration

@ExtendWith(VertxExtension::class)
class LineaSubmissionEventsClientIntTest {
  private lateinit var contractClient: LineaRollupSmartContractClient
  private val testDataDir = "testdata/coordinator/prover/v2/"

  // 1-block-per-blob test data has 3 aggregations: 1..7, 8..14, 15..21.
  // We will upgrade the contract in the middle of 2nd aggregation: 12
  // shall submit blob 12, stop submission, upgrade the contract and resume with blob 13
  // val lastSubmittedBlobs = blobs.filter { it.startBlockNumber == 7UL }
  private lateinit var aggregations: List<Aggregation>
  private lateinit var blobs: List<BlobRecord>
  private lateinit var submissionEventsFetcher: LineaRollupSubmissionEventsClient

  private fun setupTest(
    vertx: Vertx
  ) {
    val rollupDeploymentFuture = ContractsManager.get()
      .deployLineaRollup(numberOfOperators = 2, contractVersion = LineaContractVersion.V5)
    // load files from FS while smc deploy
    loadBlobsAndAggregations(
      blobsResponsesDir = "$testDataDir/compression/responses",
      aggregationsResponsesDir = "$testDataDir/aggregation/responses"
    )
      .let { (blobs, aggregations) ->
        this.blobs = blobs.sortedBy { it.startBlockNumber }
        this.aggregations = aggregations.sortedBy { it.startBlockNumber }
      }
    // wait smc deployment finishes
    val rollupDeploymentResult = rollupDeploymentFuture.get()
    contractClient = rollupDeploymentResult.rollupOperatorClient
    val eventsFetcherWeb3jClient = Web3jClientManager.buildL1Client(
      log = LogManager.getLogger("test.clients.l1.events-fetcher"),
      requestResponseLogLevel = Level.INFO,
      failuresLogLevel = Level.WARN
    )
    submissionEventsFetcher = LineaSubmissionEventsClientWeb3jIpml(
      logsClient = Web3JLogsClient(
        vertx = vertx,
        web3jClient = eventsFetcherWeb3jClient,
        config = Web3JLogsClient.Config(
          timeout = 30.seconds,
          backoffDelay = 1.seconds,
          lookBackRange = 100
        )
      ),
      smartContractAddress = rollupDeploymentResult.contractAddress
    )
    println("smart contract deployed at ${rollupDeploymentResult.contractAddress}")
  }

  @Test
  @Timeout(3, timeUnit = TimeUnit.MINUTES)
  fun `submission works with contract V5`(
    vertx: Vertx,
    testContext: VertxTestContext
  ) {
    testSubmission(vertx, testContext)
  }

  private fun testSubmission(
    vertx: Vertx,
    testContext: VertxTestContext
  ) {
    setupTest(vertx)
    val aggregationsAndBlobs = groupBlobsToAggregations(aggregations, blobs)

    val blobSubmissionFutures = aggregationsAndBlobs
      .map { (_, aggBlobs) ->
        val blobChunks = aggBlobs.chunked(6)
        blobChunks.map { blobs -> contractClient.submitBlobs(blobs, gasPriceCaps = null) }
      }

    val finalizationFutures = aggregationsAndBlobs
      .filter { it.aggregation != null }
      .map { (aggregation, aggBlobs) ->
        aggregation as Aggregation
        val parentAgg = aggregations.find { it.endBlockNumber == aggregation.startBlockNumber - 1UL }
        contractClient.finalizeBlocks(
          aggregation = aggregation.aggregationProof!!,
          aggregationLastBlob = aggBlobs.last(),
          parentShnarf = aggBlobs.first().blobCompressionProof!!.prevShnarf,
          parentL1RollingHash = parentAgg?.aggregationProof?.l1RollingHash ?: ByteArray(32),
          parentL1RollingHashMessageNumber = parentAgg?.aggregationProof?.l1RollingHashMessageNumber ?: 0L,
          gasPriceCaps = null
        )
      }
    // wait for all blobs to be submitted
    SafeFuture
      .collectAll(blobSubmissionFutures.flatten().stream()).get()

    val finalizationTxHashes = SafeFuture
      .collectAll(finalizationFutures.stream()).get()

    // wait for all finalizations Txs to be mined
    await()
      .atMost(60.seconds.toJavaDuration())
      .untilAsserted {
        val receipt =
          Web3jClientManager.l1Client
            .ethGetTransactionReceipt(finalizationTxHashes.last()).send().transactionReceipt
        assertThat(receipt.isPresent).isTrue()
      }

    val expectedSubmissionEventsToFind: List<Pair<DataFinalizedV3, List<DataSubmittedV3>>> =
      aggregationsAndBlobs
        .filter { it.aggregation != null }
        .map { (aggregation, aggBlobs) ->
          val expectedDataSubmittedEvents: List<DataSubmittedV3> = aggBlobs
            .chunked(6)
            .map { blobsChunk ->
              DataSubmittedV3(
                startBlockNumber = blobsChunk.first().startBlockNumber,
                endBlockNumber = blobsChunk.last().endBlockNumber,
                parentShnarf = blobsChunk.first().blobCompressionProof!!.prevShnarf,
                shnarf = blobsChunk.last().expectedShnarf,
                finalStateRootHash = blobsChunk.last().blobCompressionProof!!.finalStateRootHash.toBytes32()
              )
            }

          aggregation as Aggregation
          DataFinalizedV3(
            startBlockNumber = aggregation.startBlockNumber,
            endBlockNumber = aggregation.endBlockNumber,
            snarf = aggBlobs.last().expectedShnarf.toBytes32(),
            parentStateRootHash = aggBlobs.first().blobCompressionProof!!.parentStateRootHash.toBytes32(),
            finalStateRootHash = aggBlobs.last().blobCompressionProof!!.finalStateRootHash.toBytes32()
          ) to expectedDataSubmittedEvents
        }

    expectedSubmissionEventsToFind.forEach { (dataFinalizedEvent, dataSubmittedEvents) ->
      submissionEventsFetcher
        .findDataSubmittedV3EventsUtilNextFinalization(dataFinalizedEvent.startBlockNumber)
        .get()
        .also { submissionEvents ->
          assertThat(submissionEvents).isNotNull
          assertThat(submissionEvents!!.dataFinalizedEvent.event).isEqualTo(dataFinalizedEvent)
          assertThat(submissionEvents.dataSubmittedEvents.map { it.event }).isEqualTo(dataSubmittedEvents)
        }
    }

    testContext.completeNow()
  }

  data class BlobAggregationPair(
    val aggregation: Aggregation?,
    val blobs: List<BlobRecord>
  )

  private fun groupBlobsToAggregations(
    aggregations: List<Aggregation>,
    blobs: List<BlobRecord>
  ): List<BlobAggregationPair> {
    val aggBlobs = aggregations.map { agg ->
      BlobAggregationPair(agg, blobs.filter { it.startBlockNumber in agg.blocksRange })
    }.sortedBy { it.aggregation!!.startBlockNumber }

    val blobsWithoutAgg = blobs.filter { blob ->
      aggBlobs.none { it.blobs.contains(blob) }
    }
    return aggBlobs + listOf(BlobAggregationPair(null, blobsWithoutAgg))
  }
}
