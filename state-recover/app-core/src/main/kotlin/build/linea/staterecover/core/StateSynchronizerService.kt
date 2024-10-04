package build.linea.staterecover.core

import build.linea.staterecover.clients.BlobFetcher
import build.linea.staterecover.clients.DataFinalizedV3
import build.linea.staterecover.clients.DataSubmittedV3
import build.linea.staterecover.clients.ExecutionLayerClient
import build.linea.staterecover.clients.LineaRollupSubmissionEventsClient
import build.linea.staterecover.clients.TransactionDetailsClient
import io.vertx.core.Vertx
import net.consensys.encodeHex
import net.consensys.eth.EthLogEvent
import net.consensys.linea.BlockNumberAndHash
import net.consensys.linea.BlockParameter
import net.consensys.tuweni.bytes.toBytes32
import net.consensys.zkevm.PeriodicPollingService
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture

class StateSynchronizerService(
  private val vertx: Vertx,
  private val elClient: ExecutionLayerClient,
  private val submissionEventsClient: LineaRollupSubmissionEventsClient,
  private val blobsFetcher: BlobFetcher,
  private val transactionDetailsClient: TransactionDetailsClient,
  private val blobDecompressor: BlobDecompressorToDomain,
  private val blockImporterAndStateVerifier: BlockImporterAndStateVerifier,
  private val log: Logger = LogManager.getLogger(StateSynchronizerService::class.java)
) : PeriodicPollingService(
  vertx = vertx,
  log = log,
  pollingIntervalMs = 1000L
) {
  private data class DataSubmittedEventAndBlobs(
    val ethLogEvent: EthLogEvent<DataSubmittedV3>,
    val blobs: List<ByteArray>
  )

  private var lastProcessedFinalization: EthLogEvent<DataFinalizedV3>? = null
  private var stateRootMismatchFound: Boolean = false

  private fun findNextFinalization(): SafeFuture<EthLogEvent<DataFinalizedV3>?> {
    return if (lastProcessedFinalization != null) {
      submissionEventsClient
        .findDataFinalizedEventByStartBlockNumber(blockNumber = lastProcessedFinalization!!.event.endBlockNumber + 1UL)
    } else {
      // 1st, assuming head matches a prev finalization,
      elClient.getBlockNumberAndHash(blockParameter = BlockParameter.Tag.LATEST)
        .thenCompose { headBlock ->
          submissionEventsClient
            .findDataFinalizedEventByStartBlockNumber(headBlock.number + 1UL)
            .thenCompose { finalizationEvent ->
              finalizationEvent?.let { SafeFuture.completedFuture(finalizationEvent) }
              // 2nd: otherwise, local head may be in between, let's find corresponding finalization
              submissionEventsClient
                .findDataFinalizedEventContainingBlock(headBlock.number)
            }
        }
    }
  }

  override fun action(): SafeFuture<Any?> {
    return findNextFinalization()
      .thenCompose { nextFinalization ->
        if (nextFinalization == null) {
          // nothing to do for now
          SafeFuture.completedFuture(null)
        } else {
          submissionEventsClient
            .findDataSubmittedV3EventsUtilNextFinalization(
              l2StartBlockNumberInclusive = nextFinalization.event.startBlockNumber + 1UL
            )
        }
      }
      .thenCompose { submissionEvents ->
        if (submissionEvents == null) {
          SafeFuture.completedFuture("No new events")
        } else {
          getBlobsForEvents(submissionEvents.dataSubmittedEvents)
            .thenCompose { dataSubmissionsWithBlobs ->
              updateNodeWithBlobsAndVerifyState(dataSubmissionsWithBlobs, submissionEvents.dataFinalizedEvent.event)
            }
            .thenCompose { blockNumberAndHash ->
              // if stateRootHash matches, finalize block on EC side
              elClient.lineaEngineForkChoiceUpdated(
                finalizedBlockHash = blockNumberAndHash.hash.toArray(),
                headBlockHash = blockNumberAndHash.hash.toArray()
              )
            }
            .thenApply {
              lastProcessedFinalization = submissionEvents.dataFinalizedEvent
              log.info("Imported blocks={} matching stateRootHash={}")
              " ${submissionEvents.dataSubmittedEvents.size} events"
            }
        }
      }
  }

  private fun getBlobsForEvents(
    events: List<EthLogEvent<DataSubmittedV3>>
  ): SafeFuture<List<DataSubmittedEventAndBlobs>> {
    return SafeFuture.collectAll(
      events
        .map { dataSubmittedEvent ->
          transactionDetailsClient
            .getBlobVersionedHashesByTransactionHash(dataSubmittedEvent.log.transactionHash.toArray())
            .thenCompose(blobsFetcher::fetchBlobsByHash)
            .thenApply { blobs -> DataSubmittedEventAndBlobs(dataSubmittedEvent, blobs) }
        }.stream()
    )
  }

  private fun updateNodeWithBlobsAndVerifyState(
    dataSubmissions: List<DataSubmittedEventAndBlobs>,
    dataFinalizedV3: DataFinalizedV3
  ): SafeFuture<BlockNumberAndHash> {
    val decompressedBlocks: List<BlockL1RecoveredData> = blobDecompressor
      .decompress(dataSubmissions.flatMap { it.blobs })
      .sortedBy { it.blockNumber }

    return blockImporterAndStateVerifier
      .importBlocks(decompressedBlocks)
      .thenCompose { importResult -> assertStateMatches(importResult, dataFinalizedV3) }
      .thenApply {
        BlockNumberAndHash(
          number = decompressedBlocks.last().blockNumber,
          hash = decompressedBlocks.last().blockHash.toBytes32()
        )
      }
  }

  private fun assertStateMatches(
    importResult: ImportResult,
    finalizedV3: DataFinalizedV3
  ): SafeFuture<Unit> {
    return if (importResult.zkStateRootHash.contentEquals(finalizedV3.finalStateRootHash.toArray())) {
      SafeFuture.completedFuture(Unit)
    } else {
      log.error(
        "stateRootHash mismatch: finalization={} l1StateRootHash={} localStateRootHash={}",
        finalizedV3.intervalString(),
        finalizedV3.finalStateRootHash,
        importResult.zkStateRootHash.encodeHex()
      )
      stateRootMismatchFound = true
      SafeFuture.failedFuture(IllegalStateException("stateRootHash mismatch"))
    }
  }
}
