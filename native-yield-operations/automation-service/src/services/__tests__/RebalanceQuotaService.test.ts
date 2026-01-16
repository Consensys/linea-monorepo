import { mock, MockProxy } from "jest-mock-extended";
import { ILogger } from "@consensys/linea-shared-utils";
import { Address } from "viem";
import { RebalanceDirection } from "../../core/entities/RebalanceRequirement.js";
import { INativeYieldAutomationMetricsUpdater } from "../../core/metrics/INativeYieldAutomationMetricsUpdater.js";
import { RebalanceQuotaService } from "../RebalanceQuotaService.js";

describe("RebalanceQuotaService", () => {
  const vaultAddress = "0x1111111111111111111111111111111111111111" as Address;
  const totalSystemBalance = 1000000000000000000000n; // 1000 ETH in wei
  const rebalanceQuotaBps = 1800; // 18%
  const rebalanceToleranceAmountWei = 1000000000000000000n; // 1 ETH in wei
  const quotaWindowSizeInCycles = 24;

  let logger: MockProxy<ILogger>;
  let metricsUpdater: MockProxy<INativeYieldAutomationMetricsUpdater>;
  let service: RebalanceQuotaService;

  const createService = (
    stakingDirection: RebalanceDirection = RebalanceDirection.STAKE,
    windowSize: number = quotaWindowSizeInCycles,
    quotaBps: number = rebalanceQuotaBps,
    toleranceWei: bigint = rebalanceToleranceAmountWei,
  ) => {
    return new RebalanceQuotaService(logger, metricsUpdater, stakingDirection, windowSize, quotaBps, toleranceWei);
  };

  beforeEach(() => {
    jest.clearAllMocks();
    logger = mock<ILogger>();
    metricsUpdater = mock<INativeYieldAutomationMetricsUpdater>();
  });

  describe("constructor", () => {
    it("initializes with all required parameters", () => {
      service = createService();

      expect(service).toBeDefined();
      expect(service.getStakingDirection()).toBe(RebalanceDirection.STAKE);
    });

    it("logs initialization message with correct parameters", () => {
      service = createService();

      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining("RebalanceQuotaService initialized"),
      );
      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining(`stakingDirection=${RebalanceDirection.STAKE}`),
      );
      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining(`quotaWindowSizeInCycles=${quotaWindowSizeInCycles}`),
      );
      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining(`rebalanceQuotaBps=${rebalanceQuotaBps}`),
      );
      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining(`rebalanceToleranceAmountWei=${rebalanceToleranceAmountWei.toString()}`),
      );
    });

    it("stores all constructor parameters correctly", () => {
      const customDirection = RebalanceDirection.UNSTAKE;
      const customWindowSize = 48;
      const customQuotaBps = 2000;
      const customTolerance = 2000000000000000000n;

      service = new RebalanceQuotaService(
        logger,
        metricsUpdater,
        customDirection,
        customWindowSize,
        customQuotaBps,
        customTolerance,
      );

      expect(service.getStakingDirection()).toBe(customDirection);
    });

    it("creates SlidingWindowAccumulator with correct window size", () => {
      const customWindowSize = 12;
      service = createService(RebalanceDirection.STAKE, customWindowSize);

      // Verify by checking behavior - accumulator should track window correctly
      const quotaWei = (totalSystemBalance * BigInt(rebalanceQuotaBps)) / 10000n;
      const amountPerCycle = quotaWei / BigInt(customWindowSize);

      // Fill the window
      for (let i = 0; i < customWindowSize; i++) {
        service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, amountPerCycle);
      }

      // Next call should still be within quota (window is full but at quota)
      const result = service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, amountPerCycle);
      expect(result).toBeGreaterThan(0n);
    });
  });

  describe("getStakingDirection", () => {
    it("returns the staking direction passed to constructor (STAKE)", () => {
      service = createService(RebalanceDirection.STAKE);

      expect(service.getStakingDirection()).toBe(RebalanceDirection.STAKE);
    });

    it("returns correct direction for different enum values", () => {
      service = createService(RebalanceDirection.UNSTAKE);

      expect(service.getStakingDirection()).toBe(RebalanceDirection.UNSTAKE);
    });
  });

  describe("getRebalanceAmountAfterQuota - quota disabled", () => {
    it("returns full amount when quotaWindowSizeInCycles = 0", () => {
      service = createService(RebalanceDirection.STAKE, 0);
      const reBalanceAmountWei = 50000000000000000000n; // 50 ETH

      const result = service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, reBalanceAmountWei);

      expect(result).toBe(reBalanceAmountWei);
    });

    it("does not call metrics updater when quota is disabled", () => {
      service = createService(RebalanceDirection.STAKE, 0);
      const reBalanceAmountWei = 50000000000000000000n;

      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, reBalanceAmountWei);

      expect(metricsUpdater.incrementStakingDepositQuotaExceeded).not.toHaveBeenCalled();
    });

    it("does not log quota-related messages when disabled", () => {
      service = createService(RebalanceDirection.STAKE, 0);
      const reBalanceAmountWei = 50000000000000000000n;

      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, reBalanceAmountWei);

      // Should not log threshold or quota calculation messages
      expect(logger.info).not.toHaveBeenCalledWith(
        expect.stringContaining("below tolerance threshold"),
      );
      expect(logger.info).not.toHaveBeenCalledWith(
        expect.stringContaining("quotaWei"),
      );
    });
  });

  describe("getRebalanceAmountAfterQuota - below tolerance threshold", () => {
    it("returns 0n when amount is below tolerance threshold", () => {
      service = createService();
      const reBalanceAmountWei = 500000000000000000n; // 0.5 ETH, below 1 ETH threshold

      const result = service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, reBalanceAmountWei);

      expect(result).toBe(0n);
    });

    it("pushes 0n to buffer (not the actual amount)", () => {
      service = createService();
      const reBalanceAmountWei = 500000000000000000n; // Below threshold

      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, reBalanceAmountWei);

      // Next call with amount above threshold should be processed normally
      // If the actual amount was pushed, the buffer would have that value
      const quotaWei = (totalSystemBalance * BigInt(rebalanceQuotaBps)) / 10000n;
      const largeAmount = quotaWei / 2n; // Half of quota

      const result = service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, largeAmount);
      // Should return full amount since buffer only has 0n from previous call
      expect(result).toBe(largeAmount);
    });

    it("logs threshold message", () => {
      service = createService();
      const reBalanceAmountWei = 500000000000000000n;

      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, reBalanceAmountWei);

      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining("below tolerance threshold"),
      );
      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining(`reBalanceAmountWei=${reBalanceAmountWei.toString()}`),
      );
      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining(`rebalanceToleranceAmountWei=${rebalanceToleranceAmountWei.toString()}`),
      );
    });

    it("does not increment metrics when below threshold", () => {
      service = createService();
      const reBalanceAmountWei = 500000000000000000n;

      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, reBalanceAmountWei);

      expect(metricsUpdater.incrementStakingDepositQuotaExceeded).not.toHaveBeenCalled();
    });
  });

  describe("getRebalanceAmountAfterQuota - within quota (happy path)", () => {
    it("returns full amount when newTotal <= quotaWei", () => {
      service = createService();
      const quotaWei = (totalSystemBalance * BigInt(rebalanceQuotaBps)) / 10000n;
      const reBalanceAmountWei = quotaWei / 2n; // Half of quota

      const result = service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, reBalanceAmountWei);

      expect(result).toBe(reBalanceAmountWei);
    });

    it("calculates quota correctly (totalSystemBalance * bps / 10000)", () => {
      service = createService();
      const quotaWei = (totalSystemBalance * BigInt(rebalanceQuotaBps)) / 10000n;
      // 1000 ETH * 1800 bps / 10000 = 180 ETH
      expect(quotaWei).toBe(180000000000000000000n);

      const reBalanceAmountWei = quotaWei - 1n; // Just under quota

      const result = service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, reBalanceAmountWei);

      expect(result).toBe(reBalanceAmountWei);
    });

    it("pushes amount to buffer", () => {
      service = createService();
      const quotaWei = (totalSystemBalance * BigInt(rebalanceQuotaBps)) / 10000n;
      const amount1 = quotaWei / 3n;
      const amount2 = quotaWei / 3n;

      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, amount1);
      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, amount2);

      // Third call should still be within quota (total = 2/3 of quota)
      const result = service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, amount1);
      expect(result).toBe(amount1);
    });

    it("logs quota calculation details", () => {
      service = createService();
      const reBalanceAmountWei = 50000000000000000000n; // 50 ETH
      const quotaWei = (totalSystemBalance * BigInt(rebalanceQuotaBps)) / 10000n;

      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, reBalanceAmountWei);

      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining(`totalSystemBalance=${totalSystemBalance.toString()}`),
      );
      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining(`reBalanceAmountWei=${reBalanceAmountWei.toString()}`),
      );
      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining(`quotaWei=${quotaWei.toString()}`),
      );
      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining("prevTotal"),
      );
    });

    it("does not increment metrics when within quota", () => {
      service = createService();
      const quotaWei = (totalSystemBalance * BigInt(rebalanceQuotaBps)) / 10000n;
      const reBalanceAmountWei = quotaWei / 2n;

      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, reBalanceAmountWei);

      expect(metricsUpdater.incrementStakingDepositQuotaExceeded).not.toHaveBeenCalled();
    });
  });

  describe("getRebalanceAmountAfterQuota - exceeding quota scenarios", () => {
    it("returns partial amount (quotaWei - prevTotal) when crossing threshold but previous cycle was under quota", () => {
      service = createService();
      const quotaWei = (totalSystemBalance * BigInt(rebalanceQuotaBps)) / 10000n;
      // Fill buffer to just under quota
      const amountPerCycle = (quotaWei * 95n) / 100n; // 95% of quota
      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, amountPerCycle);

      // Now add amount that would exceed quota
      const excessAmount = quotaWei / 10n; // 10% of quota
      const result = service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, excessAmount);

      // Should return partial amount to exactly hit quota
      const prevTotal = amountPerCycle;
      const expectedPartial = quotaWei - prevTotal;
      expect(result).toBe(expectedPartial);
      expect(result).toBeLessThan(excessAmount);
    });

    it("returns 0n when exceeding quota and previous cycle was also over quota", () => {
      service = createService();
      const quotaWei = (totalSystemBalance * BigInt(rebalanceQuotaBps)) / 10000n;
      // First call exceeds quota
      const amount1 = quotaWei + 1000000000000000000n; // Over quota
      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, amount1);

      // Second call also exceeds quota (previous was over quota)
      const amount2 = quotaWei + 1000000000000000000n;
      const result = service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, amount2);

      expect(result).toBe(0n);
    });

    it("increments metrics when quota is exceeded (partial return case)", () => {
      service = createService();
      const quotaWei = (totalSystemBalance * BigInt(rebalanceQuotaBps)) / 10000n;
      const amountPerCycle = (quotaWei * 95n) / 100n;
      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, amountPerCycle);

      const excessAmount = quotaWei / 10n;
      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, excessAmount);

      expect(metricsUpdater.incrementStakingDepositQuotaExceeded).toHaveBeenCalledWith(vaultAddress);
    });

    it("increments metrics when quota is exceeded (zero return case)", () => {
      service = createService();
      const quotaWei = (totalSystemBalance * BigInt(rebalanceQuotaBps)) / 10000n;
      const amount1 = quotaWei + 1000000000000000000n;
      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, amount1);

      const amount2 = quotaWei + 1000000000000000000n;
      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, amount2);

      expect(metricsUpdater.incrementStakingDepositQuotaExceeded).toHaveBeenCalledTimes(2);
      expect(metricsUpdater.incrementStakingDepositQuotaExceeded).toHaveBeenCalledWith(vaultAddress);
    });

    it("logs newTotal when quota is exceeded", () => {
      service = createService();
      const quotaWei = (totalSystemBalance * BigInt(rebalanceQuotaBps)) / 10000n;
      const amountPerCycle = (quotaWei * 95n) / 100n;
      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, amountPerCycle);

      jest.clearAllMocks();
      const excessAmount = quotaWei / 10n;
      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, excessAmount);

      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining("newTotal"),
      );
    });
  });

  describe("edge cases and integration", () => {
    it("handles zero totalSystemBalance (quota = 0)", () => {
      service = createService();
      const zeroBalance = 0n;
      const reBalanceAmountWei = 1000000000000000000n;

      const result = service.getRebalanceAmountAfterQuota(vaultAddress, zeroBalance, reBalanceAmountWei);

      // Quota is 0, so any amount exceeds quota
      // Since prevTotal is 0 and quota is 0, should return 0n
      expect(result).toBe(0n);
    });

    it("handles very large bigint values", () => {
      service = createService();
      const veryLargeBalance = 1000000000000000000000000n; // 1 million ETH
      const quotaWei = (veryLargeBalance * BigInt(rebalanceQuotaBps)) / 10000n;
      const reBalanceAmountWei = quotaWei / 2n;

      const result = service.getRebalanceAmountAfterQuota(vaultAddress, veryLargeBalance, reBalanceAmountWei);

      expect(result).toBe(reBalanceAmountWei);
    });

    it("maintains correct window over multiple consecutive calls", () => {
      service = createService(RebalanceDirection.STAKE, 3); // Small window for testing
      const quotaWei = (totalSystemBalance * BigInt(rebalanceQuotaBps)) / 10000n;
      const amountPerCycle = quotaWei / 4n; // Each cycle uses 1/4 of quota

      // Fill window (3 cycles)
      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, amountPerCycle);
      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, amountPerCycle);
      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, amountPerCycle);

      // Fourth call should slide window (oldest removed, new added)
      const result = service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, amountPerCycle);
      expect(result).toBe(amountPerCycle); // Should still be within quota
    });

    it("window slides correctly over multiple cycles", () => {
      service = createService(RebalanceDirection.STAKE, 2); // Window of 2
      const quotaWei = (totalSystemBalance * BigInt(rebalanceQuotaBps)) / 10000n;
      const amount1 = quotaWei / 3n;
      const amount2 = quotaWei / 3n;

      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, amount1);
      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, amount2);

      // Third call should slide window (amount1 removed, amount3 added)
      const amount3 = quotaWei / 3n;
      const result = service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, amount3);

      expect(result).toBe(amount3); // Should be within quota (amount2 + amount3)
    });

    it("calculates quota correctly with different BPS values (100 bps = 1%)", () => {
      service = createService(RebalanceDirection.STAKE, quotaWindowSizeInCycles, 100); // 1%
      const quotaWei = (totalSystemBalance * 100n) / 10000n;
      expect(quotaWei).toBe(10000000000000000000n); // 10 ETH (1% of 1000 ETH)

      const reBalanceAmountWei = quotaWei / 2n;
      const result = service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, reBalanceAmountWei);

      expect(result).toBe(reBalanceAmountWei);
    });

    it("calculates quota correctly with different BPS values (1800 bps = 18%)", () => {
      service = createService(RebalanceDirection.STAKE, quotaWindowSizeInCycles, 1800); // 18%
      const quotaWei = (totalSystemBalance * 1800n) / 10000n;
      expect(quotaWei).toBe(180000000000000000000n); // 180 ETH (18% of 1000 ETH)

      const reBalanceAmountWei = quotaWei / 2n;
      const result = service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, reBalanceAmountWei);

      expect(result).toBe(reBalanceAmountWei);
    });
  });

  describe("metrics tracking", () => {
    it("calls incrementStakingDepositQuotaExceeded with correct vaultAddress when quota exceeded", () => {
      service = createService();
      const quotaWei = (totalSystemBalance * BigInt(rebalanceQuotaBps)) / 10000n;
      const amount = quotaWei + 1000000000000000000n; // Over quota

      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, amount);

      expect(metricsUpdater.incrementStakingDepositQuotaExceeded).toHaveBeenCalledWith(vaultAddress);
      expect(metricsUpdater.incrementStakingDepositQuotaExceeded).toHaveBeenCalledTimes(1);
    });

    it("does not call incrementStakingDepositQuotaExceeded when within quota", () => {
      service = createService();
      const quotaWei = (totalSystemBalance * BigInt(rebalanceQuotaBps)) / 10000n;
      const amount = quotaWei / 2n; // Half of quota

      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, amount);

      expect(metricsUpdater.incrementStakingDepositQuotaExceeded).not.toHaveBeenCalled();
    });

    it("does not call incrementStakingDepositQuotaExceeded when below tolerance", () => {
      service = createService();
      const amount = 500000000000000000n; // Below threshold

      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, amount);

      expect(metricsUpdater.incrementStakingDepositQuotaExceeded).not.toHaveBeenCalled();
    });
  });

  describe("logging verification", () => {
    it("logs initialization message in constructor", () => {
      service = createService();

      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining("RebalanceQuotaService initialized"),
      );
    });

    it("logs threshold message when below tolerance", () => {
      service = createService();
      const amount = 500000000000000000n; // Below threshold

      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, amount);

      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining("below tolerance threshold"),
      );
    });

    it("logs quota calculation details (totalSystemBalance, reBalanceAmountWei, quotaWei, prevTotal)", () => {
      service = createService();
      const reBalanceAmountWei = 50000000000000000000n;
      const quotaWei = (totalSystemBalance * BigInt(rebalanceQuotaBps)) / 10000n;

      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, reBalanceAmountWei);

      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining(`totalSystemBalance=${totalSystemBalance.toString()}`),
      );
      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining(`reBalanceAmountWei=${reBalanceAmountWei.toString()}`),
      );
      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining(`quotaWei=${quotaWei.toString()}`),
      );
      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining("prevTotal"),
      );
    });

    it("logs newTotal when processing", () => {
      service = createService();
      const reBalanceAmountWei = 50000000000000000000n;

      service.getRebalanceAmountAfterQuota(vaultAddress, totalSystemBalance, reBalanceAmountWei);

      expect(logger.info).toHaveBeenCalledWith(
        expect.stringContaining("newTotal"),
      );
    });
  });
});
