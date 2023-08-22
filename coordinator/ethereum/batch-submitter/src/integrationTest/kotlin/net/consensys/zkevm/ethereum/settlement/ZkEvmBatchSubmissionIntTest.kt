package net.consensys.zkevm.ethereum.settlement

import io.vertx.junit5.VertxExtension
import net.consensys.linea.contract.ZkEvmV2AsyncFriendly
import net.consensys.zkevm.coordinator.clients.GetProofResponse
import net.consensys.zkevm.ethereum.coordination.conflation.Batch
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.TestInstance
import org.junit.jupiter.api.extension.ExtendWith
import org.web3j.crypto.Credentials
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.methods.response.TransactionReceipt
import org.web3j.protocol.http.HttpService
import org.web3j.tx.gas.StaticEIP1559GasProvider
import org.web3j.tx.response.PollingTransactionReceiptProcessor
import org.web3j.utils.Async
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import java.math.BigInteger
import java.time.Instant
import java.util.*

@ExtendWith(VertxExtension::class)
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class ZkEvmBatchSubmissionIntTest {
  private val gasEstimationPercentile = 0.5
  private val gasLimit = BigInteger.valueOf(10000000)
  private val gwei = BigInteger.valueOf(1000000000L)
  private val maxFeePerGas = gwei.multiply(BigInteger.valueOf(5L))
  private val currentTimestamp = System.currentTimeMillis()
  private val zkProofVerifierVersion = 1

  // WARNING: FOR LOCAL DEV ONLY - DO NOT REUSE THESE KEYS ELSEWHERE
  private val privateKey = "202454d1b4e72c41ebf58150030f649648d3cf5590297fb6718e27039ed9c86d"
  private val contractAddress = System.getProperty("ContractAddress")
  private val web3j =
    Web3j.build(HttpService("http://localhost:8445"), 1000, Async.defaultExecutorService())
  private var firstRootHash: Bytes32? = null
  private var zkEvmV2Contract: ZkEvmV2AsyncFriendly? = null
  private val validTransactionRlp =
    this::class.java.getResource("/valid-transaction.rlp")!!.readText().trim()

  @BeforeAll
  fun beforeAll() {
    // To warmup assertions
    assertThat(true).isTrue()
    zkEvmV2Contract = connectToZkevmContract()
  }

  @BeforeEach
  fun beforeEach() {
    firstRootHash =
      Bytes32.wrap(zkEvmV2Contract!!.stateRootHashes(zkEvmV2Contract!!.currentL2BlockNumber().send()).send())
  }

  @Test
  fun `batch submitter doesn't fail when submitting fake batches`() {
    val verifierIndex = 0L
    val batchSubmitter = ZkEvmBatchSubmitter(zkEvmV2Contract!!)
    val batchToSubmit = createBatchToSubmit(verifierIndex, firstRootHash!!)

    val actualTransactionReceipt =
      batchSubmitter.submitBatch(batchToSubmit).get() as TransactionReceipt
    assertThat(actualTransactionReceipt.status).isEqualTo("0x1")

    assertRootHash(
      batchToSubmit.proverResponse.blocksData.last().zkRootHash
    )
  }

  @Test
  fun `batch submitter submits multiple batches with increasing nonces`() {
    val verifierIndex = 0L
    val batchSubmitter = ZkEvmBatchSubmitter(zkEvmV2Contract!!)
    val batchToSubmit1 = createBatchToSubmit(verifierIndex, firstRootHash!!)
    val batchToSubmit2 = createBatchToSubmit(verifierIndex, batchToSubmit1.proverResponse.blocksData[0].zkRootHash)
    val batchToSubmit3 = createBatchToSubmit(verifierIndex, batchToSubmit2.proverResponse.blocksData[0].zkRootHash)

    val receipt1 = batchSubmitter.submitBatch(batchToSubmit1).get() as TransactionReceipt
    val receipt2 = batchSubmitter.submitBatch(batchToSubmit2).get() as TransactionReceipt
    val receipt3 = batchSubmitter.submitBatch(batchToSubmit3).get() as TransactionReceipt

    assertThat(receipt1.status).isEqualTo("0x1")
    assertThat(receipt2.status).isEqualTo("0x1")
    assertThat(receipt3.status).isEqualTo("0x1")

    assertRootHash(batchToSubmit3.proverResponse.blocksData.last().zkRootHash)
  }

  private fun assertRootHash(expectedRootHash: Bytes32) {
    val resultingRootHash =
      Bytes32.wrap(zkEvmV2Contract!!.stateRootHashes(zkEvmV2Contract!!.currentL2BlockNumber().send()).send())
    assertThat(resultingRootHash).isEqualTo(
      expectedRootHash
    )
  }

  private fun connectToZkevmContract(): ZkEvmV2AsyncFriendly {
    val gasProvider = StaticEIP1559GasProvider(
      web3j.ethChainId().send().chainId.toLong(),
      maxFeePerGas,
      maxFeePerGas.minus(BigInteger.valueOf(100)),
      gasLimit
    )
    val pollingTransactionReceiptProcessor = PollingTransactionReceiptProcessor(web3j, 1000, 40)
    return ZkEvmV2AsyncFriendly.load(
      contractAddress,
      web3j,
      Credentials.create(privateKey),
      pollingTransactionReceiptProcessor,
      gasProvider
    )
  }

  private fun createBatchToSubmit(verifierIndex: Long, parentRootHash: Bytes32): Batch {
    val proof = Bytes.random(100)
    val zkStateRootHash = Bytes32.random()
    val batchReceptionIndices = emptyList<UShort>()
    val l2ToL1MsgHashes = emptyList<String>()
    val fromAddresses = Bytes.EMPTY
    val blocks =
      listOf(
        GetProofResponse.BlockData(
          zkStateRootHash,
          Instant.ofEpochMilli(currentTimestamp),
          listOf(validTransactionRlp, validTransactionRlp),
          batchReceptionIndices,
          l2ToL1MsgHashes,
          fromAddresses
        )
      )
    val proverResponse =
      GetProofResponse(
        proof,
        verifierIndex,
        parentRootHash,
        blocks,
        zkProofVerifierVersion.toString()
      )
    return Batch(UInt64.ONE, UInt64.valueOf(2), proverResponse)
  }
}
