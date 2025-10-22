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

  private async _process(): Promise<void> {}
}
