package linea.coordinator.config.v2.toml

import linea.coordinator.config.v2.CoordinatorConfig
import linea.web3j.SmartContractErrors
import net.consensys.linea.ethereum.gaspricing.dynamiccap.TimeOfDayMultipliers
import net.consensys.linea.traces.TracesCountersV2
import net.consensys.linea.traces.TracingModuleV2

data class CoordinatorConfigFileToml(
  val defaults: DefaultsToml = DefaultsToml(),
  val protocol: ProtocolToml,
  val conflation: ConflationToml = ConflationToml(),
  val prover: ProverToml,
  val traces: TracesToml,
  val stateManager: StateManagerToml,
  val type2StateProofProvider: Type2StateProofManagerToml,
  val l1FinalizationMonitor: L1FinalizationMonitorConfigToml,
  val l1Submission: L1SubmissionConfigToml,
  val messageAnchoring: MessageAnchoringConfigToml,
  val l2NetworkGasPricing: L2NetworkGasPricingConfigToml,
  val database: DatabaseToml,
  val api: ApiConfigToml = ApiConfigToml(),
)

data class TracesLimitsConfigFileToml(
  val tracesLimits: Map<TracingModuleV2, UInt>,
)

data class GasPriceCapTimeOfDayMultipliersConfigFileToml(
  val gasPriceCapTimeOfDayMultipliers: TimeOfDayMultipliers,
)

data class SmartContractErrorCodesConfigFileToml(val smartContractErrors: SmartContractErrors)

data class CoordinatorConfigToml(
  val configs: CoordinatorConfigFileToml,
  val tracesLimitsV2: TracesLimitsConfigFileToml,
  val l1DynamicGasPriceCapTimeOfDayMultipliers: GasPriceCapTimeOfDayMultipliersConfigFileToml? = null,
  val smartContractErrors: SmartContractErrorCodesConfigFileToml? = null,
) {
  fun reified(): CoordinatorConfig {
    return CoordinatorConfig(
      protocol = configs.protocol.reified(),
      conflation = configs.conflation.reified(
        defaults = configs.defaults,
        tracesCountersLimitsV2 = TracesCountersV2(tracesLimitsV2.tracesLimits),
      ),
      proversConfig = this.configs.prover.reified(),
      traces = this.configs.traces.reified(),
      stateManager = this.configs.stateManager.reified(),
      type2StateProofProvider = this.configs.type2StateProofProvider.reified(),
      l1FinalizationMonitor = this.configs.l1FinalizationMonitor.reified(
        defaults = this.configs.defaults,
      ),
      l1Submission = this.configs.l1Submission.reified(
        l1DefaultEndpoint = this.configs.defaults.l1Endpoint,
        timeOfDayMultipliers = l1DynamicGasPriceCapTimeOfDayMultipliers
          ?.gasPriceCapTimeOfDayMultipliers
          ?: emptyMap(),
      ),
      messageAnchoring = this.configs.messageAnchoring.reified(
        l1DefaultEndpoint = this.configs.defaults.l1Endpoint,
        l2DefaultEndpoint = this.configs.defaults.l2Endpoint,
      ),
      l2NetworkGasPricing = this.configs.l2NetworkGasPricing.reified(
        l1DefaultEndpoint = this.configs.defaults.l1Endpoint,
      ),
      database = this.configs.database.reified(),
      api = this.configs.api.reified(),
      smartContractErrors = smartContractErrors?.smartContractErrors ?: emptyMap(),
    )
  }
}
