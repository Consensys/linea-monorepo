import { mock, MockProxy } from "jest-mock-extended";
import type { ILogger, IBeaconNodeAPIClient } from "@consensys/linea-shared-utils";
import type { IValidatorDataClient } from "../../core/clients/IValidatorDataClient.js";
import type { INativeYieldAutomationMetricsUpdater } from "../../core/metrics/INativeYieldAutomationMetricsUpdater.js";
import type { IYieldManager } from "../../core/clients/contracts/IYieldManager.js";
import type { IVaultHub } from "../../core/clients/contracts/IVaultHub.js";
import type { ValidatorBalanceWithPendingWithdrawal, ValidatorBalance } from "../../core/entities/ValidatorBalance.js";
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
  let vaultHubContractClient: MockProxy<IVaultHub<TransactionReceipt>>;
  let beaconNodeApiClient: MockProxy<IBeaconNodeAPIClient>;
  let poller: GaugeMetricsPoller;

  beforeEach(() => {
    jest.clearAllMocks();
    logger = mock<ILogger>();
    validatorDataClient = mock<IValidatorDataClient>();
    metricsUpdater = mock<INativeYieldAutomationMetricsUpdater>();
    yieldManagerContractClient = mock<IYieldManager<TransactionReceipt>>();
    vaultHubContractClient = mock<IVaultHub<TransactionReceipt>>();
    beaconNodeApiClient = mock<IBeaconNodeAPIClient>();

    poller = new GaugeMetricsPoller(
      logger,
      validatorDataClient,
      metricsUpdater,
      yieldManagerContractClient,
      vaultHubContractClient,
      yieldProvider,
      beaconNodeApiClient,
    );

    // Default mocks
    validatorDataClient.getActiveValidators.mockResolvedValue([]);
    beaconNodeApiClient.getPendingPartialWithdrawals.mockResolvedValue([]);
    validatorDataClient.joinValidatorsWithPendingWithdrawals.mockReturnValue([]);
    validatorDataClient.getTotalPendingPartialWithdrawalsWei.mockReturnValue(0n);
    validatorDataClient.getFilteredAndAggregatedPendingWithdrawals.mockReturnValue([]);
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
      lastReportedNegativeYield: 0n,
    } as YieldProviderData);
    yieldManagerContractClient.getLidoStakingVaultAddress.mockResolvedValue(vaultAddress);
    vaultHubContractClient.getLatestVaultReportTimestamp.mockResolvedValue(0n);
  });

  describe("poll", () => {
    it("updates LastTotalPendingPartialWithdrawalsGwei gauge", async () => {
      validatorDataClient.getFilteredAndAggregatedPendingWithdrawals.mockReturnValue([]);
      const allValidators: ValidatorBalance[] = [
        {
          balance: 32n,
          effectiveBalance: 32n,
          publicKey: "validator-1",
          validatorIndex: 1n,
        },
        {
          balance: 32n,
          effectiveBalance: 32n,
          publicKey: "validator-2",
          validatorIndex: 2n,
        },
      ];

      const pendingWithdrawalsQueue = [
        { validator_index: 1, amount: 3n * ONE_GWEI, withdrawable_epoch: 0 },
        { validator_index: 2, amount: 1n * ONE_GWEI, withdrawable_epoch: 0 },
      ];

      const joinedValidators: ValidatorBalanceWithPendingWithdrawal[] = [
        {
          balance: 32n,
          effectiveBalance: 32n,
          publicKey: "validator-1",
          validatorIndex: 1n,
          pendingWithdrawalAmount: 3n * ONE_GWEI,
          withdrawableAmount: 0n,
        },
        {
          balance: 32n,
          effectiveBalance: 32n,
          publicKey: "validator-2",
          validatorIndex: 2n,
          pendingWithdrawalAmount: 1n * ONE_GWEI,
          withdrawableAmount: 0n,
        },
      ];

      validatorDataClient.getActiveValidators.mockResolvedValue(allValidators);
      beaconNodeApiClient.getPendingPartialWithdrawals.mockResolvedValue(pendingWithdrawalsQueue);
      validatorDataClient.joinValidatorsWithPendingWithdrawals.mockReturnValue(joinedValidators);
      validatorDataClient.getTotalPendingPartialWithdrawalsWei.mockReturnValue(4n * ONE_GWEI);

      await poller.poll();

      expect(validatorDataClient.getActiveValidators).toHaveBeenCalled();
      expect(beaconNodeApiClient.getPendingPartialWithdrawals).toHaveBeenCalled();
      expect(validatorDataClient.joinValidatorsWithPendingWithdrawals).toHaveBeenCalledWith(
        allValidators,
        pendingWithdrawalsQueue,
      );
      expect(validatorDataClient.getTotalPendingPartialWithdrawalsWei).toHaveBeenCalledWith(joinedValidators);
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
        lastReportedNegativeYield: 0n,
      } as YieldProviderData);

      await poller.poll();

      expect(yieldManagerContractClient.getYieldProviderData).toHaveBeenCalledWith(yieldProvider);
      expect(yieldManagerContractClient.getLidoStakingVaultAddress).toHaveBeenCalledTimes(1);
      expect(yieldManagerContractClient.getLidoStakingVaultAddress).toHaveBeenCalledWith(yieldProvider);
      expect(metricsUpdater.setYieldReportedCumulative).toHaveBeenCalledWith(vaultAddress, 1000);
    });

    it("updates LastVaultReportTimestamp gauge", async () => {
      const expectedTimestamp = 1704067200n; // Unix timestamp
      vaultHubContractClient.getLatestVaultReportTimestamp.mockResolvedValue(expectedTimestamp);

      await poller.poll();

      expect(yieldManagerContractClient.getLidoStakingVaultAddress).toHaveBeenCalledTimes(1);
      expect(yieldManagerContractClient.getLidoStakingVaultAddress).toHaveBeenCalledWith(yieldProvider);
      expect(vaultHubContractClient.getLatestVaultReportTimestamp).toHaveBeenCalledWith(vaultAddress);
      expect(metricsUpdater.setLastVaultReportTimestamp).toHaveBeenCalledWith(vaultAddress, Number(expectedTimestamp));
    });

    it("updates PendingPartialWithdrawalQueueAmountGwei gauge for each aggregated withdrawal", async () => {
      const allValidators: ValidatorBalance[] = [
        {
          balance: 32n,
          effectiveBalance: 32n,
          publicKey: "validator-1-pubkey",
          validatorIndex: 1n,
        },
        {
          balance: 32n,
          effectiveBalance: 32n,
          publicKey: "validator-2-pubkey",
          validatorIndex: 2n,
        },
      ];

      const pendingWithdrawalsQueue = [
        { validator_index: 1, amount: 3n * ONE_GWEI, withdrawable_epoch: 100 },
        { validator_index: 1, amount: 2n * ONE_GWEI, withdrawable_epoch: 100 }, // Same validator and epoch - should aggregate
        { validator_index: 1, amount: 5n * ONE_GWEI, withdrawable_epoch: 200 }, // Same validator, different epoch
        { validator_index: 2, amount: 1n * ONE_GWEI, withdrawable_epoch: 100 },
      ];

      const aggregatedWithdrawals = [
        {
          validator_index: 1,
          withdrawable_epoch: 100,
          amount: 5n, // 3 + 2 aggregated (amounts are in gwei)
          pubkey: "validator-1-pubkey",
        },
        {
          validator_index: 1,
          withdrawable_epoch: 200,
          amount: 5n, // amounts are in gwei
          pubkey: "validator-1-pubkey",
        },
        {
          validator_index: 2,
          withdrawable_epoch: 100,
          amount: 1n, // amounts are in gwei
          pubkey: "validator-2-pubkey",
        },
      ];

      validatorDataClient.getActiveValidators.mockResolvedValue(allValidators);
      beaconNodeApiClient.getPendingPartialWithdrawals.mockResolvedValue(pendingWithdrawalsQueue);
      validatorDataClient.getFilteredAndAggregatedPendingWithdrawals.mockReturnValue(aggregatedWithdrawals);

      await poller.poll();

      expect(validatorDataClient.getFilteredAndAggregatedPendingWithdrawals).toHaveBeenCalledWith(
        allValidators,
        pendingWithdrawalsQueue,
      );
      expect(metricsUpdater.setPendingPartialWithdrawalQueueAmountGwei).toHaveBeenCalledTimes(3);
      expect(metricsUpdater.setPendingPartialWithdrawalQueueAmountGwei).toHaveBeenNthCalledWith(
        1,
        "validator-1-pubkey",
        100,
        5,
      );
      expect(metricsUpdater.setPendingPartialWithdrawalQueueAmountGwei).toHaveBeenNthCalledWith(
        2,
        "validator-1-pubkey",
        200,
        5,
      );
      expect(metricsUpdater.setPendingPartialWithdrawalQueueAmountGwei).toHaveBeenNthCalledWith(
        3,
        "validator-2-pubkey",
        100,
        1,
      );
    });

    it("handles undefined aggregated withdrawals gracefully for queue gauge", async () => {
      validatorDataClient.getActiveValidators.mockResolvedValue([]);
      beaconNodeApiClient.getPendingPartialWithdrawals.mockResolvedValue([]);
      validatorDataClient.getFilteredAndAggregatedPendingWithdrawals.mockReturnValue(undefined);

      await poller.poll();

      expect(validatorDataClient.getFilteredAndAggregatedPendingWithdrawals).toHaveBeenCalled();
      expect(metricsUpdater.setPendingPartialWithdrawalQueueAmountGwei).not.toHaveBeenCalled();
    });

    it("handles empty aggregated withdrawals array for queue gauge", async () => {
      validatorDataClient.getActiveValidators.mockResolvedValue([]);
      beaconNodeApiClient.getPendingPartialWithdrawals.mockResolvedValue([]);
      validatorDataClient.getFilteredAndAggregatedPendingWithdrawals.mockReturnValue([]);

      await poller.poll();

      expect(validatorDataClient.getFilteredAndAggregatedPendingWithdrawals).toHaveBeenCalled();
      expect(metricsUpdater.setPendingPartialWithdrawalQueueAmountGwei).not.toHaveBeenCalled();
    });

    it("handles undefined validator list gracefully", async () => {
      validatorDataClient.getActiveValidators.mockResolvedValue(undefined);
      beaconNodeApiClient.getPendingPartialWithdrawals.mockResolvedValue([]);
      validatorDataClient.joinValidatorsWithPendingWithdrawals.mockReturnValue(undefined);
      validatorDataClient.getFilteredAndAggregatedPendingWithdrawals.mockReturnValue(undefined);

      await poller.poll();

      expect(validatorDataClient.getActiveValidators).toHaveBeenCalled();
      expect(beaconNodeApiClient.getPendingPartialWithdrawals).toHaveBeenCalled();
      expect(validatorDataClient.joinValidatorsWithPendingWithdrawals).toHaveBeenCalled();
      expect(validatorDataClient.getFilteredAndAggregatedPendingWithdrawals).toHaveBeenCalled();
      expect(validatorDataClient.getTotalPendingPartialWithdrawalsWei).not.toHaveBeenCalled();
      expect(metricsUpdater.setLastTotalPendingPartialWithdrawalsGwei).not.toHaveBeenCalled();
      expect(metricsUpdater.setPendingPartialWithdrawalQueueAmountGwei).not.toHaveBeenCalled();
      // YieldReportedCumulative and LastVaultReportTimestamp should still be updated
      expect(metricsUpdater.setYieldReportedCumulative).toHaveBeenCalled();
      expect(metricsUpdater.setLastVaultReportTimestamp).toHaveBeenCalled();
    });

    it("handles validator data client failure gracefully for pending partial withdrawals", async () => {
      validatorDataClient.getActiveValidators.mockRejectedValue(new Error("Validator data fetch failed"));
      beaconNodeApiClient.getPendingPartialWithdrawals.mockResolvedValue([]);
      // When allValidators is undefined, joinValidatorsWithPendingWithdrawals should return undefined
      validatorDataClient.joinValidatorsWithPendingWithdrawals.mockReturnValue(undefined);
      validatorDataClient.getFilteredAndAggregatedPendingWithdrawals.mockReturnValue(undefined);

      await poller.poll();

      // Should not throw, and other metrics should still be updated
      expect(yieldManagerContractClient.getYieldProviderData).toHaveBeenCalled();
      expect(metricsUpdater.setLastTotalPendingPartialWithdrawalsGwei).not.toHaveBeenCalled();
      expect(metricsUpdater.setPendingPartialWithdrawalQueueAmountGwei).not.toHaveBeenCalled();
      // Fetch failure is logged at fetch time
      expect(logger.error).toHaveBeenCalledWith("Failed to fetch active validators", {
        error: expect.any(Error),
      });
    });

    it("handles contract read failure gracefully for YieldReportedCumulative", async () => {
      yieldManagerContractClient.getYieldProviderData.mockRejectedValue(new Error("Contract read failed"));

      await poller.poll();

      // Should not throw, and other metrics should still be updated
      expect(validatorDataClient.getActiveValidators).toHaveBeenCalled();
      expect(metricsUpdater.setYieldReportedCumulative).not.toHaveBeenCalled();
      expect(logger.error).toHaveBeenCalledWith(
        "Failed to update yield reported cumulative gauge metric",
        { error: expect.any(Error) },
      );
    });

    it("handles contract read failure gracefully for LastVaultReportTimestamp", async () => {
      vaultHubContractClient.getLatestVaultReportTimestamp.mockRejectedValue(new Error("Contract read failed"));

      await poller.poll();

      // Should not throw, and other metrics should still be updated
      expect(validatorDataClient.getActiveValidators).toHaveBeenCalled();
      expect(yieldManagerContractClient.getYieldProviderData).toHaveBeenCalled();
      expect(metricsUpdater.setLastVaultReportTimestamp).not.toHaveBeenCalled();
      expect(logger.error).toHaveBeenCalledWith(
        "Failed to update last vault report timestamp gauge metric",
        { error: expect.any(Error) },
      );
    });

    it("updates all four metrics in parallel and fetches vault address only once", async () => {
      const allValidators: ValidatorBalance[] = [
        {
          balance: 32n,
          effectiveBalance: 32n,
          publicKey: "validator-1",
          validatorIndex: 1n,
        },
      ];

      const pendingWithdrawalsQueue = [{ validator_index: 1, amount: 2n * ONE_GWEI, withdrawable_epoch: 0 }];

      const joinedValidators: ValidatorBalanceWithPendingWithdrawal[] = [
        {
          balance: 32n,
          effectiveBalance: 32n,
          publicKey: "validator-1",
          validatorIndex: 1n,
          pendingWithdrawalAmount: 2n * ONE_GWEI,
          withdrawableAmount: 0n,
        },
      ];

      validatorDataClient.getActiveValidators.mockResolvedValue(allValidators);
      beaconNodeApiClient.getPendingPartialWithdrawals.mockResolvedValue(pendingWithdrawalsQueue);
      validatorDataClient.joinValidatorsWithPendingWithdrawals.mockReturnValue(joinedValidators);
      validatorDataClient.getTotalPendingPartialWithdrawalsWei.mockReturnValue(2n * ONE_GWEI);
      validatorDataClient.getFilteredAndAggregatedPendingWithdrawals.mockReturnValue([
        {
          validator_index: 1,
          withdrawable_epoch: 0,
          amount: 2n, // amounts are in gwei
          pubkey: "validator-1",
        },
      ]);

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
        lastReportedNegativeYield: 0n,
      } as YieldProviderData);

      const expectedTimestamp = 1704067200n;
      vaultHubContractClient.getLatestVaultReportTimestamp.mockResolvedValue(expectedTimestamp);

      await poller.poll();

      // Verify vault address is only fetched once, even though it's used by two metrics
      expect(yieldManagerContractClient.getLidoStakingVaultAddress).toHaveBeenCalledTimes(1);
      expect(yieldManagerContractClient.getLidoStakingVaultAddress).toHaveBeenCalledWith(yieldProvider);
      expect(metricsUpdater.setLastTotalPendingPartialWithdrawalsGwei).toHaveBeenCalledWith(2);
      expect(metricsUpdater.setPendingPartialWithdrawalQueueAmountGwei).toHaveBeenCalledWith("validator-1", 0, 2);
      expect(metricsUpdater.setYieldReportedCumulative).toHaveBeenCalledWith(vaultAddress, 500);
      expect(metricsUpdater.setLastVaultReportTimestamp).toHaveBeenCalledWith(vaultAddress, Number(expectedTimestamp));
    });

    it("converts wei to gwei correctly and timestamp to number", async () => {
      const allValidators: ValidatorBalance[] = [];
      const pendingWithdrawalsQueue: any[] = [];
      const joinedValidators: ValidatorBalanceWithPendingWithdrawal[] = [];

      validatorDataClient.getActiveValidators.mockResolvedValue(allValidators);
      beaconNodeApiClient.getPendingPartialWithdrawals.mockResolvedValue(pendingWithdrawalsQueue);
      validatorDataClient.joinValidatorsWithPendingWithdrawals.mockReturnValue(joinedValidators);
      validatorDataClient.getTotalPendingPartialWithdrawalsWei.mockReturnValue(1500000000n); // 1.5 gwei in wei
      validatorDataClient.getFilteredAndAggregatedPendingWithdrawals.mockReturnValue([]);

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
        lastReportedNegativeYield: 0n,
      } as YieldProviderData);

      const expectedTimestamp = 1704067200n;
      vaultHubContractClient.getLatestVaultReportTimestamp.mockResolvedValue(expectedTimestamp);

      await poller.poll();

      expect(metricsUpdater.setLastTotalPendingPartialWithdrawalsGwei).toHaveBeenCalledWith(1); // Rounded down
      expect(metricsUpdater.setYieldReportedCumulative).toHaveBeenCalledWith(vaultAddress, 2); // Rounded down
      expect(metricsUpdater.setLastVaultReportTimestamp).toHaveBeenCalledWith(vaultAddress, Number(expectedTimestamp));
    });

    it("handles all metrics failing gracefully", async () => {
      // Mock all metrics to fail
      validatorDataClient.getActiveValidators.mockRejectedValue(new Error("Validator data fetch failed"));
      beaconNodeApiClient.getPendingPartialWithdrawals.mockResolvedValue([]);
      // Mock the update functions to throw errors
      validatorDataClient.joinValidatorsWithPendingWithdrawals.mockImplementation(() => {
        throw new Error("Join failed");
      });
      validatorDataClient.getFilteredAndAggregatedPendingWithdrawals.mockImplementation(() => {
        throw new Error("Filter failed");
      });
      yieldManagerContractClient.getYieldProviderData.mockRejectedValue(new Error("Contract read failed"));
      vaultHubContractClient.getLatestVaultReportTimestamp.mockRejectedValue(new Error("Contract read failed"));

      await poller.poll();

      // Verify all errors were logged
      // 1. Fetch failure for active validators
      // 2. Update failure for total pending partial withdrawals (index 0)
      // 3. Update failure for pending partial withdrawals queue (index 1)
      // 4. Update failure for yield reported cumulative (index 2)
      // 5. Update failure for last vault report timestamp (index 3)
      expect(logger.error).toHaveBeenCalledWith("Failed to fetch active validators", {
        error: expect.any(Error),
      });
      expect(logger.error).toHaveBeenCalledWith(
        "Failed to update total pending partial withdrawals gauge metric",
        { error: expect.any(Error) },
      );
      expect(logger.error).toHaveBeenCalledWith(
        "Failed to update pending partial withdrawals queue gauge metric",
        { error: expect.any(Error) },
      );
      expect(logger.error).toHaveBeenCalledWith(
        "Failed to update yield reported cumulative gauge metric",
        { error: expect.any(Error) },
      );
      expect(logger.error).toHaveBeenCalledWith(
        "Failed to update last vault report timestamp gauge metric",
        { error: expect.any(Error) },
      );
    });

    it("handles vault address fetch failure gracefully", async () => {
      // Mock vault address fetch to fail
      yieldManagerContractClient.getLidoStakingVaultAddress.mockRejectedValue(new Error("Vault address fetch failed"));

      await poller.poll();

      // Verify vault address fetch was attempted
      expect(yieldManagerContractClient.getLidoStakingVaultAddress).toHaveBeenCalledTimes(1);
      expect(yieldManagerContractClient.getLidoStakingVaultAddress).toHaveBeenCalledWith(yieldProvider);

      // Verify error was logged
      expect(logger.error).toHaveBeenCalledWith("Failed to fetch vault address, skipping vault-dependent metrics", {
        error: expect.any(Error),
      });

      // Verify vault-dependent metrics were not called
      expect(metricsUpdater.setYieldReportedCumulative).not.toHaveBeenCalled();
      expect(metricsUpdater.setLastVaultReportTimestamp).not.toHaveBeenCalled();

      // Verify other metrics still work
      expect(validatorDataClient.getActiveValidators).toHaveBeenCalled();
    });

    it("handles out of bounds index by using unknown metric name", async () => {
      // This test covers the fallback case where index >= metricNames.length
      // We need to simulate a scenario where we have more rejected promises than metric names
      // Since we can't naturally create this, we'll test the logic directly by mocking Promise.allSettled
      const originalAllSettled = Promise.allSettled;
      const mockAllSettled = jest.fn().mockResolvedValue([
        { status: "rejected" as const, reason: new Error("Error 1") },
        { status: "rejected" as const, reason: new Error("Error 2") },
        { status: "rejected" as const, reason: new Error("Error 3") },
        { status: "rejected" as const, reason: new Error("Error 4") },
        { status: "rejected" as const, reason: new Error("Error 5") }, // Index 4 triggers "unknown"
      ]);

      // Temporarily replace Promise.allSettled
      (global as any).Promise.allSettled = mockAllSettled;

      await poller.poll();

      // Verify that the 5th error uses "unknown" as the metric name
      expect(logger.error).toHaveBeenCalledWith(
        "Failed to update unknown gauge metric",
        { error: expect.any(Error) },
      );

      // Restore original Promise.allSettled
      (global as any).Promise.allSettled = originalAllSettled;
    });
  });
});

