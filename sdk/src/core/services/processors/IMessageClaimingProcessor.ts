import { Direction } from "../../enums/MessageEnums";

export interface IMessageClaimingProcessor {
  process(): Promise<void>;
}

export type MessageClaimingProcessorConfig = {
  maxNonceDiff: number;
  feeRecipientAddress?: string;
  profitMargin: number;
  maxNumberOfRetries: number;
  retryDelayInSeconds: number;
  maxClaimGasLimit: bigint;
  direction: Direction;
  originContractAddress: string;
};
