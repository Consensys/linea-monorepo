import { DataSource } from "typeorm";
import { LoggerOptions } from "winston";
import { L2MessageServiceContract } from "../../contracts";
import { getLogger } from "../../logger";
import { L2NetworkConfig } from "../utils/types";
import { Direction } from "../utils/enums";
import { ClaimStatusWatcher } from "./ClaimStatusWatcher";

export class L2ClaimStatusWatcher extends ClaimStatusWatcher<L2MessageServiceContract> {
  constructor(
    dataSource: DataSource,
    l2MessageServiceContract: L2MessageServiceContract,
    config: L2NetworkConfig,
    loggerOptions?: LoggerOptions,
  ) {
    super(dataSource, l2MessageServiceContract, config, Direction.L1_TO_L2);
    this.logger = getLogger(L2ClaimStatusWatcher.name, loggerOptions);
  }

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  override async isRateLimitExceededError(_transactionHash: string): Promise<boolean> {
    return false;
  }
}
