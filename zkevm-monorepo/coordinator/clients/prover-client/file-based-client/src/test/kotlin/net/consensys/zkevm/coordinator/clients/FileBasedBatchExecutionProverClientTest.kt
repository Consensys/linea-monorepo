package net.consensys.zkevm.coordinator.clients
import com.fasterxml.jackson.databind.node.ArrayNode
import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import io.vertx.core.Vertx
import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import net.consensys.linea.contract.Web3JL2MessageServiceLogsClient
import net.consensys.linea.errors.ErrorResponse
import net.consensys.zkevm.coordinator.clients.prover.CommonTestData.bridgeLogs
import net.consensys.zkevm.coordinator.clients.prover.CommonTestData.ethLogs
import net.consensys.zkevm.coordinator.clients.prover.FileBasedExecutionProverClient
import net.consensys.zkevm.coordinator.clients.prover.ProverFilesNameProvider
import net.consensys.zkevm.coordinator.clients.prover.randomExecutionPayloads
import net.consensys.zkevm.coordinator.clients.prover.serialization.JsonSerialization.proofResponseMapperV1
import net.consensys.zkevm.coordinator.clients.prover.validateRequest
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.extension.ExtendWith
import org.junit.jupiter.api.io.TempDir
import org.junit.jupiter.params.ParameterizedTest
import org.junit.jupiter.params.provider.Arguments
import org.junit.jupiter.params.provider.MethodSource
import org.mockito.Mockito
import org.mockito.kotlin.any
import org.mockito.kotlin.eq
import org.mockito.kotlin.mock
import org.mockito.kotlin.whenever
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.methods.response.EthBlock
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.io.File
import java.nio.file.Path
import java.util.concurrent.TimeUnit
import java.util.stream.Stream
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.seconds

@ExtendWith(VertxExtension::class)
class FileBasedBatchExecutionProverClientTest {
  private val tracesFileName = "/some/path/1-3-conflated-traces.json.gz"
  private val tracesEngineVersion = "0.2.3"
  private val zkEvmStateManagerVersion = "0.3.4"
  private val proverVersion = "0.4.5"
  private val requestSubdirectory = "request"
  private val responseSubdirectory = "response"
  private val mapper = proofResponseMapperV1
  private val pollingInterval = 10.milliseconds
  private val mockBridgeLogsClient = mock<Web3JL2MessageServiceLogsClient>(defaultAnswer = Mockito.RETURNS_DEEP_STUBS)
  private val mockL2Client = mock<Web3j>(defaultAnswer = Mockito.RETURNS_DEEP_STUBS)
  private val previousStateRoot = Bytes32.random().toHexString()

  private val merkleProofJson: ArrayNode = let {
    val testFilePath = "$testdataPath/type2state-manager/state-proof.json"
    mapper.readTree(Path.of(testFilePath).toFile()).let {
      val merkleProof = it.get("zkStateMerkleProof")
      assert(merkleProof.isArray)
      merkleProof as ArrayNode
    }
  }

  companion object {
    private val testdataPath = "../../../../testdata"
    private val proverOutputs: Array<File> = File("$testdataPath/prover/output/").listFiles()!!

    @JvmStatic
    private fun proofFiles(): Stream<Arguments> {
      return proverOutputs.map { Arguments.of(it) }.stream()
    }

    @JvmStatic
    private fun blocksAndProofs(): Stream<Arguments> {
      return proverOutputs.mapIndexed { index, file -> Arguments.of(index + 1, file) }.stream()
    }
  }

