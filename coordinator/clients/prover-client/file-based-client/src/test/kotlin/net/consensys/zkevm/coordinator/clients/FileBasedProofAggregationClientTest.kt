package net.consensys.zkevm.coordinator.clients

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import io.vertx.core.Vertx
import io.vertx.junit5.VertxExtension
import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.encodeHex
import net.consensys.linea.async.get
import net.consensys.linea.errors.ErrorResponse
import net.consensys.trimToSecondPrecision
import net.consensys.zkevm.coordinator.clients.prover.AggregationProofFileNameProvider
import net.consensys.zkevm.coordinator.clients.prover.FileBasedProofAggregationClient
import net.consensys.zkevm.coordinator.clients.prover.serialization.JsonSerialization.proofResponseMapperV1
import net.consensys.zkevm.coordinator.clients.prover.serialization.ProofToFinalizeJsonResponse
import net.consensys.zkevm.domain.BlockIntervals
import net.consensys.zkevm.domain.ProofIndex
import net.consensys.zkevm.domain.ProofToFinalize
import net.consensys.zkevm.domain.ProofsToAggregate
import net.consensys.zkevm.fileio.FileMonitor
import net.consensys.zkevm.fileio.FileWriter
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.Assertions
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import java.nio.file.Files
import java.nio.file.Path
import kotlin.random.Random
import kotlin.time.DurationUnit
import kotlin.time.toDuration

@ExtendWith(VertxExtension::class)
class FileBasedProofAggregationClientTest {
  lateinit var tmpRequestDirectory: Path
  lateinit var tmpResponseDirectory: Path
  lateinit var config: FileBasedProofAggregationClient.Config
  lateinit var aggregation: ProofsToAggregate
  lateinit var responseFileName: String
  lateinit var fileWriter: FileWriter
  lateinit var fileMonitor: FileMonitor
  lateinit var fileBasedProofAggregationClient: FileBasedProofAggregationClient

  val sampleResponse = ProofToFinalize(
    aggregatedProof = "mock_aggregatedProof".toByteArray(),
    aggregatedVerifierIndex = 1,
    aggregatedProofPublicInput = "mock_aggregatedProofPublicInput".toByteArray(),
    dataHashes = listOf("mock_dataHashes_1".toByteArray()),
    dataParentHash = "mock_dataParentHash".toByteArray(),
    parentStateRootHash = "mock_parentStateRootHash".toByteArray(),
    parentAggregationLastBlockTimestamp = Clock.System.now().trimToSecondPrecision(),
    finalTimestamp = Clock.System.now().trimToSecondPrecision(),
    firstBlockNumber = 1,
    finalBlockNumber = 23,
    l1RollingHash = "mock_l1RollingHash".toByteArray(),
    l1RollingHashMessageNumber = 4,
    l2MerkleRoots = listOf("mock_l2MerkleRoots".toByteArray()),
    l2MerkleTreesDepth = 5,
    l2MessagingBlocksOffsets = "mock_l2MessagingBlocksOffsets".toByteArray()
  )

  private val requestHash = "request-hash".toByteArray()

  @BeforeEach
  fun setup(vertx: Vertx) {
    tmpRequestDirectory =
      Files.createTempDirectory(FileBasedProofAggregationClientTest::class.toString() + "-request")
    tmpResponseDirectory =
      Files.createTempDirectory(FileBasedProofAggregationClientTest::class.toString() + "-response")
    config = FileBasedProofAggregationClient.Config(
      requestFileDirectory = tmpRequestDirectory,
      responseFileDirectory = tmpResponseDirectory,
      responseFilePollingInterval = 200.toDuration(DurationUnit.MILLISECONDS),
      responseFileMonitorTimeout = 2.toDuration(DurationUnit.SECONDS),
      inprogressRequestFileSuffix = "inp",
      proverInProgressSuffixPattern = "\\.prover-inprogress"
    )

    aggregation = ProofsToAggregate(
      compressionProofIndexes = listOf(
        ProofIndex(11u, 23u, Random.nextBytes(32)),
        ProofIndex(24u, 27u, Random.nextBytes(32))
      ),
      executionProofs = BlockIntervals(11u, listOf(23u, 27u)),
      parentAggregationLastBlockTimestamp = Instant.parse("2024-01-21T16:08:22Z"),
      parentAggregationLastL1RollingHashMessageNumber = 1u.toULong(),
      parentAggregationLastL1RollingHash = ByteArray(32)
    )

    responseFileName = AggregationProofFileNameProvider.getFileName(
      ProofIndex(
        startBlockNumber = 11u,
        endBlockNumber = 27u,
        hash = requestHash
      )
    )
    fileWriter = FileWriter(vertx, proofResponseMapperV1)
    fileMonitor = FileMonitor(
      vertx,
      FileMonitor.Config(50.toDuration(DurationUnit.MILLISECONDS), 2.toDuration(DurationUnit.SECONDS))
    )

    fileBasedProofAggregationClient = FileBasedProofAggregationClient(
      vertx = vertx,
      config = config,
      hashFunction = { _ -> requestHash }
    )
  }

