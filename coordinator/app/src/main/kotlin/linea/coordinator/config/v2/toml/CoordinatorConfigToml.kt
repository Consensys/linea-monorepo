package linea.coordinator.config.v2.toml

import linea.domain.BlockParameter
import java.net.URL
import kotlin.time.Duration
import kotlin.time.Duration.Companion.seconds

data class DefaultsToml(
  val l1Endpoint: URL,
  val l2Endpoint: URL
)

data class ProtocolToml(
  val genesis: Genesis,
  val l1: LayerConfig,
  val l2: LayerConfig
) {
  data class Genesis(
    val genesisStateRootHash: ByteArray,
    val genesisShnarf: ByteArray
  ) {

    override fun equals(other: Any?): Boolean {
      if (this === other) return true
      if (javaClass != other?.javaClass) return false

      other as Genesis

      if (!genesisStateRootHash.contentEquals(other.genesisStateRootHash)) return false
      if (!genesisShnarf.contentEquals(other.genesisShnarf)) return false

      return true
    }

    override fun hashCode(): Int {
      var result = genesisStateRootHash.contentHashCode()
      result = 31 * result + genesisShnarf.contentHashCode()
      return result
    }
  }

  data class LayerConfig(
    val contractAddress: String,
    val contractDeploymentBlockNumber: BlockParameter?
  )
}

data class ConflationToml(
  val disabled: Boolean = false,
  val blocksLimit: Int? = null,
  val conflationCalculatorVersion: String? = null,
  val conflationDeadline: Duration? = null,
  // val conflationDeadlineCheckInterval: Duration = Duration.ofSeconds(30),
  // val conflationDeadlineLastBlockConfirmationDelay: Duration = Duration.ofSeconds(24),
  val consistentNumberOfBlocksOnL1ToWait: Int = 32, // 1 epoch
  val l2FetchBlocksLimit: UInt? = null,
  val l2BlockCreationEndpoint: URL? = null,
  val l2LogsEndpoint: URL? = null,
  val blobCompression: BlobCompressionToml,
  val proofAggregation: ProofAggregationToml
) {
  data class BlobCompressionToml(
    val blobSizeLimit: ULong = 102400u,
    val handlerPollingInterval: Duration = 1.seconds,
    val batchesLimit: UInt? = null
  )

  data class ProofAggregationToml(
    val proofsLimit: ULong = 300u,
    val deadline: Duration? = null,
    val deadlineCheckInterval: Duration? = 30.seconds,
    val coordinatorPollingInterval: Duration? = 3.seconds,
    val targetEndBlocks: List<ULong>? = null
  )
}

data class ProverToml(
  val version: String,
  val fsInprogressRequestWritingSuffix: String = ".inprogress_coordinator_writing",
  val fsInprogressProvingSuffixPattern: String = "\\.inprogress\\.prover.*",
  val fsPollingInterval: Duration = 15.seconds,
  val fsPollingTimeout: Duration? = null,
  val execution: ProverDirectoriesToml,
  val blobCompression: ProverDirectoriesToml,
  val proofAggregation: ProverDirectoriesToml,
  val new: ProverToml? = null
) {
  data class ProverDirectoriesToml(
    val fsRequestsDirectory: String,
    val fsResponsesDirectory: String
  )
}

data class TracesToml(
  val expectedTracesApiVersion: String,
  val counters: ClientApiConfigToml,
  val conflation: ClientApiConfigToml,
  val switchBlockNumberInclusive: UInt? = null,
  val new: TracesToml? = null
) {
  data class ClientApiConfigToml(
    val endpoints: List<URL>,
    val requestLimitPerEndpoint: UInt? = null,
    val requestRetries: RequestRetriesToml? = null
  ) {
    override fun toString(): String {
      return "ClientApiConfigToml(" +
        "endpoints=$endpoints, " +
        "requestLimitPerEndpoint=$requestLimitPerEndpoint, " +
        "requestRetries=$requestRetries" +
        ")"
    }
  }
}

