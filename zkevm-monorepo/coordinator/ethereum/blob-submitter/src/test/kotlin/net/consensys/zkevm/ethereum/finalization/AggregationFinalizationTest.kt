package net.consensys.zkevm.ethereum.finalization

import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import kotlinx.datetime.Clock
import kotlinx.datetime.toKotlinInstant
import net.consensys.linea.contract.LineaRollupAsyncFriendly
import net.consensys.zkevm.coordinator.clients.BlobCompressionProof
import net.consensys.zkevm.domain.BlobRecord
import net.consensys.zkevm.domain.BlobStatus
import net.consensys.zkevm.domain.BlockIntervals
import net.consensys.zkevm.domain.createProofToFinalize
import net.consensys.zkevm.domain.defaultGasPriceCaps
import net.consensys.zkevm.ethereum.gaspricing.GasPriceCapProvider
import net.consensys.zkevm.ethereum.signing.ECKeypairSigner
import net.consensys.zkevm.ethereum.signing.ECKeypairSignerAdapter
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.Mockito
import org.mockito.kotlin.any
import org.mockito.kotlin.doReturn
import org.mockito.kotlin.mock
import org.mockito.kotlin.whenever
import org.web3j.crypto.Credentials
import org.web3j.crypto.Hash
import org.web3j.crypto.Keys
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.Request
import org.web3j.protocol.core.methods.response.EthGasPrice
import org.web3j.protocol.core.methods.response.EthGetTransactionCount
import org.web3j.protocol.core.methods.response.EthSendTransaction
import org.web3j.protocol.core.methods.response.TransactionReceipt
import org.web3j.tx.gas.StaticEIP1559GasProvider
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger
import java.time.Instant
import java.util.Optional
import java.util.concurrent.TimeUnit
import kotlin.random.Random
import kotlin.time.Duration.Companion.seconds

@ExtendWith(VertxExtension::class)
class AggregationFinalizationTest {
  private val testContractAddress = "0x6d976c9b8ceee705d4fe8699b44e5eb58242f484"
  private val expectedConflationCalculatorVersion = "0.1.0"
  private val gasLimit = BigInteger.valueOf(100)
  private val maxFeePerGas = BigInteger.valueOf(10000)
  private val currentTimestamp = System.currentTimeMillis()
  private val blockNumber = 13
  private val keyPair = Keys.createEcKeyPair()
  private val signer = ECKeypairSigner(keyPair)
  private val fixedClock =
    mock<Clock> { on { now() } doReturn Instant.ofEpochMilli(currentTimestamp).toKotlinInstant() }
  private val expectedStartBlockTime = kotlinx.datetime.Instant.fromEpochMilliseconds(
    fixedClock.now().toEpochMilliseconds()
  )
  private val expectedBatchesCount = 1U
  private val gasPriceCapProvider = mock<GasPriceCapProvider> {
    on { this.getGasPriceCaps(any()) } doReturn SafeFuture.completedFuture(defaultGasPriceCaps)
  }

  @BeforeEach
  fun beforeEach() {
    // To warmup assertions otherwise first test may fail
    assertThat(true).isTrue()
    mock<LineaRollupAsyncFriendly>().currentNonce()
  }

  @Test
  @Timeout(10, timeUnit = TimeUnit.SECONDS)
  fun `aggregation finalization doesn't fail when submitting fake aggregations`() {
    val expectedTransactionReceipt = mock<TransactionReceipt>()
    whenever(expectedTransactionReceipt.isStatusOK).thenReturn(true)
    val web3jClient = createMockedWeb3jClient(expectedTransactionReceipt)

    val lineaContract = createZkEvmContractWithSimpleKeypairSigner(web3jClient)
    val aggregationFinalization = AggregationFinalizationAsCallData(lineaContract, gasPriceCapProvider)
    val blobRecords = listOf(
      createBlobToSubmit(1U, 5U),
      createBlobToSubmit(6U, 10U)
    )
    val proofToFinalize = createProofToFinalize(
      firstBlockNumber = blobRecords.first().startBlockNumber.toLong(),
      finalBlockNumber = blobRecords.last().endBlockNumber.toLong(),
      startBlockTime = expectedStartBlockTime
    )
    lineaContract.resetNonce().get()

    val actualTransactionReceipt =
      aggregationFinalization.finalizeAggregation(aggregationProof = proofToFinalize).get() as TransactionReceipt
    assertThat(actualTransactionReceipt).isEqualTo(expectedTransactionReceipt)
  }

