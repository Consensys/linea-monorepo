package net.consensys.zkevm.coordinator.clients

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import io.vertx.core.Vertx
import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import io.vertx.junit5.VertxTestContext
import net.consensys.encodeHex
import net.consensys.linea.traces.TracesCountersV1
import net.consensys.zkevm.coordinator.clients.prover.CompressionProofRequestFileNameProvider
import net.consensys.zkevm.coordinator.clients.prover.FileBasedBlobCompressionProverClient
import net.consensys.zkevm.coordinator.clients.prover.serialization.BlobCompressionProofJsonResponse
import net.consensys.zkevm.domain.BlockIntervals
import net.consensys.zkevm.domain.ConflationCalculationResult
import net.consensys.zkevm.domain.ConflationTrigger
import net.consensys.zkevm.domain.ProofIndex
import net.consensys.zkevm.ethereum.coordination.blob.ShnarfResult
import net.consensys.zkevm.fileio.FileMonitor
import net.consensys.zkevm.fileio.FileReader
import net.consensys.zkevm.fileio.FileWriter
import org.assertj.core.api.Assertions
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.TestInstance
import org.junit.jupiter.api.extension.ExtendWith
import org.junit.jupiter.api.io.TempDir
import org.mockito.Mockito
import org.mockito.kotlin.any
import org.mockito.kotlin.mock
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.nio.file.Path
import java.util.concurrent.TimeUnit
import kotlin.random.Random
import kotlin.time.Duration.Companion.milliseconds

