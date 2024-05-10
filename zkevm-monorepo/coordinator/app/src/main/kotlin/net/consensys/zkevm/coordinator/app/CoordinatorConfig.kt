package net.consensys.zkevm.coordinator.app

import com.sksamuel.hoplite.ConfigAlias
import com.sksamuel.hoplite.Masked
import net.consensys.assertIs32Bytes
import net.consensys.decodeHex
import net.consensys.linea.assertIsValidAddress
import net.consensys.linea.jsonrpc.client.RequestRetryConfig
import net.consensys.linea.traces.TracesCounters
import net.consensys.linea.traces.TracingModule
import net.consensys.linea.web3j.SmartContractErrors
import net.consensys.zkevm.ethereum.gaspricing.MAX_FEE_HISTORIES_STORAGE_PERIOD
import net.consensys.zkevm.ethereum.gaspricing.MAX_FEE_HISTORY_BLOCK_COUNT
import net.consensys.zkevm.ethereum.gaspricing.MAX_REWARD_PERCENTILES_SIZE
import net.consensys.zkevm.ethereum.gaspricing.TimeOfDayMultipliers
import net.consensys.zkevm.ethereum.gaspricing.getAllTimeOfDayKeys
import java.math.BigDecimal
import java.math.BigInteger
import java.net.URL
import java.nio.file.Path
import java.time.Duration
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.toJavaDuration
import kotlin.time.toKotlinDuration

data class ApiConfig(
  val observabilityPort: UInt
)

data class ConflationConfig(
  val consistentNumberOfBlocksOnL1ToWait: Int,
  val conflationCalculatorVersion: String,
  val conflationDeadline: Duration,
  val conflationDeadlineCheckInterval: Duration,
  val conflationDeadlineLastBlockConfirmationDelay: Duration,
  val blocksLimit: Long? = null,
  private var _tracesLimits: TracesCounters?,
  val smartContractErrors: SmartContractErrors?,
  val fetchBlocksLimit: Int
) {

  init {
    require(conflationDeadlineCheckInterval <= conflationDeadline) {
      "Clock ticker interval must be smaller than conflation deadline"
    }
    consistentNumberOfBlocksOnL1ToWait.let {
      require(it > 0) { "consistentNumberOfBlocksOnL1ToWait must be grater than 0" }
    }
    blocksLimit?.let { require(it > 0) { "blocksLimit must be greater than 0" } }
    _tracesLimits?.also { limits ->
      if (limits.keys != TracingModule.allModules) {
        val missingModules = TracingModule.allModules - limits.keys
        val extraModules = limits.keys - TracingModule.allModules
        val errorMessage =
          "Invalid traces limits: missing modules: " +
            "${missingModules.joinToString(",", "[", "]")}, " +
            "unsupported modules=${extraModules.joinToString(",", "[", "]")}"
        throw IllegalStateException(errorMessage)
      }
    }
  }

  val tracesLimits: TracesCounters
    get() = _tracesLimits ?: throw IllegalStateException("Traces limits not defined!")
}

data class ZkGethTraces(
  val ethApi: URL,
  val newBlockPollingInterval: Duration
)

data class ProverConfig(
  val version: String,
  val fsRequestsDirectory: Path,
  val fsResponsesDirectory: Path,
  val fsPollingInterval: Duration,
  val fsPollingTimeout: Duration,
  val fsInprogessProvingSuffixPattern: String,
  val fsInprogessRequestWritingSuffix: String
)

/**
 * This class is used to parse the requestRetry config from the toml file.
 * If we use UInt toml parser throws an exception because it does not support SOMETIMES UInt values: ¯\_(ツ)_/¯
 *
 *  kotlin.reflect.jvm.internal.KotlinReflectionInternalError:
 *  This callable does not support a default call: public constructor
 *  RequestRetryConfigTomlFriendly(maxAttempts: kotlin.UInt? = ...
 */
