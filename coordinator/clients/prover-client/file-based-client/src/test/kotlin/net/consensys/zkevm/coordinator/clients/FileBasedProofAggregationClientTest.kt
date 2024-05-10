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
import net.consensys.zkevm.coordinator.clients.prover.AggregationProofResponseFileNameProviderV2
import net.consensys.zkevm.coordinator.clients.prover.CompressionProofFileNameProvider
import net.consensys.zkevm.coordinator.clients.prover.ExecutionProofResponseFileNameProvider
import net.consensys.zkevm.coordinator.clients.prover.FileBasedProofAggregationClient
import net.consensys.zkevm.coordinator.clients.prover.serialization.JsonSerialization.proofResponseMapperV1
import net.consensys.zkevm.coordinator.clients.prover.serialization.ProofToFinalizeJsonResponse
import net.consensys.zkevm.domain.BlockIntervals
import net.consensys.zkevm.domain.ExecutionProofVersions
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
import java.security.MessageDigest
import java.util.*
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
      inProgressRequestFileSuffix = "inp",
      proverInProgressSuffixPattern = "\\.prover-inprogress"
    )

    aggregation = ProofsToAggregate(
      compressionProofs = BlockIntervals(11u, listOf(23u, 27u)),
      executionProofs = BlockIntervals(11u, listOf(23u, 27u)),
      executionVersion = listOf(
        ExecutionProofVersions("ccv2", "epv2"),
        ExecutionProofVersions("ccv2", "epv2")
      ),
      parentAggregationLastBlockTimestamp = Instant.parse("2024-01-21T16:08:22Z"),
      parentAggregationLastL1RollingHashMessageNumber = 1u.toULong(),
      parentAggregationLastL1RollingHash = ByteArray(32)
    )

    responseFileName = AggregationProofResponseFileNameProviderV2.getResponseFileName(11u, 27u)
    fileWriter = FileWriter(vertx, proofResponseMapperV1)
    fileMonitor = FileMonitor(
      vertx,
      FileMonitor.Config(50.toDuration(DurationUnit.MILLISECONDS), 2.toDuration(DurationUnit.SECONDS))
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
  fun test_getAggregatedProof_proofExists(vertx: Vertx) {
    val responseFile = config.responseFileDirectory.resolve(responseFileName).toFile()
    proofResponseMapperV1.writeValue(responseFile, ProofToFinalizeJsonResponse.fromDomainObject(sampleResponse))

    val fileBasedProofAggregationClient = FileBasedProofAggregationClient(
      vertx,
      config
    )

    val result = fileBasedProofAggregationClient.getAggregatedProof(aggregation).get()
    Assertions.assertEquals(
      Ok(sampleResponse),
      result
    )
  }

  @Test
  fun test_getAggregatedProof_proofCreated(vertx: Vertx) {
    fileMonitor.fileExists(config.requestFileDirectory, ".*json").thenApply {
      val responseFile = config.responseFileDirectory.resolve(responseFileName).toFile()
      proofResponseMapperV1.writeValue(responseFile, ProofToFinalizeJsonResponse.fromDomainObject(sampleResponse))
    }

    val fileBasedProofAggregationClient = FileBasedProofAggregationClient(
      vertx,
      config
    )

    val result = fileBasedProofAggregationClient.getAggregatedProof(aggregation).get()
    Assertions.assertEquals(
      Ok(sampleResponse),
      result
    )
  }

  @Test
  fun test_getAggregatedProof_proofNotCreated(vertx: Vertx) {
    val fileBasedProofAggregationClient = FileBasedProofAggregationClient(
      vertx,
      config
    )
    val error = Err(ErrorResponse(ProverErrorType.ResponseNotFound, ""))
    val result = fileBasedProofAggregationClient.getAggregatedProof(aggregation).get()
    Assertions.assertEquals(error, result)
  }

  @Test
  fun test_getAggregatedProof_requestWriteInProgress(vertx: Vertx) {
    val fileBasedProofAggregationClient = FileBasedProofAggregationClient(
      vertx,
      config
    )

    val requestFileName = fileBasedProofAggregationClient.getZkAggregatedProofRequestFileName(
      fileBasedProofAggregationClient.buildRequest(aggregation),
      aggregation
    )
    val requestFilePath = config.requestFileDirectory.resolve(requestFileName)
    val inProgressRequestFilePath = Path.of(
      requestFilePath.toAbsolutePath().toString() + ".${config.inProgressRequestFileSuffix}"
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
    val fileBasedProofAggregationClient = FileBasedProofAggregationClient(
      vertx,
      config
    )

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
  fun test_getRequestFileName(vertx: Vertx) {
    val proofsToAggregate = ProofsToAggregate(
      compressionProofs = BlockIntervals(startingBlockNumber = 11u, upperBoundaries = listOf(20u, 27u)),
      executionProofs = BlockIntervals(startingBlockNumber = 11u, upperBoundaries = listOf(20u, 27u)),
      executionVersion = listOf(
        ExecutionProofVersions("ccv2", "ev2"),
        ExecutionProofVersions("ccv2", "ev2")
      ),
      parentAggregationLastBlockTimestamp = Instant.parse("2024-01-21T16:08:22Z"),
      parentAggregationLastL1RollingHashMessageNumber = 1u.toULong(),
      parentAggregationLastL1RollingHash = ByteArray(32)
    )
    val fileBasedProofAggregationClient = FileBasedProofAggregationClient(
      vertx,
      config
    )

    val compressionProofFileNameProvider = CompressionProofFileNameProvider
    val compressionProofs = proofsToAggregate.compressionProofs
      .toIntervalList()
      .map {
        compressionProofFileNameProvider.getResponseFileName(
          it.startBlockNumber,
          it.endBlockNumber
        )
      }

    val executionProofFileNameProvider = ExecutionProofResponseFileNameProvider
    val executionProofs = proofsToAggregate.executionProofs
      .toIntervalList()
      .zip(proofsToAggregate.executionVersion)
      .map {
        val blockInterval = it.first
        executionProofFileNameProvider.getResponseFileName(
          blockInterval.startBlockNumber,
          blockInterval.endBlockNumber
        )
      }

    val contentBytes = (compressionProofs + executionProofs).joinToString().toByteArray()

    val contentHash = HexFormat.of().formatHex(
      MessageDigest.getInstance("SHA-256").digest(contentBytes)
    )
    val request = FileBasedProofAggregationClient.Request(
      compressionProofs = compressionProofs,
      executionProofs = executionProofs,
      parentAggregationLastBlockTimestamp = proofsToAggregate.parentAggregationLastBlockTimestamp.epochSeconds,
      parentAggregationLastL1RollingHashMessageNumber =
      proofsToAggregate.parentAggregationLastL1RollingHashMessageNumber.toLong(),
      parentAggregationLastL1RollingHash = proofsToAggregate.parentAggregationLastL1RollingHash.encodeHex()
    )

    Assertions.assertEquals(
      "11-27-$contentHash-getZkAggregatedProof.json",
      fileBasedProofAggregationClient.getZkAggregatedProofRequestFileName(request, proofsToAggregate)
    )
  }
}
