package net.consensys.linea.ethereum.gaspricing.staticcap

import linea.domain.FeeHistory
import linea.web3j.ExtendedWeb3J
import net.consensys.linea.ethereum.gaspricing.FeesCalculator
import net.consensys.linea.ethereum.gaspricing.HistoricVariableCostProvider
import net.consensys.linea.ethereum.gaspricing.L2CalldataSizeAccumulator
import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatThrownBy
import org.junit.jupiter.api.Test
import org.mockito.Mockito
import org.mockito.kotlin.any
import org.mockito.kotlin.doReturn
import org.mockito.kotlin.eq
import org.mockito.kotlin.mock
import org.mockito.kotlin.whenever
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.math.BigInteger

class L2CalldataBasedVariableFeesCalculatorTest {
  private class FakeHistoricVariableCostProviderImpl : HistoricVariableCostProvider {
    private var latestVariableCost: Double = 0.0

    fun setLatestVariableCost(latestVariableCost: Double) {
      this.latestVariableCost = latestVariableCost
    }

    override fun getVariableCost(latestBlockNumber: ULong): SafeFuture<Double> {
      return SafeFuture.completedFuture(latestVariableCost)
    }
  }

  private val config = L2CalldataBasedVariableFeesCalculator.Config(
    feeChangeDenominator = 32u,
    calldataSizeBlockCount = 5u,
    maxBlockCalldataSize = 109000u,
  )
  private val feeHistory = FeeHistory(
    oldestBlock = 100uL,
    baseFeePerGas = listOf(100UL),
    reward = listOf(listOf(1000UL)),
    gasUsedRatio = listOf(0.25),
    baseFeePerBlobGas = listOf(100UL),
    blobGasUsedRatio = listOf(0.25),
  )

  // mocked VariableFeesCalculator
  private val variableFee = 15000.0
  private val mockVariableFeesCalculator = mock<FeesCalculator> {
    on { calculateFees(eq(feeHistory)) } doReturn variableFee
  }

  // mocked L2 Web3jClient
  private val mockWeb3jClient = mock<ExtendedWeb3J> {
    on { ethBlockNumber() } doReturn SafeFuture.completedFuture(100.toBigInteger())
  }

  // mocked L2CalldataSizeAccumulator
  private val sumOfCalldataSize = (109000 * 5).toBigInteger() // maxBlockCalldataSize * calldataSizeBlockCount
  private val mockL2CalldataSizeAccumulator = mock<L2CalldataSizeAccumulator> {
    on { getSumOfL2CalldataSize(any()) } doReturn SafeFuture.completedFuture(sumOfCalldataSize)
  }

  private val fakeHistoricVariableCostProvider = FakeHistoricVariableCostProviderImpl()

  @Test
  fun test_calculateFees_past_blocks_calldata_at_max_target() {
    // delta would be 1.0
    val delta = 1.0
    val expectedVariableFees = 15000.0 * (1.0 + delta / 32.0)

    val feesCalculator = L2CalldataBasedVariableFeesCalculator(
      config = config,
      web3jClient = mockWeb3jClient,
      variableFeesCalculator = mockVariableFeesCalculator,
      l2CalldataSizeAccumulator = mockL2CalldataSizeAccumulator,
      historicVariableCostProvider = fakeHistoricVariableCostProvider,
    )

    // call calculateFees first to instantiate the lastVariableCost
    feesCalculator.calculateFees(feeHistory).let {
      fakeHistoricVariableCostProvider.setLatestVariableCost(it)
    }

    assertThat(feesCalculator.calculateFees(feeHistory))
      .isEqualTo(expectedVariableFees)
  }