  private fun buildProverClient(
    vertx: Vertx,
    tempDir: Path
  ): FileBasedExecutionProverClient {
    val responseDirectory = Path.of(tempDir.toString(), responseSubdirectory)
    return FileBasedExecutionProverClient(
      config = FileBasedExecutionProverClient.Config(
        requestDirectory = Path.of(tempDir.toString(), requestSubdirectory),
        responseDirectory = responseDirectory,
        inprogressProvingSuffixPattern = "\\.inprogress\\.prover.*",
        pollingInterval = pollingInterval,
        timeout = 1.seconds,
        tracesVersion = tracesEngineVersion,
        stateManagerVersion = zkEvmStateManagerVersion,
        proverVersion = proverVersion
      ),
      l2MessageServiceLogsClient = mockBridgeLogsClient,
      vertx = vertx,
      l2Web3jClient = mockL2Client,
      mapper = mapper,
      fileNamesProvider = SimpleFileNameProvider()
    )
  }

  private class SimpleFileNameProvider : ProverFilesNameProvider {
    override fun getRequestFileName(startBlockNumber: ULong, endBlockNumber: ULong): String {
      return "$startBlockNumber-$endBlockNumber-getZkProof.json"
    }

    override fun getResponseFileName(startBlockNumber: ULong, endBlockNumber: ULong): String {
      return "$startBlockNumber-$endBlockNumber-proof.json"
    }
  }

  @BeforeEach
  fun beforeEach() {
    // To warmup assertions otherwise first test may fail
    Assertions.assertThat(true).isTrue()
    whenever(mockBridgeLogsClient.getBridgeLogs(any())).thenAnswer { SafeFuture.completedFuture(bridgeLogs) }
    whenever(mockL2Client.ethGetLogs(any()).send().logs).thenAnswer { ethLogs }
    whenever(mockL2Client.ethGetBlockByNumber(any(), eq(false)).sendAsync()).thenAnswer {
      val blockResponse = EthBlock()
      blockResponse.result = EthBlock.Block()
      blockResponse.block.stateRoot = previousStateRoot
      SafeFuture.completedFuture(blockResponse)
    }
  }

  @Timeout(1, timeUnit = TimeUnit.SECONDS)
  @ParameterizedTest
  @MethodSource("proofFiles")
  fun responseFileMonitor_discoversResponseFile(
    proofFile: File,
    vertx: Vertx,
    @TempDir tempDir: Path,
    testContext: VertxTestContext
  ) {
    val outputDirectory = Path.of(tempDir.toString(), responseSubdirectory)
    val proverClient = buildProverClient(vertx, tempDir)
    val expectedProverOutput = mapper.readValue(proofFile, GetProofResponse::class.java)

    val startBlockNumber = 3123UL
    val endBlockNumber = 3129UL
    val fileMonitor = proverClient.ResponseFileMonitor(startBlockNumber, endBlockNumber)
    fileMonitor
      .monitor()
      .thenApply { response: Result<GetProofResponse, ErrorResponse<ProverErrorType>> ->
        testContext
          .verify { Assertions.assertThat(response).isEqualTo(Ok(expectedProverOutput)) }
          .completeNow()
      }
      .exceptionally { testContext.failNow(it) }

    val outputFileInprogress =
      "$outputDirectory/$startBlockNumber-$endBlockNumber-proof.json.inprogress"
    val outputFile = "$outputDirectory/$startBlockNumber-$endBlockNumber-proof.json"
    Thread.sleep(pollingInterval.inWholeMilliseconds * 2)
    val inprogressOutputFile = File(outputFileInprogress)
    proofFile.copyTo(inprogressOutputFile, true)
    inprogressOutputFile.renameTo(File(outputFile))
  }

