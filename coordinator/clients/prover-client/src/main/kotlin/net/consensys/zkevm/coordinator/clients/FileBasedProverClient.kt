package net.consensys.zkevm.coordinator.clients

import com.fasterxml.jackson.databind.DeserializationFeature
import com.fasterxml.jackson.databind.MapperFeature
import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.databind.module.SimpleModule
import com.fasterxml.jackson.databind.node.ArrayNode
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule
import com.fasterxml.jackson.module.kotlin.jacksonMapperBuilder
import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import io.vertx.core.Vertx
import net.consensys.linea.CommonDomainFunctions.batchIntervalString
import net.consensys.linea.async.toSafeFuture
import net.consensys.linea.contract.L2MessageService
import net.consensys.linea.errors.ErrorResponse
import net.consensys.zkevm.coordinator.clients.prover.ProverFilesNameProvider
import net.consensys.zkevm.coordinator.clients.prover.ProverFilesNameProviderImplV1
import net.consensys.zkevm.coordinator.clients.prover.RequestFileWriter
import net.consensys.zkevm.coordinator.clients.response.FileBasedProverResponsesRepository
import net.consensys.zkevm.coordinator.clients.response.ProverResponsesRepository
import net.consensys.zkevm.toULong
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes
import org.web3j.abi.EventEncoder
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameter
import org.web3j.protocol.core.methods.request.EthFilter
import org.web3j.protocol.core.methods.response.EthLog.LogObject
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.ethereum.executionclient.serialization.BytesSerializer
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.bytes.Bytes20
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import java.math.BigInteger
import java.nio.file.Path
import kotlin.io.path.notExists
import kotlin.time.Duration

/**
 * Implementation of interface with the Prover trough Files.
 *
 * Prover will ingest file like
 * path/to/prover-input-dir/<startBlockNumber>-<endBlockNumber>-getZkProof.json
 *
 * When done prover will output file
 * path/to/prover-output-dir/<startBlockNumber>-<endBlockNumber>-proof.json
 *
 * So, this class will need to watch the file system and wait for the output proof to be generated
 */