data class RequestRetryConfigTomlFriendly(
  val maxAttempts: Int? = null,
  val timeout: Duration? = null,
  val backoffDelay: Duration,
  val failuresWarningThreshold: Int = 0
) {
  init {
    require(maxAttempts != null || timeout != null) { "maxRetries or timeout must be specified" }
    maxAttempts?.also {
      require(maxAttempts > 0) { "maxRetries must be greater than zero. value=$maxAttempts" }
    }
    timeout?.also {
      require(timeout.toKotlinDuration() > 0.milliseconds) { "timeout must be >= 1ms. value=$timeout" }
    }
  }

  internal val asDomain = RequestRetryConfig(
    maxAttempts = maxAttempts?.toUInt(),
    timeout = timeout?.toKotlinDuration(),
    backoffDelay = backoffDelay.toKotlinDuration(),
    failuresWarningThreshold = failuresWarningThreshold.toUInt()
  )
}

private interface RequestRetryConfigurable {
  val requestRetry: RequestRetryConfigTomlFriendly
  val requestRetryConfig: RequestRetryConfig
    get() = requestRetry.asDomain
}

data class BlobCompressionConfig(
  val blobSizeLimit: Int,
  val handlerPollingInterval: Duration,
  val prover: ProverConfig
)

data class AggregationConfig(
  val aggregationCalculatorVersion: String,
  val aggregationProofsLimit: Int,
  val aggregationDeadline: Duration,
  val pollingInterval: Duration,
  val prover: ProverConfig
)

data class TracesConfig(
  val rawExecutionTracesVersion: String,
  val expectedTracesApiVersion: String,
  val counters: FunctionalityEndpoint,
  val conflation: FunctionalityEndpoint,
  val fileManager: FileManager
) {
  data class FunctionalityEndpoint(
    val endpoints: List<URL>,
    val requestLimitPerEndpoint: UInt,
    override val requestRetry: RequestRetryConfigTomlFriendly
  ) : RequestRetryConfigurable {
    init {
      require(requestLimitPerEndpoint > 0u) { "requestLimitPerEndpoint must be greater than 0" }
    }
  }

  data class FileManager(
    val tracesFileExtension: String,
    val rawTracesDirectory: Path,
    val nonCanonicalRawTracesDirectory: Path,
    val createNonCanonicalDirectory: Boolean,
    val pollingInterval: Duration,
    val tracesFileCreationWaitTimeout: Duration
  )
}

data class StateManagerClientConfig(
  val version: String,
  val endpoints: List<URL>,
  val requestLimitPerEndpoint: UInt,
  override val requestRetry: RequestRetryConfigTomlFriendly
) : RequestRetryConfigurable {
  init {
    require(requestLimitPerEndpoint > 0u) { "requestLimitPerEndpoint must be greater than 0" }
  }
}

data class BlobSubmissionConfig(
  // defaults to Keccak256("0x000000...")
  private val _genesisShnarf: String = "0x4f64fe1ce613546d34d666d8258c13c6296820fd13114d784203feb91276e838",
  val dbPollingInterval: Duration,
  val maxBlobsToReturn: Int,
  val proofSubmissionDelay: Duration,
  val priorityFeePerGasUpperBound: BigInteger,
  val priorityFeePerGasLowerBound: BigInteger,
  val maxBlobsToSubmitPerTick: Int = maxBlobsToReturn,
  // defaults to 6, not supported atm, preparatory work
  val targetBlobsToSendPerTransaction: Int = 6,
  override var disabled: Boolean = false
) : FeatureToggleable {
  val genesisShnarf: ByteArray = _genesisShnarf.decodeHex().assertIs32Bytes("genesisShnarf")

  init {
    require(maxBlobsToReturn > 0) { "maxBlobsToReturn must be greater than 0" }
    require(maxBlobsToSubmitPerTick >= 0) { "submissionLimit must be greater or equal to 0" }
    require(priorityFeePerGasUpperBound >= BigInteger.ZERO) { "priorityFeePerGasUpperBound must be at least 0" }
    require(priorityFeePerGasLowerBound >= BigInteger.ZERO) { "priorityFeePerGasLowerBound must be at least 0" }
    require(targetBlobsToSendPerTransaction in 1..6) {
      "targetBlobsToSendPerTransaction must be between 1 and 6, value=$targetBlobsToSendPerTransaction"
    }
  }

  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as BlobSubmissionConfig

    if (!_genesisShnarf.contentEquals(other._genesisShnarf)) return false
    if (dbPollingInterval != other.dbPollingInterval) return false
    if (maxBlobsToReturn != other.maxBlobsToReturn) return false
    if (proofSubmissionDelay != other.proofSubmissionDelay) return false
    if (priorityFeePerGasUpperBound != other.priorityFeePerGasUpperBound) return false
    if (priorityFeePerGasLowerBound != other.priorityFeePerGasLowerBound) return false
    if (disabled != other.disabled) return false

    return true
  }

  override fun hashCode(): Int {
    var result = genesisShnarf.contentHashCode()
    result = 31 * result + dbPollingInterval.hashCode()
    result = 31 * result + maxBlobsToReturn
    result = 31 * result + proofSubmissionDelay.hashCode()
    result = 31 * result + priorityFeePerGasUpperBound.hashCode()
    result = 31 * result + priorityFeePerGasLowerBound.hashCode()
    result = 31 * result + disabled.hashCode()
    return result
  }
}

