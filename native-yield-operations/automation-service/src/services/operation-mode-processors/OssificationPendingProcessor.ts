import { Address, TransactionReceipt } from "viem";
import { IYieldManager } from "../../core/clients/contracts/IYieldManager.js";
import { IOperationModeProcessor } from "../../core/services/operation-mode/IOperationModeProcessor.js";
import { ILogger, tryResult } from "@consensys/linea-shared-utils";
import { wait } from "@consensys/linea-sdk";
import { ILazyOracle } from "../../core/clients/contracts/ILazyOracle.js";
import { ILidoAccountingReportClient } from "../../core/clients/ILidoAccountingReportClient.js";
import { IBeaconChainStakingClient } from "../../core/clients/IBeaconChainStakingClient.js";

export class OssificationPendingProcessor implements IOperationModeProcessor {
  constructor(
    private readonly logger: ILogger,
    private readonly yieldManagerContractClient: IYieldManager<TransactionReceipt>,
    private readonly lazyOracleContractClient: ILazyOracle<TransactionReceipt>,
    private readonly lidoAccountingReportClient: ILidoAccountingReportClient,
    private readonly beaconChainStakingClient: IBeaconChainStakingClient,
    private readonly maxInactionMs: number,
    private readonly yieldProvider: Address,
  ) {}

  /**
   * Executes one processing cycle:
   * - Waits for the next `VaultsReportDataUpdated` event **or** a timeout, whichever happens first.
   * - Once triggered, runs the main processing logic (`_process()`).
   * - Always cleans up the event watcher afterward.
   */
  public async process(): Promise<void> {
    const { unwatch, waitForEvent } = await this.lazyOracleContractClient.waitForVaultsReportDataUpdatedEvent();
    try {
      this.logger.info(`process - Waiting for VaultsReportDataUpdated event vs timeout race, timeout=${this.maxInactionMs}ms`)
      // Race: event vs. timeout
      const winner = await Promise.race([
        waitForEvent.then(() => "event" as const),
        wait(this.maxInactionMs).then(() => "timeout" as const),
      ]);
      this.logger.info(
        `process - race won by ${winner === "timeout" ?  `time out after ${this.maxInactionMs}ms` : "VaultsReportDataUpdated event"}`,
      );
      await this._process();
    } finally {
      this.logger.debug("Cleaning up VaultsReportDataUpdated event watcher")
      // clean up watcher
      unwatch();
    }
  }

  /**
   * Main processing loop:
   * 1. Submit vault report if available
   * 2. Perform processPendingOssifcation
   * 3. Max withdraw
   * 4. Max unstake
   */
  private async _process(): Promise<void> {
    this.logger.info("_process - Fetching latest vault report")
    // Submit vault report if available
    await this.lidoAccountingReportClient.getLatestSubmitVaultReportParams();
    const isSimulateSubmitLatestVaultReportSuccessful =
      await this.lidoAccountingReportClient.isSimulateSubmitLatestVaultReportSuccessful();
    if (isSimulateSubmitLatestVaultReportSuccessful) {
      this.logger.info("_process - Submitting latest vault report")
      await this.lidoAccountingReportClient.submitLatestVaultReport();
    }

    // Process Pending Ossification
    await this.yieldManagerContractClient.progressPendingOssification(this.yieldProvider);

    // Max withdraw if ossified
    this.logger.info("_process - Ossification completed, performing max safe withdrawal")
    if (await this.yieldManagerContractClient.isOssified(this.yieldProvider)) {
      await tryResult(() =>
        this.yieldManagerContractClient.safeMaxAddToWithdrawalReserve(this.yieldProvider)
      ).mapErr((error) => {
        this.logger.warn(
          "safeMaxAddToWithdrawalReserve failed (tolerated)", { error }
        );
        return error; // required by mapErr
      });
    }

    // Max unstake
    this.logger.info("_process - performing max unstake from beacon chain")
    await this.beaconChainStakingClient.submitMaxAvailableWithdrawalRequests();
  }
}
