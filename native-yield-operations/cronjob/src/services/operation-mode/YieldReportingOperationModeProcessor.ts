import { TransactionReceipt } from "viem";
import { IYieldManager } from "../../core/services/contracts/IYieldManager";
import { IOperationModeProcessor } from "../../core/services/operation-mode/IOperationModeProcessor";
import { ILogger } from "ts-libs/linea-shared-utils/dist";
import { wait } from "sdk/sdk-ethers/dist";
import { ILazyOracle } from "../../core/services/contracts/ILazyOracle";

export class YieldReportingOperationModeProcessor implements IOperationModeProcessor {
  constructor(
    private readonly yieldManagerContractClient: IYieldManager<TransactionReceipt>,
    private readonly lazyOracleContractClient: ILazyOracle<TransactionReceipt>,
    private readonly logger: ILogger,
    private readonly maxInactionMs: number,
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
   * 1. Determine whether a new report should be submitted.
   * 2. Decide whether a rebalance is required.
   * 3. If rebalancing, calculate the amount and direction.
   * 4. Execute.
   * 5. Evaluate execution result, amend if needed.
   */
  private async _process(): Promise<void> {}
}
