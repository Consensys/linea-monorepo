import { PublicClient, TransactionReceipt } from "viem";
import { IBaseContractClient } from "../../core/clients/IBaseContractClient";
import { IYieldManager } from "../../core/services/contracts/IYieldManager";
import { IOperationModeProcessor } from "../../core/services/operation-mode/IOperationModeProcessor";
import { IContractClientLibrary } from "ts-libs/linea-shared-utils/src/core/client/IContractClientLibrary";
import { ILogger } from "ts-libs/linea-shared-utils/dist";
import { NativeYieldCronJobClientConfig } from "../../application/main/config/NativeYieldCronJobClientConfig";
import { wait } from "sdk/sdk-ethers/dist";

export class YieldReportingOperationModeProcessor implements IOperationModeProcessor {
  constructor(
    private readonly yieldManagerContractClient: IYieldManager<TransactionReceipt> & IBaseContractClient,
    private readonly ethereumMainnetClientLibrary: IContractClientLibrary<PublicClient, TransactionReceipt>,
    private readonly logger: ILogger,
    private readonly triggerTimingConfig: NativeYieldCronJobClientConfig["timing"]["trigger"],
  ) {}

  // Poll for VaultsReportDataUpdated event
  // Fallback to 'cronjob' logic if it is not found
  public async poll(): Promise<void> {
    const blockchainClient = await this.ethereumMainnetClientLibrary.getBlockchainClient();
    const maxInactionMs = this.triggerTimingConfig.maxInactionSeconds * 1000;
    const pollMs = this.triggerTimingConfig.pollIntervalSeconds * 1000;

    // Promise resolves on the first *non-removed* matching log
    let stop: (() => void) | undefined;
    const onFirstEvent = new Promise<void>((resolve) => {
      stop = blockchainClient.watchContractEvent({
        address: this.yieldManagerContractClient.getAddress(),
        abi: this.yieldManagerContractClient.getContract().abi,
        eventName: "VaultsReportDataUpdated",
        pollingInterval: pollMs,
        onLogs: (logs) => {
          // Reorg guard
          const valid = logs.find((l) => !l.removed);
          if (!valid) return;
          this.logger.info("VaultsReportDataUpdated detected");
          resolve();
        },
        onError: (err) => {
          this.logger.error({ err }, "watchContractEvent error");
          // You might choose to resolve here to trigger processing on errors
        },
      });
    });

    try {
      // Race: event vs. timeout
      const winner = await Promise.race([
        onFirstEvent.then(() => "event" as const),
        wait(maxInactionMs).then(() => "timeout" as const),
      ]);

      // Either way, run one processing pass
      await this.process();
      this.logger.info(`poll(): finished via ${winner}`);
    } finally {
      // clean up watcher
      if (stop) stop();
    }
  }

  public async process(): Promise<void> {}
}