class FileBasedProverClient(
  private val config: Config,
  private val vertx: Vertx,
  private val l2Web3jClient: Web3j,
  private val mapper: ObjectMapper = objectMapperV1,
  private val fileNamesProvider: ProverFilesNameProvider = ProverFilesNameProviderImplV1(
    config.tracesVersion,
    config.stateManagerVersion,
    config.proverVersion,
    "json"
  ),
  val proverResponsesRepository: FileBasedProverResponsesRepository = FileBasedProverResponsesRepository(
    FileBasedProverResponsesRepository.Config(config.responseDirectory, config.proverVersion),
    mapper = mapper,
    proofFileNameSupplier = fileNamesProvider
  )
) : ProverClient {
  private val log: Logger = LogManager.getLogger(this::class.java)
  private val timeoutAttempts = config.timeout / config.pollingInterval
  private val requestFileWriter: RequestFileWriter = RequestFileWriter(
    vertx = vertx,
    config = RequestFileWriter.Config(
      requestDirectory = config.requestDirectory,
      writingInprogressSuffix = ".coordinator_writing_inprogress",
      proverInprogressSuffixPattern = config.inprogessProvingSuffixPattern
    ),
    mapper = mapper,
    fileNamesProvider = fileNamesProvider,
    log = log
  )

  init {
    if (config.requestDirectory.notExists()) {
      config.requestDirectory.toFile().mkdirs()
    }
    if (config.responseDirectory.notExists()) {
      config.responseDirectory.toFile().mkdirs()
    }
  }

  data class Config(
    val requestDirectory: Path,
    val responseDirectory: Path,
    val inprogessProvingSuffixPattern: String,
    val pollingInterval: Duration,
    val timeout: Duration,
    val tracesVersion: String,
    val stateManagerVersion: String,
    val proverVersion: String,
    val l2MessageServiceAddress: Bytes20
  )

  internal data class RlpBridgeLogsData(val rlp: String, val bridgeLogs: List<BridgeLogsData>)

  data class BridgeLogsData(
    val removed: Boolean,
    val logIndex: String,
    val transactionIndex: String,
    val transactionHash: String,
    val blockHash: String,
    val blockNumber: String,
    val address: String,
    val data: String,
    val topics: List<String>
  )

  internal data class GetProofRequest(
    val zkParentStateRootHash: String?,
    val keccakParentStateRootHash: String,
    val conflatedExecutionTracesFile: String,
    val tracesEngineVersion: String,
    val type2StateManagerVersion: String?,
    val zkStateMerkleProof: ArrayNode,
    val blocksData: List<RlpBridgeLogsData>
  )

  internal inner class ResponseFileMonitor(
    startBlockNumber: UInt64,
    endBlockNumber: UInt64
  ) {
    private val proverResponseIndex = ProverResponsesRepository.ProverResponseIndex(
      startBlockNumber,
      endBlockNumber,
      config.proverVersion
    )

    fun findResponse(): SafeFuture<GetProofResponse?> {
      return proverResponsesRepository
        .find(proverResponseIndex)
        .thenApply { response -> response.component1() }
    }

    fun monitor(): SafeFuture<Result<GetProofResponse, ErrorResponse<ProverErrorType>>> {
      val result = SafeFuture<Result<GetProofResponse, ErrorResponse<ProverErrorType>>>()
      var attempts = 0
      val monitorStream = vertx.periodicStream(config.pollingInterval.inWholeMilliseconds)

      monitorStream.handler {
        attempts += 1
        monitorStream.pause()
        proverResponsesRepository
          .find(proverResponseIndex)
          .thenApply { response ->
            when (response) {
              is Ok -> {
                monitorStream.cancel()
                result.complete(response)
              }

              is Err -> {
                if (attempts > timeoutAttempts) {
                  val error =
                    Err(
                      ErrorResponse(
                        ProverErrorType.ResponseTimeout,
                        "Monitoring timed out after $attempts attempts," +
                          " ${(attempts * config.pollingInterval.inWholeMilliseconds) / 1000f} seconds"
                      )
                    )
                  result.complete(error)
                  monitorStream.cancel()
                }
                Unit
              }
            }
          }
          .whenComplete { _, _ -> monitorStream.resume() }
      }
      return result
    }
  }

  private fun getBridgeLogsByBlockNumber(
    blockNumber: Long,
    l2MessageServiceAddress: String
  ): SafeFuture<List<BridgeLogsData>> {
    return vertx
      .executeBlocking(
        { promise ->
          try {
            val ethFilter =
              EthFilter(
                DefaultBlockParameter.valueOf(BigInteger.valueOf(blockNumber)),
                DefaultBlockParameter.valueOf(BigInteger.valueOf(blockNumber)),
                l2MessageServiceAddress
              )
            ethFilter.addOptionalTopics(
              EventEncoder.encode(L2MessageService.L1L2MESSAGEHASHESADDEDTOINBOX_EVENT),
              EventEncoder.encode(L2MessageService.MESSAGESENT_EVENT)
            )

            val ethLogs = l2Web3jClient.ethGetLogs(ethFilter).send()
            promise.complete(
              ethLogs.logs.map { log -> parseBridgeLogsData(log.get() as LogObject) }
            )
          } catch (th: Throwable) {
            promise.fail(th)
          }
        },
        false
      ) // Setting "ordered" to false for parallel execution in worker pool
      .toSafeFuture()
  }

  private fun getPreviousBlockKeccakStateRootHash(blockNumber: Long): SafeFuture<String> {
    return SafeFuture.of(
      l2Web3jClient
        .ethGetBlockByNumber(
          DefaultBlockParameter.valueOf(BigInteger.valueOf(blockNumber - 1)),
          false
        )
        .sendAsync()
    )
      .thenApply { previousBlock -> previousBlock.block.stateRoot }
  }

  override fun getZkProof(
    blocks: List<ExecutionPayloadV1>,
    tracesResponse: GenerateTracesResponse,
    type2StateData: GetZkEVMStateMerkleProofResponse
  ): SafeFuture<Result<GetProofResponse, ErrorResponse<ProverErrorType>>> {
    val responseMonitor = ResponseFileMonitor(blocks.first().blockNumber, blocks.last().blockNumber)

    // Check if the request is already proven. If so, return it.
    // This happens when coordinator is restarted and the request is already proven.
    val alreadyProvenResponse = responseMonitor.findResponse().get()
    if (alreadyProvenResponse != null) {
      log.info(
        "Reusing proof already done for batch={}. Found file {}",
        batchIntervalString(blocks.first().blockNumber.toULong(), blocks.last().blockNumber.toULong()),
        fileNamesProvider.getResponseFileName(blocks.first().blockNumber.toULong(), blocks.last().blockNumber.toULong())
      )
      return SafeFuture.completedFuture(Ok(alreadyProvenResponse))
    }

    // Load all the requests to the worker pool to run in parallel
    val bridgeLogsSfList =
      blocks.map { block ->
        getBridgeLogsByBlockNumber(
          block.blockNumber.longValue(),
          config.l2MessageServiceAddress.toHexString()
        )
      }

    return SafeFuture.collectAll(bridgeLogsSfList.stream())
      .thenComposeCombined(
        getPreviousBlockKeccakStateRootHash(blocks.first().blockNumber.longValue())
      ) { blocksLogs,
        previousKeccakStateRootHash ->
        val bundledBlocks = blocks.zip(blocksLogs)
        requestFileWriter.write(
          bundledBlocks,
          tracesResponse,
          type2StateData,
          previousKeccakStateRootHash
        )
      }
      .thenCompose { responseMonitor.monitor() }
  }

  companion object {
    val objectMapperV1: ObjectMapper =
      jacksonMapperBuilder()
        .enable(MapperFeature.ACCEPT_CASE_INSENSITIVE_ENUMS)
        .configure(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, false)
        .addModule(JavaTimeModule())
        .addModule(SimpleModule().addSerializer(Bytes::class.java, BytesSerializer()))
        .build()

    fun parseBridgeLogsData(logObject: LogObject): BridgeLogsData {
      return BridgeLogsData(
        removed = logObject.isRemoved,
        logIndex = logObject.logIndexRaw,
        transactionIndex = logObject.transactionIndexRaw,
        transactionHash = logObject.transactionHash,
        blockHash = logObject.blockHash,
        blockNumber = logObject.blockNumberRaw,
        address = logObject.address,
        data = logObject.data,
        topics = logObject.topics
      )
    }
  }
}