  @AfterEach
  fun tearDown(vertx: Vertx) {
    vertx.fileSystem().deleteRecursiveBlocking(tmpRequestDirectory.toString(), true)
    vertx.fileSystem().deleteRecursiveBlocking(tmpResponseDirectory.toString(), true)
    val vertxStopFuture = vertx.close()
    vertxStopFuture.get()
  }

  @Test
  fun test_getAggregatedProof_proofExists() {
    val responseFile = config.responseFileDirectory.resolve(responseFileName).toFile()
    proofResponseMapperV1.writeValue(responseFile, ProofToFinalizeJsonResponse.fromDomainObject(sampleResponse))

    val result = fileBasedProofAggregationClient.getAggregatedProof(aggregation).get()
    Assertions.assertEquals(
      Ok(sampleResponse),
      result
    )
  }

  @Test
  fun test_getAggregatedProof_proofCreated() {
    fileMonitor.fileExists(config.requestFileDirectory, ".*json").thenApply {
      val responseFile = config.responseFileDirectory.resolve(responseFileName).toFile()
      proofResponseMapperV1.writeValue(responseFile, ProofToFinalizeJsonResponse.fromDomainObject(sampleResponse))
    }

    val result = fileBasedProofAggregationClient.getAggregatedProof(aggregation).get()
    Assertions.assertEquals(
      Ok(sampleResponse),
      result
    )
  }

  @Test
  fun test_getAggregatedProof_proofNotCreated() {
    val error = Err(ErrorResponse(ProverErrorType.ResponseNotFound, ""))
    val result = fileBasedProofAggregationClient.getAggregatedProof(aggregation).get()
    Assertions.assertEquals(error, result)
  }

  @Test
  fun test_getAggregatedProof_requestWriteInProgress(vertx: Vertx) {
    val requestFileName = fileBasedProofAggregationClient.getZkAggregatedProofRequestFileName(
      fileBasedProofAggregationClient.buildRequest(aggregation),
      aggregation
    )
    val requestFilePath = config.requestFileDirectory.resolve(requestFileName)
    val inProgressRequestFilePath = Path.of(
      requestFilePath.toAbsolutePath().toString() + ".${config.inprogressRequestFileSuffix}"
    )
    inProgressRequestFilePath.toFile().createNewFile()

    vertx.setTimer(500L) {
      val responseFile = config.responseFileDirectory.resolve(responseFileName).toFile()
      proofResponseMapperV1.writeValue(responseFile, ProofToFinalizeJsonResponse.fromDomainObject(sampleResponse))
    }

    val result = fileBasedProofAggregationClient.getAggregatedProof(aggregation).get()
    Assertions.assertEquals(
      Ok(sampleResponse),
      result
    )
  }

  @Test
  fun test_getAggregatedProof_provingInProgress(vertx: Vertx) {
    val requestFileName = fileBasedProofAggregationClient.getZkAggregatedProofRequestFileName(
      fileBasedProofAggregationClient.buildRequest(aggregation),
      aggregation
    )
    val requestFilePath = config.requestFileDirectory.resolve(requestFileName)
    val provingInProgressFilePath = Path.of(
      requestFilePath.toAbsolutePath().toString() + ".prover-inprogress"
    )
    provingInProgressFilePath.toFile().createNewFile()

    vertx.setTimer(500L) {
      val responseFile = config.responseFileDirectory.resolve(responseFileName).toFile()
      proofResponseMapperV1.writeValue(responseFile, ProofToFinalizeJsonResponse.fromDomainObject(sampleResponse))
    }

    val result = fileBasedProofAggregationClient.getAggregatedProof(aggregation).get()
    Assertions.assertEquals(
      Ok(sampleResponse),
      result
    )
  }

  @Test
  fun test_getRequestFileName() {
    val proofsToAggregate = ProofsToAggregate(
      compressionProofIndexes = listOf(
        ProofIndex(11u, 20u, Random.nextBytes(32)),
        ProofIndex(21u, 27u, Random.nextBytes(32))
      ),
      executionProofs = BlockIntervals(startingBlockNumber = 11u, upperBoundaries = listOf(20u, 27u)),
      parentAggregationLastBlockTimestamp = Instant.parse("2024-01-21T16:08:22Z"),
      parentAggregationLastL1RollingHashMessageNumber = 1u.toULong(),
      parentAggregationLastL1RollingHash = ByteArray(32)
    )

    val blockInterval = proofsToAggregate.getStartEndBlockInterval()

    Assertions.assertEquals(
      "11-27-${requestHash.encodeHex(prefix = false)}-getZkAggregatedProof.json",
      AggregationProofFileNameProvider.getFileName(
        ProofIndex(
          startBlockNumber = blockInterval.startBlockNumber,
          endBlockNumber = blockInterval.endBlockNumber,
          hash = requestHash
        )
      )
    )
  }
}