  private fun createMockedWeb3jClient(expectedTransactionReceipt: TransactionReceipt): Web3j {
    val web3jClient = mock<Web3j>(defaultAnswer = Mockito.RETURNS_DEEP_STUBS)
    whenever(web3jClient.ethBlockNumber().send().blockNumber)
      .thenReturn(BigInteger.valueOf(blockNumber.toLong()))
    val nonceResponse = EthGetTransactionCount()
    nonceResponse.result = BigInteger.TWO.toString(16)
    whenever(
      web3jClient.ethGetTransactionCount(
        any(),
        any()
      ).sendAsync()
    )
      .thenReturn(
        SafeFuture.completedFuture(nonceResponse)
      )
    whenever(web3jClient.ethGasPrice().sendAsync()).thenAnswer {
      val gasPriceResponse = EthGasPrice()
      gasPriceResponse.result = "0x100"
      SafeFuture.completedFuture(gasPriceResponse)
    }
    val sendTransactionResponse = EthSendTransaction()
    val expectedTransactionHash =
      "0xfa41235fcc064e57ab2566d65732a25a24b36ff6edba3cdd5eb482071b435906"
    sendTransactionResponse.result = expectedTransactionHash
    whenever(web3jClient.ethSendRawTransaction(any())).thenAnswer {
      val hashToReturn = Hash.sha3(it.arguments[0] as String)
      sendTransactionResponse.result = hashToReturn
      val requestMock = mock<Request<*, EthSendTransaction>>()
      whenever(requestMock.send()).thenReturn(sendTransactionResponse)
      whenever(requestMock.sendAsync()).thenReturn(SafeFuture.completedFuture(sendTransactionResponse))
      requestMock
    }
    whenever(web3jClient.ethGetTransactionReceipt(any()).send().transactionReceipt)
      .thenReturn(Optional.of(expectedTransactionReceipt))

    return web3jClient
  }

  private fun createZkEvmContractWithSimpleKeypairSigner(web3jClient: Web3j): LineaRollupAsyncFriendly {
    val signerAdapter = ECKeypairSignerAdapter(signer, keyPair.publicKey)
    val credentials = Credentials.create(signerAdapter)
    return LineaRollupAsyncFriendly.load(
      testContractAddress,
      web3jClient,
      credentials,
      StaticEIP1559GasProvider(1, maxFeePerGas, maxFeePerGas.minus(BigInteger.ONE), gasLimit),
      emptyMap()
    )
  }

  private fun createBlobToSubmit(
    startBlockNumber: ULong,
    endBlockNumber: ULong,
    conflationCalculationVersion: String = expectedConflationCalculatorVersion,
    blobHash: ByteArray = Random.nextBytes(32),
    status: BlobStatus = BlobStatus.COMPRESSION_PROVEN,
    expectedShnarf: ByteArray = Random.nextBytes(32),
    blobCompressionProof: BlobCompressionProof? = BlobCompressionProof(
      compressedData = Random.nextBytes(128),
      conflationOrder = BlockIntervals(startBlockNumber, listOf(endBlockNumber)),
      prevShnarf = Random.nextBytes(32),
      parentStateRootHash = Random.nextBytes(32),
      finalStateRootHash = Random.nextBytes(32),
      parentDataHash = Random.nextBytes(32),
      dataHash = blobHash,
      snarkHash = Random.nextBytes(32),
      expectedX = Random.nextBytes(32),
      expectedY = Random.nextBytes(32),
      expectedShnarf = expectedShnarf,
      decompressionProof = Random.nextBytes(512),
      proverVersion = "mock-0.0.0",
      verifierID = 6789,
      eip4844Enabled = false,
      commitment = ByteArray(0),
      kzgProofContract = ByteArray(0),
      kzgProofSidecar = ByteArray(0)
    )
  ): BlobRecord {
    val startBlockTime = expectedStartBlockTime.plus(
      ((startBlockNumber - 1UL).toLong() * 12).seconds
    )
    val endBlockTime = startBlockTime.plus(
      ((endBlockNumber - startBlockNumber).toLong() * 12).seconds
    )
    return BlobRecord(
      startBlockNumber = startBlockNumber,
      endBlockNumber = endBlockNumber,
      conflationCalculatorVersion = conflationCalculationVersion,
      blobHash = blobHash,
      startBlockTime = startBlockTime,
      endBlockTime = endBlockTime,
      batchesCount = expectedBatchesCount,
      status = status,
      expectedShnarf = expectedShnarf,
      blobCompressionProof = blobCompressionProof!!
    )
  }
}