  @Test
  fun test_calculateFees_past_blocks_calldata_exceed_max_target() {
    // This could happen as the calldata from L2CalldataSizeAccumulator is just approximation
    val sumOfCalldataSize = (200000 * 5).toBigInteger()
    val mockL2CalldataSizeAccumulator = mock<L2CalldataSizeAccumulator> {
      on { getSumOfL2CalldataSize(any()) } doReturn SafeFuture.completedFuture(sumOfCalldataSize)
    }

    // delta would be 1.0
    val delta = 1.0
    val expectedVariableFees = 15000.0 * (1.0 + delta / 32.0)

    val feesCalculator = L2CalldataBasedVariableFeesCalculator(
      config = config,
      web3jClient = mockWeb3jClient,
      variableFeesCalculator = mockVariableFeesCalculator,
      l2CalldataSizeAccumulator = mockL2CalldataSizeAccumulator,
      historicVariableCostProvider = fakeHistoricVariableCostProvider,
    )

    // call calculateFees first to instantiate the lastVariableCost
    feesCalculator.calculateFees(feeHistory).let {
      fakeHistoricVariableCostProvider.setLatestVariableCost(it)
    }

    assertThat(feesCalculator.calculateFees(feeHistory))
      .isEqualTo(expectedVariableFees)
  }

  @Test
  fun test_calculateFees_past_blocks_calldata_size_at_zero() {
    val mockL2CalldataSizeAccumulator = mock<L2CalldataSizeAccumulator> {
      on { getSumOfL2CalldataSize(any()) } doReturn SafeFuture.completedFuture(BigInteger.ZERO)
    }

    val feesCalculator = L2CalldataBasedVariableFeesCalculator(
      config = config,
      web3jClient = mockWeb3jClient,
      variableFeesCalculator = mockVariableFeesCalculator,
      l2CalldataSizeAccumulator = mockL2CalldataSizeAccumulator,
      historicVariableCostProvider = fakeHistoricVariableCostProvider,
    )

    // call calculateFees first to instantiate the lastVariableCost
    feesCalculator.calculateFees(feeHistory).let {
      fakeHistoricVariableCostProvider.setLatestVariableCost(it)
    }

    assertThat(feesCalculator.calculateFees(feeHistory))
      .isEqualTo(15000.0)
  }

  @Test
  fun test_calculateFees_past_blocks_calldata_above_half_max() {
    val sumOfCalldataSize = (81750 * 5).toBigInteger()
    val mockL2CalldataSizeAccumulator = mock<L2CalldataSizeAccumulator> {
      on { getSumOfL2CalldataSize(any()) } doReturn SafeFuture.completedFuture(sumOfCalldataSize)
    }

    // delta would be 0.5
    val delta = 0.5
    val expectedVariableFees = 15000.0 * (1.0 + delta / 32.0)

    val feesCalculator = L2CalldataBasedVariableFeesCalculator(
      config = config,
      web3jClient = mockWeb3jClient,
      variableFeesCalculator = mockVariableFeesCalculator,
      l2CalldataSizeAccumulator = mockL2CalldataSizeAccumulator,
      historicVariableCostProvider = fakeHistoricVariableCostProvider,
    )

    // call calculateFees first to instantiate the lastVariableCost
    feesCalculator.calculateFees(feeHistory).let {
      fakeHistoricVariableCostProvider.setLatestVariableCost(it)
    }

    assertThat(feesCalculator.calculateFees(feeHistory))
      .isEqualTo(expectedVariableFees)
  }

  @Test
  fun test_calculateFees_past_blocks_calldata_below_half_max() {
    val sumOfCalldataSize = (27250 * 5).toBigInteger()
    val mockL2CalldataSizeAccumulator = mock<L2CalldataSizeAccumulator> {
      on { getSumOfL2CalldataSize(any()) } doReturn SafeFuture.completedFuture(sumOfCalldataSize)
    }

    val feesCalculator = L2CalldataBasedVariableFeesCalculator(
      config = config,
      web3jClient = mockWeb3jClient,
      variableFeesCalculator = mockVariableFeesCalculator,
      l2CalldataSizeAccumulator = mockL2CalldataSizeAccumulator,
      historicVariableCostProvider = fakeHistoricVariableCostProvider,
    )

    // call calculateFees first to instantiate the lastVariableCost
    feesCalculator.calculateFees(feeHistory).let {
      fakeHistoricVariableCostProvider.setLatestVariableCost(it)
    }

    assertThat(feesCalculator.calculateFees(feeHistory))
      .isEqualTo(15000.0)
  }

