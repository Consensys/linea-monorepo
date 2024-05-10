package net.consensys.zkevm.ethereum.coordination.dynamicgaspricecap

import io.vertx.core.Vertx
import io.vertx.junit5.VertxExtension
import kotlinx.datetime.Clock
import kotlinx.datetime.Instant
import net.consensys.zkevm.coordinator.blockcreation.ExtendedWeb3J
import net.consensys.zkevm.ethereum.gaspricing.FeeHistoriesRepositoryWithCache
import net.consensys.zkevm.ethereum.gaspricing.GasPriceCapCalculator
import net.consensys.zkevm.ethereum.gaspricing.GasPriceCaps
import net.consensys.zkevm.ethereum.gaspricing.PercentileGasFees
import net.consensys.zkevm.ethereum.gaspricing.TimeOfDayMultipliers
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32
import org.apache.tuweni.units.bigints.UInt256
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import org.junit.jupiter.api.extension.ExtendWith
import org.mockito.Mockito.RETURNS_DEEP_STUBS
import org.mockito.kotlin.any
import org.mockito.kotlin.doReturn
import org.mockito.kotlin.mock
import org.mockito.kotlin.whenever
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.methods.response.EthBlockNumber
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.infrastructure.bytes.Bytes20
import tech.pegasys.teku.infrastructure.unsigned.UInt64
import java.math.BigInteger
import java.util.concurrent.CompletableFuture
import kotlin.time.Duration
import kotlin.time.Duration.Companion.hours

@ExtendWith(VertxExtension::class)
class GasPriceCapProviderImplTest {
  private val currentTime = Instant.parse("2024-03-20T00:00:00Z") // Wednesday
  private val latestL1BlockNumber = 200L
  private val baseFeePerGasPercentile = 10.0
  private val baseFeePerGasPercentileWindowInBlocks = 100U
  private val baseFeePerGasPercentileWindowLeewayInBlocks = 20U
  private val timeOfDayMultipliers = mapOf(
    "WEDNESDAY_0" to 1.0,
    "WEDNESDAY_1" to 2.0,
    "WEDNESDAY_2" to 3.0,
    "WEDNESDAY_3" to 4.0
  )
  private val p10BaseFeeGas = BigInteger("1000000000") // 1GWei
  private val p10BaseFeeBlobGas = BigInteger("100000000") // 0.1GWei
  private val avgP10Reward = BigInteger("200000000") // 2GWei
  private val maxFeePerGasCap = BigInteger("10000000000") // 10GWei
  private val maxFeePerBlobGasCap = BigInteger("1000000000") // 1GWei
  private val storedFeeHistoriesNum = 100
  private val adjustmentConstant = 25U
  private val finalizationTargetMaxDelay = 6.hours
  private val gasPriceCapCalculator = GasPriceCapCalculatorImpl()

  private lateinit var targetBlockTime: Instant
  private lateinit var mockedL1Web3jClient: Web3j
  private lateinit var mockedL2ExtendedWeb3JClient: ExtendedWeb3J
  private lateinit var mockedL1FeeHistoriesRepository: FeeHistoriesRepositoryWithCache
  private lateinit var mockedClock: Clock

  private fun executionPayloadV1(
    blockNumber: Long = 0,
    parentHash: Bytes32 = Bytes32.random(),
    feeRecipient: Bytes20 = Bytes20(Bytes.random(20)),
    stateRoot: Bytes32 = Bytes32.random(),
    receiptsRoot: Bytes32 = Bytes32.random(),
    logsBloom: Bytes = Bytes32.random(),
    prevRandao: Bytes32 = Bytes32.random(),
    gasLimit: UInt64 = UInt64.valueOf(0),
    gasUsed: UInt64 = UInt64.valueOf(0),
    timestamp: UInt64 = UInt64.valueOf(0),
    extraData: Bytes = Bytes32.random(),
    baseFeePerGas: UInt256 = UInt256.valueOf(256),
    blockHash: Bytes32 = Bytes32.random(),
    transactions: List<Bytes> = emptyList()
  ): ExecutionPayloadV1 {
    return ExecutionPayloadV1(
      parentHash,
      feeRecipient,
      stateRoot,
      receiptsRoot,
      logsBloom,
      prevRandao,
      UInt64.valueOf(blockNumber),
      gasLimit,
      gasUsed,
      timestamp,
      extraData,
      baseFeePerGas,
      blockHash,
      transactions
    )
  }

