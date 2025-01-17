package linea.staterecover

import build.linea.clients.StateManagerClientV1
import build.linea.contract.l1.LineaRollupSmartContractClientReadOnly
import build.linea.domain.EthLogEvent
import io.vertx.core.Vertx
import linea.EthLogsSearcher
import net.consensys.linea.BlockParameter
import net.consensys.linea.BlockParameter.Companion.toBlockParameter
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
  private val ethLogsSearcher: EthLogsSearcher,
  // Driven Ports
  private val blobFetcher: BlobFetcher,
  private val elClient: ExecutionLayerClient,
  private val stateManagerClient: StateManagerClientV1,
  private val transactionDetailsClient: TransactionDetailsClient,
  private val blockHeaderStaticFields: BlockHeaderStaticFields,
  // configs
  private val config: Config = Config.lineaMainnet
) : LongRunningService {
  data class Config(
    val smartContractAddress: String,
    val l1EarliestSearchBlock: BlockParameter = BlockParameter.Tag.EARLIEST,
    val l1LatestSearchBlock: BlockParameter = BlockParameter.Tag.FINALIZED,
    val l1PollingInterval: Duration = 12.seconds,
    val executionClientPollingInterval: Duration = 1.seconds,
    val blobDecompressorVersion: BlobDecompressorVersion = BlobDecompressorVersion.V1_1_0,
    val logsBlockChunkSize: UInt = 1000u,
    /**
     * The block number at which the recovery mode will start overriding the recovery start block number
     * this is meant for testing purposes, not production
     */
    val overridingRecoveryStartBlockNumber: ULong? = null
  ) {
    companion object {
      val lineaMainnet = Config(
        smartContractAddress = "0xd19d4b5d358258f05d7b411e21a1460d11b0876f",
        // TODO: set block of V6 Upgrade
        l1EarliestSearchBlock = 1UL.toBlockParameter(),
        l1LatestSearchBlock = BlockParameter.Tag.FINALIZED,
        executionClientPollingInterval = 10.seconds,
        l1PollingInterval = 12.seconds
      )
      val lineaSepolia = Config(
        smartContractAddress = "0xb218f8a4bc926cf1ca7b3423c154a0d627bdb7e5",
        l1EarliestSearchBlock = 7164537UL.toBlockParameter(),
        l1LatestSearchBlock = BlockParameter.Tag.FINALIZED,
        executionClientPollingInterval = 10.seconds,
        l1PollingInterval = 12.seconds
      )
    }
  }

  init {
    require(config.smartContractAddress.lowercase() == lineaContractClient.getAddress().lowercase()) {
      "contract address mismatch: config=${config.smartContractAddress} client=${lineaContractClient.getAddress()}"
    }
  }
  private val l1EventsClient = LineaSubmissionEventsClientImpl(
    logsSearcher = ethLogsSearcher,
    smartContractAddress = config.smartContractAddress,
    l1EarliestSearchBlock = config.l1EarliestSearchBlock,
    l1LatestSearchBlock = config.l1LatestSearchBlock,
    logsBlockChunkSize = config.logsBlockChunkSize.toInt()
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
    staticFields = blockHeaderStaticFields,
    vertx = vertx
  )
  private val stateSynchronizerService = StateSynchronizerService(
    vertx = vertx,
    elClient = elClient,
    submissionEventsClient = l1EventsClient,
    blobsFetcher = blobFetcher,
    transactionDetailsClient = transactionDetailsClient,
    blobDecompressor = blobDecompressor,
    blockImporterAndStateVerifier = blockImporterAndStateVerifier,
    pollingInterval = config.l1PollingInterval
  )
  val lastSuccessfullyRecoveredFinalization: EthLogEvent<DataFinalizedV3>?
    get() = stateSynchronizerService.lastSuccessfullyProcessedFinalization
  val stateRootMismatchFound: Boolean
    get() = stateSynchronizerService.stateRootMismatchFound

  fun trySetRecoveryModeAtBlockHeight(stateRecoverStartBlockNumber: ULong): SafeFuture<StateRecoveryStatus> {
    return elClient
      .lineaGetStateRecoveryStatus()
      .thenCompose { statusBeforeUpdate ->
        elClient
          .lineaEnableStateRecovery(stateRecoverStartBlockNumber)
          .thenPeek { newStatus ->
            val updateLabel = if (statusBeforeUpdate.stateRecoverStartBlockNumber == null) "Enabled" else "Updated"
            log.info(
              "Recovery mode was {}: headBlockNumber={} " +
                "prevStartBlockNumber={} newStartBlockNumber={}",
              updateLabel,
              newStatus.headBlockNumber,
              statusBeforeUpdate.stateRecoverStartBlockNumber,
              newStatus.stateRecoverStartBlockNumber
            )
          }
      }
  }

  private fun enableRecoveryMode(): SafeFuture<*> {
    if (config.overridingRecoveryStartBlockNumber != null) {
      return trySetRecoveryModeAtBlockHeight(config.overridingRecoveryStartBlockNumber)
    }

    return elClient
      .lineaGetStateRecoveryStatus()
      .thenCompose { status ->
        if (status.stateRecoverStartBlockNumber != null) {
          // already enabled, let's just resume from where we left off
          log.info(
            "starting recovery mode already enabled: stateRecoverStartBlockNumber={} headBlockNumber={}",
            status.stateRecoverStartBlockNumber,
            status.headBlockNumber
          )
          SafeFuture.completedFuture(Unit)
        } else {
          lineaContractClient.finalizedL2BlockNumber(blockParameter = config.l1LatestSearchBlock)
            .thenCompose { lastFinalizedBlockNumber ->
              val stateRecoverStartBlockNumber = when {
                status.headBlockNumber >= lastFinalizedBlockNumber -> status.headBlockNumber + 1UL
                else -> lastFinalizedBlockNumber + 1UL
              }
              log.info(
                "Starting enabling recovery mode: stateRecoverStartBlockNumber={} headBlockNumber={} " +
                  "L1 lastFinalizedBlockNumber={}",
                stateRecoverStartBlockNumber,
                status.headBlockNumber,
                lastFinalizedBlockNumber
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
      stopRetriesPredicate = { recoveryStatus ->
        log.debug(
          "waiting for node to sync until stateRecoverStartBlockNumber={} headBlockNumber={}",
          recoveryStatus.stateRecoverStartBlockNumber,
          recoveryStatus.headBlockNumber
        )
        // headBlockNumber shall be at least 1 block behind of stateRecoverStartBlockNumber
        // if it is after it means it was already enabled
        recoveryStatus.stateRecoverStartBlockNumber?.let { startBlockNumber ->
          recoveryStatus.headBlockNumber + 1u >= startBlockNumber
        } ?: false
      }
    ) {
      elClient.lineaGetStateRecoveryStatus()
    }
  }

  override fun start(): CompletableFuture<Unit> {
    log.warn("Starting StateRecoverApp")
    val enablementFuture = enableRecoveryMode()

    enablementFuture
      .thenCompose { waitForSyncUntilStateRecoverBlock() }
      .thenCompose { stateSynchronizerService.start() }

    return enablementFuture.thenApply { }
  }

  override fun stop(): CompletableFuture<Unit> {
    return stateSynchronizerService.stop()
  }
}
