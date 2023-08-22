import { DataSource } from "typeorm";
import { LoggerOptions } from "winston";
import { L2MessageServiceContract } from "../../contracts";
import { getLogger } from "../../logger";
import { L2NetworkConfig } from "../utils/types";
import { Direction } from "../utils/enums";
import { AnchoredEventListener } from "./AnchoredEventListener";

export class L2AnchoredEventListener extends AnchoredEventListener<L2MessageServiceContract> {
  constructor(
    dataSource: DataSource,
    public readonly l2MessageServiceContract: L2MessageServiceContract,
    config: L2NetworkConfig,
    originContractAddress: string,
    loggerOptions?: LoggerOptions,
  ) {
    super(dataSource, l2MessageServiceContract, config, Direction.L1_TO_L2);
    this.logger = getLogger(L2AnchoredEventListener.name, loggerOptions);
    this.originContractAddress = originContractAddress;
  }
}
