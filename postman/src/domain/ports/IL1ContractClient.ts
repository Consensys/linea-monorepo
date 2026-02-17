import { MessageProps } from "../message/Message";
import { MessageSent } from "../types";
import { IMessageServiceContract } from "./IMessageServiceContract";

export interface IL1ContractClient extends IMessageServiceContract {
  estimateClaimGas(
    message: (MessageSent | MessageProps) & { feeRecipient?: string; messageBlockNumber?: number },
    opts?: { claimViaAddress?: string },
  ): Promise<bigint>;
}
