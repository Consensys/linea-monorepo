import { Direction } from "../../enums";

export interface IMessageClaimingProcessor {
  process(): Promise<void>;
}

export type MessageClaimingProcessorConfig = {
  feeRecipientAddress?: string;
  profitMargin: number;
  maxNumberOfRetries: number;
  retryDelayInSeconds: number;
  maxClaimGasLimit: bigint;
  direction: Direction;
  originContractAddress: string;
  claimViaAddress?: string;
};
