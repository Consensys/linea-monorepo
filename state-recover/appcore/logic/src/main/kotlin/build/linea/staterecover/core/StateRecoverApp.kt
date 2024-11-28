package build.linea.staterecover.core

import build.linea.clients.StateManagerClientV1
import build.linea.contract.l1.LineaRollupSmartContractClientReadOnly
import build.linea.domain.EthLogEvent
import build.linea.staterecover.clients.BlobFetcher
import build.linea.staterecover.clients.DataFinalizedV3
import build.linea.staterecover.clients.ExecutionLayerClient
import build.linea.staterecover.clients.LineaRollupSubmissionEventsClient
import build.linea.staterecover.clients.StateRecoveryStatus
import build.linea.staterecover.clients.TransactionDetailsClient
import io.vertx.core.Vertx
import net.consensys.linea.BlockParameter
import net.consensys.linea.async.AsyncRetryer
import net.consensys.linea.blob.BlobDecompressorVersion
import net.consensys.linea.blob.GoNativeBlobDecompressorFactory
import net.consensys.zkevm.LongRunningService
import org.apache.logging.log4j.LogManager
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.CompletableFuture
import kotlin.time.Duration
import kotlin.time.Duration.Companion.seconds

class StateRecoverApp(
  private val vertx: Vertx,
  // Driving Ports
  private val lineaContractClient: LineaRollupSmartContractClientReadOnly,
  private val l1EventsClient: LineaRollupSubmissionEventsClient,
  // Driven Ports
  private val blobFetcher: BlobFetcher,
  private val elClient: ExecutionLayerClient,
  private val stateManagerClient: StateManagerClientV1,
  private val transactionDetailsClient: TransactionDetailsClient,
  private val l1EventsPollingInterval: Duration,
  private val blockHeaderStaticFields: BlockHeaderStaticFields,
  // configs
  private val config: Config = Config()
) : LongRunningService {
  data class Config(
    val l1HighestBlockTag: BlockParameter = BlockParameter.Tag.FINALIZED,
    val l1PollingInterval: Duration = 12.seconds,
    val executionClientPollingInterval: Duration = 1.seconds,
    val blobDecompressorVersion: BlobDecompressorVersion = BlobDecompressorVersion.V1_1_0
  )

  private val log = LogManager.getLogger(this::class.java)
  private val blockImporterAndStateVerifier = BlockImporterAndStateVerifierV1(
    vertx = vertx,
    elClient = elClient,
    stateManagerClient = stateManagerClient,
    stateManagerImportTimeoutPerBlock = 2.seconds
  )
  private val blobDecompressor: BlobDecompressorAndDeserializer = BlobDecompressorToDomainV1(
    decompressor = GoNativeBlobDecompressorFactory.getInstance(config.blobDecompressorVersion),
    staticFields = blockHeaderStaticFields
  )
  private val stateSynchronizerService = StateSynchronizerService(
    vertx = vertx,
    elClient = elClient,
    submissionEventsClient = l1EventsClient,
    blobsFetcher = blobFetcher,
    transactionDetailsClient = transactionDetailsClient,
    blobDecompressor = blobDecompressor,
    blockImporterAndStateVerifier = blockImporterAndStateVerifier,
    pollingInterval = l1EventsPollingInterval
  )
  val lastProcessedFinalization: EthLogEvent<DataFinalizedV3>?
    get() = stateSynchronizerService.lastProcessedFinalization

  private fun enableRecoveryMode(): SafeFuture<*> {
    /* start up scenarios:
    1. when state recovery disabled:
      1.1 when there are no finalizations - start from genesis or V6 upgrade
      1.2 when there is finalization - start from the last finalized block
    2. when state recovery enabled:
      2.1 recoveryStartBlockNumber > headBlockNumber:
        pull for head block number until is reached and start recovery there
      2.2 recoveryStartBlockNumber <= headBlockNumber:
        resume recovery from headBlockNumber
    */
    return elClient
      .lineaGetStateRecoveryStatus()
      .thenCompose { status ->
        if (status.stateRecoverStartBlockNumber != null) {
          // already enabled, let's just resume from where we left off
          log.info(
            "Starting recovery mode already enabled: stateRecoverStartBlockNumber={} headBlockNumber={}",
            status.stateRecoverStartBlockNumber,
            status.headBlockNumber
          )
          SafeFuture.completedFuture(Unit)
        } else {
          lineaContractClient.finalizedL2BlockNumber(blockParameter = config.l1HighestBlockTag)
            .thenCompose { lastFinalizedBlockNumber ->
              val stateRecoverStartBlockNumber = lastFinalizedBlockNumber + 1UL
              log.info(
                "Starting enabling recovery mode: stateRecoverStartBlockNumber={} headBlockNumber={}",
                stateRecoverStartBlockNumber,
                status.headBlockNumber
              )
              elClient.lineaEnableStateRecovery(stateRecoverStartBlockNumber)
            }.thenApply { }
        }
      }
  }

  private fun waitForSyncUntilStateRecoverBlock(): SafeFuture<StateRecoveryStatus> {
    return AsyncRetryer.retry(
      vertx = vertx,
      backoffDelay = config.executionClientPollingInterval,
      stopRetriesPredicate = {
        // headBlockNumber shall be at least 1 block behind of stateRecoverStartBlockNumber
        // if it is after is means it has already enabled
        (it.headBlockNumber + 1UL) >= (it.stateRecoverStartBlockNumber ?: 0UL)
      }
    ) {
      elClient.lineaGetStateRecoveryStatus()
    }
  }

  override fun start(): CompletableFuture<Unit> {
    return enableRecoveryMode()
      .thenCompose { waitForSyncUntilStateRecoverBlock() }
      .thenCompose { stateSynchronizerService.start() }
  }

  override fun stop(): CompletableFuture<Unit> {
    return stateSynchronizerService.stop()
  }
}
