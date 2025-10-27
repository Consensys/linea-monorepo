export enum LineaNativeYieldAutomationServiceMetrics {
  // Counter that increments each time a rebalance between L1MessageService and YieldProvider is performed
  // Labels:
  // i.) direction = STAKE | UNSTAKE
  // ii.) type = INITIAL | POST_REPORT
  RebalanceAmountTotal = "linea_native_yield_automation_service_rebalance_amount_total",

  // Counter that increments each time a partial beacon chain withdrawal is made
  // Single label - validator_pubkey
  ValidatorPartialUnstakeAmountTotal = "linea_native_yield_automation_service_validator_partial_unstake_amount_total",

  // Counter that increments each time a validator exit is made
  // Single label - validator_pubkey
  ValidatorExitTotal = "linea_native_yield_automation_service_validator_exit_total",

  // Counter that increments each time the _process() fn in YieldReportingProcessor is triggered
  // Single label `trigger` - VaultsReportDataUpdated_event vs timeout
  YieldReportingModeProcessorTriggerTotal = "linea_native_yield_automation_service_yield_reporting_mode_processor_trigger_total",

  // Counter that increment each time a vault accounting report is submitted
  // Single label `vault_address`
  LidoVaultAccountingReportSubmittedTotal = "linea_native_yield_automation_service_...",

  // Counter that increment each time YieldManager.reportYield is called
  // Single label `vault_address`
  ReportYieldTotal = "linea_native_yield_automation_service_report_yield_total",

  // Counter that increments by the yield amount reported
  // Single label `vault_address`
  ReportYieldAmountTotal = "linea_native_yield_automation_service_report_yield_amount_total",

  // Gauge representing outstanding negative yield as of the last yield report
  // Single label `vault_address`
  CurrentNegativeYieldLastReport = "linea_native_yield_automation_service_current_negative_yield",

  // Counter that increments by the node operator fees paid
  // Single label `vault_address`
  // N.B. Only accounts for payments by the automation service, but external actors can also trigger payment
  NodeOperatorFeesPaidTotal = "linea_native_yield_automation_service_node_operator_fees_paid_total",

  // Counter that increments by the node operator fees paid
  // Single label `vault_address`
  // N.B. Only accounts for payments by the automation service, but external actors can also trigger payment
  LiabilitiesPaidTotal = "linea_native_yield_automation_service_liabilities_paid_total",

  // Counter that increments by Lido fees paid
  // Single label `vault_address`
  // N.B. Only accounts for payments by the automation service, but external actors can also trigger payment
  LidoFeesPaidTotal = "linea_native_yield_automation_service_lido_fees_paid_total",

  // Counter representing tx fees paid by automation service (rounded down in gwei - bigint is not supported in Prometheus)
  // Single label `vault_address`
  TransactionFeesGwei = "linea_native_yield_automation_service_transaction_fees_gwei",

  // Counter that increments each time an operation mode runs.
  // Single label `mode`
  OperationModeExecutionTotal = "linea_native_yield_automation_service_operation_mode_execution_total",

  // Histogram that tracks time for each operation mode run.
  // Single label `mode`
  OperationModeExecutionDurationSeconds = "linea_native_yield_automation_service_operation_mode_execution_duration_seconds"
}
