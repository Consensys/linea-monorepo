// NB - All amounts rounded down in gwei. Due to limitation that PromQL does not support bigint.

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

  // Counter that increment each time a vault accounting report is submitted
  // Single label `vault_address`
  LidoVaultAccountingReportSubmittedTotal = "linea_native_yield_automation_service_lido_vault_accounting_report_submitted_total",

  // Counter that increment each time YieldManager.reportYield is called
  // Single label `vault_address`
  ReportYieldTotal = "linea_native_yield_automation_service_report_yield_total",

  // Counter that increments by the yield amount reported
  // Single label `vault_address`
  ReportYieldAmountTotal = "linea_native_yield_automation_service_report_yield_amount_total",

  // Gauge representing cumulative yield reported from the YieldManager contract
  // Single label `vault_address`
  YieldReportedCumulative = "linea_native_yield_automation_service_yield_reported_cumulative",

  // Gauge representing outstanding negative yield from the last peeked yield report
  // Single label `vault_address`
  LastPeekedNegativeYieldReport = "linea_native_yield_automation_service_last_peeked_negative_yield_report",

  // Gauge representing positive yield amount from the last peeked yield report
  // Single label `vault_address`
  LastPeekedPositiveYieldReport = "linea_native_yield_automation_service_last_peeked_positive_yield_report",

  // Gauge representing settleable Lido protocol fees from the last query
  // Single label `vault_address`
  LastSettleableLidoFees = "linea_native_yield_automation_service_last_settleable_lido_fees",

  // Gauge representing timestamp from the latest vault report
  // Single label `vault_address`
  LastVaultReportTimestamp = "linea_native_yield_automation_service_last_vault_report_timestamp",

  // Gauge representing total pending partial withdrawals in gwei
  LastTotalPendingPartialWithdrawalsGwei = "linea_native_yield_automation_service_last_total_pending_partial_withdrawals_gwei",

  // Gauge representing pending partial withdrawal queue amount in gwei
  // Labels:
  // i.) `pubkey` - Validator public key
  // ii.) `withdrawable_epoch` - Epoch when withdrawal becomes available
  PendingPartialWithdrawalQueueAmountGwei = "linea_native_yield_automation_service_pending_partial_withdrawal_queue_amount_gwei",

  // Gauge representing pending partial withdrawal queue withdrawable epoch
  // Single label `pubkey`
  PendingPartialWithdrawalQueueWithdrawableEpoch = "linea_native_yield_automation_service_pending_partial_withdrawal_queue_withdrawable_epoch",

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

  // Counter that increments each time an operation mode completes execution.
  // Labels:
  // i.) `mode`
  // ii.) `status` - OperationModeExecutionStatus.Success | OperationModeExecutionStatus.Failure
  OperationModeExecutionTotal = "linea_native_yield_automation_service_operation_mode_execution_total",

  // Histogram that tracks time for each operation mode run.
  // Single label `mode`
  OperationModeExecutionDurationSeconds = "linea_native_yield_automation_service_operation_mode_execution_duration_seconds",
}

export enum OperationTrigger {
  VAULTS_REPORT_DATA_UPDATED_EVENT = "VAULTS_REPORT_DATA_UPDATED_EVENT",
  TIMEOUT = "TIMEOUT",
}

export enum OperationModeExecutionStatus {
  Success = "success",
  Failure = "failure",
}
