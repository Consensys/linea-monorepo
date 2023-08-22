import { DataSource } from "typeorm";
import { LoggerOptions } from "winston";
import { SentEventListener } from "./SentEventListener";
import { L2MessageServiceContract } from "../../contracts";
import { getLogger } from "../../logger";
import { L2NetworkConfig } from "../utils/types";
import { Direction } from "../utils/enums";

export class L2SentEventListener extends SentEventListener<L2MessageServiceContract> {
  constructor(
    dataSource: DataSource,
    public readonly l2MessageServiceContract: L2MessageServiceContract,
    config: L2NetworkConfig,
    loggerOptions?: LoggerOptions,
  ) {
    super(dataSource, l2MessageServiceContract, config, Direction.L2_TO_L1);
    this.logger = getLogger(L2SentEventListener.name, loggerOptions);
  }
}