  private fun createGasPriceCapProvider(
    enabled: Boolean = true,
    maxFeePerGasCap: BigInteger = this.maxFeePerGasCap,
    maxFeePerBlobGasCap: BigInteger = this.maxFeePerBlobGasCap,
    baseFeePerGasPercentile: Double = this.baseFeePerGasPercentile,
    baseFeePerGasPercentileWindowInBlocks: UInt = this.baseFeePerGasPercentileWindowInBlocks,
    baseFeePerGasPercentileWindowLeewayInBlocks: UInt = this.baseFeePerGasPercentileWindowLeewayInBlocks,
    timeOfDayMultipliers: TimeOfDayMultipliers = this.timeOfDayMultipliers,
    adjustmentConstant: UInt = this.adjustmentConstant,
    blobAdjustmentConstant: UInt = this.adjustmentConstant,
    finalizationTargetMaxDelay: Duration = this.finalizationTargetMaxDelay,
    l1Web3jClient: Web3j = mockedL1Web3jClient,
    l2ExtendedWeb3JClient: ExtendedWeb3J = mockedL2ExtendedWeb3JClient,
    feeHistoriesRepository: FeeHistoriesRepositoryWithCache = mockedL1FeeHistoriesRepository,
    gasPriceCapCalculator: GasPriceCapCalculator = this.gasPriceCapCalculator,
    clock: Clock = mockedClock
  ): GasPriceCapProviderImpl {
    return GasPriceCapProviderImpl(
      config = GasPriceCapProviderImpl.Config(
        enabled = enabled,
        maxFeePerGasCap = maxFeePerGasCap,
        maxFeePerBlobGasCap = maxFeePerBlobGasCap,
        baseFeePerGasPercentile = baseFeePerGasPercentile,
        baseFeePerGasPercentileWindowInBlocks = baseFeePerGasPercentileWindowInBlocks,
        baseFeePerGasPercentileWindowLeewayInBlocks = baseFeePerGasPercentileWindowLeewayInBlocks,
        timeOfDayMultipliers = timeOfDayMultipliers,
        adjustmentConstant = adjustmentConstant,
        blobAdjustmentConstant = blobAdjustmentConstant,
        finalizationTargetMaxDelay = finalizationTargetMaxDelay
      ),
      l1Web3jClient = l1Web3jClient,
      l2ExtendedWeb3JClient = l2ExtendedWeb3JClient,
      feeHistoriesRepository = feeHistoriesRepository,
      gasPriceCapCalculator = gasPriceCapCalculator,
      clock = clock
    )
  }

  @BeforeEach
  fun beforeEach(vertx: Vertx) {
    mockedL1Web3jClient = mock<Web3j>(defaultAnswer = RETURNS_DEEP_STUBS)
    val mockBlockNumberReturn = mock<EthBlockNumber>()
    whenever(mockBlockNumberReturn.blockNumber).thenReturn(BigInteger.valueOf(latestL1BlockNumber))
    whenever(mockedL1Web3jClient.ethBlockNumber().sendAsync())
      .thenReturn(CompletableFuture.completedFuture(mockBlockNumberReturn))

    targetBlockTime = currentTime - 1.hours
    mockedL2ExtendedWeb3JClient = mock<ExtendedWeb3J> {
      on { ethGetExecutionPayloadByNumber(any()) } doReturn SafeFuture.completedFuture(
        executionPayloadV1(timestamp = UInt64.valueOf(targetBlockTime.epochSeconds))
      )
    }

    mockedL1FeeHistoriesRepository = mock<FeeHistoriesRepositoryWithCache> {
      on { getCachedNumOfFeeHistoriesFromBlockNumber() } doReturn storedFeeHistoriesNum
      on { getCachedPercentileGasFees() } doReturn PercentileGasFees(
        percentileBaseFeePerGas = p10BaseFeeGas,
        percentileBaseFeePerBlobGas = p10BaseFeeBlobGas,
        percentileAvgReward = avgP10Reward
      )
    }

    mockedClock = mock<Clock> {
      on { now() } doReturn currentTime
    }
  }

  @Test
  fun `constructor throws error if config variables are invalid`() {
    val negativeMaxFeeCap = BigInteger.valueOf(-1000000000)
    assertThrows<IllegalArgumentException> {
      createGasPriceCapProvider(
        maxFeePerGasCap = negativeMaxFeeCap,
        maxFeePerBlobGasCap = BigInteger.valueOf(1000000000)
      )
    }.also { exception ->
      assertThat(exception.message)
        .isEqualTo(
          "maxFeePerGasCap must be no less than 0. Value=$negativeMaxFeeCap"
        )
    }

    assertThrows<IllegalArgumentException> {
      createGasPriceCapProvider(
        maxFeePerGasCap = BigInteger.valueOf(1000000000),
        maxFeePerBlobGasCap = negativeMaxFeeCap
      )
    }.also { exception ->
      assertThat(exception.message)
        .isEqualTo(
          "maxFeePerBlobGasCap must be no less than 0. Value=$negativeMaxFeeCap"
        )
    }

    val negativePercentile = -10.0
    assertThrows<IllegalArgumentException> {
      createGasPriceCapProvider(
        baseFeePerGasPercentile = negativePercentile
      )
    }.also { exception ->
      assertThat(exception.message)
        .isEqualTo(
          "baseFeePerGasPercentile must be no less than 0.0. Value=$negativePercentile"
        )
    }

    val negativeDuration = (-1).hours
    assertThrows<IllegalArgumentException> {
      createGasPriceCapProvider(
        finalizationTargetMaxDelay = negativeDuration
      )
    }.also { exception ->
      assertThat(exception.message)
        .isEqualTo(
          "finalizationTargetMaxDelay duration must be longer than zero second. Value=$negativeDuration"
        )
    }
  }

