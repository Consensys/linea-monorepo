package net.consensys.zkevm.coordinator.clients.prover

import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.databind.node.ArrayNode
import io.vertx.core.Vertx
import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import net.consensys.zkevm.coordinator.clients.GenerateTracesResponse
import net.consensys.zkevm.coordinator.clients.GetZkEVMStateMerkleProofResponse
import net.consensys.zkevm.coordinator.clients.prover.CommonTestData.bridgeLogs
import net.consensys.zkevm.coordinator.clients.prover.CommonTestData.validTransactionRlp
import net.consensys.zkevm.coordinator.clients.prover.serialization.JsonSerialization.proofResponseMapperV1
import net.consensys.zkevm.domain.BridgeLogsData
import net.consensys.zkevm.fileio.FileMonitor
import net.consensys.zkevm.toULong
import org.apache.logging.log4j.Logger
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import org.junit.jupiter.api.io.TempDir
import org.junit.jupiter.params.ParameterizedTest
import org.junit.jupiter.params.provider.Arguments
import org.junit.jupiter.params.provider.MethodSource
import org.mockito.kotlin.mock
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import java.io.File
import java.nio.file.Files
import java.nio.file.Path
import java.util.concurrent.TimeUnit
import java.util.stream.Stream
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.seconds

@ExtendWith(VertxExtension::class)
class RequestFileWriterTest {
  private val tracesFileName = "/some/path/1-3-conflated-traces.json.gz"
  private val tracesEngineVersion = "0.2.3"
  private val zkEvmStateManagerVersion = "0.3.4"
  private val mapper = proofResponseMapperV1
  private val previousStateRoot = Bytes32.random().toHexString()
  private val fileMonitorConfig = FileMonitor.Config(
    pollingInterval = 50.milliseconds,
    timeout = 10.seconds
  )

  private val testdataPath = "../../../../testdata"
  private val merkleProofJson: ArrayNode = let {
    val testFilePath = "$testdataPath/type2state-manager/state-proof.json"
    mapper.readTree(Path.of(testFilePath).toFile()).let {
      val merkleProof = it.get("zkStateMerkleProof")
      assert(merkleProof.isArray)
      merkleProof as ArrayNode
    }
  }
  val tracesResponse = GenerateTracesResponse(tracesFileName, tracesEngineVersion)

  companion object {
    @JvmStatic
    private fun blocksToGenerate(): Stream<Arguments?>? {
      return Stream.of(Arguments.of(1), Arguments.of(5))
    }
  }

  @BeforeEach
  fun beforeEach() {
    // To warmup assertions otherwise first test may fail
    assertThat(true).isTrue()
  }

  @ParameterizedTest
  @Timeout(10, timeUnit = TimeUnit.SECONDS)
  @MethodSource("blocksToGenerate")
  fun requestFileWriter_writesValidFile(
    blocksToGenerate: Int,
    vertx: Vertx,
    testContext: VertxTestContext,
    @TempDir tempDir: Path
  ) {
    val zkParentStateRootHash = Bytes32.random()
    val blocks = randomExecutionPayloads(blocksToGenerate)
    val tracesResponse = GenerateTracesResponse(tracesFileName, tracesEngineVersion)
    val stateManagerResponse = GetZkEVMStateMerkleProofResponse(
      zkStateMerkleProof = merkleProofJson,
      zkParentStateRootHash = zkParentStateRootHash,
      zkEndStateRootHash = Bytes32.random(),
      zkStateManagerVersion = zkEvmStateManagerVersion
    )

    val fileWriter = RequestFileWriter(
      vertx,
      SimpleFileNameProvider(),
      RequestFileWriter.Config(
        requestDirectory = tempDir,
        writingInprogressSuffix = "coordinator_writting_inprogress",
        proverInprogressSuffixPattern = "\\.inprogress\\.prover.*"
      ),
      mapper = mapper,
      log = mock<Logger>(),
      fileMonitor = FileMonitor(vertx, fileMonitorConfig)
    )
    fileWriter
      .write(
        blocks.map { it to bridgeLogs },
        tracesResponse,
        stateManagerResponse,
        previousStateRoot
      )
      .thenApply { requestFilePath ->
        testContext
          .verify {
            assertThat(requestFilePath.toString())
              .isEqualTo(
                Path.of(
                  tempDir.toString(),
                  "${blocks.first().blockNumber}-${blocks.last().blockNumber}-getZkProof.json"
                )
                  .toString()
              )
            assertThat(requestFilePath).exists()
            validateRequest(
              mapper,
              requestFilePath.toFile(),
              stateManagerResponse,
              blocks,
              bridgeLogs,
              tracesFileName,
              tracesEngineVersion,
              previousStateRoot
            )
          }
          .completeNow()
      }
      .exceptionally { testContext.failNow(it) }
  }

