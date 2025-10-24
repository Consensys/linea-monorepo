import { Address, TransactionReceipt } from "viem";
import { YieldReportingOperationModeProcessor } from "../YieldReportingOperationModeProcessor.js";
import { RebalanceDirection, RebalanceRequirement } from "../../../core/entities/RebalanceRequirement.js";
import { wait } from "@consensys/linea-sdk";
import type { IYieldManager } from "../../../core/services/contracts/IYieldManager.js";
import type { ILazyOracle } from "../../../core/services/contracts/ILazyOracle.js";
import type { ILineaRollupYieldExtension } from "../../../core/services/contracts/ILineaRollupYieldExtension.js";
import type { ILidoAccountingReportClient } from "../../../core/clients/ILidoAccountingReportClient.js";
import type { IBeaconChainStakingClient } from "../../../core/clients/IBeaconChainStakingClient.js";
import type { ILogger } from "@consensys/linea-shared-utils";

jest.mock("@consensys/linea-sdk", () => ({
  wait: jest.fn(),
}));

const waitMock = wait as jest.MockedFunction<typeof wait>;

const YIELD_PROVIDER = "0x0000000000000000000000000000000000000010" as Address;
const L2_RECIPIENT = "0x0000000000000000000000000000000000000020" as Address;
const MAX_INACTION_MS = 5_000;

type Mocks = {
  yieldManager: {
    getRebalanceRequirements: jest.Mock;
    pauseStakingIfNotAlready: jest.Mock;
    unpauseStakingIfNotAlready: jest.Mock;
    fundYieldProvider: jest.Mock;
    reportYield: jest.Mock;
    safeAddToWithdrawalReserveIfAboveThreshold: jest.Mock;
  };
  lazyOracle: {
    waitForVaultsReportDataUpdatedEvent: jest.Mock;
  };
  lineaRollup: {
    transferFundsForNativeYield: jest.Mock;
  };
  lidoAccounting: {
    getLatestSubmitVaultReportParams: jest.Mock;
    isSimulateSubmitLatestVaultReportSuccessful: jest.Mock;
    submitLatestVaultReport: jest.Mock;
  };
  beacon: {
    submitWithdrawalRequestsToFulfilAmount: jest.Mock;
  };
  logger: {
    name: string;
    info: jest.Mock;
    error: jest.Mock;
    warn: jest.Mock;
    debug: jest.Mock;
    warnOrError: jest.Mock;
  };
};

const createProcessor = () => {
  const mocks: Mocks = {
    yieldManager: {
      getRebalanceRequirements: jest.fn(),
      pauseStakingIfNotAlready: jest.fn(),
      unpauseStakingIfNotAlready: jest.fn(),
      fundYieldProvider: jest.fn(),
      reportYield: jest.fn(),
      safeAddToWithdrawalReserveIfAboveThreshold: jest.fn(),
    },
    lazyOracle: {
      waitForVaultsReportDataUpdatedEvent: jest.fn(),
    },
    lineaRollup: {
      transferFundsForNativeYield: jest.fn(),
    },
    lidoAccounting: {
      getLatestSubmitVaultReportParams: jest.fn(),
      isSimulateSubmitLatestVaultReportSuccessful: jest.fn(),
      submitLatestVaultReport: jest.fn(),
    },
    beacon: {
      submitWithdrawalRequestsToFulfilAmount: jest.fn(),
    },
    logger: {
      name: "test-logger",
      info: jest.fn(),
      error: jest.fn(),
      warn: jest.fn(),
      debug: jest.fn(),
      warnOrError: jest.fn(),
    },
  };

  const processor = new YieldReportingOperationModeProcessor(
    mocks.logger as unknown as ILogger,
    mocks.yieldManager as unknown as IYieldManager<TransactionReceipt>,
    mocks.lazyOracle as unknown as ILazyOracle<TransactionReceipt>,
    mocks.lineaRollup as unknown as ILineaRollupYieldExtension<TransactionReceipt>,
    mocks.lidoAccounting as unknown as ILidoAccountingReportClient,
    mocks.beacon as unknown as IBeaconChainStakingClient,
    MAX_INACTION_MS,
    YIELD_PROVIDER,
    L2_RECIPIENT,
  );

  return { processor, mocks };
};

