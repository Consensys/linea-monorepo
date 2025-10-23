import { Address, TransactionReceipt } from "viem";
import { IYieldManager } from "../../core/services/contracts/IYieldManager.js";
import { IOperationModeProcessor } from "../../core/services/operation-mode/IOperationModeProcessor.js";
import { wait } from "sdk/sdk-ethers";
import { IBeaconChainStakingClient } from "../../core/clients/IBeaconChainStakingClient.js";

export class OssificationCompleteOperationModeProcessor implements IOperationModeProcessor {
  constructor(
    private readonly yieldManagerContractClient: IYieldManager<TransactionReceipt>,
    private readonly beaconChainStakingClient: IBeaconChainStakingClient,
    private readonly maxInactionMs: number,
    private readonly yieldProvider: Address,
  ) {}

  public async process(): Promise<void> {
    await wait(this.maxInactionMs);
    await this._process();
  }

  private async _process(): Promise<void> {
    // Max withdraw
    await this.yieldManagerContractClient.safeMaxAddToWithdrawalReserve(this.yieldProvider);

    // Max unstake
    await this.beaconChainStakingClient.submitMaxAvailableWithdrawalRequests();
  }
}