data class AggregationFinalizationConfig(
  val dbPollingInterval: Duration,
  val maxAggregationsToFinalizePerTick: Int,
  val proofSubmissionDelay: Duration,
  override var disabled: Boolean = false
) : FeatureToggleable {
  init {
    require(maxAggregationsToFinalizePerTick > 0) {
      "maxAggregationsToFinalizePerIteration must be greater than 0"
    }
  }
}

data class DatabaseConfig(
  val host: String,
  val port: Int,
  val username: String,
  val password: Masked,
  val schema: String,
  val readPoolSize: Int,
  val readPipeliningLimit: Int,
  val transactionalPoolSize: Int
)

data class L1Config(
  val zkEvmContractAddress: String,
  val rpcEndpoint: URL,
  val finalizationPollingInterval: Duration,
  val newBatchPollingInterval: Duration,
  val blocksToFinalization: UInt,
  val gasLimit: BigInteger,
  val feeHistoryBlockCount: Int,
  val feeHistoryRewardPercentile: Double,
  val maxFeePerGasCap: BigInteger,
  val maxFeePerBlobGasCap: BigInteger,
  val gasPriceCapMultiplierForFinalization: Double,
  val earliestBlock: BigInteger,
  val sendMessageEventPollingInterval: Duration,
  val maxEventScrapingTime: Duration,
  val maxMessagesToCollect: UInt,
  val finalizedBlockTag: String,
  val blockRangeLoopLimit: UInt = 0U,
  @ConfigAlias("eth-fee-history-endpoint") private val _ethFeeHistoryEndpoint: URL?
) {
  val ethFeeHistoryEndpoint: URL
    get() = _ethFeeHistoryEndpoint ?: rpcEndpoint

  init {
    require(gasPriceCapMultiplierForFinalization > 0.0) {
      "gas price cap multiplier for finalization must be greater than 0.0." +
        " Value=$gasPriceCapMultiplierForFinalization"
    }
  }
}

data class L2Config(
  val messageServiceAddress: String,
  val rpcEndpoint: URL,
  val gasLimit: BigInteger,
  val maxFeePerGasCap: BigInteger,
  val feeHistoryBlockCount: UInt,
  val feeHistoryRewardPercentile: Double,
  val blocksToFinalization: UInt,
  val lastHashSearchWindow: UInt,
  val lastHashSearchMaxBlocksBack: UInt,
  val anchoringReceiptPollingInterval: Duration,
  val maxReceiptRetries: UInt
) {
  init {
    messageServiceAddress.assertIsValidAddress("messageServiceAddress")
  }
}

data class SignerConfig(
  val type: Type,
  val web3signer: Web3SignerConfig?,
  val web3j: Web3jConfig?
) {
  enum class Type {
    Web3j,
    Web3Signer
  }

  init {
    when (type) {
      Type.Web3Signer -> {
        if (web3signer == null) throw IllegalStateException("Signer $type configuration is null.")
      }

      Type.Web3j -> {
        if (web3j == null) throw IllegalStateException("Signer $type configuration is null.")
      }
    }
  }
}

data class Web3jConfig(
  val privateKey: Masked
)

data class Web3SignerConfig(
  val endpoint: String,
  val maxPoolSize: UInt,
  val keepAlive: Boolean,
  val publicKey: String
)

