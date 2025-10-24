import { Address, TransactionReceipt } from "viem";
import { ILogger, attempt } from "@consensys/linea-shared-utils";
import { IYieldManager } from "../../core/clients/contracts/IYieldManager.js";
import { IOperationModeProcessor } from "../../core/services/operation-mode/IOperationModeProcessor.js";
import { wait } from "@consensys/linea-sdk";
import { IBeaconChainStakingClient } from "../../core/clients/IBeaconChainStakingClient.js";

export class OssificationCompleteProcessor implements IOperationModeProcessor {
  constructor(
    private readonly logger: ILogger,
    private readonly yieldManagerContractClient: IYieldManager<TransactionReceipt>,
    private readonly beaconChainStakingClient: IBeaconChainStakingClient,
    private readonly maxInactionMs: number,
    private readonly yieldProvider: Address,
  ) {}

  public async process(): Promise<void> {
    this.logger.info(
      `Waiting ${this.maxInactionMs}ms before executing actions`,
    );
    await wait(this.maxInactionMs);
    await this._process();
  }

  private async _process(): Promise<void> {
    // Max withdraw
    this.logger.info("_process - Performing max withdrawal from YieldProvider");
    await attempt(this.logger, () =>
      this.yieldManagerContractClient.safeMaxAddToWithdrawalReserve(this.yieldProvider), "_process - safeMaxAddToWithdrawalReserve failed (tolerated)");

    // Max unstake
    this.logger.info("_process - Performing max unstake from beacon chain");
    await this.beaconChainStakingClient.submitMaxAvailableWithdrawalRequests();
  }
}
