import { MessageProps } from "../message/Message";
import { LineaGasFees, MessageSent } from "../types";
import { IMessageServiceContract } from "./IMessageServiceContract";

export interface IL2ContractClient extends IMessageServiceContract {
  encodeClaimMessageTransactionData(message: MessageProps & { feeRecipient?: string }): string;

  estimateClaimGasFees(
    message: (MessageSent | MessageProps) & { feeRecipient?: string },
    opts?: { claimViaAddress?: string },
  ): Promise<LineaGasFees>;

  getSigner(): string | undefined;

  getContractAddress(): string;
}
