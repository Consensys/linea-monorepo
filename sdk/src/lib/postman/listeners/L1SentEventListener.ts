import { DataSource } from "typeorm";
import { LoggerOptions } from "winston";
import { L1MessageServiceContract } from "../../contracts";
import { SentEventListener } from "./SentEventListener";
import { getLogger } from "../../logger";
import { L1NetworkConfig } from "../utils/types";
import { Direction } from "../utils/enums";

export class L1SentEventListener extends SentEventListener<L1MessageServiceContract> {
  constructor(
    dataSource: DataSource,
    public readonly l1MessageServiceContract: L1MessageServiceContract,
    config: L1NetworkConfig,
    loggerOptions?: LoggerOptions,
  ) {
    super(dataSource, l1MessageServiceContract, config, Direction.L1_TO_L2);
    this.logger = getLogger(L1SentEventListener.name, loggerOptions);
  }
}
