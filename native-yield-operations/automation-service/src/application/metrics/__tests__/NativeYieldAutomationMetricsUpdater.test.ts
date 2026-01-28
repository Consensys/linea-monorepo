import { jest } from "@jest/globals";
import { IMetricsService } from "@consensys/linea-shared-utils";
import { NativeYieldAutomationMetricsUpdater } from "../NativeYieldAutomationMetricsUpdater.js";
import {
  LineaNativeYieldAutomationServiceMetrics,
  OperationModeExecutionStatus,
} from "../../../core/metrics/LineaNativeYieldAutomationServiceMetrics.js";
import { RebalanceDirection } from "../../../core/entities/RebalanceRequirement.js";
import { OperationMode } from "../../../core/enums/OperationModeEnums.js";
import { Address, Hex } from "viem";

const createMetricsServiceMock = (): jest.Mocked<IMetricsService<LineaNativeYieldAutomationServiceMetrics>> =>
  ({
    getRegistry: jest.fn(),
    createCounter: jest.fn(),
    createGauge: jest.fn(),
    incrementCounter: jest.fn(),
    setGauge: jest.fn(),
    incrementGauge: jest.fn(),
    decrementGauge: jest.fn(),
    getGaugeValue: jest.fn(),
    getCounterValue: jest.fn(),
    createHistogram: jest.fn(),
    addValueToHistogram: jest.fn(),
    getHistogramMetricsValues: jest.fn(),
  }) as unknown as jest.Mocked<IMetricsService<LineaNativeYieldAutomationServiceMetrics>>;

