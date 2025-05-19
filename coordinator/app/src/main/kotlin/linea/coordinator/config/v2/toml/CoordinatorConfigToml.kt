package linea.coordinator.config.v2.toml

import linea.domain.BlockParameter
import java.net.URI
import kotlin.time.Duration
import kotlin.time.Duration.Companion.seconds

data class Protocol(
    val genesis: Genesis,
    val l1: LayerConfig,
    val l2: LayerConfig,
) {
    data class Genesis(
        val genesisStateRootHash: ByteArray,
        val genesisShnarf: ByteArray,
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
        val contractDeploymentBlockNumber: BlockParameter.BlockNumber?,
    )
}

data class ConflationToml(
    val disabled: Boolean = false,
    val blocksLimit: Int? = null,
    val conflationCalculatorVersions: String? = null,
    val conflationDeadline: Duration? = null,
    // val conflationDeadlineCheckInterval: Duration = Duration.ofSeconds(30),
    // val conflationDeadlineLastBlockConfirmationDelay: Duration = Duration.ofSeconds(24),
    val consistentNumberOfBlocksOnL1ToWait: Int = 32, // 1 epoch
    val fetchBlocksLimit: Int? = null,
    val l2BlockCreationEndpoint: URI? = null,
    val l2LogsEndpoint: URI? = null,
    val blobCompression: BlobCompressionToml,
    val proofAggregation: ProofAggregationToml,
) {
    // data class BatchesToml(
    //     val blocksLimit: UInt? = null,
    //     val conflationDeadline: Duration? = null,
    // )

    data class BlobCompressionToml(
        val blobSizeLimit: ULong = 102400u,
        val handlerPollingInterval: Duration = 1.seconds,
        val batchesLimit: UInt? = null,
    )

    data class ProofAggregationToml(
        val aggregationProofLimit: ULong = 300u,
        val aggregationDeadline: Duration? = null,
        val aggregationDeadlineCheckInterval: Duration? = 30.seconds,
        val aggregationCoordinatorPollingInterval: Duration? = 3.seconds,
        val aggregationTragetEndBlockNumbers: List<ULong>? = null,
    )
}

data class ProverToml(
    val version: String,
    val fsInprogressRequestWritingSuffix: String = ".inprogress_coordinator_writing",
    val fsInprogressProvingSuffixPattern: String = "\\.inprogress\\.prover.*",
    val fsPollingInterval: Duration = 15.seconds,
    val fsPollingTimeout: Duration?,
    val execution: ProverDirectoriesToml,
    val compression: ProverDirectoriesToml,
    val aggregation: ProverDirectoriesToml,
    val new: ProverToml? = null,
) {
    data class ProverDirectoriesToml(
        val fsRequestsDirectory: String,
        val fsResponsesDirectory: String
    )
}


data class RequestRetriesToml(
    val maxAttempts: UInt? = null,
    val timeout: Duration? = null,
    val backoffDelay: Duration = 1.seconds,
    val failuresWarningThreshold: UInt? = null,
)

open class ClientApiConfigToml(
    open val endpoints: List<URI>,
    open val requestLimitPerEndpoint: Unit? = null,
    open val requestRetries: RequestRetriesToml? = null,
) {
    override fun toString(): String {
        return "ClientApiConfigToml(" +
            "endpoints=$endpoints, " +
            "requestLimitPerEndpoint=$requestLimitPerEndpoint, " +
            "requestRetries=$requestRetries" +
            ")"
    }
}

data class TracesToml(
    val rawExecutionTracesVersion: String,
    val expectedTracesApiVersion: String,
    val counters: ClientApiConfigToml,
    val conflation: ClientApiConfigToml
)

class StateManagerToml(
    val version: String,
    override val endpoints: List<URI>,
    override val requestLimitPerEndpoint: Unit? = null,
    override val requestRetries: RequestRetriesToml? = null,
): ClientApiConfigToml(endpoints, requestLimitPerEndpoint, requestRetries)


class Type2StateProofManagerToml(
    val disabled: Boolean = false,
    override val endpoints: List<URI>,
    override val requestRetries: RequestRetriesToml? = null,
    val l1HighestQueryBlockTag: BlockParameter? = null,
    val l1PollingInterval: Duration = 6.seconds,
): ClientApiConfigToml(endpoints, null, requestRetries)


data class L1FinalizationMonitor(
    val l1Endpoint: URI?,
    val l2Endpoint: URI?,
    val l1PollingInterval: Duration = 6.seconds,
    val l1HighestQueryBlockTag: BlockParameter? = null,
    val requestRetries: RequestRetriesToml
)


data class L1SubmissionToml(
    val disabled: Boolean,
    val dynamicGasPriceCap: DynamicGasPriceCapToml,
    val fallbackGasPrice: FallbackGasPriceToml,
    val blob: BlobConfig,
    val aggregation: AggregationToml
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
            val finalizationTargetMaxDelay: String,
            val baseFeePerGasPercentileWindow: String,
            val baseFeePerGasPercentileWindowLeeway: String,
            val baseFeePerGasPercentile: Int,
            val gasPriceCapsCheckCoefficient: Double
        )

        data class FeeHistoryFetcherConfig(
            val fetchInterval: String,
            val maxBlockCount: Int,
            val rewardPercentiles: List<Int>
        )

        data class FeeHistoryStorageConfig(
            val storagePeriod: String
        )
    }

    data class FallbackGasPriceToml(
        val feeHistoryBlockCount: Int,
        val feeHistoryRewardPercentile: Int
    )

    data class BlobConfig(
        val disabled: Boolean,
        val endpoint: String,
        val submissionDelay: String,
        val submissionTickInterval: String,
        val maxSubmissionsPerTick: Int,
        val dbMaxBlobsToReturn: Int,
        val gas: GasConfig,
        val signer: SignerConfig
    ) {
        data class GasConfig(
            val gasLimit: Int,
            val maxFeePerGasCap: Long,
            val maxFeePerBlobGasCap: Long,
            val fallback: FallbackGasConfig
        ) {
            data class FallbackGasConfig(
                val priorityFeePerGasUpperBound: Long,
                val priorityFeePerGasLowerBound: Long
            )
        }

        data class SignerConfig(
            val type: String,
            val web3j: Web3jConfig?,
            val web3signer: Web3signerConfig?
        ) {
            data class Web3jConfig(
                val privateKey: String
            )

            data class Web3signerConfig(
                val endpoint: String,
                val maxPoolSize: Int,
                val keepAlive: Boolean,
                val publicKey: String
            )
        }
    }

    data class AggregationToml(
        val disabled: Boolean,
        val endpoint: String,
        val submissionDelay: String,
        val submissionTickInterval: String,
        val maxSubmissionsPerTick: Int,
        val gas: GasConfig,
        val signer: SignerConfig
    ) {
        data class GasConfig(
            val gasLimit: Int,
            val maxFeePerGasCap: Long,
            val fallback: FallbackGasConfig
        ) {
            data class FallbackGasConfig(
                val priorityFeePerGasUpperBound: Long,
                val priorityFeePerGasLowerBound: Long
            )
        }

        data class SignerConfig(
            val type: String,
            val web3j: Web3jConfig?,
            val web3signer: Web3signerConfig?
        ) {
            data class Web3jConfig(
                val privateKey: String
            )

            data class Web3signerConfig(
                val endpoint: String,
                val maxPoolSize: Int,
                val keepAlive: Boolean,
                val publicKey: String
            )
        }
    }
}


class CoordinatorConfigToml {
}
