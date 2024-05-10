package net.consensys.zkevm.ethereum.finalization

import kotlinx.datetime.Clock
import kotlinx.datetime.toKotlinInstant
import net.consensys.decodeHex
import net.consensys.linea.contract.LineaRollupAsyncFriendly
import net.consensys.toULong
import net.consensys.zkevm.domain.createBlobRecord
import net.consensys.zkevm.domain.createProofToFinalize
import net.consensys.zkevm.domain.createProofToFinalizeFromBlobs
import net.consensys.zkevm.domain.defaultGasPriceCaps
import net.consensys.zkevm.ethereum.ContractsManager
import net.consensys.zkevm.ethereum.L1AccountManager
import net.consensys.zkevm.ethereum.gaspricing.GasPriceCapProvider
import net.consensys.zkevm.ethereum.submission.BlobSubmitterAsCallData
import net.consensys.zkevm.ethereum.waitForTransactionExecution
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.mockito.kotlin.any
import org.mockito.kotlin.doReturn
import org.mockito.kotlin.mock
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.methods.response.TransactionReceipt
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.time.Instant
import kotlin.time.Duration.Companion.seconds

class AggregationFinalizationIntTest {
  private val currentTimestamp = System.currentTimeMillis()
  private val fixedClock = mock<Clock> {
    on { now() } doReturn Instant.ofEpochMilli(currentTimestamp).toKotlinInstant()
  }
  private val expectedStartBlockTime = fixedClock.now()
  private var startingParentStateRootHash: ByteArray =
    "0x113e9977cebf08f3b271d121342540fce95530c206c2a878ae957925e1d0fc02".decodeHex()
  private var currentL2BlockNumber = 0UL
  private var emptyRootHash: ByteArray = Bytes32.ZERO.toArray()
  private lateinit var lineaRollupContract: LineaRollupAsyncFriendly
  private lateinit var web3j: Web3j
  private val gasPriceCapProvider = mock<GasPriceCapProvider> {
    on { this.getGasPriceCaps(any()) } doReturn SafeFuture.completedFuture(defaultGasPriceCaps)
  }

  @BeforeEach
  fun beforeEach() {
    val deploymentResult = ContractsManager.get().deployLineaRollup().get()
    lineaRollupContract = deploymentResult.rollupOperatorClient
    currentL2BlockNumber = lineaRollupContract.currentL2BlockNumber().send().toULong()
    startingParentStateRootHash =
      lineaRollupContract.stateRootHashes(BigInteger.valueOf(currentL2BlockNumber.toLong())).send()
    web3j = L1AccountManager.web3jClient
  }

