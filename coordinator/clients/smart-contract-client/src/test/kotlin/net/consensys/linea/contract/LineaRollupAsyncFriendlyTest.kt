package net.consensys.linea.contract

import net.consensys.linea.Constants
import net.consensys.linea.contract.LineaRollup.SupportingSubmissionData
import net.consensys.linea.web3j.EIP1559GasFees
import net.consensys.linea.web3j.EIP4844GasFees
import net.consensys.zkevm.ethereum.gaspricing.GasPriceCaps
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.mockito.Mockito
import org.mockito.kotlin.any
import org.mockito.kotlin.eq
import org.mockito.kotlin.mock
import org.mockito.kotlin.verify
import org.mockito.kotlin.whenever
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.methods.response.EthSendTransaction
import java.math.BigInteger
import java.util.concurrent.CompletableFuture

class LineaRollupAsyncFriendlyTest {
  private val contractAddress = "0x6d976c9b8ceee705d4fe8699b44e5eb58242f484"
  private val smartContractErrors = mapOf<String, String>()
  private val supportingSubmissionData = SupportingSubmissionData(
    /*parentStateRootHash*/ Bytes32.random().toArray(),
    /*dataParentHash*/ Bytes32.random().toArray(),
    /*finalStateRootHash*/ Bytes32.random().toArray(),
    /*firstBlockInData*/ BigInteger.ONE,
    /*finalBlockInData*/ BigInteger.TWO,
    /*snarkHash*/ Bytes32.random().toArray()
  )

  private lateinit var mockedWeb3jClient: Web3j
  private lateinit var lineaRollupAsyncFriendly: LineaRollupAsyncFriendly
  private lateinit var mockedAsyncTransactionManager: AsyncFriendlyTransactionManager
  private lateinit var mockedWMAGasProvider: WMAGasProvider

  @BeforeEach
  fun beforeEach() {
    mockedWeb3jClient = mock<Web3j>(defaultAnswer = Mockito.RETURNS_DEEP_STUBS)
    whenever(mockedWeb3jClient.ethSendRawTransaction(any()).sendAsync())
      .thenReturn(CompletableFuture.completedFuture(EthSendTransaction()))

    mockedAsyncTransactionManager = mock<AsyncFriendlyTransactionManager>(defaultAnswer = Mockito.RETURNS_DEEP_STUBS)

    mockedWMAGasProvider = mock<WMAGasProvider>(defaultAnswer = Mockito.RETURNS_DEEP_STUBS)

    lineaRollupAsyncFriendly = LineaRollupAsyncFriendly.load(
      contractAddress = contractAddress,
      web3j = mockedWeb3jClient,
      transactionManager = mockedAsyncTransactionManager,
      contractGasProvider = mockedWMAGasProvider,
      smartContractErrors = smartContractErrors
    )
  }

  @Test
  fun `finalizeAggregation maxFeePerGas of createRawTransaction should be capped by the given gasPriceCaps`() {
    whenever(mockedWMAGasProvider.isEIP1559Enabled)
      .thenReturn(true)
    whenever(mockedWMAGasProvider.getEIP1559GasFees())
      .thenReturn(
        EIP1559GasFees(
          BigInteger.valueOf(20000000000), // 20GWei
          BigInteger.valueOf(20000000000) // 20GWei
        )
      )

    lineaRollupAsyncFriendly.finalizeAggregation(
      aggregatedProof = Bytes32.random().toArray(),
      proofType = BigInteger.ONE,
      finalizationData = mock<LineaRollup.FinalizationData>(),
      gasPriceCaps = GasPriceCaps(
        maxPriorityFeePerGasCap = BigInteger.valueOf(10000000000), // 10GWei
        maxFeePerGasCap = BigInteger.valueOf(10000000000), // 10GWei
        maxFeePerBlobGasCap = BigInteger.valueOf(1000000000) // 1Gwei
      )
    )

    verify(mockedAsyncTransactionManager)
      .createRawTransaction(
        any(),
        any(),
        maxPriorityFeePerGas = eq(BigInteger.valueOf(10000000000)), // 10GWei
        maxFeePerGas = eq(BigInteger.valueOf(10000000000)), // 10GWei
        any(),
        any(),
        any(),
        any()
      )
  }

  @Test
  fun `finalizeAggregation maxFeePerGas of createRawTransaction should not be capped by the given gasPriceCaps`() {
    whenever(mockedWMAGasProvider.isEIP1559Enabled)
      .thenReturn(true)
    whenever(mockedWMAGasProvider.getEIP1559GasFees())
      .thenReturn(
        EIP1559GasFees(
          BigInteger.valueOf(4000000000), // 4GWei
          BigInteger.valueOf(4000000000) // 4GWei
        )
      )

    lineaRollupAsyncFriendly.finalizeAggregation(
      aggregatedProof = Bytes32.random().toArray(),
      proofType = BigInteger.ONE,
      finalizationData = mock<LineaRollup.FinalizationData>(),
      gasPriceCaps = GasPriceCaps(
        maxPriorityFeePerGasCap = BigInteger.valueOf(10000000000), // 10GWei
        maxFeePerGasCap = BigInteger.valueOf(10000000000), // 10GWei
        maxFeePerBlobGasCap = BigInteger.valueOf(1000000000) // 1Gwei
      )
    )

    verify(mockedAsyncTransactionManager)
      .createRawTransaction(
        any(),
        any(),
        maxPriorityFeePerGas = eq(BigInteger.valueOf(4000000000)), // 4GWei
        maxFeePerGas = eq(BigInteger.valueOf(4000000000)), // 4GWei
        any(),
        any(),
        any(),
        any()
      )
  }

