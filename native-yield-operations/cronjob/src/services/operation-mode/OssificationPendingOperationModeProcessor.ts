import { Address, TransactionReceipt } from "viem";
import { IYieldManager } from "../../core/services/contracts/IYieldManager";
import { IOperationModeProcessor } from "../../core/services/operation-mode/IOperationModeProcessor";
import { ILogger } from "@consensys/linea-shared-utils";
import { wait } from "sdk/sdk-ethers";
import { ILazyOracle } from "../../core/services/contracts/ILazyOracle";
import { ILidoAccountingReportClient } from "../../core/clients/ILidoAccountingReportClient";
import { IBeaconChainStakingClient } from "../../core/clients/IBeaconChainStakingClient";

export class OssificationPendingOperationModeProcessor implements IOperationModeProcessor {
  constructor(
    private readonly yieldManagerContractClient: IYieldManager<TransactionReceipt>,
    private readonly lazyOracleContractClient: ILazyOracle<TransactionReceipt>,
    private readonly lidoAccountingReportClient: ILidoAccountingReportClient,
    private readonly beaconChainStakingClient: IBeaconChainStakingClient,
    private readonly logger: ILogger,
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
      // Race: event vs. timeout
      const winner = await Promise.race([
        waitForEvent.then(() => "event" as const),
        wait(this.maxInactionMs).then(() => "timeout" as const),
      ]);
      await this._process();
      this.logger.info(`poll(): finished via ${winner}`);
    } finally {
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
    // Submit vault report if available
    await this.lidoAccountingReportClient.getLatestSubmitVaultReportParams();
    const isSimulateSubmitLatestVaultReportSuccessful =
      await this.lidoAccountingReportClient.isSimulateSubmitLatestVaultReportSuccessful();
    if (isSimulateSubmitLatestVaultReportSuccessful) {
      await this.lidoAccountingReportClient.submitLatestVaultReport();
    }

    // Process Pending Ossification
    await this.yieldManagerContractClient.progressPendingOssification(this.yieldProvider);

    // Max withdraw if ossified
    if (await this.yieldManagerContractClient.isOssified(this.yieldProvider)) {
      await this.yieldManagerContractClient.safeMaxAddToWithdrawalReserve(this.yieldProvider);
    }

    // Max unstake
    await this.beaconChainStakingClient.submitMaxAvailableWithdrawalRequests();
  }
}