  @Timeout(15, timeUnit = TimeUnit.SECONDS)
  @ParameterizedTest
  @MethodSource("blocksAndProofs")
  fun fileBasedProverClient_returnsProofs(
    blocksToGenerate: Int,
    proofFile: File,
    vertx: Vertx,
    @TempDir tempDir: Path,
    testContext: VertxTestContext
  ) {
    val outputDirectory = Path.of(tempDir.toString(), responseSubdirectory)
    val inputDirectory = Path.of(tempDir.toString(), requestSubdirectory)
    val proverClient = buildProverClient(vertx, tempDir)
    val zkParentStateRootHash = Bytes32.random()
    val blocks = randomExecutionPayloads(blocksToGenerate)
    val startBlockNumber = blocks.first().blockNumber
    val endBlockNumber = blocks.last().blockNumber
    val tracesResponse = GenerateTracesResponse(tracesFileName, tracesEngineVersion)
    val stateManagerResponse =
      GetZkEVMStateMerkleProofResponse(
        zkStateMerkleProof = merkleProofJson,
        zkParentStateRootHash = zkParentStateRootHash,
        zkStateManagerVersion = zkEvmStateManagerVersion,
        zkEndStateRootHash = Bytes32.random()
      )

    proverClient
      .getBatchProof(blocks, tracesResponse, stateManagerResponse)
      .thenApply { response ->
        testContext
          .verify {
            if (response is Err) {
              testContext.failNow(response.error.asException())
            }
            val expectedRequestPath =
              Path.of(
                inputDirectory.toString(),
                "$startBlockNumber-$endBlockNumber-getZkProof.json"
              )

            Assertions.assertThat(expectedRequestPath).exists()
            validateRequest(
              mapper,
              expectedRequestPath.toFile(),
              stateManagerResponse,
              blocks,
              bridgeLogs,
              tracesFileName,
              tracesEngineVersion,
              previousStateRoot
            )

            val expectedProverOutput = mapper.readValue(proofFile, GetProofResponse::class.java)
            Assertions.assertThat(response).isEqualTo(Ok(expectedProverOutput))
          }
          .completeNow()
      }
      .exceptionally { testContext.failNow(it) }

    val outputFileInprogress =
      "$outputDirectory/$startBlockNumber-$endBlockNumber-proof.json.inprogress"
    val outputFile = "$outputDirectory/$startBlockNumber-$endBlockNumber-proof.json"

    Thread.sleep(pollingInterval.inWholeMilliseconds * 2)
    val inprogressOutputFile = File(outputFileInprogress)
    proofFile.copyTo(inprogressOutputFile, true)
    inprogressOutputFile.renameTo(File(outputFile))
  }

  @Timeout(2, timeUnit = TimeUnit.SECONDS)
  @MethodSource("blocksAndProofs")
  @ParameterizedTest
  fun fileBasedProverClient_reusesAlreadyCreatedProofs_doesntRequestAgain(
    blocksToGenerate: Int,
    proofFile: File,
    vertx: Vertx,
    @TempDir tempDir: Path,
    testContext: VertxTestContext
  ) {
    val requestDirectory = Path.of(tempDir.toString(), requestSubdirectory)
    val responseDirectory = Path.of(tempDir.toString(), responseSubdirectory)
    val proverClient = buildProverClient(vertx, tempDir)
    val blocks = randomExecutionPayloads(blocksToGenerate)
    val startBlockNumber = blocks.first().blockNumber
    val endBlockNumber = blocks.last().blockNumber

    val tracesResponse = GenerateTracesResponse(tracesFileName, tracesEngineVersion)
    val stateManagerResponse =
      GetZkEVMStateMerkleProofResponse(
        zkStateMerkleProof = merkleProofJson,
        zkParentStateRootHash = Bytes32.random(),
        zkEndStateRootHash = Bytes32.random(),
        zkStateManagerVersion = zkEvmStateManagerVersion
      )

    proofFile.copyTo(File("$responseDirectory/$startBlockNumber-$endBlockNumber-proof.json"), true)

    proverClient
      .getBatchProof(blocks, tracesResponse, stateManagerResponse)
      .thenApply { response ->
        testContext
          .verify {
            if (response is Err) {
              testContext.failNow(response.error.asException())
            }
            Assertions.assertThat(requestDirectory).isEmptyDirectory()

            val expectedProverOutput = mapper.readValue(proofFile, GetProofResponse::class.java)
            Assertions.assertThat(response).isEqualTo(Ok(expectedProverOutput))
          }
          .completeNow()
      }
      .exceptionally { testContext.failNow(it) }
  }
}
