import { DataSource } from "typeorm";
import { LoggerOptions } from "winston";
import { L2MessageServiceContract } from "../../contracts";
import { getLogger } from "../../logger";
import { L2NetworkConfig } from "../utils/types";
import { Direction } from "../utils/enums";
import { ClaimTxSender } from "./ClaimTxSender";

export class L2ClaimTxSender extends ClaimTxSender<L2MessageServiceContract> {
  constructor(
    dataSource: DataSource,
    l2MessageServiceContract: L2MessageServiceContract,
    config: L2NetworkConfig,
    originContractAddress: string,
    loggerOptions?: LoggerOptions,
  ) {
    super(dataSource, l2MessageServiceContract, config, Direction.L1_TO_L2);
    this.logger = getLogger(L2ClaimTxSender.name, loggerOptions);
    this.originContractAddress = originContractAddress;
  }

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  override async isRateLimitExceeded(_messageFee: string, _messageValue: string): Promise<boolean> {
    return false;
  }
}
