import { mock, MockProxy } from "jest-mock-extended";
import type { ILogger, IBeaconNodeAPIClient } from "@consensys/linea-shared-utils";
import type { IValidatorDataClient } from "../../core/clients/IValidatorDataClient.js";
import type { INativeYieldAutomationMetricsUpdater } from "../../core/metrics/INativeYieldAutomationMetricsUpdater.js";
import type { IYieldManager } from "../../core/clients/contracts/IYieldManager.js";
import type { IVaultHub } from "../../core/clients/contracts/IVaultHub.js";
import type { ExitingValidator, ValidatorBalanceWithPendingWithdrawal, ValidatorBalance } from "../../core/entities/ValidatorBalance.js";
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
    validatorDataClient.getExitingValidators.mockResolvedValue([]);
    beaconNodeApiClient.getPendingPartialWithdrawals.mockResolvedValue([]);
    beaconNodeApiClient.getPendingDeposits.mockResolvedValue([]);
    validatorDataClient.joinValidatorsWithPendingWithdrawals.mockReturnValue([]);
    validatorDataClient.getTotalPendingPartialWithdrawalsWei.mockReturnValue(0n);
    validatorDataClient.getFilteredAndAggregatedPendingWithdrawals.mockReturnValue([]);
    validatorDataClient.getTotalValidatorBalanceGwei.mockReturnValue(undefined);
    validatorDataClient.getTotalBalanceOfExitingValidators.mockReturnValue(undefined);
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

    it("updates LastTotalValidatorBalanceGwei gauge", async () => {
      const allValidators: ValidatorBalance[] = [
        {
          balance: 32n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "validator-1",
          validatorIndex: 1n,
        },
        {
          balance: 40n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "validator-2",
          validatorIndex: 2n,
        },
      ];

      validatorDataClient.getActiveValidators.mockResolvedValue(allValidators);
      validatorDataClient.getTotalValidatorBalanceGwei.mockReturnValue(72n * ONE_GWEI);

      await poller.poll();

      expect(validatorDataClient.getActiveValidators).toHaveBeenCalled();
      expect(validatorDataClient.getTotalValidatorBalanceGwei).toHaveBeenCalledWith(allValidators);
      expect(metricsUpdater.setLastTotalValidatorBalanceGwei).toHaveBeenCalledWith(Number(72n * ONE_GWEI));
    });

    it("handles undefined validator list gracefully for total validator balance", async () => {
      validatorDataClient.getActiveValidators.mockResolvedValue(undefined);
      validatorDataClient.getTotalValidatorBalanceGwei.mockReturnValue(undefined);

      await poller.poll();

      expect(validatorDataClient.getActiveValidators).toHaveBeenCalled();
      expect(validatorDataClient.getTotalValidatorBalanceGwei).toHaveBeenCalledWith(undefined);
      expect(metricsUpdater.setLastTotalValidatorBalanceGwei).not.toHaveBeenCalled();
      expect(logger.warn).toHaveBeenCalledWith("Skipping total validator balance gauge update: validator balance unavailable");
    });

    it("handles empty validator array gracefully for total validator balance", async () => {
      validatorDataClient.getActiveValidators.mockResolvedValue([]);
      validatorDataClient.getTotalValidatorBalanceGwei.mockReturnValue(undefined);

      await poller.poll();

      expect(validatorDataClient.getActiveValidators).toHaveBeenCalled();
      expect(validatorDataClient.getTotalValidatorBalanceGwei).toHaveBeenCalledWith([]);
      expect(metricsUpdater.setLastTotalValidatorBalanceGwei).not.toHaveBeenCalled();
      expect(logger.warn).toHaveBeenCalledWith("Skipping total validator balance gauge update: validator balance unavailable");
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

    it("updates LstLiabilityPrincipalGwei gauge", async () => {
      const lstLiabilityPrincipalWei = 5000n * ONE_GWEI;
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
        lstLiabilityPrincipal: lstLiabilityPrincipalWei,
        lastReportedNegativeYield: 0n,
      } as YieldProviderData);

      await poller.poll();

      expect(yieldManagerContractClient.getYieldProviderData).toHaveBeenCalledWith(yieldProvider);
      expect(yieldManagerContractClient.getLidoStakingVaultAddress).toHaveBeenCalledTimes(1);
      expect(yieldManagerContractClient.getLidoStakingVaultAddress).toHaveBeenCalledWith(yieldProvider);
      expect(metricsUpdater.setLstLiabilityPrincipalGwei).toHaveBeenCalledWith(vaultAddress, 5000);
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
      expect(logger.warn).toHaveBeenCalledWith("Skipping pending partial withdrawals queue gauge update: aggregated withdrawals unavailable");
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
      expect(validatorDataClient.getTotalValidatorBalanceGwei).toHaveBeenCalledWith(undefined);
      expect(metricsUpdater.setLastTotalPendingPartialWithdrawalsGwei).not.toHaveBeenCalled();
      expect(metricsUpdater.setLastTotalValidatorBalanceGwei).not.toHaveBeenCalled();
      expect(metricsUpdater.setPendingPartialWithdrawalQueueAmountGwei).not.toHaveBeenCalled();
      // YieldReportedCumulative, LstLiabilityPrincipal, and LastVaultReportTimestamp should still be updated
      expect(metricsUpdater.setYieldReportedCumulative).toHaveBeenCalled();
      expect(metricsUpdater.setLstLiabilityPrincipalGwei).toHaveBeenCalled();
      expect(metricsUpdater.setLastVaultReportTimestamp).toHaveBeenCalled();
      // Verify warnings are logged for skipped metrics
      expect(logger.warn).toHaveBeenCalledWith("Skipping total pending partial withdrawals gauge update: validator data unavailable");
      expect(logger.warn).toHaveBeenCalledWith("Skipping pending partial withdrawals queue gauge update: aggregated withdrawals unavailable");
      expect(logger.warn).toHaveBeenCalledWith("Skipping total validator balance gauge update: validator balance unavailable");
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
      expect(metricsUpdater.setLastTotalValidatorBalanceGwei).not.toHaveBeenCalled();
      expect(metricsUpdater.setPendingPartialWithdrawalQueueAmountGwei).not.toHaveBeenCalled();
      // Fetch failure is logged at fetch time
      expect(logger.error).toHaveBeenCalledWith("Failed to fetch active validators", {
        error: expect.any(Error),
      });
      // Verify warnings are logged for skipped metrics
      expect(logger.warn).toHaveBeenCalledWith("Skipping total pending partial withdrawals gauge update: validator data unavailable");
      expect(logger.warn).toHaveBeenCalledWith("Skipping pending partial withdrawals queue gauge update: aggregated withdrawals unavailable");
    });

    it("handles exiting validators fetch failure gracefully", async () => {
      validatorDataClient.getActiveValidators.mockResolvedValue([]);
      validatorDataClient.getExitingValidators.mockRejectedValue(new Error("Exiting validators fetch failed"));
      beaconNodeApiClient.getPendingPartialWithdrawals.mockResolvedValue([]);

      await poller.poll();

      // Should not throw, and other metrics should still be updated
      expect(metricsUpdater.setLastTotalPendingPartialWithdrawalsGwei).toHaveBeenCalled();
      // Fetch failure is logged at fetch time
      expect(logger.error).toHaveBeenCalledWith("Failed to fetch exiting validators", {
        error: expect.any(Error),
      });
    });

    it("handles pending partial withdrawals fetch failure gracefully", async () => {
      beaconNodeApiClient.getPendingPartialWithdrawals.mockRejectedValue(new Error("Failed to fetch pending partial withdrawals"));

      await poller.poll();

      // Should not throw, and other metrics should still be updated
      expect(validatorDataClient.getActiveValidators).toHaveBeenCalled();
      expect(logger.error).toHaveBeenCalledWith("Failed to fetch pending partial withdrawals", {
        error: expect.any(Error),
      });
    });

    it("handles yield provider data fetch failure gracefully", async () => {
      yieldManagerContractClient.getYieldProviderData.mockRejectedValue(new Error("Contract read failed"));

      await poller.poll();

      // Should not throw, and other metrics should still be updated
      expect(validatorDataClient.getActiveValidators).toHaveBeenCalled();
      expect(metricsUpdater.setYieldReportedCumulative).not.toHaveBeenCalled();
      expect(metricsUpdater.setLstLiabilityPrincipalGwei).not.toHaveBeenCalled();
      expect(logger.error).toHaveBeenCalledWith("Failed to fetch yield provider data", {
        error: expect.any(Error),
      });
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

    it("updates all six metrics in parallel and fetches vault address only once", async () => {
      const allValidators: ValidatorBalance[] = [
        {
          balance: 32n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "validator-1",
          validatorIndex: 1n,
        },
      ];

      const pendingWithdrawalsQueue = [{ validator_index: 1, amount: 2n * ONE_GWEI, withdrawable_epoch: 0 }];

      const joinedValidators: ValidatorBalanceWithPendingWithdrawal[] = [
        {
          balance: 32n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
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
      validatorDataClient.getTotalValidatorBalanceGwei.mockReturnValue(32n * ONE_GWEI);
      validatorDataClient.getFilteredAndAggregatedPendingWithdrawals.mockReturnValue([
        {
          validator_index: 1,
          withdrawable_epoch: 0,
          amount: 2n, // amounts are in gwei
          pubkey: "validator-1",
        },
      ]);

      const yieldReportedCumulativeWei = 500n * ONE_GWEI;
      const lstLiabilityPrincipalWei = 3000n * ONE_GWEI;
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
        lstLiabilityPrincipal: lstLiabilityPrincipalWei,
        lastReportedNegativeYield: 0n,
      } as YieldProviderData);

      const expectedTimestamp = 1704067200n;
      vaultHubContractClient.getLatestVaultReportTimestamp.mockResolvedValue(expectedTimestamp);

      await poller.poll();

      // Verify vault address and yield provider data are only fetched once, even though they're used by multiple metrics
      expect(yieldManagerContractClient.getLidoStakingVaultAddress).toHaveBeenCalledTimes(1);
      expect(yieldManagerContractClient.getLidoStakingVaultAddress).toHaveBeenCalledWith(yieldProvider);
      expect(yieldManagerContractClient.getYieldProviderData).toHaveBeenCalledTimes(1);
      expect(yieldManagerContractClient.getYieldProviderData).toHaveBeenCalledWith(yieldProvider);
      expect(metricsUpdater.setLastTotalPendingPartialWithdrawalsGwei).toHaveBeenCalledWith(2);
      expect(metricsUpdater.setPendingPartialWithdrawalQueueAmountGwei).toHaveBeenCalledWith("validator-1", 0, 2);
      expect(metricsUpdater.setLastTotalValidatorBalanceGwei).toHaveBeenCalledWith(Number(32n * ONE_GWEI));
      expect(metricsUpdater.setYieldReportedCumulative).toHaveBeenCalledWith(vaultAddress, 500);
      expect(metricsUpdater.setLstLiabilityPrincipalGwei).toHaveBeenCalledWith(vaultAddress, 3000);
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
      // Reset and mock all metrics to fail
      validatorDataClient.getActiveValidators.mockReset();
      validatorDataClient.getActiveValidators.mockRejectedValue(new Error("Validator data fetch failed"));
      beaconNodeApiClient.getPendingPartialWithdrawals.mockResolvedValue([]);
      // Mock pending deposits with data so the functions get called and can fail
      beaconNodeApiClient.getPendingDeposits.mockResolvedValue([
        {
          pubkey: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
          withdrawal_credentials: "0x0200000000000000000000002222222222222222222222222222222222222222",
          amount: 32000000000,
          signature: "0xabcdef",
          slot: 100,
        },
      ]);
      // Ensure vault address fetch succeeds
      yieldManagerContractClient.getLidoStakingVaultAddress.mockReset();
      yieldManagerContractClient.getLidoStakingVaultAddress.mockResolvedValue(vaultAddress);
      // Mock yield provider data to fail - this should make yieldProviderData undefined
      yieldManagerContractClient.getYieldProviderData.mockReset();
      yieldManagerContractClient.getYieldProviderData.mockRejectedValue(new Error("Contract read failed"));
      // Mock the update functions to throw errors
      validatorDataClient.joinValidatorsWithPendingWithdrawals.mockImplementation(() => {
        throw new Error("Join failed");
      });
      validatorDataClient.getFilteredAndAggregatedPendingWithdrawals.mockImplementation(() => {
        throw new Error("Filter failed");
      });
      vaultHubContractClient.getLatestVaultReportTimestamp.mockRejectedValue(new Error("Contract read failed"));
      // Mock pending deposits functions to throw errors
      metricsUpdater.setPendingDepositQueueAmountGwei.mockImplementation(() => {
        throw new Error("Pending deposits queue failed");
      });
      metricsUpdater.setLastTotalPendingDepositGwei.mockImplementation(() => {
        throw new Error("Total pending deposits failed");
      });

      await poller.poll();

      // Verify all errors were logged
      // 1. Fetch failure for active validators
      // 2. Fetch failure for yield provider data
      // 3. Update failure for total pending partial withdrawals (index 0)
      // 4. Update failure for pending partial withdrawals queue (index 1)
      // 5. Update failure for last vault report timestamp (index 3)
      // 6. Update failure for pending deposits queue (index 4)
      // 7. Update failure for total pending deposits (index 5)
      // Note: yield provider data metrics are not added to updatePromises when fetch fails
      expect(logger.error).toHaveBeenCalledWith("Failed to fetch active validators", {
        error: expect.any(Error),
      });
      expect(logger.error).toHaveBeenCalledWith("Failed to fetch yield provider data", {
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
        "Failed to update last vault report timestamp gauge metric",
        { error: expect.any(Error) },
      );
      expect(logger.error).toHaveBeenCalledWith(
        "Failed to update pending deposits queue gauge metric",
        { error: expect.any(Error) },
      );
      expect(logger.error).toHaveBeenCalledWith(
        "Failed to update total pending deposits gauge metric",
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
      expect(metricsUpdater.setLstLiabilityPrincipalGwei).not.toHaveBeenCalled();
      expect(metricsUpdater.setLastVaultReportTimestamp).not.toHaveBeenCalled();

      // Verify other metrics still work
      expect(validatorDataClient.getActiveValidators).toHaveBeenCalled();
    });

    it("updates PendingDepositQueueAmountGwei gauge for matching deposits", async () => {
      // Mock pending deposits with matching withdrawal credentials
      const vaultWithdrawalCredentials = "0x0200000000000000000000002222222222222222222222222222222222222222";
      const matchingDeposit = {
        pubkey: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
        withdrawal_credentials: vaultWithdrawalCredentials,
        amount: 32000000000, // 32 ETH in gwei
        signature: "0xabcdef",
        slot: 100,
      };
      const nonMatchingDeposit = {
        pubkey: "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
        withdrawal_credentials: "0x0200000000000000000000003333333333333333333333333333333333333333",
        amount: 32000000000,
        signature: "0xfedcba",
        slot: 101,
      };

      beaconNodeApiClient.getPendingDeposits.mockResolvedValue([matchingDeposit, nonMatchingDeposit]);

      await poller.poll();

      // Verify that only the matching deposit was processed
      expect(metricsUpdater.setPendingDepositQueueAmountGwei).toHaveBeenCalledTimes(1);
      expect(metricsUpdater.setPendingDepositQueueAmountGwei).toHaveBeenCalledWith(
        matchingDeposit.pubkey,
        matchingDeposit.slot,
        matchingDeposit.amount,
      );
    });

    it("handles undefined pending deposits gracefully", async () => {
      beaconNodeApiClient.getPendingDeposits.mockResolvedValue(undefined);

      await poller.poll();

      // Verify that setPendingDepositQueueAmountGwei was not called
      expect(metricsUpdater.setPendingDepositQueueAmountGwei).not.toHaveBeenCalled();
      expect(logger.warn).toHaveBeenCalledWith("Skipping pending deposits queue gauge update: pending deposits data unavailable");
    });

    it("handles empty pending deposits array gracefully", async () => {
      beaconNodeApiClient.getPendingDeposits.mockResolvedValue([]);

      await poller.poll();

      // Verify that setPendingDepositQueueAmountGwei was not called
      expect(metricsUpdater.setPendingDepositQueueAmountGwei).not.toHaveBeenCalled();
    });

    it("handles pending deposits fetch failure gracefully", async () => {
      beaconNodeApiClient.getPendingDeposits.mockRejectedValue(new Error("Failed to fetch pending deposits"));

      await poller.poll();

      // Verify error was logged
      expect(logger.error).toHaveBeenCalledWith("Failed to fetch pending deposits", {
        error: expect.any(Error),
      });

      // Verify that setPendingDepositQueueAmountGwei was not called
      expect(metricsUpdater.setPendingDepositQueueAmountGwei).not.toHaveBeenCalled();
      expect(metricsUpdater.setLastTotalPendingDepositGwei).not.toHaveBeenCalled();
    });

    it("updates LastTotalPendingDepositGwei gauge with sum of matching deposits", async () => {
      // Mock pending deposits with matching withdrawal credentials
      const vaultWithdrawalCredentials = "0x0200000000000000000000002222222222222222222222222222222222222222";
      const matchingDeposit1 = {
        pubkey: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
        withdrawal_credentials: vaultWithdrawalCredentials,
        amount: 32000000000, // 32 ETH in gwei
        signature: "0xabcdef",
        slot: 100,
      };
      const matchingDeposit2 = {
        pubkey: "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
        withdrawal_credentials: vaultWithdrawalCredentials,
        amount: 16000000000, // 16 ETH in gwei
        signature: "0xfedcba",
        slot: 101,
      };
      const nonMatchingDeposit = {
        pubkey: "0x9876543210fedcba9876543210fedcba9876543210fedcba9876543210fedcba",
        withdrawal_credentials: "0x0200000000000000000000003333333333333333333333333333333333333333",
        amount: 32000000000,
        signature: "0x123456",
        slot: 102,
      };

      beaconNodeApiClient.getPendingDeposits.mockResolvedValue([matchingDeposit1, matchingDeposit2, nonMatchingDeposit]);

      await poller.poll();

      // Verify that setLastTotalPendingDepositGwei was called with the sum of matching deposits
      expect(metricsUpdater.setLastTotalPendingDepositGwei).toHaveBeenCalledTimes(1);
      expect(metricsUpdater.setLastTotalPendingDepositGwei).toHaveBeenCalledWith(48000000000); // 32 + 16 = 48 ETH in gwei
    });

    it("handles empty matching deposits array for total pending deposits", async () => {
      const nonMatchingDeposit = {
        pubkey: "0x9876543210fedcba9876543210fedcba9876543210fedcba9876543210fedcba",
        withdrawal_credentials: "0x0200000000000000000000003333333333333333333333333333333333333333",
        amount: 32000000000,
        signature: "0x123456",
        slot: 102,
      };

      beaconNodeApiClient.getPendingDeposits.mockResolvedValue([nonMatchingDeposit]);

      await poller.poll();

      // Verify that setLastTotalPendingDepositGwei was called with 0 (no matching deposits)
      expect(metricsUpdater.setLastTotalPendingDepositGwei).toHaveBeenCalledTimes(1);
      expect(metricsUpdater.setLastTotalPendingDepositGwei).toHaveBeenCalledWith(0);
    });

    it("handles undefined pending deposits gracefully for total pending deposits", async () => {
      beaconNodeApiClient.getPendingDeposits.mockResolvedValue(undefined);

      await poller.poll();

      // Verify that setLastTotalPendingDepositGwei was not called
      expect(metricsUpdater.setLastTotalPendingDepositGwei).not.toHaveBeenCalled();
      expect(logger.warn).toHaveBeenCalledWith("Skipping total pending deposits gauge update: pending deposits data unavailable");
    });

    it("handles out of bounds index by using unknown metric name", async () => {
      // This test covers the fallback case where index >= metricNames.length
      // We need to simulate a scenario where we have more rejected promises than metric names
      // Since we can't naturally create this, we'll test the logic directly by mocking Promise.allSettled
      // We now have 10 metrics (indices 0-9), so index 10 will trigger "unknown"
      const originalAllSettled = Promise.allSettled;
      
      // Mock Promise.allSettled to return different values for fetch (first call) and update (second call)
      let callCount = 0;
      const mockAllSettled = jest.fn().mockImplementation((promises: Promise<any>[]) => {
        callCount++;
        // First call: fetch promises (6 promises - added getExitingValidators)
        if (callCount === 1) {
          return Promise.resolve([
            { status: "fulfilled" as const, value: [] },
            { status: "fulfilled" as const, value: [] }, // getExitingValidators
            { status: "fulfilled" as const, value: [] },
            { status: "fulfilled" as const, value: [] },
            { status: "fulfilled" as const, value: vaultAddress },
            {
              status: "fulfilled" as const,
              value: {
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
              } as YieldProviderData,
            },
          ]);
        }
        // Second call: update promises (11 promises to trigger index 10 = "unknown")
        return Promise.resolve([
          { status: "rejected" as const, reason: new Error("Error 1") },
          { status: "rejected" as const, reason: new Error("Error 2") },
          { status: "rejected" as const, reason: new Error("Error 3") },
          { status: "rejected" as const, reason: new Error("Error 4") },
          { status: "rejected" as const, reason: new Error("Error 5") },
          { status: "rejected" as const, reason: new Error("Error 6") },
          { status: "rejected" as const, reason: new Error("Error 7") },
          { status: "rejected" as const, reason: new Error("Error 8") },
          { status: "rejected" as const, reason: new Error("Error 9") },
          { status: "rejected" as const, reason: new Error("Error 10") },
          { status: "rejected" as const, reason: new Error("Error 11") }, // Index 10 triggers "unknown"
        ]);
      });

      // Temporarily replace Promise.allSettled
      (global as any).Promise.allSettled = mockAllSettled;

      await poller.poll();

      // Verify that the 11th error (index 10) uses "unknown" as the metric name
      expect(logger.error).toHaveBeenCalledWith(
        "Failed to update unknown gauge metric",
        { error: expect.any(Error) },
      );

      // Restore original Promise.allSettled
      (global as any).Promise.allSettled = originalAllSettled;
    });
  });

  describe("_updatePendingExitQueueAmountGwei", () => {
    it("updates PendingExitQueueAmountGwei gauge for each exiting validator", async () => {
      const exitDateString = "2024-01-15T10:30:00Z";
      const exitingValidators: ExitingValidator[] = [
        {
          balance: 32n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "0xvalidator1",
          validatorIndex: 1n,
          exitEpoch: 100,
          exitDate: new Date(exitDateString),
          slashed: false,
        },
        {
          balance: 40n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "0xvalidator2",
          validatorIndex: 2n,
          exitEpoch: 150,
          exitDate: new Date(exitDateString),
          slashed: true,
        },
      ];

      validatorDataClient.getExitingValidators.mockResolvedValue(exitingValidators);

      await poller.poll();

      expect(metricsUpdater.setPendingExitQueueAmountGwei).toHaveBeenCalledTimes(2);
      expect(metricsUpdater.setPendingExitQueueAmountGwei).toHaveBeenCalledWith(
        "0xvalidator1",
        100,
        32000000000,
        false,
      );
      expect(metricsUpdater.setPendingExitQueueAmountGwei).toHaveBeenCalledWith(
        "0xvalidator2",
        150,
        40000000000,
        true,
      );
    });

    it("handles undefined exitingValidators with warning", async () => {
      validatorDataClient.getExitingValidators.mockResolvedValue(undefined);

      await poller.poll();

      expect(metricsUpdater.setPendingExitQueueAmountGwei).not.toHaveBeenCalled();
      expect(logger.warn).toHaveBeenCalledWith(
        "Skipping pending exit queue gauge update: exiting validators data unavailable or empty",
      );
    });

    it("handles empty exitingValidators array with warning", async () => {
      validatorDataClient.getExitingValidators.mockResolvedValue([]);

      await poller.poll();

      expect(metricsUpdater.setPendingExitQueueAmountGwei).not.toHaveBeenCalled();
      expect(logger.warn).toHaveBeenCalledWith(
        "Skipping pending exit queue gauge update: exiting validators data unavailable or empty",
      );
    });

    it("converts balance bigint to number correctly", async () => {
      const exitDateString = "2024-01-15T10:30:00Z";
      const exitingValidators: ExitingValidator[] = [
        {
          balance: 35n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "0xvalidator1",
          validatorIndex: 1n,
          exitEpoch: 100,
          exitDate: new Date(exitDateString),
          slashed: false,
        },
      ];

      validatorDataClient.getExitingValidators.mockResolvedValue(exitingValidators);

      await poller.poll();

      expect(metricsUpdater.setPendingExitQueueAmountGwei).toHaveBeenCalledWith(
        "0xvalidator1",
        100,
        35000000000,
        false,
      );
    });

    it("passes slashed boolean correctly", async () => {
      const exitDateString = "2024-01-15T10:30:00Z";
      const exitingValidators: ExitingValidator[] = [
        {
          balance: 32n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "0xvalidator-slashed",
          validatorIndex: 1n,
          exitEpoch: 100,
          exitDate: new Date(exitDateString),
          slashed: true,
        },
        {
          balance: 32n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "0xvalidator-normal",
          validatorIndex: 2n,
          exitEpoch: 150,
          exitDate: new Date(exitDateString),
          slashed: false,
        },
      ];

      validatorDataClient.getExitingValidators.mockResolvedValue(exitingValidators);

      await poller.poll();

      expect(metricsUpdater.setPendingExitQueueAmountGwei).toHaveBeenCalledWith(
        "0xvalidator-slashed",
        100,
        32000000000,
        true,
      );
      expect(metricsUpdater.setPendingExitQueueAmountGwei).toHaveBeenCalledWith(
        "0xvalidator-normal",
        150,
        32000000000,
        false,
      );
    });
  });

  describe("_updateLastTotalPendingExitGwei", () => {
    it("updates LastTotalPendingExitGwei gauge with total balance", async () => {
      const exitDateString = "2024-01-15T10:30:00Z";
      const exitingValidators: ExitingValidator[] = [
        {
          balance: 32n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "0xvalidator1",
          validatorIndex: 1n,
          exitEpoch: 100,
          exitDate: new Date(exitDateString),
          slashed: false,
        },
        {
          balance: 40n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "0xvalidator2",
          validatorIndex: 2n,
          exitEpoch: 150,
          exitDate: new Date(exitDateString),
          slashed: true,
        },
      ];

      validatorDataClient.getExitingValidators.mockResolvedValue(exitingValidators);
      validatorDataClient.getTotalBalanceOfExitingValidators.mockReturnValue(72n * ONE_GWEI);

      await poller.poll();

      expect(validatorDataClient.getTotalBalanceOfExitingValidators).toHaveBeenCalledWith(exitingValidators);
      expect(metricsUpdater.setLastTotalPendingExitGwei).toHaveBeenCalledWith(72000000000);
    });

    it("handles undefined result with warning", async () => {
      const exitDateString = "2024-01-15T10:30:00Z";
      const exitingValidators: ExitingValidator[] = [
        {
          balance: 32n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "0xvalidator1",
          validatorIndex: 1n,
          exitEpoch: 100,
          exitDate: new Date(exitDateString),
          slashed: false,
        },
      ];

      validatorDataClient.getExitingValidators.mockResolvedValue(exitingValidators);
      validatorDataClient.getTotalBalanceOfExitingValidators.mockReturnValue(undefined);

      await poller.poll();

      expect(metricsUpdater.setLastTotalPendingExitGwei).not.toHaveBeenCalled();
      expect(logger.warn).toHaveBeenCalledWith("Skipping total pending exit gauge update: total balance unavailable");
    });

    it("converts bigint to number correctly", async () => {
      const exitDateString = "2024-01-15T10:30:00Z";
      const exitingValidators: ExitingValidator[] = [
        {
          balance: 35n * ONE_GWEI,
          effectiveBalance: 32n * ONE_GWEI,
          publicKey: "0xvalidator1",
          validatorIndex: 1n,
          exitEpoch: 100,
          exitDate: new Date(exitDateString),
          slashed: false,
        },
      ];

      validatorDataClient.getExitingValidators.mockResolvedValue(exitingValidators);
      validatorDataClient.getTotalBalanceOfExitingValidators.mockReturnValue(35n * ONE_GWEI);

      await poller.poll();

      expect(metricsUpdater.setLastTotalPendingExitGwei).toHaveBeenCalledWith(35000000000);
    });

    it("handles undefined exitingValidators with warning", async () => {
      validatorDataClient.getExitingValidators.mockResolvedValue(undefined);
      validatorDataClient.getTotalBalanceOfExitingValidators.mockReturnValue(undefined);

      await poller.poll();

      expect(validatorDataClient.getTotalBalanceOfExitingValidators).toHaveBeenCalledWith(undefined);
      expect(metricsUpdater.setLastTotalPendingExitGwei).not.toHaveBeenCalled();
      expect(logger.warn).toHaveBeenCalledWith("Skipping total pending exit gauge update: total balance unavailable");
    });

    it("handles empty exitingValidators array with warning", async () => {
      validatorDataClient.getExitingValidators.mockResolvedValue([]);
      validatorDataClient.getTotalBalanceOfExitingValidators.mockReturnValue(undefined);

      await poller.poll();

      expect(validatorDataClient.getTotalBalanceOfExitingValidators).toHaveBeenCalledWith([]);
      expect(metricsUpdater.setLastTotalPendingExitGwei).not.toHaveBeenCalled();
      expect(logger.warn).toHaveBeenCalledWith("Skipping total pending exit gauge update: total balance unavailable");
    });
  });
});