  @Test
  fun test_calculateFees_increase_to_more_than_double_when_past_blocks_calldata_at_max_target() {
    val feesCalculator = L2CalldataBasedVariableFeesCalculator(
      config = config,
      web3jClient = mockWeb3jClient,
      variableFeesCalculator = mockVariableFeesCalculator,
      l2CalldataSizeAccumulator = mockL2CalldataSizeAccumulator,
      historicVariableCostProvider = fakeHistoricVariableCostProvider,
    )

    // call calculateFees first to instantiate the lastVariableCost
    feesCalculator.calculateFees(feeHistory).let {
      fakeHistoricVariableCostProvider.setLatestVariableCost(it)
    }

    // With (1 + 1/32) as the rate, after 23 calls
    // the expectedVariableFees should be increased to more than double
    var calculatedFee = 0.0
    (0..22).forEach { _ ->
      calculatedFee = feesCalculator.calculateFees(feeHistory).apply {
        fakeHistoricVariableCostProvider.setLatestVariableCost(this)
      }
    }

    assertThat(calculatedFee).isGreaterThan(15000.0 * 2.0)
  }

  @Test
  fun test_calculateFees_decrease_to_less_than_half_when_past_blocks_calldata_at_zero() {
    val mockVariableFeesCalculator = mock<FeesCalculator>()
    whenever(mockVariableFeesCalculator.calculateFees(eq(feeHistory)))
      .thenReturn(variableFee, 0.0)

    val mockL2CalldataSizeAccumulator = mock<L2CalldataSizeAccumulator> {
      on { getSumOfL2CalldataSize(any()) } doReturn SafeFuture.completedFuture(BigInteger.ZERO)
    }

    val feesCalculator = L2CalldataBasedVariableFeesCalculator(
      config = config,
      web3jClient = mockWeb3jClient,
      variableFeesCalculator = mockVariableFeesCalculator,
      l2CalldataSizeAccumulator = mockL2CalldataSizeAccumulator,
      historicVariableCostProvider = fakeHistoricVariableCostProvider,
    )

    // call calculateFees first to instantiate the lastVariableCost
    feesCalculator.calculateFees(feeHistory).let {
      fakeHistoricVariableCostProvider.setLatestVariableCost(it)
    }

    // With (1 - 1/32) as the rate, after 22 calls
    // the expectedVariableFees should be decreased to less than half
    var calculatedFee = 0.0
    (0..21).forEach { _ ->
      calculatedFee = feesCalculator.calculateFees(feeHistory).apply {
        fakeHistoricVariableCostProvider.setLatestVariableCost(this)
      }
    }

    assertThat(calculatedFee).isLessThan(15000.0 / 2.0)
  }

  @Test
  fun test_calculateFees_when_block_count_is_zero() {
    val feesCalculator = L2CalldataBasedVariableFeesCalculator(
      config = L2CalldataBasedVariableFeesCalculator.Config(
        feeChangeDenominator = 32u,
        calldataSizeBlockCount = 0u, // set zero to disable calldata-based variable fees
        maxBlockCalldataSize = 109000u,
      ),
      web3jClient = mockWeb3jClient,
      variableFeesCalculator = mockVariableFeesCalculator,
      l2CalldataSizeAccumulator = mockL2CalldataSizeAccumulator,
      historicVariableCostProvider = fakeHistoricVariableCostProvider,
    )

    // call calculateFees first to instantiate the lastVariableCost
    feesCalculator.calculateFees(feeHistory).let {
      fakeHistoricVariableCostProvider.setLatestVariableCost(it)
    }

    // The returned variable fees should always be 15000.0
    // as calldata-based variable fees is disabled
    assertThat(feesCalculator.calculateFees(feeHistory))
      .isEqualTo(15000.0)
  }