interface FeatureToggleable {
  val disabled: Boolean
  val enabled: Boolean
    get() = !disabled
}

data class MessageAnchoringServiceConfig(
  val pollingInterval: Duration,
  val maxMessagesToAnchor: UInt,
  override var disabled: Boolean = false
) : FeatureToggleable {
  init {
    require(maxMessagesToAnchor > 0u) { "maxMessagesToAnchor must be greater than 0" }
  }
}

data class DynamicGasPriceServiceConfig(
  val priceUpdateInterval: Duration,
  val feeHistoryBlockCount: Int,
  val feeHistoryRewardPercentile: Double,
  val baseFeeCoefficient: BigDecimal,
  val priorityFeeCoefficient: BigDecimal,
  val baseFeeBlobCoefficient: BigDecimal,
  val expectedBlobGas: BigDecimal,
  val blobSubmissionExpectedExecutionGas: BigDecimal,
  val gasPriceUpperBound: BigInteger,
  val gasPriceLowerBound: BigInteger,
  val gasPriceFixedCost: BigInteger,
  val gethGasPriceUpdateRecipients: List<URL>,
  val besuGasPriceUpdateRecipients: List<URL>,
  override val requestRetry: RequestRetryConfigTomlFriendly,
  @ConfigAlias("disabled") var _disabled: Boolean = false
) : FeatureToggleable, RequestRetryConfigurable {

  override val disabled: Boolean
    get() =
      _disabled || (gethGasPriceUpdateRecipients.isEmpty() && besuGasPriceUpdateRecipients.isEmpty())

  init {
    require(feeHistoryBlockCount > 0) { "feeHistoryBlockCount must be greater than 0" }
    require(gasPriceUpperBound >= gasPriceLowerBound) {
      "gasPriceUpperBound must be greater than or equal to gasPriceLowerBound"
    }
  }
}

data class L1DynamicGasPriceCapServiceConfig(
  val gasPriceCapCalculation: GasPriceCapCalculation,
  val feeHistoryFetcher: FeeHistoryFetcher,
  val feeHistoryStorage: FeeHistoryStorage,
  override var disabled: Boolean = false
) : FeatureToggleable {
  data class GasPriceCapCalculation(
    val adjustmentConstant: UInt,
    val blobAdjustmentConstant: UInt,
    val finalizationTargetMaxDelay: Duration,
    val baseFeePerGasPercentileWindow: Duration,
    val baseFeePerGasPercentileWindowLeeway: Duration,
    val baseFeePerGasPercentile: Double,
    val timeOfDayMultipliers: TimeOfDayMultipliers?
  ) {
    init {
      timeOfDayMultipliers?.also {
        val allTimeOfDayKeys = getAllTimeOfDayKeys()
        if (allTimeOfDayKeys != it.keys) {
          val missingKeys = allTimeOfDayKeys - it.keys
          val extraKeys = it.keys - allTimeOfDayKeys
          val errorMessage =
            "Invalid time of day multipliers: missing keys: " +
              "${missingKeys.joinToString(",", "[", "]")}, " +
              "unsupported keys=${extraKeys.joinToString(",", "[", "]")}"
          throw IllegalStateException(errorMessage)
        }

        it.entries.forEach { timeOfDayMultiplier ->
          require(timeOfDayMultiplier.value > 0.0) {
            throw IllegalStateException(
              "Each multiplier in timeOfDayMultipliers must be greater than 0.0." +
                " Key=${timeOfDayMultiplier.key} Value=${timeOfDayMultiplier.value}"
            )
          }
        }

        require(baseFeePerGasPercentile in 0.0..100.0) {
          "baseFeePerGasPercentile must be within 0.0 and 100.0." +
            " Value=$baseFeePerGasPercentile"
        }
      }
    }
  }

  data class FeeHistoryStorage(
    val storagePeriod: Duration
  ) {
    init {
      require(storagePeriod <= MAX_FEE_HISTORIES_STORAGE_PERIOD.toJavaDuration()) {
        "storagePeriod must be at most $MAX_FEE_HISTORIES_STORAGE_PERIOD days"
      }
    }
  }

  data class FeeHistoryFetcher(
    val fetchInterval: Duration,
    val maxBlockCount: UInt,
    val rewardPercentiles: List<Double>,
    val endpoint: URL?
  ) {
    init {
      require(
        maxBlockCount > 0U &&
          maxBlockCount <= MAX_FEE_HISTORY_BLOCK_COUNT
      ) {
        "maxBlockCount must be greater than 0 and " +
          "less than or equal to $MAX_FEE_HISTORY_BLOCK_COUNT"
      }

      require(
        rewardPercentiles.isNotEmpty() &&
          rewardPercentiles.size <= MAX_REWARD_PERCENTILES_SIZE
      ) {
        "rewardPercentiles must be a non-empty list with " +
          "maximum length as $MAX_REWARD_PERCENTILES_SIZE."
      }

      require(rewardPercentiles.zipWithNext().all { it.first < it.second }) {
        "rewardPercentiles must contain monotonically-increasing values"
      }

      rewardPercentiles.forEach { percentile ->
        require(percentile in 0.0..100.0) {
          "Each percentile in rewardPercentiles must be within 0.0 and 100.0." +
            " Value=$percentile"
        }
      }
    }
  }

  init {
    require(feeHistoryStorage.storagePeriod >= gasPriceCapCalculation.baseFeePerGasPercentileWindow) {
      "storagePeriod must be at least same length as" +
        " baseFeePerGasPercentileWindow=${gasPriceCapCalculation.baseFeePerGasPercentileWindow}." +
        " Value=${feeHistoryStorage.storagePeriod}"
    }

    require(
      gasPriceCapCalculation.baseFeePerGasPercentileWindow
        >= gasPriceCapCalculation.baseFeePerGasPercentileWindowLeeway
    ) {
      "baseFeePerGasPercentileWindow must be at least same length as" +
        " baseFeePerGasPercentileWindowLeeway=${gasPriceCapCalculation.baseFeePerGasPercentileWindowLeeway}." +
        " Value=${gasPriceCapCalculation.baseFeePerGasPercentileWindow}"
    }

    require(feeHistoryFetcher.rewardPercentiles.contains(gasPriceCapCalculation.baseFeePerGasPercentile)) {
      "rewardPercentiles must contain the given" +
        " baseFeePerGasPercentile=${gasPriceCapCalculation.baseFeePerGasPercentile}." +
        " Value=${feeHistoryFetcher.rewardPercentiles}"
    }
  }
}

