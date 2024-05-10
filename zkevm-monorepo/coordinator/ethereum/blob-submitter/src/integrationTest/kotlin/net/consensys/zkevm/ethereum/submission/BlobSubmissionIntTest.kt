package net.consensys.zkevm.ethereum.submission

import kotlinx.datetime.Clock
import kotlinx.datetime.toKotlinInstant
import net.consensys.linea.contract.LineaRollupAsyncFriendly
import net.consensys.toULong
import net.consensys.zkevm.domain.createBlobRecord
import net.consensys.zkevm.ethereum.ContractsManager
import net.consensys.zkevm.ethereum.L1AccountManager
import net.consensys.zkevm.ethereum.waitForTransactionExecution
import org.apache.tuweni.bytes.Bytes32
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.mockito.kotlin.doReturn
import org.mockito.kotlin.mock
import java.math.BigInteger
import java.time.Instant
import kotlin.time.Duration.Companion.seconds

class BlobSubmissionIntTest {
  private val currentTimestamp = System.currentTimeMillis()
  private val fixedClock =
    mock<Clock> { on { now() } doReturn Instant.ofEpochMilli(currentTimestamp).toKotlinInstant() }
  private val expectedStartBlockTime = kotlinx.datetime.Instant.fromEpochMilliseconds(
    fixedClock.now().toEpochMilliseconds()
  )
  private var currentL2BlockNumber = 0UL
  private var zeroHash: ByteArray = Bytes32.ZERO.toArray()
  private var lineaRollupContract: LineaRollupAsyncFriendly? = null
  private val web3j = L1AccountManager.web3jClient

  @BeforeEach
  fun beforeEach() {
    val deploymentResult = ContractsManager.get().deployLineaRollup().get()
    lineaRollupContract = deploymentResult.rollupOperatorClient
    currentL2BlockNumber = lineaRollupContract!!.currentL2BlockNumber().send().toULong()
  }

  @Test
  fun `blob submitter doesn't fail when submitting fake blobs`() {
    val blobSubmitter = BlobSubmitterAsCallData(lineaRollupContract!!)
    val firstParentStateRootHash =
      lineaRollupContract!!.stateRootHashes(BigInteger.valueOf(currentL2BlockNumber.toLong()))
        .send()
    val blobToSubmit = createBlobRecord(
      startBlockNumber = currentL2BlockNumber + 1UL,
      endBlockNumber = currentL2BlockNumber + 1UL,
      startBlockTime = expectedStartBlockTime,
      parentStateRootHash = firstParentStateRootHash,
      parentDataHash = zeroHash
    )
    lineaRollupContract!!.resetNonce().get()
    web3j.waitForTransactionExecution(
      blobSubmitter.submitBlob(blobToSubmit).get(),
      expectedStatus = "0x1",
      timeout = 20.seconds
    )
  }

  @Test
  fun `blob submitter submits multiple blob with increasing nonces`() {
    val blobSubmitter = BlobSubmitterAsCallData(lineaRollupContract!!)
    val firstParentStateRootHash =
      lineaRollupContract!!.stateRootHashes(BigInteger.valueOf(currentL2BlockNumber.toLong()))
        .send()
    val blobToSubmit1 = createBlobRecord(
      startBlockNumber = currentL2BlockNumber + 1UL,
      endBlockNumber = currentL2BlockNumber + 1UL,
      startBlockTime = expectedStartBlockTime,
      parentStateRootHash = firstParentStateRootHash,
      parentDataHash = zeroHash
    )
    val blobToSubmit2 = createBlobRecord(
      startBlockNumber = currentL2BlockNumber + 2UL,
      endBlockNumber = currentL2BlockNumber + 2UL,
      startBlockTime = expectedStartBlockTime,
      parentBlobRecord = blobToSubmit1
    )
    val blobToSubmit3 = createBlobRecord(
      startBlockNumber = currentL2BlockNumber + 3UL,
      endBlockNumber = currentL2BlockNumber + 3UL,
      startBlockTime = expectedStartBlockTime,
      parentBlobRecord = blobToSubmit2
    )

    lineaRollupContract!!.resetNonce().get()
    val txHash1 = blobSubmitter.submitBlob(blobToSubmit1).get()
    val txHash2 = blobSubmitter.submitBlob(blobToSubmit2).get()
    val txHash3 = blobSubmitter.submitBlob(blobToSubmit3).get()

    web3j.waitForTransactionExecution(txHash1, expectedStatus = "0x1", timeout = 20.seconds)
    web3j.waitForTransactionExecution(txHash2, expectedStatus = "0x1", timeout = 20.seconds)
    web3j.waitForTransactionExecution(txHash3, expectedStatus = "0x1", timeout = 20.seconds)
  }
}
