package net.consensys.zkevm.coordinator.app

import com.sksamuel.hoplite.ConfigAlias
import com.sksamuel.hoplite.Masked
import net.consensys.linea.traces.TracesCounters
import net.consensys.linea.traces.TracingModule
import tech.pegasys.teku.infrastructure.bytes.Bytes20
import java.math.BigDecimal
import java.math.BigInteger
import java.net.URL
import java.nio.file.Path
import java.time.Duration

data class ApiConfig(
  val observabilityPort: UInt
// @ConfigAlias("fork-choice-provider") val forkChoiceProvider: ForkChoiceStateApiConfig
)

data class ConflationConfig(
  val consistentNumberOfBlocksOnL1ToWait: Int,
  val conflationDeadline: Duration,
  val conflationDeadlineCheckInterval: Duration,
  val conflationDeadlineLastBlockConfirmationDelay: Duration,
  val totalLimitBytes: Int,
  val perBlockOverheadBytes: Int,
  val minBlockL1SizeBytes: Int,
  val blocksLimit: Long? = null,
  val forceStartingBlock: Long? = null,
  private var _tracesLimits: TracesCounters?
) {

  init {
    require(minBlockL1SizeBytes > 0) { "minBlockL1SizeBytes must be greater than 0" }
    require(perBlockOverheadBytes > 0) { "perBlockOverheadBytes must be greater than 0" }
    require(totalLimitBytes > minBlockL1SizeBytes) {
      "totalLimitBytes must be greater than minBlockL1SizeBytes: " +
        "totalLimitBytes=$totalLimitBytes, minBlockL1SizeBytes=$minBlockL1SizeBytes"
    }
    require(conflationDeadlineCheckInterval <= conflationDeadline) {
      "Clock ticker interval must be smaller than conflation deadline"
    }
    consistentNumberOfBlocksOnL1ToWait.let {
      require(it > 0) { "consistentNumberOfBlocksOnL1ToWait must be grater than 0" }
    }
    forceStartingBlock?.let { require(it > 0) { "forceStartingBlock must be grater than 0" } }
    blocksLimit?.let { require(it > 0) { "blocksLimit must be grater than 0" } }
  }

  val tracesLimits: TracesCounters
    get() = run {
      _tracesLimits?.also { limits ->
        if (limits.keys != TracingModule.allModules) {
          val missingModules = TracingModule.allModules - limits.keys
          val extraModules = limits.keys - TracingModule.allModules
          val errorMessage =
            "Invalid traces limits: Missing modules: " +
              "${missingModules.joinToString(",", "[", "]")}, " +
              "unsupported modules=${extraModules.joinToString(",", "[", "]")}"
          throw IllegalStateException(errorMessage)
        }
      } ?: throw IllegalStateException("Traces limits not defined!")
    }
}

data class ZkGethTraces(
  val ethApi: URL,
  val newBlockPollingInterval: Duration
)

data class SequencerConfig(
  val ethApi: URL
  // val version: String,
  // val engineApi: URL,
  // val suggestedFeeRecipient: String,
  // val blockInterval: Duration,
  // @ConfigAlias("jwtsecret-file") val jwtSecretFile: Path
)

data class ProverConfig(
  val version: String,
  val fsInputDirectory: Path,
  val fsOutputDirectory: Path,
  val fsPollingInterval: Duration,
  val fsInprogessProvingSuffixPattern: String,
  val timeout: Duration
)

data class TracesConfig(
  val version: String,
  val counters: FunctionalityEndpoint,
  val conflation: FunctionalityEndpoint,
  val fileManager: FileManager
) {
  data class FunctionalityEndpoint(
    val endpoints: List<URL>,
    val requestLimitPerEndpoint: UInt,
    val requestMaxRetries: UInt,
    val requestRetryInterval: Duration
  ) {
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
  val requestLimitPerEndpoint: UInt
) {
  init {
    require(requestLimitPerEndpoint > 0u) { "requestLimitPerEndpoint must be greater than 0" }
  }
}

data class BatchSubmissionConfig(
  val maxBatchesToSendPerTick: Int,
  val proofSubmissionDelay: Duration,
  override var disabled: Boolean = false
) : FeatureToggleable {
  init {
    require(maxBatchesToSendPerTick > 0) { "maxBatchesToSendPerTick must be greater than 0" }
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
  val maxFeePerGas: BigInteger,
  val earliestBlock: BigInteger,
  val sendMessageEventPollingInterval: Duration,
  val maxEventScrapingTime: Duration,
  val maxMessagesToCollect: UInt,
  val finalizedBlockTag: String,
  val blockRangeLoopLimit: UInt
)

data class L2Config(
  @ConfigAlias("message-service-address") private val _messageServiceAddress: String,
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
  // thi guarantees that no invalid address is set in the application
  val messageServiceAddress: Bytes20 = Bytes20.fromHexString(_messageServiceAddress)
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
  val pollingInterval: Duration,
  val feeHistoryBlockCount: Int,
  val feeHistoryRewardPercentile: Double,
  val baseFeeCoefficient: BigDecimal,
  val priorityFeeCoefficient: BigDecimal,
  val gasPriceCap: BigInteger,
  val minerGasPriceUpdateRecipients: List<URL>,
  override var disabled: Boolean = false
) : FeatureToggleable {

  init {
    require(feeHistoryBlockCount > 0) { "feeHistoryBlockCount must be greater than 0" }
  }
}

data class TracesLimitsConfigFile(val tracesLimits: TracesCounters)

data class CoordinatorConfig(
  val sequencer: SequencerConfig,
  val zkGethTraces: ZkGethTraces,
  val prover: ProverConfig,
  val traces: TracesConfig,
  val l1: L1Config,
  val l2: L2Config,
  val l1Signer: SignerConfig,
  val batchSubmission: BatchSubmissionConfig,
  val database: DatabaseConfig,
  val stateManager: StateManagerClientConfig,
  val conflation: ConflationConfig,
  val api: ApiConfig,
  val l2Signer: SignerConfig,
  val messageAnchoringService: MessageAnchoringServiceConfig,
  val dynamicGasPriceService: DynamicGasPriceServiceConfig,
  val testL1Disabled: Boolean = false
) {
  init {
    if (testL1Disabled) {
      messageAnchoringService.disabled = true
      batchSubmission.disabled = true
      dynamicGasPriceService.disabled = true
    }
  }
}