  @Test
  fun test_calculateFees_would_not_change_when_latest_variable_cost_stays_same() {
    val feesCalculator = L2CalldataBasedVariableFeesCalculator(
      config = config,
      web3jClient = mockWeb3jClient,
      variableFeesCalculator = mockVariableFeesCalculator,
      l2CalldataSizeAccumulator = mockL2CalldataSizeAccumulator,
      historicVariableCostProvider = fakeHistoricVariableCostProvider,
    )

    // call calculateFees first to instantiate the lastVariableCost
    var expectedVariableFees = 15000.0
    feesCalculator.calculateFees(feeHistory).let {
      assertThat(it).isEqualTo(15000.0)
      fakeHistoricVariableCostProvider.setLatestVariableCost(it)
    }

    // delta would be 1.0
    val delta = 1.0
    expectedVariableFees = expectedVariableFees * (1.0 + delta / 32.0)

    // we don't update the latest variable cost after each calculation
    // to mimic block production halts after last 5 blocks with full calldata
    (0..10).forEach { _ ->
      val calculatedFee = feesCalculator.calculateFees(feeHistory)
      // calculatedFee should be equal to the value returned from the first calculateFees call
      assertThat(calculatedFee).isEqualTo(expectedVariableFees)
    }
  }

  @Test
  fun test_calculateFees_would_return_original_variable_cost_when_current_block_number_is_less_than_5() {
    val mockExtendedWeb3J = mock<ExtendedWeb3J>()
    whenever(mockExtendedWeb3J.ethBlockNumber())
      .thenReturn(SafeFuture.completedFuture(BigInteger.TWO)) // current block number is 2

    val fakeL2CalldataSizeAccumulator = L2CalldataSizeAccumulatorImpl(
      web3jClient = mock<ExtendedWeb3J>(defaultAnswer = Mockito.RETURNS_DEEP_STUBS),
      config = L2CalldataSizeAccumulatorImpl.Config(
        blockSizeNonCalldataOverhead = 540U,
        calldataSizeBlockCount = 5U,
      ),
    )

    val feesCalculator = L2CalldataBasedVariableFeesCalculator(
      config = config,
      web3jClient = mockExtendedWeb3J,
      variableFeesCalculator = mockVariableFeesCalculator,
      l2CalldataSizeAccumulator = fakeL2CalldataSizeAccumulator,
      historicVariableCostProvider = fakeHistoricVariableCostProvider,
    )

    // call calculateFees first to instantiate the lastVariableCost
    feesCalculator.calculateFees(feeHistory).let {
      fakeHistoricVariableCostProvider.setLatestVariableCost(it)
    }

    assertThat(feesCalculator.calculateFees(feeHistory))
      .isEqualTo(15000.0)
  }

  @Test
  fun test_calculateFees_would_throw_error_when_failed_to_get_calldata_size_sum() {
    val expectedException = RuntimeException("Error while getting calldata size sum")
    val mockL2CalldataSizeAccumulator = mock<L2CalldataSizeAccumulator>()
    whenever(mockL2CalldataSizeAccumulator.getSumOfL2CalldataSize(any()))
      .thenThrow(expectedException)

    val feesCalculator = L2CalldataBasedVariableFeesCalculator(
      config = config,
      web3jClient = mockWeb3jClient,
      variableFeesCalculator = mockVariableFeesCalculator,
      l2CalldataSizeAccumulator = mockL2CalldataSizeAccumulator,
      historicVariableCostProvider = FakeHistoricVariableCostProviderImpl(),
    )

    assertThatThrownBy {
      feesCalculator.calculateFees(feeHistory)
    }.hasCause(expectedException)
  }

  @Test
  fun test_calculateFees_would_throw_error_when_failed_to_get_latest_variable_cost() {
    val expectedException = RuntimeException("Error while getting variable cost from latest block extra data")
    val mockHistoricVariableCostProvider = mock<HistoricVariableCostProvider>()
    whenever(mockHistoricVariableCostProvider.getVariableCost(any()))
      .thenThrow(expectedException)

    val feesCalculator = L2CalldataBasedVariableFeesCalculator(
      config = config,
      web3jClient = mockWeb3jClient,
      variableFeesCalculator = mockVariableFeesCalculator,
      l2CalldataSizeAccumulator = mockL2CalldataSizeAccumulator,
      historicVariableCostProvider = mockHistoricVariableCostProvider,
    )

    assertThatThrownBy {
      feesCalculator.calculateFees(feeHistory)
    }.hasCause(expectedException)
  }
}