  @Test
  fun `gas price caps should be returned correctly`() {
    val targetL2BlockNumber = 100L
    val gasPriceCapProvider = createGasPriceCapProvider()

    assertThat(
      gasPriceCapProvider.getGasPriceCaps(targetL2BlockNumber).get()
    ).isEqualTo(
      GasPriceCaps(
        maxPriorityFeePerGasCap = BigInteger.valueOf(1894444444),
        maxFeePerGasCap = BigInteger.valueOf(1894444444),
        maxFeePerBlobGasCap = BigInteger.valueOf(169444444)
      )
    )
  }

  @Test
  fun `gas price caps should be capped if exceeds max caps`() {
    // set the block time of the target block from long past, i.e. 33 hours,
    // to trigger large max fee caps
    targetBlockTime = currentTime - 33.hours
    mockedL2ExtendedWeb3JClient = mock<ExtendedWeb3J> {
      on { ethGetExecutionPayloadByNumber(any()) } doReturn SafeFuture.completedFuture(
        executionPayloadV1(timestamp = UInt64.valueOf(targetBlockTime.epochSeconds))
      )
    }
    val targetL2BlockNumber = 100L
    val gasPriceCapProvider = createGasPriceCapProvider(
      l2ExtendedWeb3JClient = mockedL2ExtendedWeb3JClient
    )

    assertThat(
      gasPriceCapProvider.getGasPriceCaps(targetL2BlockNumber).get()
    ).isEqualTo(
      GasPriceCaps(
        maxPriorityFeePerGasCap = maxFeePerGasCap,
        maxFeePerGasCap = maxFeePerGasCap,
        maxFeePerBlobGasCap = maxFeePerBlobGasCap
      )
    )
  }

  @Test
  fun `gas price caps should be default if disabled`() {
    val targetL2BlockNumber = 100L
    val gasPriceCapProvider = createGasPriceCapProvider(
      enabled = false
    )

    assertThat(
      gasPriceCapProvider.getGasPriceCaps(targetL2BlockNumber).get()
    ).isEqualTo(
      gasPriceCapProvider.getDefaultGasPriceCaps()
    )
  }

  @Test
  fun `gas price caps should be default if not enough fee history data`() {
    val targetL2BlockNumber = 100L
    val gasPriceCapProvider = createGasPriceCapProvider(
      baseFeePerGasPercentileWindowInBlocks = 200U
    )

    assertThat(
      gasPriceCapProvider.getGasPriceCaps(targetL2BlockNumber).get()
    ).isEqualTo(
      gasPriceCapProvider.getDefaultGasPriceCaps()
    )
  }

  @Test
  fun `gas price caps should be default if error on feeHistoriesRepository`() {
    val targetL2BlockNumber = 100L
    mockedL1FeeHistoriesRepository = mock<FeeHistoriesRepositoryWithCache> {
      on { getNumOfFeeHistoriesFromBlockNumber(any(), any()) } doReturn SafeFuture.failedFuture(
        Error("Throw error for testing")
      )
    }
    val gasPriceCapProvider = createGasPriceCapProvider(
      l2ExtendedWeb3JClient = mockedL2ExtendedWeb3JClient
    )

    assertThat(
      gasPriceCapProvider.getGasPriceCaps(targetL2BlockNumber).get()
    ).isEqualTo(
      gasPriceCapProvider.getDefaultGasPriceCaps()
    )
  }

  @Test
  fun `gas price caps should be default if error on l2ExtendedWeb3JClient`() {
    val targetL2BlockNumber = 100L
    mockedL2ExtendedWeb3JClient = mock<ExtendedWeb3J> {
      on { ethGetExecutionPayloadByNumber(any()) } doReturn SafeFuture.failedFuture(
        Error("Throw error for testing")
      )
    }
    val gasPriceCapProvider = createGasPriceCapProvider(
      l2ExtendedWeb3JClient = mockedL2ExtendedWeb3JClient
    )

    assertThat(
      gasPriceCapProvider.getGasPriceCaps(targetL2BlockNumber).get()
    ).isEqualTo(
      gasPriceCapProvider.getDefaultGasPriceCaps()
    )
  }
}
