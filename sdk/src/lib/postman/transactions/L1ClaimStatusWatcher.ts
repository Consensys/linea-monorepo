import { DataSource } from "typeorm";
import { LoggerOptions } from "winston";
import { L1MessageServiceContract } from "../../contracts";
import { getLogger } from "../../logger";
import { L1NetworkConfig } from "../utils/types";
import { Direction } from "../utils/enums";
import { ClaimStatusWatcher } from "./ClaimStatusWatcher";

export class L1ClaimStatusWatcher extends ClaimStatusWatcher<L1MessageServiceContract> {
  constructor(
    dataSource: DataSource,
    private readonly l1MessageServiceContract: L1MessageServiceContract,
    config: L1NetworkConfig,
    loggerOptions?: LoggerOptions,
  ) {
    super(dataSource, l1MessageServiceContract, config, Direction.L2_TO_L1);
    this.logger = getLogger(L1ClaimStatusWatcher.name, loggerOptions);
  }

  override async isRateLimitExceededError(transactionHash: string): Promise<boolean> {
    try {
      const tx = await this.l1MessageServiceContract.provider.getTransaction(transactionHash);
      const errorEncodedData = await this.l1MessageServiceContract.provider.call(
        {
          to: tx.to,
          from: tx.from,
          nonce: tx.nonce,
          gasLimit: tx.gasLimit,
          data: tx.data,
          value: tx.value,
          chainId: tx.chainId,
          accessList: tx.accessList,
          maxPriorityFeePerGas: tx.maxPriorityFeePerGas,
          maxFeePerGas: tx.maxFeePerGas,
        },
        tx.blockNumber,
      );
      const error = this.l1MessageServiceContract.contract.interface.parseError(errorEncodedData);

      return error.name === "RateLimitExceeded";
    } catch (e) {
      return false;
    }
  }
}
