import { jest } from "@jest/globals";
import { ResultAsync, ok, err } from "neverthrow";
import type { ILogger } from "@consensys/linea-shared-utils";
import type { INativeYieldAutomationMetricsUpdater } from "../../../core/metrics/INativeYieldAutomationMetricsUpdater.js";
import type { ILidoAccountingReportClient } from "../../../core/clients/ILidoAccountingReportClient.js";
import type { IVaultHub } from "../../../core/clients/contracts/IVaultHub.js";
import type { Address, TransactionReceipt } from "viem";
import { submitVaultReportIfNotFresh } from "../vaultReportSubmission.js";

jest.mock("@consensys/linea-shared-utils", () => {
  const actual = jest.requireActual("@consensys/linea-shared-utils") as typeof import("@consensys/linea-shared-utils");
  return {
    ...actual,
    attempt: jest.fn(),
  };
});

import { attempt } from "@consensys/linea-shared-utils";

const attemptMock = attempt as jest.MockedFunction<typeof attempt>;

describe("submitVaultReportIfNotFresh", () => {
  const vaultAddress = "0x2222222222222222222222222222222222222222" as Address;
  const logPrefix = "test";

  let logger: jest.Mocked<ILogger>;
  let vaultHubClient: jest.Mocked<IVaultHub<TransactionReceipt>>;
  let lidoReportClient: jest.Mocked<ILidoAccountingReportClient>;
  let metricsUpdater: jest.Mocked<INativeYieldAutomationMetricsUpdater>;

  beforeEach(() => {
    jest.clearAllMocks();

    logger = {
      name: "test",
      info: jest.fn(),
      error: jest.fn(),
      warn: jest.fn(),
      debug: jest.fn(),
    } as unknown as jest.Mocked<ILogger>;

    vaultHubClient = {
      isVaultConnected: jest.fn(),
      isReportFresh: jest.fn(),
    } as unknown as jest.Mocked<IVaultHub<TransactionReceipt>>;

    lidoReportClient = {
      getLatestSubmitVaultReportParams: jest.fn(),
      submitLatestVaultReport: jest.fn(),
    } as unknown as jest.Mocked<ILidoAccountingReportClient>;

    metricsUpdater = {
      incrementLidoVaultAccountingReport: jest.fn(),
    } as unknown as jest.Mocked<INativeYieldAutomationMetricsUpdater>;

    // Default mocks
    vaultHubClient.isVaultConnected.mockResolvedValue(true);
    vaultHubClient.isReportFresh.mockResolvedValue(false);
    lidoReportClient.getLatestSubmitVaultReportParams.mockResolvedValue({
      vault: vaultAddress,
      totalValue: 0n,
      cumulativeLidoFees: 0n,
      liabilityShares: 0n,
      maxLiabilityShares: 0n,
      slashingReserve: 0n,
      proof: [],
    });
    lidoReportClient.submitLatestVaultReport.mockResolvedValue(undefined);
    attemptMock.mockImplementation(((loggerArg: ILogger, fn: () => unknown | Promise<unknown>) =>
      ResultAsync.fromPromise((async () => fn())(), (error) => error as Error)) as typeof attempt);
  });

  it("skips submission when shouldSubmitVaultReport is false", async () => {
    await submitVaultReportIfNotFresh(
      logger,
      vaultHubClient,
      lidoReportClient,
      metricsUpdater,
      vaultAddress,
      false,
      logPrefix,
    );

    expect(vaultHubClient.isVaultConnected).not.toHaveBeenCalled();
    expect(vaultHubClient.isReportFresh).not.toHaveBeenCalled();
    expect(lidoReportClient.getLatestSubmitVaultReportParams).not.toHaveBeenCalled();
    expect(lidoReportClient.submitLatestVaultReport).not.toHaveBeenCalled();
    expect(metricsUpdater.incrementLidoVaultAccountingReport).not.toHaveBeenCalled();
    expect(logger.info).toHaveBeenCalledWith(`${logPrefix} - Skipping vault report submission (SHOULD_SUBMIT_VAULT_REPORT=false)`);
  });

  it("skips submission when vault is not connected", async () => {
    vaultHubClient.isVaultConnected.mockResolvedValueOnce(false);

    await submitVaultReportIfNotFresh(
      logger,
      vaultHubClient,
      lidoReportClient,
      metricsUpdater,
      vaultAddress,
      true,
      logPrefix,
    );

    expect(vaultHubClient.isVaultConnected).toHaveBeenCalledWith(vaultAddress);
    expect(vaultHubClient.isReportFresh).not.toHaveBeenCalled();
    expect(lidoReportClient.getLatestSubmitVaultReportParams).not.toHaveBeenCalled();
    expect(lidoReportClient.submitLatestVaultReport).not.toHaveBeenCalled();
    expect(metricsUpdater.incrementLidoVaultAccountingReport).not.toHaveBeenCalled();
    expect(logger.info).toHaveBeenCalledWith(`${logPrefix} - Skipping vault report submission (vault is not connected)`);
  });

  it("proceeds with submission when isVaultConnected check fails", async () => {
    const error = new Error("Connection check failed");
    vaultHubClient.isVaultConnected.mockRejectedValueOnce(error);

    await submitVaultReportIfNotFresh(
      logger,
      vaultHubClient,
      lidoReportClient,
      metricsUpdater,
      vaultAddress,
      true,
      logPrefix,
    );

    expect(vaultHubClient.isVaultConnected).toHaveBeenCalledWith(vaultAddress);
    expect(logger.warn).toHaveBeenCalledWith(
      `${logPrefix} - Failed to check if vault is connected, proceeding with submission attempt: ${error}`,
    );
    expect(vaultHubClient.isReportFresh).toHaveBeenCalledWith(vaultAddress);
    expect(lidoReportClient.getLatestSubmitVaultReportParams).toHaveBeenCalledWith(vaultAddress);
    expect(lidoReportClient.submitLatestVaultReport).toHaveBeenCalledWith(vaultAddress);
  });

  it("skips submission when report is fresh", async () => {
    vaultHubClient.isReportFresh.mockResolvedValueOnce(true);

    await submitVaultReportIfNotFresh(
      logger,
      vaultHubClient,
      lidoReportClient,
      metricsUpdater,
      vaultAddress,
      true,
      logPrefix,
    );

    expect(vaultHubClient.isVaultConnected).toHaveBeenCalledWith(vaultAddress);
    expect(vaultHubClient.isReportFresh).toHaveBeenCalledWith(vaultAddress);
    expect(lidoReportClient.getLatestSubmitVaultReportParams).not.toHaveBeenCalled();
    expect(lidoReportClient.submitLatestVaultReport).not.toHaveBeenCalled();
    expect(metricsUpdater.incrementLidoVaultAccountingReport).not.toHaveBeenCalled();
    expect(logger.info).toHaveBeenCalledWith(`${logPrefix} - Skipping vault report submission (report is fresh)`);
  });

  it("proceeds with submission when isReportFresh check fails", async () => {
    const error = new Error("Freshness check failed");
    vaultHubClient.isReportFresh.mockRejectedValueOnce(error);

    await submitVaultReportIfNotFresh(
      logger,
      vaultHubClient,
      lidoReportClient,
      metricsUpdater,
      vaultAddress,
      true,
      logPrefix,
    );

    expect(vaultHubClient.isVaultConnected).toHaveBeenCalledWith(vaultAddress);
    expect(vaultHubClient.isReportFresh).toHaveBeenCalledWith(vaultAddress);
    expect(logger.warn).toHaveBeenCalledWith(
      `${logPrefix} - Failed to check if report is fresh, proceeding with submission attempt: ${error}`,
    );
    expect(lidoReportClient.getLatestSubmitVaultReportParams).toHaveBeenCalledWith(vaultAddress);
    expect(lidoReportClient.submitLatestVaultReport).toHaveBeenCalledWith(vaultAddress);
  });

  it("submits vault report when vault is connected and report is not fresh", async () => {
    await submitVaultReportIfNotFresh(
      logger,
      vaultHubClient,
      lidoReportClient,
      metricsUpdater,
      vaultAddress,
      true,
      logPrefix,
    );

    expect(vaultHubClient.isVaultConnected).toHaveBeenCalledWith(vaultAddress);
    expect(vaultHubClient.isReportFresh).toHaveBeenCalledWith(vaultAddress);
    expect(lidoReportClient.getLatestSubmitVaultReportParams).toHaveBeenCalledWith(vaultAddress);
    expect(lidoReportClient.submitLatestVaultReport).toHaveBeenCalledWith(vaultAddress);
    expect(logger.info).toHaveBeenCalledWith(`${logPrefix} - Fetching latest vault report`);
    expect(logger.info).toHaveBeenCalledWith(`${logPrefix} - Submitting latest vault report`);
  });

  it("increments metrics when submission succeeds", async () => {
    const okResult = ok(undefined);
    attemptMock.mockResolvedValueOnce(okResult);

    await submitVaultReportIfNotFresh(
      logger,
      vaultHubClient,
      lidoReportClient,
      metricsUpdater,
      vaultAddress,
      true,
      logPrefix,
    );

    expect(metricsUpdater.incrementLidoVaultAccountingReport).toHaveBeenCalledWith(vaultAddress);
    expect(logger.info).toHaveBeenCalledWith(`${logPrefix} - Vault report submission succeeded`);
  });

  it("does not increment metrics when submission fails", async () => {
    const errResult = err(new Error("Submission failed"));
    attemptMock.mockResolvedValueOnce(errResult);

    await submitVaultReportIfNotFresh(
      logger,
      vaultHubClient,
      lidoReportClient,
      metricsUpdater,
      vaultAddress,
      true,
      logPrefix,
    );

    expect(metricsUpdater.incrementLidoVaultAccountingReport).not.toHaveBeenCalled();
    expect(logger.info).not.toHaveBeenCalledWith(`${logPrefix} - Vault report submission succeeded`);
  });
});