  @Test
  fun `aggregation finalization doesn't fail when finalizing multiple proofs with increasing nonces`() {
    val blobSubmitter = BlobSubmitterAsCallData(lineaRollupContract)
    val firstParentStateRootHash =
      lineaRollupContract.stateRootHashes(BigInteger.valueOf(currentL2BlockNumber.toLong()))
        .send()

    val blobToSubmit1 = createBlobRecord(
      startBlockNumber = currentL2BlockNumber + 1UL,
      endBlockNumber = currentL2BlockNumber + 1UL,
      startBlockTime = expectedStartBlockTime,
      parentStateRootHash = firstParentStateRootHash,
      parentDataHash = emptyRootHash
    )
    val blobToSubmit2 = createBlobRecord(
      startBlockNumber = currentL2BlockNumber + 2UL,
      endBlockNumber = currentL2BlockNumber + 2UL,
      startBlockTime = expectedStartBlockTime,
      parentStateRootHash = blobToSubmit1.blobCompressionProof!!.finalStateRootHash,
      parentDataHash = blobToSubmit1.blobCompressionProof!!.dataHash
    )
    val blobToSubmit3 = createBlobRecord(
      startBlockNumber = currentL2BlockNumber + 3UL,
      endBlockNumber = currentL2BlockNumber + 3UL,
      startBlockTime = expectedStartBlockTime,
      parentStateRootHash = blobToSubmit2.blobCompressionProof!!.finalStateRootHash,
      parentDataHash = blobToSubmit2.blobCompressionProof!!.dataHash
    )
    val blobToSubmit4 = createBlobRecord(
      startBlockNumber = currentL2BlockNumber + 4UL,
      endBlockNumber = currentL2BlockNumber + 4UL,
      startBlockTime = expectedStartBlockTime,
      parentStateRootHash = blobToSubmit3.blobCompressionProof!!.finalStateRootHash,
      parentDataHash = blobToSubmit3.blobCompressionProof!!.dataHash
    )

    lineaRollupContract.resetNonce().get()
    blobSubmitter.submitBlob(blobToSubmit1).get()
    blobSubmitter.submitBlob(blobToSubmit2).get()
    blobSubmitter.submitBlob(blobToSubmit3).get()
    web3j.waitForTransactionExecution(blobSubmitter.submitBlob(blobToSubmit4).get())

    val aggregationFinalization = AggregationFinalizationAsCallData(
      lineaRollupContract,
      gasPriceCapProvider
    )

    var lastFinalizedBlockTime = kotlinx.datetime.Instant.fromEpochSeconds(
      lineaRollupContract.currentTimestamp().send().toLong()
    )
    var endBlockTime = lastFinalizedBlockTime.plus(
      (blobToSubmit3.endBlockNumber.toLong()).seconds
    )

    val proofToFinalize1 = createProofToFinalize(
      firstBlockNumber = blobToSubmit3.startBlockNumber.toLong(),
      finalBlockNumber = blobToSubmit3.endBlockNumber.toLong(),
      dataParentHash = emptyRootHash,
      parentStateRootHash = startingParentStateRootHash,
      dataHashes = listOf(
        blobToSubmit1.blobCompressionProof!!.dataHash,
        blobToSubmit2.blobCompressionProof!!.dataHash,
        blobToSubmit3.blobCompressionProof!!.dataHash
      ),
      startBlockTime = expectedStartBlockTime,
      parentAggregationLastBlockTimestamp = lastFinalizedBlockTime,
      finalTimestamp = endBlockTime
    )

    lastFinalizedBlockTime = endBlockTime
    endBlockTime = lastFinalizedBlockTime.plus(
      (blobToSubmit3.endBlockNumber.toLong()).seconds
    )

    val proofToFinalize2 = createProofToFinalize(
      firstBlockNumber = blobToSubmit4.startBlockNumber.toLong(),
      finalBlockNumber = blobToSubmit4.endBlockNumber.toLong(),
      dataParentHash = blobToSubmit3.blobCompressionProof!!.dataHash,
      parentStateRootHash = blobToSubmit3.blobCompressionProof!!.finalStateRootHash,
      dataHashes = listOf(
        blobToSubmit4.blobCompressionProof!!.dataHash
      ),
      startBlockTime = expectedStartBlockTime,
      parentAggregationLastBlockTimestamp = lastFinalizedBlockTime,
      finalTimestamp = endBlockTime
    )

    lineaRollupContract.resetNonce().get()
    val actualTransactionReceipt1 =
      aggregationFinalization.finalizeAggregation(
        aggregationProof = proofToFinalize1
      )
        .get() as TransactionReceipt
    lineaRollupContract.resetNonce().get()
    val actualTransactionReceipt2 =
      aggregationFinalization.finalizeAggregation(
        aggregationProof = proofToFinalize2
      )
        .get() as TransactionReceipt

    assertThat(actualTransactionReceipt1.status).isEqualTo("0x1")
    assertThat(actualTransactionReceipt2.status).isEqualTo("0x1")

    val actualCurrentL2BlockNumber = lineaRollupContract.currentL2BlockNumber().send()
    assertThat(actualCurrentL2BlockNumber.toLong()).isEqualTo(proofToFinalize2.finalBlockNumber)
  }

  @Test
  fun `aggregation finalization with l1 messaging blocks offsets`() {
    val blobSubmitter = BlobSubmitterAsCallData(lineaRollupContract)
    val firstParentStateRootHash =
      lineaRollupContract.stateRootHashes(BigInteger.valueOf(currentL2BlockNumber.toLong()))
        .send()
    val blobToSubmit1 = createBlobRecord(
      startBlockNumber = currentL2BlockNumber + 1UL,
      endBlockNumber = currentL2BlockNumber + 3UL,
      startBlockTime = expectedStartBlockTime,
      parentStateRootHash = firstParentStateRootHash,
      parentDataHash = emptyRootHash
    )
    lineaRollupContract.resetNonce().get()
    web3j.waitForTransactionExecution(blobSubmitter.submitBlob(blobToSubmit1).get())

    val aggregationFinalization = AggregationFinalizationAsCallData(
      lineaRollupContract,
      gasPriceCapProvider
    )

    val lastFinalizedBlockTime = kotlinx.datetime.Instant.fromEpochSeconds(
      lineaRollupContract.currentTimestamp().send().toLong()
    )
    val aggregation1 = createProofToFinalizeFromBlobs(
      blobRecords = listOf(blobToSubmit1),
      lastFinalizedBlockTime = lastFinalizedBlockTime,
      dataParentHash = emptyRootHash,
      parentStateRootHash = startingParentStateRootHash
    )

    lineaRollupContract.resetNonce().get()
    val actualTransactionReceipt1 =
      aggregationFinalization.finalizeAggregation(aggregationProof = aggregation1).get() as TransactionReceipt

    assertThat(actualTransactionReceipt1.status).isEqualTo("0x1")

    val actualCurrentL2BlockNumber = lineaRollupContract.currentL2BlockNumber().send()
    assertThat(actualCurrentL2BlockNumber.toLong()).isEqualTo(aggregation1.finalBlockNumber)
  }
}