  @Test
  fun requestFileWriter_doesNotWriteFileIfProverIsInProgress(
    vertx: Vertx,
    testContext: VertxTestContext,
    @TempDir tempDir: Path
  ) {
    val blocks = randomExecutionPayloads(2)
    val startBlockNumber = blocks.first().blockNumber
    val endBlockNumber = blocks.last().blockNumber
    val stateManagerResponse = GetZkEVMStateMerkleProofResponse(
      zkStateMerkleProof = merkleProofJson,
      zkParentStateRootHash = Bytes32.random(),
      zkEndStateRootHash = Bytes32.random(),
      zkStateManagerVersion = zkEvmStateManagerVersion
    )
    val fileNameProvider = SimpleFileNameProvider()
    val fileWriter = RequestFileWriter(
      vertx,
      fileNameProvider,
      RequestFileWriter.Config(
        requestDirectory = tempDir,
        writingInprogressSuffix = "coordinator_writting_inprogress",
        proverInprogressSuffixPattern = "\\.inprogress\\.prover.*"
      ),
      mapper = mapper,
      log = mock<Logger>(),
      fileMonitor = FileMonitor(vertx, fileMonitorConfig)
    )

    val provingInprogessFileName = fileNameProvider.getRequestFileName(
      startBlockNumber.toULong(),
      endBlockNumber.toULong()
    ) + ".inprogress.prover-1"
    Files.createFile(tempDir.resolve(provingInprogessFileName))

    fileWriter
      .write(
        blocks.map { it to bridgeLogs },
        tracesResponse,
        stateManagerResponse,
        previousStateRoot
      )
      .thenApply { requestFilePath ->
        testContext
          .verify {
            assertThat(requestFilePath).doesNotExist()
          }
          .completeNow()
      }
      .exceptionally { testContext.failNow(it) }
  }
}

fun validateRequest(
  mapper: ObjectMapper,
  requestFilePath: File,
  stateManagerResponse: GetZkEVMStateMerkleProofResponse?,
  blocks: List<ExecutionPayloadV1>,
  bridgeLogs: List<BridgeLogsData>,
  expectedTracesFileName: String,
  expectedTracesVersion: String,
  expectedPreviousStateRoot: String
) {
  val writtenRequest =
    mapper.readValue(requestFilePath, FileBasedExecutionProverClient.GetProofRequest::class.java)
  assertThat(writtenRequest).isNotNull
  assertThat(writtenRequest.conflatedExecutionTracesFile).isEqualTo(expectedTracesFileName)
  assertThat(writtenRequest.tracesEngineVersion).isEqualTo(expectedTracesVersion)
  stateManagerResponse?.run {
    assertThat(writtenRequest.zkParentStateRootHash)
      .isEqualTo(stateManagerResponse.zkParentStateRootHash.toHexString())
    assertThat(writtenRequest.type2StateManagerVersion)
      .isEqualTo(stateManagerResponse.zkStateManagerVersion)
    assertThat(writtenRequest.zkStateMerkleProof)
      .isEqualTo(stateManagerResponse.zkStateMerkleProof)
  }
  assertThat(writtenRequest.keccakParentStateRootHash).isEqualTo(expectedPreviousStateRoot)
  assertThat(writtenRequest.blocksData).hasSameSizeAs(blocks)
  writtenRequest.blocksData.zip(blocks).forEach { pair ->
    val (rlpBridgeLogData, expected) = pair
    assertThat(rlpBridgeLogData.rlp).contains(validTransactionRlp.removeRange(0, 2))
    assertThat(rlpBridgeLogData.rlp)
      .contains(expected.parentHash.toUnprefixedHexString())
    assertThat(rlpBridgeLogData.rlp)
      .contains(expected.stateRoot.toUnprefixedHexString())
    assertThat(rlpBridgeLogData.rlp)
      .contains(expected.receiptsRoot.toUnprefixedHexString())
    assertThat(rlpBridgeLogData.rlp)
      .contains(expected.logsBloom.toUnprefixedHexString())
    assertThat(rlpBridgeLogData.bridgeLogs).containsAll(bridgeLogs)
  }
}
