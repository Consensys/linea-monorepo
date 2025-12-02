import { mock, MockProxy } from "jest-mock-extended";
import type { ILogger } from "@consensys/linea-shared-utils";
import type { IValidatorDataClient } from "../../core/clients/IValidatorDataClient.js";
import type { INativeYieldAutomationMetricsUpdater } from "../../core/metrics/INativeYieldAutomationMetricsUpdater.js";
import type { IYieldManager } from "../../core/clients/contracts/IYieldManager.js";
import type { ValidatorBalanceWithPendingWithdrawal } from "../../core/entities/ValidatorBalance.js";
import type { TransactionReceipt, Address } from "viem";
import { ONE_GWEI } from "@consensys/linea-shared-utils";
import { GaugeMetricsPoller } from "../GaugeMetricsPoller.js";
import { YieldProviderData } from "../../core/clients/contracts/IYieldManager.js";

describe("GaugeMetricsPoller", () => {
  const yieldProvider = "0x1111111111111111111111111111111111111111" as Address;
  const vaultAddress = "0x2222222222222222222222222222222222222222" as Address;

  let logger: MockProxy<ILogger>;
  let validatorDataClient: MockProxy<IValidatorDataClient>;
  let metricsUpdater: MockProxy<INativeYieldAutomationMetricsUpdater>;
  let yieldManagerContractClient: MockProxy<IYieldManager<TransactionReceipt>>;
  let poller: GaugeMetricsPoller;

  beforeEach(() => {
    jest.clearAllMocks();
    logger = mock<ILogger>();
    validatorDataClient = mock<IValidatorDataClient>();
    metricsUpdater = mock<INativeYieldAutomationMetricsUpdater>();
    yieldManagerContractClient = mock<IYieldManager<TransactionReceipt>>();

    poller = new GaugeMetricsPoller(
      logger,
      validatorDataClient,
      metricsUpdater,
      yieldManagerContractClient,
      yieldProvider,
    );

    // Default mocks
    validatorDataClient.getActiveValidatorsWithPendingWithdrawalsAscending.mockResolvedValue([]);
    validatorDataClient.getTotalPendingPartialWithdrawalsWei.mockReturnValue(0n);
    yieldManagerContractClient.getYieldProviderData.mockResolvedValue({
      yieldProviderVendor: 0,
      isStakingPaused: false,
      isOssificationInitiated: false,
      isOssified: false,
      primaryEntrypoint: "0x0000000000000000000000000000000000000000" as Address,
      ossifiedEntrypoint: "0x0000000000000000000000000000000000000000" as Address,
      yieldProviderIndex: 0n,
      userFunds: 0n,
      yieldReportedCumulative: 0n,
      lstLiabilityPrincipal: 0n,
    } as YieldProviderData);
    yieldManagerContractClient.getLidoStakingVaultAddress.mockResolvedValue(vaultAddress);
  });

  describe("poll", () => {
    it("updates LastTotalPendingPartialWithdrawalsGwei gauge", async () => {
      const validators: ValidatorBalanceWithPendingWithdrawal[] = [
        {
          balance: 32n,
          effectiveBalance: 32n,
          publicKey: "validator-1",
          validatorIndex: 1n,
          pendingWithdrawalAmount: 3n,
          withdrawableAmount: 0n,
        },
        {
          balance: 32n,
          effectiveBalance: 32n,
          publicKey: "validator-2",
          validatorIndex: 2n,
          pendingWithdrawalAmount: 1n,
          withdrawableAmount: 0n,
        },
      ];

      validatorDataClient.getActiveValidatorsWithPendingWithdrawalsAscending.mockResolvedValue(validators);
      validatorDataClient.getTotalPendingPartialWithdrawalsWei.mockReturnValue(4n * ONE_GWEI);

      await poller.poll();

      expect(validatorDataClient.getActiveValidatorsWithPendingWithdrawalsAscending).toHaveBeenCalled();
      expect(validatorDataClient.getTotalPendingPartialWithdrawalsWei).toHaveBeenCalledWith(validators);
      expect(metricsUpdater.setLastTotalPendingPartialWithdrawalsGwei).toHaveBeenCalledWith(4);
    });

    it("updates YieldReportedCumulative gauge", async () => {
      const yieldReportedCumulativeWei = 1000n * ONE_GWEI;
      yieldManagerContractClient.getYieldProviderData.mockResolvedValue({
        yieldProviderVendor: 0,
        isStakingPaused: false,
        isOssificationInitiated: false,
        isOssified: false,
        primaryEntrypoint: "0x0000000000000000000000000000000000000000" as Address,
        ossifiedEntrypoint: "0x0000000000000000000000000000000000000000" as Address,
        yieldProviderIndex: 0n,
        userFunds: 0n,
        yieldReportedCumulative: yieldReportedCumulativeWei,
        lstLiabilityPrincipal: 0n,
      } as YieldProviderData);

      await poller.poll();

      expect(yieldManagerContractClient.getYieldProviderData).toHaveBeenCalledWith(yieldProvider);
      expect(yieldManagerContractClient.getLidoStakingVaultAddress).toHaveBeenCalledWith(yieldProvider);
      expect(metricsUpdater.setYieldReportedCumulative).toHaveBeenCalledWith(vaultAddress, 1000);
    });

    it("handles undefined validator list gracefully", async () => {
      validatorDataClient.getActiveValidatorsWithPendingWithdrawalsAscending.mockResolvedValue(undefined);

      await poller.poll();

      expect(validatorDataClient.getActiveValidatorsWithPendingWithdrawalsAscending).toHaveBeenCalled();
      expect(validatorDataClient.getTotalPendingPartialWithdrawalsWei).not.toHaveBeenCalled();
      expect(metricsUpdater.setLastTotalPendingPartialWithdrawalsGwei).not.toHaveBeenCalled();
      // YieldReportedCumulative should still be updated
      expect(metricsUpdater.setYieldReportedCumulative).toHaveBeenCalled();
    });

    it("handles validator data client failure gracefully for pending partial withdrawals", async () => {
      validatorDataClient.getActiveValidatorsWithPendingWithdrawalsAscending.mockRejectedValue(
        new Error("Validator data fetch failed"),
      );

      await poller.poll();

      // Should not throw, and other metrics should still be updated
      expect(yieldManagerContractClient.getYieldProviderData).toHaveBeenCalled();
      expect(metricsUpdater.setLastTotalPendingPartialWithdrawalsGwei).not.toHaveBeenCalled();
      expect(logger.error).toHaveBeenCalledWith(
        "Failed to update pending partial withdrawals gauge metric",
        { error: expect.any(Error) },
      );
    });

    it("handles contract read failure gracefully for YieldReportedCumulative", async () => {
      yieldManagerContractClient.getYieldProviderData.mockRejectedValue(new Error("Contract read failed"));

      await poller.poll();

      // Should not throw, and other metrics should still be updated
      expect(validatorDataClient.getActiveValidatorsWithPendingWithdrawalsAscending).toHaveBeenCalled();
      expect(metricsUpdater.setYieldReportedCumulative).not.toHaveBeenCalled();
      expect(logger.error).toHaveBeenCalledWith(
        "Failed to update yield reported cumulative gauge metric",
        { error: expect.any(Error) },
      );
    });

    it("updates both metrics in parallel", async () => {
      const validators: ValidatorBalanceWithPendingWithdrawal[] = [
        {
          balance: 32n,
          effectiveBalance: 32n,
          publicKey: "validator-1",
          validatorIndex: 1n,
          pendingWithdrawalAmount: 2n,
          withdrawableAmount: 0n,
        },
      ];

      validatorDataClient.getActiveValidatorsWithPendingWithdrawalsAscending.mockResolvedValue(validators);
      validatorDataClient.getTotalPendingPartialWithdrawalsWei.mockReturnValue(2n * ONE_GWEI);

      const yieldReportedCumulativeWei = 500n * ONE_GWEI;
      yieldManagerContractClient.getYieldProviderData.mockResolvedValue({
        yieldProviderVendor: 0,
        isStakingPaused: false,
        isOssificationInitiated: false,
        isOssified: false,
        primaryEntrypoint: "0x0000000000000000000000000000000000000000" as Address,
        ossifiedEntrypoint: "0x0000000000000000000000000000000000000000" as Address,
        yieldProviderIndex: 0n,
        userFunds: 0n,
        yieldReportedCumulative: yieldReportedCumulativeWei,
        lstLiabilityPrincipal: 0n,
      } as YieldProviderData);

      await poller.poll();

      expect(metricsUpdater.setLastTotalPendingPartialWithdrawalsGwei).toHaveBeenCalledWith(2);
      expect(metricsUpdater.setYieldReportedCumulative).toHaveBeenCalledWith(vaultAddress, 500);
    });

    it("converts wei to gwei correctly", async () => {
      const validators: ValidatorBalanceWithPendingWithdrawal[] = [];
      validatorDataClient.getActiveValidatorsWithPendingWithdrawalsAscending.mockResolvedValue(validators);
      validatorDataClient.getTotalPendingPartialWithdrawalsWei.mockReturnValue(1500000000n); // 1.5 gwei in wei

      const yieldReportedCumulativeWei = 2500000000n; // 2.5 gwei in wei
      yieldManagerContractClient.getYieldProviderData.mockResolvedValue({
        yieldProviderVendor: 0,
        isStakingPaused: false,
        isOssificationInitiated: false,
        isOssified: false,
        primaryEntrypoint: "0x0000000000000000000000000000000000000000" as Address,
        ossifiedEntrypoint: "0x0000000000000000000000000000000000000000" as Address,
        yieldProviderIndex: 0n,
        userFunds: 0n,
        yieldReportedCumulative: yieldReportedCumulativeWei,
        lstLiabilityPrincipal: 0n,
      } as YieldProviderData);

      await poller.poll();

      expect(metricsUpdater.setLastTotalPendingPartialWithdrawalsGwei).toHaveBeenCalledWith(1); // Rounded down
      expect(metricsUpdater.setYieldReportedCumulative).toHaveBeenCalledWith(vaultAddress, 2); // Rounded down
    });
  });
});