describe("NativeYieldAutomationMetricsUpdater", () => {
  const validatorPubkey = "0xvalidator" as Hex;
  const vaultAddress = "0xvault" as Address;

  it("registers all metrics on construction", () => {
    const metricsService = createMetricsServiceMock();

    // Constructing should immediately register all metrics
    new NativeYieldAutomationMetricsUpdater(metricsService);

    expect(metricsService.createCounter).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.RebalanceAmountTotal,
      "Total rebalance amount between L1MessageService and YieldProvider",
      ["direction", "type"],
    );
    expect(metricsService.createCounter).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.ValidatorPartialUnstakeAmountTotal,
      "Total amount partially unstaked per validator",
      ["validator_pubkey"],
    );
    expect(metricsService.createCounter).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.ValidatorExitTotal,
      "Total validator exits initiated by automation",
      ["validator_pubkey"],
    );
    expect(metricsService.createGauge).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.ValidatorStakedAmountGwei,
      "Amount staked in a validator in gwei",
      ["pubkey"],
    );
    expect(metricsService.createCounter).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.LidoVaultAccountingReportSubmittedTotal,
      "Accounting reports submitted to Lido per vault",
      ["vault_address"],
    );
    expect(metricsService.createCounter).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.ReportYieldTotal,
      "Yield reports submitted to YieldManager per vault",
      ["vault_address"],
    );
    expect(metricsService.createGauge).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.LastPeekedNegativeYieldReport,
      "Outstanding negative yield from the last peeked yield report",
      ["vault_address"],
    );
    expect(metricsService.createGauge).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.LastPeekedPositiveYieldReport,
      "Positive yield amount from the last peeked yield report",
      ["vault_address"],
    );
    expect(metricsService.createGauge).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.LastSettleableLidoFees,
      "Settleable Lido protocol fees from the last query",
      ["vault_address"],
    );
    expect(metricsService.createGauge).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.LastVaultReportTimestamp,
      "Timestamp from the latest vault report",
      ["vault_address"],
    );
    expect(metricsService.createGauge).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.YieldReportedCumulative,
      "Cumulative yield reported from the YieldManager contract",
      ["vault_address"],
    );
    expect(metricsService.createGauge).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.LstLiabilityPrincipalGwei,
      "LST liability principal from the YieldManager contract",
      ["vault_address"],
    );
    expect(metricsService.createGauge).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.LastReportedNegativeYield,
      "Last reported negative yield from the YieldManager contract",
      ["vault_address"],
    );
    expect(metricsService.createGauge).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.LidoLstLiabilityGwei,
      "Lido LST liability in gwei from Lido accounting reports",
      ["vault_address"],
    );
    expect(metricsService.createGauge).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.LastTotalPendingPartialWithdrawalsGwei,
      "Total pending partial withdrawals in gwei",
      [],
    );
    expect(metricsService.createGauge).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.LastTotalValidatorBalanceGwei,
      "Total validator balance in gwei",
      [],
    );
    expect(metricsService.createGauge).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.LastTotalPendingDepositGwei,
      "Total pending deposits in gwei",
      [],
    );
    expect(metricsService.createGauge).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.PendingPartialWithdrawalQueueAmountGwei,
      "Pending partial withdrawal queue amount in gwei",
      ["pubkey", "withdrawable_epoch"],
    );
    expect(metricsService.createGauge).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.PendingExitQueueAmountGwei,
      "Pending exit queue amount in gwei",
      ["pubkey", "exit_epoch", "slashed"],
    );
    expect(metricsService.createGauge).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.LastTotalPendingExitGwei,
      "Total pending exit amount in gwei",
      [],
    );
    expect(metricsService.createGauge).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.PendingFullWithdrawalQueueAmountGwei,
      "Pending full withdrawal queue amount in gwei",
      ["pubkey", "withdrawable_epoch", "slashed"],
    );
    expect(metricsService.createGauge).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.LastTotalPendingFullWithdrawalGwei,
      "Total pending full withdrawal amount in gwei",
      [],
    );
    expect(metricsService.createCounter).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.NodeOperatorFeesPaidTotal,
      "Node operator fees paid by automation per vault",
      ["vault_address"],
    );
    expect(metricsService.createCounter).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.LiabilitiesPaidTotal,
      "Liabilities paid by automation per vault",
      ["vault_address"],
    );
    expect(metricsService.createCounter).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.LidoFeesPaidTotal,
      "Lido fees paid by automation per vault",
      ["vault_address"],
    );
    expect(metricsService.createCounter).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.OperationModeExecutionTotal,
      "Operation mode executions grouped by mode and status",
      ["mode", "status"],
    );
    expect(metricsService.createHistogram).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.OperationModeExecutionDurationSeconds,
      [1, 5, 10, 30, 60, 120, 180, 300, 600, 900, 1200],
      "Operation mode execution duration in seconds",
      ["mode"],
    );
  });

  describe("recordRebalance", () => {
    it("increments counter when amount is positive", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.recordRebalance(RebalanceDirection.STAKE, 42);

      expect(metricsService.incrementCounter).toHaveBeenCalledTimes(1);
      expect(metricsService.incrementCounter).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.RebalanceAmountTotal,
        { direction: RebalanceDirection.STAKE },
        42,
      );
    });

    it("increments counter when amount is zero", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.recordRebalance(RebalanceDirection.UNSTAKE, 0);

      expect(metricsService.incrementCounter).toHaveBeenCalledTimes(1);
      expect(metricsService.incrementCounter).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.RebalanceAmountTotal,
        { direction: RebalanceDirection.UNSTAKE },
        0,
      );
    });

    it("does not increment counter when amount is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.recordRebalance(RebalanceDirection.UNSTAKE, -10);

      expect(metricsService.incrementCounter).not.toHaveBeenCalled();
    });

    it("increments counter when direction is NONE", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.recordRebalance(RebalanceDirection.NONE, 0);

      expect(metricsService.incrementCounter).toHaveBeenCalledTimes(1);
      expect(metricsService.incrementCounter).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.RebalanceAmountTotal,
        { direction: RebalanceDirection.NONE },
        0,
      );
    });
  });

  describe("addValidatorPartialUnstakeAmount", () => {
    it("increments counter when amount is positive", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.addValidatorPartialUnstakeAmount(validatorPubkey, 100);

      expect(metricsService.incrementCounter).toHaveBeenCalledTimes(1);
      expect(metricsService.incrementCounter).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.ValidatorPartialUnstakeAmountTotal,
        { validator_pubkey: validatorPubkey },
        100,
      );
    });

    it("does not increment when amount is non-positive", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.addValidatorPartialUnstakeAmount(validatorPubkey, 0);

      expect(metricsService.incrementCounter).not.toHaveBeenCalled();
    });
  });

  describe("incrementValidatorExit", () => {
    it("defaults count to 1", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.incrementValidatorExit(validatorPubkey);

      expect(metricsService.incrementCounter).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.ValidatorExitTotal,
        { validator_pubkey: validatorPubkey },
        1,
      );
    });

    it("does not increment when count is non-positive", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.incrementValidatorExit(validatorPubkey, 0);

      expect(metricsService.incrementCounter).not.toHaveBeenCalled();
    });
  });

  describe("setValidatorStakedAmountGwei", () => {
    it("sets gauge when value is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setValidatorStakedAmountGwei(validatorPubkey, 32000000000);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.ValidatorStakedAmountGwei,
        { pubkey: validatorPubkey },
        32000000000,
      );
    });

    it("does not set gauge when value is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setValidatorStakedAmountGwei(validatorPubkey, -1);

      expect(metricsService.setGauge).not.toHaveBeenCalled();
    });

    it("sets gauge when value is zero", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setValidatorStakedAmountGwei(validatorPubkey, 0);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.ValidatorStakedAmountGwei,
        { pubkey: validatorPubkey },
        0,
      );
    });
  });

  it("increments accounting and report counters", () => {
    const metricsService = createMetricsServiceMock();
    const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
    jest.clearAllMocks();

    updater.incrementLidoVaultAccountingReport(vaultAddress);
    updater.incrementReportYield(vaultAddress);

    expect(metricsService.incrementCounter).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.LidoVaultAccountingReportSubmittedTotal,
      { vault_address: vaultAddress },
    );
    expect(metricsService.incrementCounter).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.ReportYieldTotal,
      { vault_address: vaultAddress },
    );
  });

  describe("setLastPeekedNegativeYieldReport", () => {
    it("sets gauge when value is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLastPeekedNegativeYieldReport(vaultAddress, 123);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.LastPeekedNegativeYieldReport,
        { vault_address: vaultAddress },
        123,
      );
    });

    it("does not set gauge when value is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLastPeekedNegativeYieldReport(vaultAddress, -1);

      expect(metricsService.setGauge).not.toHaveBeenCalled();
    });
  });

  describe("setLastPeekedPositiveYieldReport", () => {
    it("sets gauge when value is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLastPeekedPositiveYieldReport(vaultAddress, 456);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.LastPeekedPositiveYieldReport,
        { vault_address: vaultAddress },
        456,
      );
    });

    it("does not set gauge when value is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLastPeekedPositiveYieldReport(vaultAddress, -1);

      expect(metricsService.setGauge).not.toHaveBeenCalled();
    });
  });

  describe("setLastSettleableLidoFees", () => {
    it("sets gauge when value is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLastSettleableLidoFees(vaultAddress, 789);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.LastSettleableLidoFees,
        { vault_address: vaultAddress },
        789,
      );
    });

    it("does not set gauge when value is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLastSettleableLidoFees(vaultAddress, -1);

      expect(metricsService.setGauge).not.toHaveBeenCalled();
    });
  });

  describe("setLastVaultReportTimestamp", () => {
    it("sets gauge when timestamp is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      const timestamp = 1704067200; // Unix timestamp
      updater.setLastVaultReportTimestamp(vaultAddress, timestamp);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.LastVaultReportTimestamp,
        { vault_address: vaultAddress },
        timestamp,
      );
    });

    it("does not set gauge when timestamp is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLastVaultReportTimestamp(vaultAddress, -1);

      expect(metricsService.setGauge).not.toHaveBeenCalled();
    });
  });

  describe("setYieldReportedCumulative", () => {
    it("sets gauge when value is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setYieldReportedCumulative(vaultAddress, 1000);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.YieldReportedCumulative,
        { vault_address: vaultAddress },
        1000,
      );
    });

    it("does not set gauge when value is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setYieldReportedCumulative(vaultAddress, -1);

      expect(metricsService.setGauge).not.toHaveBeenCalled();
    });
  });

  describe("setLstLiabilityPrincipalGwei", () => {
    it("sets gauge when value is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLstLiabilityPrincipalGwei(vaultAddress, 1000);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.LstLiabilityPrincipalGwei,
        { vault_address: vaultAddress },
        1000,
      );
    });

    it("does not set gauge when value is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLstLiabilityPrincipalGwei(vaultAddress, -1);

      expect(metricsService.setGauge).not.toHaveBeenCalled();
    });
  });

  describe("setLastReportedNegativeYield", () => {
    it("sets gauge when value is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLastReportedNegativeYield(vaultAddress, 1000);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.LastReportedNegativeYield,
        { vault_address: vaultAddress },
        1000,
      );
    });

    it("does not set gauge when value is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLastReportedNegativeYield(vaultAddress, -1);

      expect(metricsService.setGauge).not.toHaveBeenCalled();
    });
  });

  describe("setLidoLstLiabilityGwei", () => {
    it("sets gauge when value is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLidoLstLiabilityGwei(vaultAddress, 1000);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.LidoLstLiabilityGwei,
        { vault_address: vaultAddress },
        1000,
      );
    });

    it("does not set gauge when value is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLidoLstLiabilityGwei(vaultAddress, -1);

      expect(metricsService.setGauge).not.toHaveBeenCalled();
    });
  });

  describe("setLastTotalPendingPartialWithdrawalsGwei", () => {
    it("sets gauge when value is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLastTotalPendingPartialWithdrawalsGwei(1000);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.LastTotalPendingPartialWithdrawalsGwei,
        {},
        1000,
      );
    });

    it("does not set gauge when value is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLastTotalPendingPartialWithdrawalsGwei(-1);

      expect(metricsService.setGauge).not.toHaveBeenCalled();
    });
  });

  describe("setLastTotalValidatorBalanceGwei", () => {
    it("sets gauge when value is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLastTotalValidatorBalanceGwei(1000);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.LastTotalValidatorBalanceGwei,
        {},
        1000,
      );
    });

    it("does not set gauge when value is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLastTotalValidatorBalanceGwei(-1);

      expect(metricsService.setGauge).not.toHaveBeenCalled();
    });
  });

  describe("setLastTotalPendingDepositGwei", () => {
    it("sets gauge when value is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLastTotalPendingDepositGwei(1000);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.LastTotalPendingDepositGwei,
        {},
        1000,
      );
    });

    it("does not set gauge when value is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLastTotalPendingDepositGwei(-1);

      expect(metricsService.setGauge).not.toHaveBeenCalled();
    });
  });

  describe("setPendingPartialWithdrawalQueueAmountGwei", () => {
    it("sets gauge when amount and epoch are non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingPartialWithdrawalQueueAmountGwei(validatorPubkey, 60001, 32000000000);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.PendingPartialWithdrawalQueueAmountGwei,
        { pubkey: validatorPubkey, withdrawable_epoch: "60001" },
        32000000000,
      );
    });

    it("converts withdrawableEpoch to string for label", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingPartialWithdrawalQueueAmountGwei(validatorPubkey, 12345, 1000);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.PendingPartialWithdrawalQueueAmountGwei,
        { pubkey: validatorPubkey, withdrawable_epoch: "12345" },
        1000,
      );
    });

    it("does not set gauge when amount is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingPartialWithdrawalQueueAmountGwei(validatorPubkey, 60001, -1);

      expect(metricsService.setGauge).not.toHaveBeenCalled();
    });

    it("does not set gauge when withdrawableEpoch is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingPartialWithdrawalQueueAmountGwei(validatorPubkey, -1, 1000);

      expect(metricsService.setGauge).not.toHaveBeenCalled();
    });

    it("sets gauge when amount is zero and epoch is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingPartialWithdrawalQueueAmountGwei(validatorPubkey, 60001, 0);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.PendingPartialWithdrawalQueueAmountGwei,
        { pubkey: validatorPubkey, withdrawable_epoch: "60001" },
        0,
      );
    });

    it("sets gauge when epoch is zero and amount is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingPartialWithdrawalQueueAmountGwei(validatorPubkey, 0, 1000);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.PendingPartialWithdrawalQueueAmountGwei,
        { pubkey: validatorPubkey, withdrawable_epoch: "0" },
        1000,
      );
    });
  });

  describe("setPendingDepositQueueAmountGwei", () => {
    it("sets gauge when amount and slot are non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingDepositQueueAmountGwei(validatorPubkey, 123456, 32000000000);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.PendingDepositQueueAmountGwei,
        { pubkey: validatorPubkey, slot: "123456" },
        32000000000,
      );
    });

    it("converts slot to string for label", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingDepositQueueAmountGwei(validatorPubkey, 789012, 1000);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.PendingDepositQueueAmountGwei,
        { pubkey: validatorPubkey, slot: "789012" },
        1000,
      );
    });

    it("does not set gauge when amount is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingDepositQueueAmountGwei(validatorPubkey, 123456, -1);

      expect(metricsService.setGauge).not.toHaveBeenCalled();
    });

    it("does not set gauge when slot is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingDepositQueueAmountGwei(validatorPubkey, -1, 1000);

      expect(metricsService.setGauge).not.toHaveBeenCalled();
    });

    it("sets gauge when amount is zero and slot is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingDepositQueueAmountGwei(validatorPubkey, 123456, 0);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.PendingDepositQueueAmountGwei,
        { pubkey: validatorPubkey, slot: "123456" },
        0,
      );
    });

    it("sets gauge when slot is zero and amount is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingDepositQueueAmountGwei(validatorPubkey, 0, 1000);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.PendingDepositQueueAmountGwei,
        { pubkey: validatorPubkey, slot: "0" },
        1000,
      );
    });
  });

  describe("vault payout counters", () => {
    const cases: Array<{
      metric: LineaNativeYieldAutomationServiceMetrics;
      invoke: (updater: NativeYieldAutomationMetricsUpdater, address: Address, amount: number) => void;
    }> = [
      {
        metric: LineaNativeYieldAutomationServiceMetrics.NodeOperatorFeesPaidTotal,
        invoke: (updater, address, amount) => updater.addNodeOperatorFeesPaid(address, amount),
      },
      {
        metric: LineaNativeYieldAutomationServiceMetrics.LiabilitiesPaidTotal,
        invoke: (updater, address, amount) => updater.addLiabilitiesPaid(address, amount),
      },
      {
        metric: LineaNativeYieldAutomationServiceMetrics.LidoFeesPaidTotal,
        invoke: (updater, address, amount) => updater.addLidoFeesPaid(address, amount),
      },
    ];

    cases.forEach(({ metric, invoke }) => {
      it(`increments ${metric} when amount is positive`, () => {
        const metricsService = createMetricsServiceMock();
        const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
        jest.clearAllMocks();

        invoke(updater, vaultAddress, 321);

        expect(metricsService.incrementCounter).toHaveBeenCalledWith(metric, { vault_address: vaultAddress }, 321);
      });

      it(`does not increment ${metric} when amount is non-positive`, () => {
        const metricsService = createMetricsServiceMock();
        const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
        jest.clearAllMocks();

        invoke(updater, vaultAddress, 0);

        expect(metricsService.incrementCounter).not.toHaveBeenCalled();
      });
    });
  });

  it("increments operation mode execution counter with default success status", () => {
    const metricsService = createMetricsServiceMock();
    const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
    jest.clearAllMocks();

    updater.incrementOperationModeExecution(OperationMode.OSSIFICATION_PENDING_MODE);

    expect(metricsService.incrementCounter).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.OperationModeExecutionTotal,
      { mode: OperationMode.OSSIFICATION_PENDING_MODE, status: OperationModeExecutionStatus.Success },
    );
  });

  it("increments operation mode execution counter with explicit success status", () => {
    const metricsService = createMetricsServiceMock();
    const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
    jest.clearAllMocks();

    updater.incrementOperationModeExecution(OperationMode.YIELD_REPORTING_MODE, OperationModeExecutionStatus.Success);

    expect(metricsService.incrementCounter).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.OperationModeExecutionTotal,
      { mode: OperationMode.YIELD_REPORTING_MODE, status: OperationModeExecutionStatus.Success },
    );
  });

  it("increments operation mode execution counter with failure status", () => {
    const metricsService = createMetricsServiceMock();
    const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
    jest.clearAllMocks();

    updater.incrementOperationModeExecution(OperationMode.YIELD_REPORTING_MODE, OperationModeExecutionStatus.Failure);

    expect(metricsService.incrementCounter).toHaveBeenCalledWith(
      LineaNativeYieldAutomationServiceMetrics.OperationModeExecutionTotal,
      { mode: OperationMode.YIELD_REPORTING_MODE, status: OperationModeExecutionStatus.Failure },
    );
  });

  describe("recordOperationModeDuration", () => {
    it("records duration when value is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.recordOperationModeDuration(OperationMode.OSSIFICATION_COMPLETE_MODE, 0);

      expect(metricsService.addValueToHistogram).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.OperationModeExecutionDurationSeconds,
        0,
        { mode: OperationMode.OSSIFICATION_COMPLETE_MODE },
      );
    });

    it("does not record when duration is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.recordOperationModeDuration(OperationMode.YIELD_REPORTING_MODE, -1);

      expect(metricsService.addValueToHistogram).not.toHaveBeenCalled();
    });
  });

  describe("setPendingExitQueueAmountGwei", () => {
    it("sets gauge when amount and exitEpoch are non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingExitQueueAmountGwei(validatorPubkey, 60001, 32000000000, false);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.PendingExitQueueAmountGwei,
        { pubkey: validatorPubkey, exit_epoch: "60001", slashed: "false" },
        32000000000,
      );
    });

    it("converts exitEpoch to string for label", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingExitQueueAmountGwei(validatorPubkey, 12345, 1000, true);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.PendingExitQueueAmountGwei,
        { pubkey: validatorPubkey, exit_epoch: "12345", slashed: "true" },
        1000,
      );
    });

    it("converts slashed boolean to string for label", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingExitQueueAmountGwei(validatorPubkey, 60001, 1000, true);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.PendingExitQueueAmountGwei,
        { pubkey: validatorPubkey, exit_epoch: "60001", slashed: "true" },
        1000,
      );

      jest.clearAllMocks();

      updater.setPendingExitQueueAmountGwei(validatorPubkey, 60001, 1000, false);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.PendingExitQueueAmountGwei,
        { pubkey: validatorPubkey, exit_epoch: "60001", slashed: "false" },
        1000,
      );
    });

    it("does not set gauge when amount is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingExitQueueAmountGwei(validatorPubkey, 60001, -1, false);

      expect(metricsService.setGauge).not.toHaveBeenCalled();
    });

    it("does not set gauge when exitEpoch is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingExitQueueAmountGwei(validatorPubkey, -1, 1000, true);

      expect(metricsService.setGauge).not.toHaveBeenCalled();
    });

    it("sets gauge when amount is zero and exitEpoch is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingExitQueueAmountGwei(validatorPubkey, 60001, 0, false);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.PendingExitQueueAmountGwei,
        { pubkey: validatorPubkey, exit_epoch: "60001", slashed: "false" },
        0,
      );
    });

    it("sets gauge when exitEpoch is zero and amount is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingExitQueueAmountGwei(validatorPubkey, 0, 1000, true);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.PendingExitQueueAmountGwei,
        { pubkey: validatorPubkey, exit_epoch: "0", slashed: "true" },
        1000,
      );
    });
  });

  describe("setLastTotalPendingExitGwei", () => {
    it("sets gauge when value is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLastTotalPendingExitGwei(1000);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.LastTotalPendingExitGwei,
        {},
        1000,
      );
    });

    it("does not set gauge when value is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLastTotalPendingExitGwei(-1);

      expect(metricsService.setGauge).not.toHaveBeenCalled();
    });

    it("sets gauge when value is zero", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLastTotalPendingExitGwei(0);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.LastTotalPendingExitGwei,
        {},
        0,
      );
    });
  });

  describe("setPendingFullWithdrawalQueueAmountGwei", () => {
    it("sets gauge when amount and withdrawableEpoch are non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingFullWithdrawalQueueAmountGwei(validatorPubkey, 60001, 32000000000, false);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.PendingFullWithdrawalQueueAmountGwei,
        { pubkey: validatorPubkey, withdrawable_epoch: "60001", slashed: "false" },
        32000000000,
      );
    });

    it("converts withdrawableEpoch to string for label", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingFullWithdrawalQueueAmountGwei(validatorPubkey, 12345, 1000, true);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.PendingFullWithdrawalQueueAmountGwei,
        { pubkey: validatorPubkey, withdrawable_epoch: "12345", slashed: "true" },
        1000,
      );
    });

    it("converts slashed boolean to string for label", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingFullWithdrawalQueueAmountGwei(validatorPubkey, 60001, 1000, true);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.PendingFullWithdrawalQueueAmountGwei,
        { pubkey: validatorPubkey, withdrawable_epoch: "60001", slashed: "true" },
        1000,
      );

      jest.clearAllMocks();

      updater.setPendingFullWithdrawalQueueAmountGwei(validatorPubkey, 60001, 1000, false);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.PendingFullWithdrawalQueueAmountGwei,
        { pubkey: validatorPubkey, withdrawable_epoch: "60001", slashed: "false" },
        1000,
      );
    });

    it("does not set gauge when amount is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingFullWithdrawalQueueAmountGwei(validatorPubkey, 60001, -1, false);

      expect(metricsService.setGauge).not.toHaveBeenCalled();
    });

    it("does not set gauge when withdrawableEpoch is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingFullWithdrawalQueueAmountGwei(validatorPubkey, -1, 1000, true);

      expect(metricsService.setGauge).not.toHaveBeenCalled();
    });

    it("sets gauge when amount is zero and withdrawableEpoch is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingFullWithdrawalQueueAmountGwei(validatorPubkey, 60001, 0, false);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.PendingFullWithdrawalQueueAmountGwei,
        { pubkey: validatorPubkey, withdrawable_epoch: "60001", slashed: "false" },
        0,
      );
    });

    it("sets gauge when withdrawableEpoch is zero and amount is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setPendingFullWithdrawalQueueAmountGwei(validatorPubkey, 0, 1000, true);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.PendingFullWithdrawalQueueAmountGwei,
        { pubkey: validatorPubkey, withdrawable_epoch: "0", slashed: "true" },
        1000,
      );
    });
  });

  describe("setLastTotalPendingFullWithdrawalGwei", () => {
    it("sets gauge when value is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLastTotalPendingFullWithdrawalGwei(1000);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.LastTotalPendingFullWithdrawalGwei,
        {},
        1000,
      );
    });

    it("does not set gauge when value is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLastTotalPendingFullWithdrawalGwei(-1);

      expect(metricsService.setGauge).not.toHaveBeenCalled();
    });

    it("sets gauge when value is zero", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setLastTotalPendingFullWithdrawalGwei(0);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.LastTotalPendingFullWithdrawalGwei,
        {},
        0,
      );
    });
  });

  describe("incrementStakingDepositQuotaExceeded", () => {
    it("increments counter with vault address", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.incrementStakingDepositQuotaExceeded(vaultAddress);

      expect(metricsService.incrementCounter).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.StakingDepositQuotaExceeded,
        { vault_address: vaultAddress },
      );
    });
  });

  describe("incrementContractEstimateGasError", () => {
    it("increments counter with contract address, raw revert data, and error name", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      const contractAddress = "0x1234567890123456789012345678901234567890" as Address;
      const rawRevertData = "0xf2ed496c000000000000000000000000000000000000000000000025dffc6dedca6c668800000000000000000000000000000000000000000000000ac3b0cfe3a6daf2d1";
      const errorName = "ExceedsWithdrawable";

      updater.incrementContractEstimateGasError(contractAddress, rawRevertData, errorName);

      expect(metricsService.incrementCounter).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.ContractEstimateGasError,
        {
          contract_address: contractAddress,
          rawRevertData: rawRevertData,
          errorName: errorName,
        },
      );
    });

    it("increments counter with 'unknown' error name when errorName is undefined", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      const contractAddress = "0x1234567890123456789012345678901234567890" as Address;
      const rawRevertData = "0xf2ed496c000000000000000000000000000000000000000000000025dffc6dedca6c668800000000000000000000000000000000000000000000000ac3b0cfe3a6daf2d1";

      updater.incrementContractEstimateGasError(contractAddress, rawRevertData);

      expect(metricsService.incrementCounter).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.ContractEstimateGasError,
        {
          contract_address: contractAddress,
          rawRevertData: rawRevertData,
          errorName: "unknown",
        },
      );
    });
  });

  describe("setActualRebalanceRequirement", () => {
    it("sets gauge when value is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setActualRebalanceRequirement(vaultAddress, 1000, RebalanceDirection.STAKE);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.ActualRebalanceRequirementGwei,
        { vault_address: vaultAddress, staking_direction: "STAKING" },
        1000,
      );
    });

    it("does not set gauge when value is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setActualRebalanceRequirement(vaultAddress, -1, RebalanceDirection.STAKE);

      expect(metricsService.setGauge).not.toHaveBeenCalled();
    });

    it("sets gauge when value is zero", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setActualRebalanceRequirement(vaultAddress, 0, RebalanceDirection.NONE);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.ActualRebalanceRequirementGwei,
        { vault_address: vaultAddress, staking_direction: "NONE" },
        0,
      );
    });
  });

  describe("setReportedRebalanceRequirement", () => {
    it("sets gauge when value is non-negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setReportedRebalanceRequirement(vaultAddress, 1000, RebalanceDirection.UNSTAKE);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.ReportedRebalanceRequirementGwei,
        { vault_address: vaultAddress, staking_direction: "UNSTAKING" },
        1000,
      );
    });

    it("does not set gauge when value is negative", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setReportedRebalanceRequirement(vaultAddress, -1, RebalanceDirection.STAKE);

      expect(metricsService.setGauge).not.toHaveBeenCalled();
    });

    it("sets gauge when value is zero", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      updater.setReportedRebalanceRequirement(vaultAddress, 0, RebalanceDirection.NONE);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.ReportedRebalanceRequirementGwei,
        { vault_address: vaultAddress, staking_direction: "NONE" },
        0,
      );
    });

    it("handles unknown direction by defaulting to NONE", () => {
      const metricsService = createMetricsServiceMock();
      const updater = new NativeYieldAutomationMetricsUpdater(metricsService);
      jest.clearAllMocks();

      // Test the default case by passing an invalid direction value using type assertion
      updater.setActualRebalanceRequirement(vaultAddress, 1000, "INVALID" as RebalanceDirection);

      expect(metricsService.setGauge).toHaveBeenCalledWith(
        LineaNativeYieldAutomationServiceMetrics.ActualRebalanceRequirementGwei,
        { vault_address: vaultAddress, staking_direction: "NONE" },
        1000,
      );
    });
  });
});