const requirement = (rebalanceDirection: RebalanceDirection, rebalanceAmount: bigint): RebalanceRequirement => ({
  rebalanceDirection,
  rebalanceAmount,
});

beforeEach(() => {
  jest.clearAllMocks();
});

describe("YieldReportingOperationModeProcessor", () => {
  it("performs a no-op cycle when no rebalance is required and simulation fails", async () => {
    const { processor, mocks } = createProcessor();
    const unwatch = jest.fn();

    waitMock.mockImplementation(() => new Promise(() => {}));
    mocks.lazyOracle.waitForVaultsReportDataUpdatedEvent.mockResolvedValue({
      unwatch,
      waitForEvent: Promise.resolve(),
    });
    mocks.lidoAccounting.getLatestSubmitVaultReportParams.mockResolvedValue(undefined);
    mocks.lidoAccounting.isSimulateSubmitLatestVaultReportSuccessful.mockResolvedValue(false);
    mocks.yieldManager.getRebalanceRequirements
      .mockResolvedValueOnce(requirement(RebalanceDirection.NONE, 0n))
      .mockResolvedValueOnce(requirement(RebalanceDirection.NONE, 0n))
      .mockResolvedValueOnce(requirement(RebalanceDirection.NONE, 0n));

    await processor.process();

    expect(mocks.lidoAccounting.submitLatestVaultReport).not.toHaveBeenCalled();
    expect(mocks.yieldManager.reportYield).not.toHaveBeenCalled();
    expect(mocks.yieldManager.pauseStakingIfNotAlready).not.toHaveBeenCalled();
    expect(mocks.yieldManager.unpauseStakingIfNotAlready).not.toHaveBeenCalled();
    expect(mocks.beacon.submitWithdrawalRequestsToFulfilAmount).not.toHaveBeenCalled();
    expect(mocks.logger.info).toHaveBeenCalledWith("poll(): finished via event");
    expect(unwatch).toHaveBeenCalled();
  });

  it("logs timeout when the trigger event does not arrive before inaction window", async () => {
    const { processor, mocks } = createProcessor();
    const unwatch = jest.fn();

    waitMock.mockImplementation(() => Promise.resolve());
    mocks.lazyOracle.waitForVaultsReportDataUpdatedEvent.mockResolvedValue({
      unwatch,
      waitForEvent: new Promise<never>(() => {}),
    });
    mocks.lidoAccounting.getLatestSubmitVaultReportParams.mockResolvedValue(undefined);
    mocks.lidoAccounting.isSimulateSubmitLatestVaultReportSuccessful.mockResolvedValue(false);
    mocks.yieldManager.getRebalanceRequirements
      .mockResolvedValueOnce(requirement(RebalanceDirection.NONE, 0n))
      .mockResolvedValueOnce(requirement(RebalanceDirection.NONE, 0n))
      .mockResolvedValueOnce(requirement(RebalanceDirection.NONE, 0n));

    await processor.process();

    expect(mocks.logger.info).toHaveBeenCalledWith("poll(): finished via timeout");
    expect(unwatch).toHaveBeenCalled();
  });

  it("stakes surplus funds and reports yield when simulation succeeds", async () => {
    const { processor, mocks } = createProcessor();
    const unwatch = jest.fn();

    waitMock.mockImplementation(() => new Promise(() => {}));
    mocks.lazyOracle.waitForVaultsReportDataUpdatedEvent.mockResolvedValue({
      unwatch,
      waitForEvent: Promise.resolve(),
    });
    mocks.lidoAccounting.getLatestSubmitVaultReportParams.mockResolvedValue(undefined);
    mocks.lidoAccounting.isSimulateSubmitLatestVaultReportSuccessful.mockResolvedValue(true);
    mocks.yieldManager.getRebalanceRequirements
      .mockResolvedValueOnce(requirement(RebalanceDirection.STAKE, 100n))
      .mockResolvedValueOnce(requirement(RebalanceDirection.NONE, 0n))
      .mockResolvedValueOnce(requirement(RebalanceDirection.NONE, 0n));

    await processor.process();

    expect(mocks.lineaRollup.transferFundsForNativeYield).toHaveBeenCalledWith(100n);
    expect(mocks.yieldManager.fundYieldProvider).toHaveBeenCalledWith(YIELD_PROVIDER, 100n);
    expect(mocks.lidoAccounting.submitLatestVaultReport).toHaveBeenCalledTimes(1);
    expect(mocks.yieldManager.reportYield).toHaveBeenCalledWith(YIELD_PROVIDER, L2_RECIPIENT);
    expect(mocks.yieldManager.unpauseStakingIfNotAlready).toHaveBeenCalledWith(YIELD_PROVIDER);
    expect(mocks.beacon.submitWithdrawalRequestsToFulfilAmount).not.toHaveBeenCalled();
    expect(unwatch).toHaveBeenCalled();
  });

  it("pauses staking and executes deficit workflow including beacon withdrawals", async () => {
    const { processor, mocks } = createProcessor();
    const unwatch = jest.fn();

    waitMock.mockImplementation(() => new Promise(() => {}));
    mocks.lazyOracle.waitForVaultsReportDataUpdatedEvent.mockResolvedValue({
      unwatch,
      waitForEvent: Promise.resolve(),
    });
    mocks.lidoAccounting.getLatestSubmitVaultReportParams.mockResolvedValue(undefined);
    mocks.lidoAccounting.isSimulateSubmitLatestVaultReportSuccessful.mockResolvedValue(true);
    mocks.yieldManager.getRebalanceRequirements
      .mockResolvedValueOnce(requirement(RebalanceDirection.UNSTAKE, 50n))
      .mockResolvedValueOnce(requirement(RebalanceDirection.UNSTAKE, 40n))
      .mockResolvedValueOnce(requirement(RebalanceDirection.UNSTAKE, 40n));

    await processor.process();

    expect(mocks.yieldManager.pauseStakingIfNotAlready).toHaveBeenCalledWith(YIELD_PROVIDER);
    expect(mocks.lidoAccounting.submitLatestVaultReport).toHaveBeenCalledTimes(1);
    expect(mocks.yieldManager.reportYield).toHaveBeenCalledWith(YIELD_PROVIDER, L2_RECIPIENT);
    expect(mocks.yieldManager.safeAddToWithdrawalReserveIfAboveThreshold).toHaveBeenCalledWith(YIELD_PROVIDER, 50n);
    expect(mocks.beacon.submitWithdrawalRequestsToFulfilAmount).toHaveBeenCalledWith(40n);
    expect(mocks.yieldManager.unpauseStakingIfNotAlready).not.toHaveBeenCalled();
    expect(unwatch).toHaveBeenCalled();
  });

  it("performs amendment unstake when mid-cycle drift flips from surplus to deficit", async () => {
    const { processor, mocks } = createProcessor();
    const unwatch = jest.fn();

    waitMock.mockImplementation(() => new Promise(() => {}));
    mocks.lazyOracle.waitForVaultsReportDataUpdatedEvent.mockResolvedValue({
      unwatch,
      waitForEvent: Promise.resolve(),
    });
    mocks.lidoAccounting.getLatestSubmitVaultReportParams.mockResolvedValue(undefined);
    mocks.lidoAccounting.isSimulateSubmitLatestVaultReportSuccessful.mockResolvedValue(false);
    mocks.yieldManager.getRebalanceRequirements
      .mockResolvedValueOnce(requirement(RebalanceDirection.STAKE, 75n))
      .mockResolvedValueOnce(requirement(RebalanceDirection.UNSTAKE, 25n))
      .mockResolvedValueOnce(requirement(RebalanceDirection.NONE, 0n));

    await processor.process();

    expect(mocks.lineaRollup.transferFundsForNativeYield).toHaveBeenCalledWith(75n);
    expect(mocks.yieldManager.fundYieldProvider).toHaveBeenCalledWith(YIELD_PROVIDER, 75n);
    expect(mocks.lidoAccounting.submitLatestVaultReport).not.toHaveBeenCalled();
    expect(mocks.yieldManager.reportYield).not.toHaveBeenCalled();
    expect(mocks.yieldManager.safeAddToWithdrawalReserveIfAboveThreshold).toHaveBeenCalledWith(YIELD_PROVIDER, 25n);
    expect(mocks.yieldManager.unpauseStakingIfNotAlready).not.toHaveBeenCalled();
    expect(mocks.beacon.submitWithdrawalRequestsToFulfilAmount).not.toHaveBeenCalled();
    expect(unwatch).toHaveBeenCalled();
  });
});
