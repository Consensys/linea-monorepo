import { DataSource } from "typeorm";
import { LoggerOptions } from "winston";
import { L1MessageServiceContract } from "../../contracts";
import { getLogger } from "../../logger";
import { L1NetworkConfig } from "../utils/types";
import { Direction } from "../utils/enums";
import { AnchoredEventListener } from "./AnchoredEventListener";

export class L1AnchoredEventListener extends AnchoredEventListener<L1MessageServiceContract> {
  constructor(
    dataSource: DataSource,
    public readonly l1MessageServiceContract: L1MessageServiceContract,
    config: L1NetworkConfig,
    originContractAddress: string,
    loggerOptions?: LoggerOptions,
  ) {
    super(dataSource, l1MessageServiceContract, config, Direction.L2_TO_L1);
    this.logger = getLogger(L1AnchoredEventListener.name, loggerOptions);
    this.originContractAddress = originContractAddress;
  }
}