data class SmartContractErrorCodesConfig(val smartContractErrors: SmartContractErrors)

data class GasPriceCapTimeOfDayMultipliersConfig(val gasPriceCapTimeOfDayMultipliers: TimeOfDayMultipliers)

data class Type2StateProofProviderConfig(
  val endpoints: List<URL>,
  override val requestRetry: RequestRetryConfigTomlFriendly
) : RequestRetryConfigurable

data class TracesLimitsConfigFile(val tracesLimits: TracesCounters)

data class CoordinatorConfig(
  val duplicatedLogsDebounceTime: Duration = Duration.ofSeconds(10),
  val zkGethTraces: ZkGethTraces,
  val prover: ProverConfig,
  val blobCompression: BlobCompressionConfig,
  val proofAggregation: AggregationConfig,
  val traces: TracesConfig,
  val type2StateProofProvider: Type2StateProofProviderConfig,
  val l1: L1Config,
  val l2: L2Config,
  val finalizationSigner: SignerConfig,
  val dataSubmissionSigner: SignerConfig,
  val blobSubmission: BlobSubmissionConfig,
  val aggregationFinalization: AggregationFinalizationConfig,
  val database: DatabaseConfig,
  val stateManager: StateManagerClientConfig,
  val conflation: ConflationConfig,
  val api: ApiConfig,
  val l2Signer: SignerConfig,
  val messageAnchoringService: MessageAnchoringServiceConfig,
  val dynamicGasPriceService: DynamicGasPriceServiceConfig,
  val l1DynamicGasPriceCapService: L1DynamicGasPriceCapServiceConfig,
  val eip4844SwitchL2BlockNumber: Long,
  val testL1Disabled: Boolean = false
) {
  init {
    if (testL1Disabled) {
      messageAnchoringService.disabled = true
      dynamicGasPriceService._disabled = true
      l1DynamicGasPriceCapService.disabled = true
    }
  }
}
