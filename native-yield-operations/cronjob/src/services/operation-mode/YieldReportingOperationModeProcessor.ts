import { PublicClient, TransactionReceipt } from "viem";
import { IBaseContractClient } from "../../core/clients/IBaseContractClient";
import { IYieldManager } from "../../core/services/contracts/IYieldManager";
import { IOperationModeProcessor } from "../../core/services/operation-mode/IOperationModeProcessor";
import { IContractClientLibrary } from "ts-libs/linea-shared-utils/src/core/client/IContractClientLibrary";
import { ILogger } from "ts-libs/linea-shared-utils/dist";
import { NativeYieldCronJobClientConfig } from "../../application/main/config/NativeYieldCronJobClientConfig";

export class YieldReportingOperationModeProcessor implements IOperationModeProcessor {
  constructor(
    private readonly yieldManagerContractClient: IYieldManager<TransactionReceipt> & IBaseContractClient,
    private readonly ethereumMainnetClientLibrary: IContractClientLibrary<PublicClient, TransactionReceipt>,
    private readonly logger: ILogger,
    private readonly triggerTimingConfig: NativeYieldCronJobClientConfig["timing"]["trigger"],
  ) {}

  public async process(): Promise<void> {
    const blockchainClient = await this.ethereumMainnetClientLibrary.getBlockchainClient();

    // Watch for VaultsReportDataUpdated
    const unwatch = blockchainClient.watchContractEvent({
      address: this.yieldManagerContractClient.getAddress(),
      abi: this.yieldManagerContractClient.getContract().abi,
      eventName: "VaultsReportDataUpdated",
      args: {},
      onLogs: (logs) => {
        console.log(logs);
        unwatch();
      },
    });
  }
}