@ExtendWith(VertxExtension::class)
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class FileBasedBlobCompressionProverClientTest {
  private val requestSubdirectory = "request"
  private val responseSubdirectory = "response"
  private val pollingInterval = 10.milliseconds
  private val mockFileWriter = mock<FileWriter>(defaultAnswer = Mockito.RETURNS_DEEP_STUBS)
  private val mockFileReader = mock<FileReader<BlobCompressionProofJsonResponse>>(
    defaultAnswer = Mockito.RETURNS_DEEP_STUBS
  )
  private val mockFileMonitor = mock<FileMonitor>(defaultAnswer = Mockito.RETURNS_DEEP_STUBS)
  private val conflations = listOf(
    ConflationCalculationResult(
      startBlockNumber = 100UL,
      endBlockNumber = 110UL,
      tracesCounters = TracesCountersV1.EMPTY_TRACES_COUNT, /* = Map<TracingModule, UInt> */
      conflationTrigger = ConflationTrigger.TRACES_LIMIT
    ),
    ConflationCalculationResult(
      startBlockNumber = 111UL,
      endBlockNumber = 122UL,
      tracesCounters = TracesCountersV1.EMPTY_TRACES_COUNT, /* = Map<TracingModule, UInt> */
      conflationTrigger = ConflationTrigger.DATA_LIMIT
    )
  )
  val parentStateRootHash = Random.nextBytes(32)
  val finalStateRootHash = Random.nextBytes(32)
  val parentDataHash = Random.nextBytes(32)
  val prevShnarf = Random.nextBytes(32)
  val expectedShnarfResult = ShnarfResult(
    dataHash = Random.nextBytes(32),
    snarkHash = Random.nextBytes(32),
    expectedX = Random.nextBytes(32),
    expectedY = Random.nextBytes(32),
    expectedShnarf = Random.nextBytes(32),
    commitment = ByteArray(0),
    kzgProofContract = ByteArray(0),
    kzgProofSideCar = ByteArray(0)
  )
  val expectedShnarfResultWithEip4844 = ShnarfResult(
    dataHash = Random.nextBytes(32),
    snarkHash = Random.nextBytes(32),
    expectedX = Random.nextBytes(32),
    expectedY = Random.nextBytes(32),
    expectedShnarf = Random.nextBytes(32),
    commitment = Random.nextBytes(48),
    kzgProofContract = Random.nextBytes(48),
    kzgProofSideCar = Random.nextBytes(48)
  )
  val compressedData = Random.nextBytes(32)
  private val blobCompressionProofResponse = BlobCompressionProofJsonResponse(
    compressedData = Random.nextBytes(128),
    conflationOrder = BlockIntervals(100UL, listOf(110UL, 122UL)),
    prevShnarf = Random.nextBytes(32),
    parentStateRootHash = Random.nextBytes(32),
    finalStateRootHash = Random.nextBytes(32),
    parentDataHash = Random.nextBytes(32),
    dataHash = Random.nextBytes(32),
    snarkHash = Random.nextBytes(32),
    expectedX = Random.nextBytes(32),
    expectedY = Random.nextBytes(32),
    expectedShnarf = Random.nextBytes(32),
    decompressionProof = Random.nextBytes(512),
    proverVersion = "mock-0.0.0",
    verifierID = 6789
  )
  private val blobCompressionProofResponseWithEip4844 = BlobCompressionProofJsonResponse(
    compressedData = Random.nextBytes(128),
    conflationOrder = BlockIntervals(100UL, listOf(110UL, 122UL)),
    prevShnarf = Random.nextBytes(32),
    parentStateRootHash = Random.nextBytes(32),
    finalStateRootHash = Random.nextBytes(32),
    parentDataHash = Random.nextBytes(32),
    dataHash = Random.nextBytes(32),
    snarkHash = Random.nextBytes(32),
    expectedX = Random.nextBytes(32),
    expectedY = Random.nextBytes(32),
    expectedShnarf = Random.nextBytes(32),
    decompressionProof = Random.nextBytes(512),
    proverVersion = "mock-0.0.0",
    verifierID = 6789,
    eip4844Enabled = true,
    commitment = expectedShnarfResultWithEip4844.commitment,
    kzgProofContract = expectedShnarfResultWithEip4844.kzgProofContract,
    kzgProofSidecar = expectedShnarfResultWithEip4844.kzgProofSideCar
  )

  private fun buildProverClient(
    vertx: Vertx,
    requestFolderPath: Path,
    responseFolderPath: Path
  ): FileBasedBlobCompressionProverClient {
    return FileBasedBlobCompressionProverClient(
      FileBasedBlobCompressionProverClient.Config(
        requestFileDirectory = requestFolderPath,
        responseFileDirectory = responseFolderPath,
        inprogressProvingSuffixPattern = ".*\\.inprogress\\.prover.*",
        inprogressRequestFileSuffix = ".coordinator_writing_inprogress",
        pollingInterval = pollingInterval,
        timeout = 100.milliseconds
      ),
      vertx,
      mockFileWriter,
      mockFileReader,
      mockFileMonitor
    )
  }

  @BeforeAll
  fun init() {
    // To warmup assertions otherwise first test may fail
    Assertions.assertThat(true).isTrue()
  }

  @Timeout(15, timeUnit = TimeUnit.SECONDS)
  @Test
  fun fileBasedBlobCompressionProverClient_returnsProofs(
    vertx: Vertx,
    @TempDir tempDir: Path,
    testContext: VertxTestContext
  ) {
    val inputDirectory = Path.of(tempDir.toString(), requestSubdirectory)
    val outputDirectory = Path.of(tempDir.toString(), responseSubdirectory)
    val proverClient = buildProverClient(vertx, inputDirectory, outputDirectory)

    whenever(
      mockFileMonitor.fileExists(
        any()
      )
    ).thenAnswer {
      SafeFuture.completedFuture(false)
    }

    whenever(
      mockFileMonitor.fileExists(
        any(),
        any()
      )
    ).thenAnswer {
      SafeFuture.completedFuture(false)
    }

    whenever(
      mockFileMonitor.monitor(
        any()
      )
    ).thenAnswer {
      SafeFuture.completedFuture<Result<Path, FileMonitor.ErrorType>>(
        Ok(
          it.getArgument(0)
        )
      )
    }

    whenever(
      mockFileWriter.write(
        any(),
        any(),
        any()
      )
    ).thenAnswer {
      SafeFuture.completedFuture<Path>(
        it.getArgument(1)
      )
    }

    whenever(
      mockFileReader.read(
        any()
      )
    ).thenAnswer {
      SafeFuture.completedFuture(
        Ok(blobCompressionProofResponse)
      )
    }

    proverClient
      .requestBlobCompressionProof(
        compressedData = compressedData,
        conflations = conflations,
        parentStateRootHash = parentStateRootHash,
        finalStateRootHash = finalStateRootHash,
        parentDataHash = parentDataHash,
        prevShnarf = prevShnarf,
        expectedShnarfResult = expectedShnarfResult,
        commitment = expectedShnarfResult.commitment,
        kzgProofContract = expectedShnarfResult.kzgProofContract,
        kzgProofSideCar = expectedShnarfResult.kzgProofSideCar
      )
      .thenApply { response ->
        testContext
          .verify {
            if (response is Err) {
              testContext.failNow(response.error.asException())
            }
            Assertions.assertThat(response).isEqualTo(Ok(blobCompressionProofResponse.toDomainObject()))
          }
          .completeNow()
      }
      .exceptionally { testContext.failNow(it) }
  }

  @Timeout(2, timeUnit = TimeUnit.SECONDS)
  @Test
  fun fileBasedBlobCompressionProverClient_reusesAlreadyCreatedProofs_doesntRequestAgain(
    vertx: Vertx,
    @TempDir tempDir: Path,
    testContext: VertxTestContext
  ) {
    val inputDirectory = Path.of(tempDir.toString(), requestSubdirectory)
    val outputDirectory = Path.of(tempDir.toString(), responseSubdirectory)
    val proverClient = buildProverClient(vertx, inputDirectory, outputDirectory)

    whenever(
      mockFileWriter.write(
        any(),
        any(),
        any()
      )
    ).thenAnswer {
      SafeFuture.failedFuture<Path>(
        Exception("Failed to write request")
      )
    }

    whenever(
      mockFileMonitor.monitor(
        any()
      )
    ).thenAnswer {
      SafeFuture.failedFuture<Result<Path, FileMonitor.ErrorType>>(
        Exception("Failed to monitor file")
      )
    }

    whenever(
      mockFileMonitor.fileExists(
        any()
      )
    ).thenAnswer {
      SafeFuture.completedFuture(true)
    }

    whenever(
      mockFileReader.read(
        any()
      )
    ).thenAnswer {
      SafeFuture.completedFuture(
        Ok(blobCompressionProofResponse)
      )
    }

    proverClient
      .requestBlobCompressionProof(
        compressedData = compressedData,
        conflations = conflations,
        parentStateRootHash = parentStateRootHash,
        finalStateRootHash = finalStateRootHash,
        parentDataHash = parentDataHash,
        prevShnarf = prevShnarf,
        expectedShnarfResult = expectedShnarfResult,
        commitment = expectedShnarfResult.commitment,
        kzgProofContract = expectedShnarfResult.kzgProofContract,
        kzgProofSideCar = expectedShnarfResult.kzgProofSideCar
      )
      .thenApply { response ->
        testContext
          .verify {
            if (response is Err) {
              testContext.failNow(response.error.asException())
            }

            Assertions.assertThat(response).isEqualTo(Ok(blobCompressionProofResponse.toDomainObject()))
          }
          .completeNow()
      }
      .exceptionally { testContext.failNow(it) }
  }

  @Timeout(15, timeUnit = TimeUnit.SECONDS)
  @Test
  fun fileBasedBlobCompressionProverClient_requestsWithEip4844(
    vertx: Vertx,
    @TempDir tempDir: Path,
    testContext: VertxTestContext
  ) {
    val inputDirectory = Path.of(tempDir.toString(), requestSubdirectory)
    val outputDirectory = Path.of(tempDir.toString(), responseSubdirectory)
    val proverClient = buildProverClient(vertx, inputDirectory, outputDirectory)

    whenever(
      mockFileMonitor.fileExists(
        any()
      )
    ).thenAnswer {
      SafeFuture.completedFuture(false)
    }

    whenever(
      mockFileMonitor.fileExists(
        any(),
        any()
      )
    ).thenAnswer {
      SafeFuture.completedFuture(false)
    }

    whenever(
      mockFileMonitor.monitor(
        any()
      )
    ).thenAnswer {
      SafeFuture.completedFuture<Result<Path, FileMonitor.ErrorType>>(
        Ok(
          it.getArgument(0)
        )
      )
    }

    whenever(
      mockFileWriter.write(
        any(),
        any(),
        any()
      )
    ).thenAnswer {
      SafeFuture.completedFuture<Path>(
        it.getArgument(1)
      )
    }

    whenever(
      mockFileReader.read(
        any()
      )
    ).thenAnswer {
      SafeFuture.completedFuture(
        Ok(blobCompressionProofResponseWithEip4844)
      )
    }

    proverClient
      .requestBlobCompressionProof(
        compressedData = compressedData,
        conflations = conflations,
        parentStateRootHash = parentStateRootHash,
        finalStateRootHash = finalStateRootHash,
        parentDataHash = parentDataHash,
        prevShnarf = prevShnarf,
        expectedShnarfResult = expectedShnarfResultWithEip4844,
        commitment = expectedShnarfResultWithEip4844.commitment,
        kzgProofContract = expectedShnarfResultWithEip4844.kzgProofContract,
        kzgProofSideCar = expectedShnarfResultWithEip4844.kzgProofSideCar
      )
      .thenApply { response ->
        testContext
          .verify {
            if (response is Err) {
              testContext.failNow(response.error.asException())
            }

            Assertions.assertThat(response).isEqualTo(Ok(blobCompressionProofResponseWithEip4844.toDomainObject()))
          }
          .completeNow()
      }
      .exceptionally { testContext.failNow(it) }
  }

  @Test
  fun `test request filename`() {
    val requestHash = expectedShnarfResult.expectedShnarf
    val requestHashString = requestHash.encodeHex(prefix = false)
    val requestFileName = CompressionProofRequestFileNameProvider.getFileName(
      ProofIndex(
        startBlockNumber = 1uL,
        endBlockNumber = 11uL,
        hash = requestHash
      )
    )
    Assertions.assertThat(requestFileName).isEqualTo(
      "1-11-bcv0.0-ccv0.0-$requestHashString-getZkBlobCompressionProof.json"
    )
  }
}
