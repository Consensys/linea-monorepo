package net.consensys.zkevm.ethereum.settlement

import io.vertx.junit5.Timeout
import io.vertx.junit5.VertxExtension
import net.consensys.linea.contract.ZkEvmV2AsyncFriendly
import net.consensys.zkevm.coordinator.clients.GetProofResponse
import net.consensys.zkevm.ethereum.coordination.conflation.Batch
import net.consensys.zkevm.ethereum.crypto.ECKeypairSignerAdapter
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.TestInstance
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.ArgumentMatchers.eq
import org.mockito.Mockito
import org.mockito.kotlin.any
import org.mockito.kotlin.mock
import org.mockito.kotlin.whenever
import org.web3j.crypto.Credentials
import org.web3j.crypto.Hash
import org.web3j.crypto.Keys
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameter
import org.web3j.protocol.core.Request
import org.web3j.protocol.core.methods.response.EthFeeHistory
import org.web3j.protocol.core.methods.response.EthFeeHistory.FeeHistory
import org.web3j.protocol.core.methods.response.EthGasPrice
import org.web3j.protocol.core.methods.response.EthGetTransactionCount
import org.web3j.protocol.core.methods.response.EthSendTransaction
import org.web3j.protocol.core.methods.response.TransactionReceipt
import org.web3j.tx.gas.StaticEIP1559GasProvider
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import java.math.BigInteger
import java.time.Instant
import java.util.*
import java.util.concurrent.TimeUnit

@ExtendWith(VertxExtension::class)
@TestInstance(TestInstance.Lifecycle.PER_CLASS)
class ZkEvmBatchSubmissionTest {
  private val testContractAddress = "0x6d976c9b8ceee705d4fe8699b44e5eb58242f484"
  private val gasEstimationPercentile = 0.5
  private val gasLimit = BigInteger.valueOf(100)
  private val maxFeePerGas = BigInteger.valueOf(10000)
  private val currentTimestamp = System.currentTimeMillis()
  private val zkProofVerifierVersion = 1
  private val blockNumber = 13
  private val keyPair = Keys.createEcKeyPair()
  private val signer = ECKeypairSigner(keyPair)

  @BeforeAll
  fun warmup() {
    // To warmup assertions
    assertThat(true).isTrue()
  }

  @Test
  @Timeout(10, timeUnit = TimeUnit.SECONDS)
  fun `batch submitter doesn't fail when submitting fake batches`() {
    val expectedTransactionReceipt = mock<TransactionReceipt>()
    whenever(expectedTransactionReceipt.isStatusOK).thenReturn(true)
    val web3jClient = createMockedWeb3jClient(expectedTransactionReceipt)

    val zkEvmV2Contract = createZkEvmContractWithSimpleKeypairSigner(web3jClient)
    val verifierIndex = 0L
    val batchSubmitter = ZkEvmBatchSubmitter(zkEvmV2Contract)
    val batchToSubmit = createBatchToSubmit(verifierIndex)
    zkEvmV2Contract.resetNonce().get()

    val actualTransactionReceipt =
      batchSubmitter.submitBatch(batchToSubmit).get() as TransactionReceipt
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
    whenever(
      web3jClient
        .ethFeeHistory(
          eq(1),
          eq(DefaultBlockParameter.valueOf("pending")),
          eq(listOf(gasEstimationPercentile))
        )
        .sendAsync()
    )
      .thenAnswer {
        val feeHistoryResponse = EthFeeHistory()
        val feeHistory = FeeHistory()
        feeHistory.setReward(mutableListOf(mutableListOf("0x1000")))
        feeHistory.setBaseFeePerGas(mutableListOf("0x100"))
        feeHistoryResponse.result = feeHistory
        SafeFuture.completedFuture(feeHistoryResponse)
      }
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
      requestMock
    }
    whenever(web3jClient.ethGetTransactionReceipt(any()).send().transactionReceipt)
      .thenReturn(Optional.of(expectedTransactionReceipt))

    return web3jClient
  }

  private fun createZkEvmContractWithSimpleKeypairSigner(web3jClient: Web3j): ZkEvmV2AsyncFriendly {
    val signerAdapter = ECKeypairSignerAdapter(signer, keyPair.publicKey)
    val credentials = Credentials.create(signerAdapter)
    return ZkEvmV2AsyncFriendly.load(
      testContractAddress,
      web3jClient,
      credentials,
      StaticEIP1559GasProvider(1, maxFeePerGas, maxFeePerGas.minus(BigInteger.ONE), gasLimit)
    )
  }

  private fun createBatchToSubmit(verifierIndex: Long): Batch {
    val proof = Bytes.random(100)
    val parentStateRootHash = Bytes32.random()
    val zkStateRootHash = Bytes32.random()
    val batchReceptionIndices: List<UShort> = listOf(0u)
    val l2ToL1MsgHashes = listOf("0x33f9b98e9afb2b613a3c235dc7d8ea78b4d30b3a1eb249568884794de921cae7")
    val fromAddresses = Bytes.fromHexString("0x03dfa322A95039BB679771346Ee2dBfEa0e2B773")
    val blocks =
      listOf(
        GetProofResponse.BlockData(
          zkStateRootHash,
          Instant.ofEpochMilli(currentTimestamp),
          listOf("0x01", "0x02"),
          batchReceptionIndices,
          l2ToL1MsgHashes,
          fromAddresses
        )
      )
    val proverResponse =
      GetProofResponse(
        proof,
        verifierIndex,
        parentStateRootHash,
        blocks,
        zkProofVerifierVersion.toString()
      )
    return Batch(UInt64.ONE, UInt64.valueOf(2), proverResponse)
  }
}
