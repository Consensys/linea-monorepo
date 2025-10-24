import { Address, TransactionReceipt } from "viem";
import { ILogger } from "@consensys/linea-shared-utils";
import { IYieldManager } from "../../core/clients/contracts/IYieldManager.js";
import { IOperationModeProcessor } from "../../core/services/operation-mode/IOperationModeProcessor.js";
import { wait } from "@consensys/linea-sdk";
import { IBeaconChainStakingClient } from "../../core/clients/IBeaconChainStakingClient.js";

export class OssificationCompleteOperationModeProcessor implements IOperationModeProcessor {
  constructor(
    private readonly logger: ILogger,
    private readonly yieldManagerContractClient: IYieldManager<TransactionReceipt>,
    private readonly beaconChainStakingClient: IBeaconChainStakingClient,
    private readonly maxInactionMs: number,
    private readonly yieldProvider: Address,
  ) {}

  public async process(): Promise<void> {
    this.logger.info(
      `OssificationCompleteOperationModeProcessor: waiting ${this.maxInactionMs}ms before executing actions`,
    );
    await wait(this.maxInactionMs);
    await this._process();
  }

  private async _process(): Promise<void> {
    this.logger.info("OssificationCompleteOperationModeProcessor: executing withdrawal and unstake cycle");
    // Max withdraw
    await this.yieldManagerContractClient.safeMaxAddToWithdrawalReserve(this.yieldProvider);

    // Max unstake
    await this.beaconChainStakingClient.submitMaxAvailableWithdrawalRequests();
  }
}