data class StateManagerToml(
  val version: String,
  val endpoints: List<URL>,
  val requestLimitPerEndpoint: UInt? = null,
  val requestRetries: RequestRetriesToml? = null
)

data class Type2StateProofManagerToml(
  val disabled: Boolean = false,
  val endpoints: List<URL>,
  val requestRetries: RequestRetriesToml? = null,
  val l1QueryBlockTag: BlockParameter? = null,
  val l1PollingInterval: Duration = 6.seconds
)

data class L1FinalizationMonitor(
  val l1Endpoint: URL?,
  val l2Endpoint: URL?,
  val l1PollingInterval: Duration = 6.seconds,
  val l1QueryBlockTag: BlockParameter? = null,
  val requestRetries: RequestRetriesToml
)

data class L1SubmissionToml(
  val disabled: Boolean,
  val dynamicGasPriceCap: DynamicGasPriceCapToml,
  val fallbackGasPrice: FallbackGasPriceToml,
  val blob: BlobSubmissionConfigToml,
  val aggregation: AggregationSubmissionToml
) {
  data class DynamicGasPriceCapToml(
    val disabled: Boolean,
    val gasPriceCapCalculation: GasPriceCapCalculationToml,
    val feeHistoryFetcher: FeeHistoryFetcherConfig,
    val feeHistoryStorage: FeeHistoryStorageConfig
  ) {
    data class GasPriceCapCalculationToml(
      val adjustmentConstant: Int,
      val blobAdjustmentConstant: Int,
      val finalizationTargetMaxDelay: Duration,
      val baseFeePerGasPercentileWindow: Duration,
      val baseFeePerGasPercentileWindowLeeway: Duration,
      val baseFeePerGasPercentile: UInt,
      val gasPriceCapsCheckCoefficient: Double
    )

    data class FeeHistoryFetcherConfig(
      val fetchInterval: Duration,
      val maxBlockCount: UInt,
      val rewardPercentiles: List<UInt>
    )

    data class FeeHistoryStorageConfig(
      val storagePeriod: Duration
    )
  }

  data class FallbackGasPriceToml(
    val feeHistoryBlockCount: UInt,
    val feeHistoryRewardPercentile: UInt
  )

  data class GasConfigToml(
    val gasLimit: UInt,
    val maxFeePerGasCap: ULong,
    val maxFeePerBlobGasCap: ULong? = null,
    val fallback: FallbackGasConfig
  ) {
    data class FallbackGasConfig(
      val priorityFeePerGasUpperBound: ULong,
      val priorityFeePerGasLowerBound: ULong
    )
  }

  data class BlobSubmissionConfigToml(
    val disabled: Boolean,
    val endpoint: URL,
    val submissionDelay: Duration,
    val submissionTickInterval: Duration,
    val maxSubmissionTransactionsPerTick: UInt,
    val targetBlobsPerTransaction: UInt,
    val dbMaxBlobsToReturn: UInt,
    val gas: GasConfigToml,
    val signer: SignerConfigToml
  )

  data class AggregationSubmissionToml(
    val disabled: Boolean,
    val endpoint: URL,
    val submissionDelay: Duration,
    val submissionTickInterval: Duration,
    val maxSubmissionsPerTick: UInt,
    val gas: GasConfigToml,
    val signer: SignerConfigToml
  )
}

data class ApiConfigToml(
  val observabilityPort: UInt = 9545u
)

data class CoordinatorConfigFileToml(
  val defaults: DefaultsToml?,
  val protocol: ProtocolToml,
  val conflation: ConflationToml,
  val prover: ProverToml,
  val traces: TracesToml,
  val stateManager: StateManagerToml,
  val type2StateProofProvider: Type2StateProofManagerToml,
  val l1FinalizationMonitor: L1FinalizationMonitor,
  val l1Submission: L1SubmissionToml,
  val messageAnchoring: MessageAnchoringConfigToml,
  val l2NetworkGasPricing: L2NetworkGasPricingConfigToml,
  val database: DataBaseToml,
  val api: ApiConfigToml
)
