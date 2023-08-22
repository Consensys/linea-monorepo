import { DataSource } from "typeorm";
import { LoggerOptions } from "winston";
import { L1MessageServiceContract } from "../../contracts";
import { getLogger } from "../../logger";
import { L1NetworkConfig } from "../utils/types";
import { Direction } from "../utils/enums";
import { ClaimTxSender } from "./ClaimTxSender";
import { DEFAULT_RATE_LIMIT_MARGIN } from "../../utils/constants";

export class L1ClaimTxSender extends ClaimTxSender<L1MessageServiceContract> {
  constructor(
    dataSource: DataSource,
    private readonly l1MessageServiceContract: L1MessageServiceContract,
    config: L1NetworkConfig,
    originContractAddress: string,
    loggerOptions?: LoggerOptions,
  ) {
    super(dataSource, l1MessageServiceContract, config, Direction.L2_TO_L1);
    this.logger = getLogger(L1ClaimTxSender.name, loggerOptions);
    this.originContractAddress = originContractAddress;
  }

  override async isRateLimitExceeded(messageFee: string, messageValue: string): Promise<boolean> {
    const rateLimitInWei = await this.l1MessageServiceContract.contract.limitInWei();
    const currentPeriodAmountInWei = await this.l1MessageServiceContract.contract.currentPeriodAmountInWei();

    return (
      parseFloat(currentPeriodAmountInWei.add(messageFee).add(messageValue).toString()) >
      parseFloat(rateLimitInWei.toString()) * DEFAULT_RATE_LIMIT_MARGIN
    );
  }
}