  @Test
  fun
  `finalizeAggregation maxFeePerGas of createRawTransaction should not be capped if the given gasPriceCaps is null`() {
    whenever(mockedWMAGasProvider.isEIP1559Enabled)
      .thenReturn(true)
    whenever(mockedWMAGasProvider.getEIP1559GasFees())
      .thenReturn(
        EIP1559GasFees(
          BigInteger.valueOf(4000000000), // 4GWei
          BigInteger.valueOf(4000000000) // 4GWei
        )
      )

    lineaRollupAsyncFriendly.finalizeAggregation(
      aggregatedProof = Bytes32.random().toArray(),
      proofType = BigInteger.ONE,
      finalizationData = mock<LineaRollup.FinalizationData>(),
      gasPriceCaps = null
    )

    verify(mockedAsyncTransactionManager)
      .createRawTransaction(
        any(),
        any(),
        maxPriorityFeePerGas = eq(BigInteger.valueOf(4000000000)), // 4GWei
        maxFeePerGas = eq(BigInteger.valueOf(4000000000)), // 4GWei
        any(),
        any(),
        any(),
        any()
      )
  }

  @Test
  fun `sendBlobData maxFeePerGas of createRawTransaction should be capped by the given gasPriceCaps`() {
    whenever(mockedWMAGasProvider.getEIP4844GasFees())
      .thenReturn(
        EIP4844GasFees(
          EIP1559GasFees(
            BigInteger.valueOf(20000000000), // 20GWei
            BigInteger.valueOf(20000000000) // 20GWei
          ),
          BigInteger.valueOf(2000000000) // 2GWei
        )
      )

    lineaRollupAsyncFriendly.sendBlobData(
      supportingSubmissionData = supportingSubmissionData,
      dataEvaluationClaim = BigInteger.ONE,
      kzgCommitment = Bytes32.random().toArray(),
      kzgProof = Bytes32.random().toArray(),
      blob = Bytes.random(Constants.Eip4844BlobSize).toArray(),
      gasPriceCaps = GasPriceCaps(
        maxPriorityFeePerGasCap = BigInteger.valueOf(5000000000), // 5GWei
        maxFeePerGasCap = BigInteger.valueOf(5000000000), // 5GWei
        maxFeePerBlobGasCap = BigInteger.valueOf(1000000000) // 1GWei
      )
    )

    verify(mockedAsyncTransactionManager)
      .createRawTransaction(
        any(), any(), any(),
        maxPriorityFeePerGas = eq(BigInteger.valueOf(5000000000)), // 5GWei
        maxFeePerGas = eq(BigInteger.valueOf(5000000000)), // 5GWei
        any(), any(), any(), any(),
        maxFeePerBlobGas = eq(BigInteger.valueOf(1000000000)) // 1GWei
      )
  }

  @Test
  fun `sendBlobData maxFeePerGas of createRawTransaction should not be capped by the given gasPriceCaps`() {
    whenever(mockedWMAGasProvider.getEIP4844GasFees())
      .thenReturn(
        EIP4844GasFees(
          EIP1559GasFees(
            BigInteger.valueOf(4000000000), // 4GWei
            BigInteger.valueOf(4000000000) // 4GWei
          ),
          BigInteger.valueOf(500000000) // 0.5GWei
        )
      )

    lineaRollupAsyncFriendly.sendBlobData(
      supportingSubmissionData = supportingSubmissionData,
      dataEvaluationClaim = BigInteger.ONE,
      kzgCommitment = Bytes32.random().toArray(),
      kzgProof = Bytes32.random().toArray(),
      blob = Bytes.random(Constants.Eip4844BlobSize).toArray(),
      gasPriceCaps = GasPriceCaps(
        maxPriorityFeePerGasCap = BigInteger.valueOf(5000000000), // 5GWei
        maxFeePerGasCap = BigInteger.valueOf(5000000000), // 5GWei
        maxFeePerBlobGasCap = BigInteger.valueOf(1000000000) // 1GWei
      )
    )

    verify(mockedAsyncTransactionManager)
      .createRawTransaction(
        any(), any(), any(),
        maxPriorityFeePerGas = eq(BigInteger.valueOf(4000000000)), // 4GWei
        maxFeePerGas = eq(BigInteger.valueOf(4000000000)), // 4GWei
        any(), any(), any(), any(),
        maxFeePerBlobGas = eq(BigInteger.valueOf(500000000)) // 0.5GWei
      )
  }

  @Test
  fun `sendBlobData maxFeePerGas of createRawTransaction should not be capped if the given gasPriceCaps is null`() {
    whenever(mockedWMAGasProvider.getEIP4844GasFees())
      .thenReturn(
        EIP4844GasFees(
          EIP1559GasFees(
            BigInteger.valueOf(20000000000), // 20GWei
            BigInteger.valueOf(20000000000) // 20GWei
          ),
          BigInteger.valueOf(2000000000) // 2GWei
        )
      )

    lineaRollupAsyncFriendly.sendBlobData(
      supportingSubmissionData = supportingSubmissionData,
      dataEvaluationClaim = BigInteger.ONE,
      kzgCommitment = Bytes32.random().toArray(),
      kzgProof = Bytes32.random().toArray(),
      blob = Bytes.random(Constants.Eip4844BlobSize).toArray(),
      gasPriceCaps = null
    )

    verify(mockedAsyncTransactionManager)
      .createRawTransaction(
        any(), any(), any(),
        maxPriorityFeePerGas = eq(BigInteger.valueOf(20000000000)), // 20GWei
        maxFeePerGas = eq(BigInteger.valueOf(20000000000)), // 20GWei
        any(), any(), any(), any(),
        maxFeePerBlobGas = eq(BigInteger.valueOf(2000000000)) // 2GWei
      )
  }
}
